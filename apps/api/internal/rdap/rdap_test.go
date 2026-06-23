package rdap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	t.Cleanup(srv.Close)
	return NewClientWithHTTPClient(srv.Client(), srv.URL+"/domain/")
}

func TestLookup(t *testing.T) {
	t.Run("success extracts registrar and expiry from a realistic RDAP body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/rdap+json")
			_, _ = fmt.Fprint(w, `{
				"events": [
					{"eventAction": "registration", "eventDate": "2010-01-01T00:00:00Z"},
					{"eventAction": "expiration", "eventDate": "2027-06-01T00:00:00Z"}
				],
				"entities": [
					{
						"roles": ["registrar"],
						"handle": "REG-123",
						"vcardArray": ["vcard", [["version", {}, "text", "4.0"], ["fn", {}, "text", "Example Registrar, LLC"]]]
					}
				]
			}`)
		}))
		c := testClient(t, srv)

		result, err := c.Lookup("example.com")
		if err != nil {
			t.Fatalf("Lookup: %v", err)
		}
		if result.Registrar != "Example Registrar, LLC" {
			t.Fatalf("want registrar %q, got %q", "Example Registrar, LLC", result.Registrar)
		}
		want := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
		if !result.ExpiresAt.Equal(want) {
			t.Fatalf("want expiresAt %v, got %v", want, result.ExpiresAt)
		}
	})

	t.Run("falls back to the entity handle when vcardArray has no fn", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprint(w, `{
				"events": [{"eventAction": "expiration", "eventDate": "2027-01-01T00:00:00Z"}],
				"entities": [{"roles": ["registrar"], "handle": "REG-456"}]
			}`)
		}))
		c := testClient(t, srv)

		result, err := c.Lookup("example.com")
		if err != nil {
			t.Fatalf("Lookup: %v", err)
		}
		if result.Registrar != "REG-456" {
			t.Fatalf("want registrar fallback to handle REG-456, got %q", result.Registrar)
		}
	})

	t.Run("404 is a not-found error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		c := testClient(t, srv)

		if _, err := c.Lookup("this-domain-does-not-exist.invalid"); err == nil {
			t.Fatal("want an error for a 404 response")
		}
	})

	t.Run("non-2xx, non-404 status is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		c := testClient(t, srv)

		if _, err := c.Lookup("example.com"); err == nil {
			t.Fatal("want an error for a 500 response")
		}
	})

	t.Run("malformed JSON body is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprint(w, "not json")
		}))
		c := testClient(t, srv)

		if _, err := c.Lookup("example.com"); err == nil {
			t.Fatal("want an error for a malformed body")
		}
	})

	t.Run("missing expiration event is an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprint(w, `{"events": [{"eventAction": "registration", "eventDate": "2010-01-01T00:00:00Z"}]}`)
		}))
		c := testClient(t, srv)

		if _, err := c.Lookup("example.com"); err == nil {
			t.Fatal("want an error when the response has no expiration event")
		}
	})

	t.Run("connection failure is an error", func(t *testing.T) {
		c := NewClientWithHTTPClient(http.DefaultClient, "http://127.0.0.1:0/domain/")
		if _, err := c.Lookup("example.com"); err == nil {
			t.Fatal("want an error for an unreachable server")
		}
	})
}
