// Package whois looks up a domain's registrar and expiry over WHOIS (port
// 43) for the TLDs whose registries publish no RDAP service, the fallback
// the domain expiry monitor uses next to internal/rdap (ADR domain-lookups).
//
// WHOIS answers are free text in a different shape per registry, so each
// supported TLD has its own server and parser, written against that
// registry's real answers; there is no generic parsing of unknown formats.
package whois

import (
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"
)

// Result is the subset of a WHOIS answer checkmeup cares about, the same as
// rdap.Result.
type Result struct {
	Registrar string
	ExpiresAt time.Time
}

// registry describes one TLD's WHOIS server and how to read its answers.
type registry struct {
	server string
	// query turns the domain into what the server expects; nil sends it as is.
	query func(domain string) string
	// expiry patterns, tried in order; group 1 is the date, read with layout.
	expiry []expiryPattern
	// registrar's group 1 is the registrar's name; nil when the registry
	// doesn't publish one.
	registrar *regexp.Regexp
	// notFound matches the registry's answer for a domain that isn't
	// registered.
	notFound *regexp.Regexp
}

type expiryPattern struct {
	re     *regexp.Regexp
	layout string
}

// icann is the answer shape of gTLD-style registries (and ccTLDs run on the
// same software): "Registry Expiry Date: 2027-05-01T00:00:02Z".
func icann(server string) registry {
	return registry{
		server:    server,
		expiry:    []expiryPattern{{regexp.MustCompile(`(?m)^\s*Registry Expiry Date:\s*(\S+)`), time.RFC3339}},
		registrar: regexp.MustCompile(`(?m)^\s*Registrar:\s*(.+?)\s*$`),
		notFound:  regexp.MustCompile(`(?mi)^\s*(Domain not found\.|No Data Found|The queried object does not exist)`),
	}
}

// registries are the TLDs with no RDAP service whose WHOIS publishes an
// expiry date, keyed by the domain's last label.
var registries = map[string]registry{
	"il": {
		server:    "whois.isoc.org.il",
		expiry:    []expiryPattern{{regexp.MustCompile(`(?m)^validity:\s*(\d{2}-\d{2}-\d{4})`), "02-01-2006"}},
		registrar: regexp.MustCompile(`(?m)^registrar name:\s*(.+?)\s*$`),
		notFound:  regexp.MustCompile(`No data was found to match the request criteria`),
	},
	"io": icann("whois.nic.io"),
	"me": icann("whois.nic.me"),
	"co": icann("whois.registry.co"),
	"us": icann("whois.nic.us"),
	"it": {
		server:    "whois.nic.it",
		expiry:    []expiryPattern{{regexp.MustCompile(`(?m)^Expire Date:\s*(\d{4}-\d{2}-\d{2})`), "2006-01-02"}},
		registrar: regexp.MustCompile(`(?m)^Registrar\s*\n\s+Organization:\s*(.+?)\s*$`),
		notFound:  regexp.MustCompile(`(?m)^Status:\s*AVAILABLE`),
	},
	"ru": {
		server:    "whois.tcinet.ru",
		expiry:    []expiryPattern{{regexp.MustCompile(`(?m)^paid-till:\s*(\S+)`), time.RFC3339}},
		registrar: regexp.MustCompile(`(?m)^registrar:\s*(.+?)\s*$`),
		notFound:  regexp.MustCompile(`No entries found`),
	},
	"jp": {
		server: "whois.jprs.jp",
		// "/e" asks JPRS for English labels instead of Japanese ones.
		query: func(domain string) string { return domain + "/e" },
		expiry: []expiryPattern{
			{regexp.MustCompile(`(?m)^\[Expires on\]\s*(\d{4}/\d{2}/\d{2})`), "2006/01/02"},
			// Second-level names (co.jp, ne.jp) show the date in [State].
			{regexp.MustCompile(`(?m)^\[State\]\s*Connected \((\d{4}/\d{2}/\d{2})\)`), "2006/01/02"},
		},
		notFound: regexp.MustCompile(`No match!!`),
	},
	"cn": {
		server: "whois.cnnic.cn",
		// China Standard Time; a day's precision is all the monitor uses.
		expiry:    []expiryPattern{{regexp.MustCompile(`(?m)^Expiration Time:\s*(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`), "2006-01-02 15:04:05"}},
		registrar: regexp.MustCompile(`(?m)^Sponsoring Registrar:\s*(.+?)\s*$`),
		notFound:  regexp.MustCompile(`No matching record`),
	},
	"se": {
		server:    "whois.iis.se",
		expiry:    []expiryPattern{{regexp.MustCompile(`(?m)^expires:\s*(\d{4}-\d{2}-\d{2})`), "2006-01-02"}},
		registrar: regexp.MustCompile(`(?m)^registrar:\s*(.+?)\s*$`),
		notFound:  regexp.MustCompile(`(?m)^domain "[^"]*" not found`),
	},
}

// noExpiry are TLDs with no RDAP service whose registries don't publish an
// expiry date over WHOIS either (or, for .es, don't answer WHOIS from
// outside an allow list), so a domain under them can't be monitored.
var noExpiry = map[string]bool{"de": true, "eu": true, "ch": true, "li": true, "at": true, "es": true}

// ErrNotRegistered is the registry's answer for a domain nobody registered.
var ErrNotRegistered = errors.New("domain is not registered")

// tld is the domain's last label: "il" for "example.co.il".
func tld(domain string) string {
	domain = strings.TrimSuffix(strings.ToLower(domain), ".")
	return domain[strings.LastIndex(domain, ".")+1:]
}

// Supports reports whether domain is looked up over WHOIS rather than RDAP.
func Supports(domain string) bool {
	_, ok := registries[tld(domain)]
	return ok
}

// Unsupported returns why domain can't be monitored, or "" when it can as
// far as this package knows: its registry publishes no expiry date.
func Unsupported(domain string) string {
	t := tld(domain)
	if noExpiry[t] {
		return fmt.Sprintf("the .%s registry doesn't publish domain expiry dates, so .%s domains can't be monitored", t, t)
	}
	return ""
}

// Client queries WHOIS servers over TCP port 43.
type Client struct {
	dialer  net.Dialer
	timeout time.Duration
	// addr maps a WHOIS server to the address dialed; tests point it at a
	// local listener. nil dials server:43.
	addr func(server string) string
}

func NewClient() *Client {
	return &Client{dialer: net.Dialer{Timeout: 10 * time.Second}, timeout: 15 * time.Second}
}

// NewClientWithAddr builds a Client that dials addr(server) instead of
// server:43, for tests.
func NewClientWithAddr(addr func(server string) string) *Client {
	c := NewClient()
	c.addr = addr
	return c
}

// maxAnswer bounds how much of an answer is read; real ones are a few KB.
const maxAnswer = 64 * 1024

// Lookup asks the domain's registry over WHOIS for its registrar and expiry.
// It returns ErrNotRegistered for a domain the registry doesn't know, and an
// error for an unsupported TLD, a network failure, or an answer without a
// readable expiry date.
func (c *Client) Lookup(domain string) (Result, error) {
	reg, ok := registries[tld(domain)]
	if !ok {
		return Result{}, fmt.Errorf("no WHOIS lookup for .%s domains", tld(domain))
	}
	answer, err := c.query(reg, domain)
	if err != nil {
		return Result{}, err
	}
	return parse(reg, answer)
}

func (c *Client) query(reg registry, domain string) (string, error) {
	addr := net.JoinHostPort(reg.server, "43")
	if c.addr != nil {
		addr = c.addr(reg.server)
	}
	conn, err := c.dialer.Dial("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("whois request to %s failed: %w", reg.server, err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	q := domain
	if reg.query != nil {
		q = reg.query(domain)
	}
	if _, err := io.WriteString(conn, q+"\r\n"); err != nil {
		return "", fmt.Errorf("whois request to %s failed: %w", reg.server, err)
	}
	b, err := io.ReadAll(io.LimitReader(conn, maxAnswer))
	if err != nil {
		return "", fmt.Errorf("whois answer from %s failed: %w", reg.server, err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n"), nil
}

func parse(reg registry, answer string) (Result, error) {
	if reg.notFound.MatchString(answer) {
		return Result{}, ErrNotRegistered
	}
	var res Result
	for _, p := range reg.expiry {
		m := p.re.FindStringSubmatch(answer)
		if m == nil {
			continue
		}
		t, err := time.Parse(p.layout, m[1])
		if err != nil {
			return Result{}, fmt.Errorf("whois expiry date %q unreadable: %w", m[1], err)
		}
		res.ExpiresAt = t
		break
	}
	if res.ExpiresAt.IsZero() {
		return Result{}, errors.New("whois answer has no expiry date")
	}
	if reg.registrar != nil {
		if m := reg.registrar.FindStringSubmatch(answer); m != nil {
			res.Registrar = m[1]
		}
	}
	return res, nil
}
