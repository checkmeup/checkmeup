# Free SSL and domain expiry checkers

## Problem

Checkmeup's domain is young and gets almost no search traffic: Search Console showed 0 clicks from 2026-07-15 to 2026-10-14, nearly all impressions for the brand name. Developers do search for one-off checks ("ssl checker", "domain expiry check"), and nothing on the site answers them without an account.

## Users

Developers, freelancers and agencies checking one certificate or domain right now, who may not know Checkmeup. Not for bulk or scripted lookups: that's the public API and paid monitors.

## Success signal

By 2026-12-15, eight weeks after launch, Search Console shows impressions for non-brand queries on the tool pages (baseline 0), and GA4 counts at least one `sign_up` key event from a session that landed on a tool page.

## Scope

- `/tools/ssl-checker`: enter a hostname, see the certificate's expiry date, days left and issuer, or why the check failed.
- `/tools/domain-expiry-checker`: enter a domain, see its expiry date, days left and registrar, or why the lookup failed; the same RDAP-then-WHOIS lookup a domain monitor uses (ADR domain-lookups).
- No sign-up needed. Under each result, a link to sign up and monitor it automatically.
- Both pages prerendered, with their own title, description and `<h1>`, listed in `sitemap.xml` and linked from the site footer.

## Non-goals

- Carrying the checked host into the sign-up flow or a prefilled monitor form.
- DNS, port and cron expression tools (Inbox "Free tools without sign-up").
- Checking several hosts at once, history, or a result URL to share.
- Caching results.

## Acceptance criteria

- A hostname, or a URL with scheme, path or port, gets the certificate's expiry date, days left and issuer of `<host>:443`.
  → test: apps/api/internal/handler/tools_test.go "TestCheckSSL_ReadsTheCertificateOfTheHostInAURL"
- A host that resolves to a loopback, private, link-local or cloud-metadata address is refused, as for an SSL monitor.
  → test: apps/api/internal/handler/tools_test.go "TestCheckSSL_RefusesAPrivateAddress"
- A domain gets its expiry date, days left and registrar, through WHOIS for the TLDs ADR domain-lookups lists and RDAP for the rest.
  → test: apps/api/internal/handler/tools_test.go "TestCheckDomain_ReadsExpiryAndRegistrar"
- A domain under a TLD that publishes no expiry date (.de, .eu, .at, .ch, .li, .es) gets that reason, with no lookup made.
  → test: apps/api/internal/handler/tools_test.go "TestCheckDomain_ATLDWithNoPublishedExpiryIsRefusedWithoutALookup"
- An empty input gets `400` with no lookup made.
  → test: apps/api/internal/handler/tools_test.go "TestCheckDomain_AnEmptyInputIsRefusedWithoutALookup"
  → test: apps/api/internal/handler/tools_test.go "TestCheckSSL_RefusesWhatIsNotAHostname"
- One IP gets at most 5 checks a minute per tool, and all visitors together at most 30 domain lookups a minute; past either, `429` with `Retry-After`, and the page says to try again in a minute.
  → test: apps/api/internal/server/server_test.go "TestRoutes_FreeToolsAreRateLimited"
  → test: apps/web/src/lib/toolCheck.test.ts "says to try again in a minute when rate limited"
- A failed check shows its reason on the page, not a blank result.
  → test: apps/web/src/components/ToolChecker.test.ts "shows why a check failed instead of a result"
  → test: apps/api/internal/handler/tools_test.go "TestCheckSSL_AFailedCheckGivesItsReason"
- Each result shows a sign-up link offering to monitor it.
  → test: apps/web/src/components/ToolChecker.test.ts "shows the expiry, days left and registrar, with an offer to monitor it"
- Fetching either page without JavaScript returns its `<h1>`, title, description and Open Graph tags; both are in `sitemap.xml`.
- Both pages are linked from the footer of every marketing page.
  → test: apps/web/src/layouts/LandingLayout.test.ts "links the free checkers from the footer"

## Open questions

- An expired or untrusted certificate shows the TLS error, not its expiry date and issuer, since the check verifies the chain as monitors do. Settled when it's judged fine, or the tool reads the certificate without verifying it.
