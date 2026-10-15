package whois

// testdata/*.txt are real answers from each registry's WHOIS server, saved
// on 2026-10-15: <tld>.txt for a registered domain, <tld>-missing.txt for
// one nobody registered. Parsers are checked against them rather than
// against hand-written samples, since the formats are the registries'.

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

func TestParse_RegisteredDomains(t *testing.T) {
	cases := []struct {
		tld, file, expires, registrar string
	}{
		{"il", "il.txt", "2029-01-11", "Israel Internet Association ISOC-IL"},
		{"io", "io.txt", "2027-05-01", "Registry Operator acts as Registrar (9999)"},
		{"me", "me.txt", "2027-04-29", "Registry Operator acts as Registrar (9999)"},
		{"co", "co.txt", "2027-04-25", "GoDaddy.com, LLC"},
		{"us", "us.txt", "2027-04-17", "usTLD Registry"},
		{"it", "it.txt", "2026-12-31", "ccTLD 'it' Registry"},
		{"ru", "ru.txt", "2027-09-30", "RU-CENTER-RU"},
		{"jp", "jp.txt", "2026-11-30", ""},
		{"jp", "co.jp.txt", "2027-03-31", ""},
		{"cn", "cn.txt", "2028-03-10", "北京新网数码信息技术有限公司"},
		{"se", "se.txt", "2031-03-09", "Internetstift"},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			res, err := parse(registries[c.tld], fixture(t, c.file))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := res.ExpiresAt.Format("2006-01-02"); got != c.expires {
				t.Errorf("expires %s, want %s", got, c.expires)
			}
			if res.Registrar != c.registrar {
				t.Errorf("registrar %q, want %q", res.Registrar, c.registrar)
			}
		})
	}
}

func TestParse_UnregisteredDomains(t *testing.T) {
	for _, tld := range []string{"il", "io", "me", "co", "us", "it", "ru", "jp", "cn", "se"} {
		t.Run(tld, func(t *testing.T) {
			_, err := parse(registries[tld], fixture(t, tld+"-missing.txt"))
			if !errors.Is(err, ErrNotRegistered) {
				t.Fatalf("want ErrNotRegistered, got %v", err)
			}
		})
	}
}

func TestParse_AnswerWithoutExpiry(t *testing.T) {
	_, err := parse(registries["il"], "% rate limit exceeded, try again later\n")
	if err == nil || errors.Is(err, ErrNotRegistered) {
		t.Fatalf("want a no-expiry error, got %v", err)
	}
}

func TestSupportsAndUnsupported(t *testing.T) {
	for _, d := range []string{"sagsuget.co.il", "example.io", "EXAMPLE.CO.JP", "example.com.cn."} {
		if !Supports(d) {
			t.Errorf("want %s looked up over WHOIS", d)
		}
	}
	for _, d := range []string{"example.com", "example.de", "example.nl"} {
		if Supports(d) {
			t.Errorf("want %s not looked up over WHOIS", d)
		}
	}
	if msg := Unsupported("example.de"); !strings.Contains(msg, ".de registry doesn't publish") {
		t.Errorf("Unsupported(example.de) = %q", msg)
	}
	if msg := Unsupported("example.ch"); !strings.Contains(msg, ".ch registry only answers WHOIS for clients on its allow list") {
		t.Errorf("Unsupported(example.ch) = %q", msg)
	}
	if msg := Unsupported("example.co.il"); msg != "" {
		t.Errorf("Unsupported(example.co.il) = %q, want empty", msg)
	}
}

// serveWHOIS answers one WHOIS query on a local listener with answer, and
// records the query line it got.
func serveWHOIS(t *testing.T, answer string) (addr string, query <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	got := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		line, _ := bufio.NewReader(conn).ReadString('\n')
		got <- line
		_, _ = conn.Write([]byte(strings.ReplaceAll(answer, "\n", "\r\n")))
	}()
	return ln.Addr().String(), got
}

func TestLookup(t *testing.T) {
	t.Run("asks the TLD's server and reads its answer", func(t *testing.T) {
		addr, query := serveWHOIS(t, fixture(t, "il.txt"))
		var dialed string
		c := NewClientWithAddr(func(server string) string { dialed = server; return addr })

		res, err := c.Lookup("isoc.org.il")
		if err != nil {
			t.Fatalf("lookup: %v", err)
		}
		if dialed != "whois.isoc.org.il" {
			t.Errorf("dialed %s, want whois.isoc.org.il", dialed)
		}
		if q := <-query; q != "isoc.org.il\r\n" {
			t.Errorf("query %q, want the domain and CRLF", q)
		}
		if !res.ExpiresAt.Equal(time.Date(2029, 1, 11, 0, 0, 0, 0, time.UTC)) {
			t.Errorf("expires %v", res.ExpiresAt)
		}
	})

	t.Run("asks JPRS for English answers", func(t *testing.T) {
		addr, query := serveWHOIS(t, fixture(t, "jp.txt"))
		c := NewClientWithAddr(func(string) string { return addr })
		if _, err := c.Lookup("yandex.jp"); err != nil {
			t.Fatalf("lookup: %v", err)
		}
		if q := <-query; q != "yandex.jp/e\r\n" {
			t.Errorf("query %q, want yandex.jp/e", q)
		}
	})

	t.Run("a TLD without a WHOIS parser is an error, not a guess", func(t *testing.T) {
		c := NewClientWithAddr(func(string) string { t.Fatal("dialed"); return "" })
		if _, err := c.Lookup("example.com"); err == nil {
			t.Fatal("want an error")
		}
	})

	t.Run("a server that can't be reached is an error", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		addr := ln.Addr().String()
		_ = ln.Close()
		c := NewClientWithAddr(func(string) string { return addr })
		if _, err := c.Lookup("example.co.il"); err == nil {
			t.Fatal("want an error")
		}
	})
}
