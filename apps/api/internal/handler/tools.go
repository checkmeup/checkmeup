package handler

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/whois"
	"github.com/checkmeup/checkmeup/internal/worker"
)

// ToolsHandler serves the free, no-sign-up SSL, domain expiry and DNS
// checkers and the cron expression explainer (PRD free-tools): one live
// check per request, through the same TLS check, RDAP/WHOIS lookup, DNS
// resolver and schedule parser monitors use. Nothing is stored.
type ToolsHandler struct {
	rdap      *rdap.Client
	whois     *whois.Client
	checkTLS  func(hostname string) (time.Time, string, int, error)
	lookupDNS func(ctx context.Context, hostname string) []worker.DNSRecordSet
	now       func() time.Time
}

func NewToolsHandler(rd *rdap.Client, wi *whois.Client) *ToolsHandler {
	return &ToolsHandler{rdap: rd, whois: wi, checkTLS: worker.CheckTLS,
		lookupDNS: func(ctx context.Context, hostname string) []worker.DNSRecordSet {
			return worker.LookupAllDNS(ctx, net.DefaultResolver, hostname)
		},
		now: time.Now}
}

type sslToolResponse struct {
	Host      string    `json:"host"`
	ExpiresAt time.Time `json:"expiresAt"`
	DaysLeft  int       `json:"daysLeft"`
	Issuer    string    `json:"issuer"`
}

type dnsToolResponse struct {
	Host    string                `json:"host"`
	Records []worker.DNSRecordSet `json:"records"`
}

type cronToolResponse struct {
	Expression  string      `json:"expression"`
	Description string      `json:"description"`
	Runs        []time.Time `json:"runs"`
}

type domainToolResponse struct {
	Domain    string    `json:"domain"`
	ExpiresAt time.Time `json:"expiresAt"`
	DaysLeft  int       `json:"daysLeft"`
	Registrar string    `json:"registrar"`
	// Stage is where the domain is after expiry (rdap.LifecycleStage), left
	// out when the registry listed no status.
	Stage string `json:"stage,omitempty"`
}

// hostnamePattern is a DNS name's character set: anything else (spaces,
// "?", "@", "%") can't be a host and isn't worth a lookup.
var hostnamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

// toolHost reads the query parameter name as a hostname, accepting a full
// URL as monitors do, and writes a 400 when it isn't one.
func toolHost(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	host, err := parseHostname(r.URL.Query().Get(name))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return "", false
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if !hostnamePattern.MatchString(host) || !strings.Contains(host, ".") {
		respond.Error(w, http.StatusBadRequest, "enter a hostname such as example.com", "bad_request")
		return "", false
	}
	return host, true
}

// CheckSSL handles GET /api/v1/tools/ssl?host=.
func (h *ToolsHandler) CheckSSL(w http.ResponseWriter, r *http.Request) {
	host, ok := toolHost(w, r, "host")
	if !ok {
		return
	}
	expiresAt, issuer, daysLeft, err := h.checkTLS(host)
	if err != nil {
		respond.Error(w, http.StatusUnprocessableEntity, err.Error(), "check_failed")
		return
	}
	respond.JSON(w, http.StatusOK, sslToolResponse{Host: host, ExpiresAt: expiresAt, DaysLeft: daysLeft, Issuer: issuer})
}

// CheckDomain handles GET /api/v1/tools/domain?domain=.
func (h *ToolsHandler) CheckDomain(w http.ResponseWriter, r *http.Request) {
	domain, ok := toolHost(w, r, "domain")
	if !ok {
		return
	}
	res, err := worker.LookupDomain(h.rdap, h.whois, domain)
	if err != nil {
		respond.Error(w, http.StatusUnprocessableEntity, err.Error(), "check_failed")
		return
	}
	daysLeft := int(time.Until(res.ExpiresAt).Hours() / 24)
	respond.JSON(w, http.StatusOK, domainToolResponse{
		Domain: domain, ExpiresAt: res.ExpiresAt, DaysLeft: daysLeft, Registrar: res.Registrar,
		Stage: rdap.LifecycleStage(res.Statuses),
	})
}

// CheckDNS handles GET /api/v1/tools/dns?host=. A name with records of some
// types and failed lookups for others still answers 200, each type with its
// own reason; only a name with no records at all is a failed check.
func (h *ToolsHandler) CheckDNS(w http.ResponseWriter, r *http.Request) {
	host, ok := toolHost(w, r, "host")
	if !ok {
		return
	}
	sets := h.lookupDNS(r.Context(), host)
	reason := host + " has no DNS records"
	for _, set := range sets {
		if len(set.Values) > 0 {
			respond.JSON(w, http.StatusOK, dnsToolResponse{Host: host, Records: sets})
			return
		}
		if set.Error != "" {
			reason = set.Error
		}
	}
	respond.Error(w, http.StatusUnprocessableEntity, reason, "check_failed")
}

// cronToolRuns is how many upcoming run times the cron explainer lists.
const cronToolRuns = 5

// ExplainCron handles GET /api/v1/tools/cron?expression=: what a cron
// expression means and when it next fires in UTC, the times a cron monitor
// with that schedule expects, before its grace period.
func (h *ToolsHandler) ExplainCron(w http.ResponseWriter, r *http.Request) {
	sched, fields, err := parseCronExpression(r.URL.Query().Get("expression"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	runs := make([]time.Time, 0, cronToolRuns)
	t := h.now().UTC()
	for range cronToolRuns {
		t = sched.Next(t)
		runs = append(runs, t)
	}
	respond.JSON(w, http.StatusOK, cronToolResponse{
		Expression:  strings.Join(fields, " "),
		Description: describeCron(fields),
		Runs:        runs,
	})
}
