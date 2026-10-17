package handler

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/whois"
	"github.com/checkmeup/checkmeup/internal/worker"
)

// ToolsHandler serves the free, no-sign-up SSL and domain expiry checkers
// (PRD free-tools): one live check per request, through the same TLS check
// and RDAP/WHOIS lookup the worker runs for monitors. Nothing is stored.
type ToolsHandler struct {
	rdap     *rdap.Client
	whois    *whois.Client
	checkTLS func(hostname string) (time.Time, string, int, error)
}

func NewToolsHandler(rd *rdap.Client, wi *whois.Client) *ToolsHandler {
	return &ToolsHandler{rdap: rd, whois: wi, checkTLS: worker.CheckTLS}
}

type sslToolResponse struct {
	Host      string    `json:"host"`
	ExpiresAt time.Time `json:"expiresAt"`
	DaysLeft  int       `json:"daysLeft"`
	Issuer    string    `json:"issuer"`
}

type domainToolResponse struct {
	Domain    string    `json:"domain"`
	ExpiresAt time.Time `json:"expiresAt"`
	DaysLeft  int       `json:"daysLeft"`
	Registrar string    `json:"registrar"`
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
	respond.JSON(w, http.StatusOK, domainToolResponse{Domain: domain, ExpiresAt: res.ExpiresAt, DaysLeft: daysLeft, Registrar: res.Registrar})
}
