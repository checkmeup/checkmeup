// Package httpsafe provides shared SSRF hardening for any outbound client
// that connects to a user-supplied host — notification delivery (Slack,
// generic webhooks) and monitor checks (uptime, port, SSL) alike. Without it,
// a customer could point a monitor or webhook at internal infrastructure
// (e.g. a cloud metadata endpoint) and have checkmeup's own server make the
// request for them.
package httpsafe

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// IsRestrictedIP reports whether ip falls outside the public internet:
// loopback, private, link-local (which includes the 169.254.169.254
// cloud-metadata address), unspecified, or multicast.
func IsRestrictedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

// BlockPrivateDial is a net.Dialer.Control hook that rejects connections to
// a restricted address. It runs after DNS resolution but before the TCP
// handshake, on the actual address being connected to — so a hostname that
// resolves differently between validation and connect (DNS rebinding) can't
// bypass it the way a pre-flight URL/IP check could. It also re-fires on
// every redirect hop an http.Client follows, since each hop dials again
// through the same Control-equipped Dialer.
func BlockPrivateDial(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("httpsafe: refusing to dial non-IP address %q", host)
	}
	if IsRestrictedIP(ip) {
		return fmt.Errorf("httpsafe: refusing to dial restricted address %s", ip)
	}
	return nil
}

// RefuseRedirects is an http.Client.CheckRedirect hook that stops a 3xx
// response from retargeting a one-shot delivery request (Slack/webhook) to a
// different address after the initial URL passed muster. A client that
// legitimately needs to follow redirects (e.g. an uptime checker) should
// rely on Dialer's per-hop BlockPrivateDial instead of this hook.
func RefuseRedirects(*http.Request, []*http.Request) error {
	return errors.New("httpsafe: redirects are not followed")
}

// Dialer returns a *net.Dialer hardened with BlockPrivateDial, suitable for
// direct TCP/TLS dials (a port or certificate check) or as the DialContext
// backing an *http.Transport.
func Dialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{Timeout: timeout, Control: BlockPrivateDial}
}
