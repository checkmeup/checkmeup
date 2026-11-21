# Free SSL, domain expiry and DNS checkers and a cron expression explainer

## Problem

Checkmeup's domain is young and gets almost no search traffic: Search Console showed 0 clicks from 2026-07-15 to 2026-10-14, nearly all impressions for the brand name. Developers do search for one-off checks ("ssl checker", "domain expiry check", "cron expression"), and nothing on the site answers them without an account.

## Users

Developers, freelancers and agencies checking one certificate or domain, or reading one cron expression, right now, who may not know Checkmeup. Not for bulk or scripted lookups: that's the public API and paid monitors.

## Success signal

By 2026-12-15, eight weeks after launch, Search Console shows impressions for non-brand queries on the tool pages (baseline 0), and GA4 counts at least one `sign_up` key event from a session that landed on a tool page.

## Scope

- `/tools/ssl-checker`: enter a hostname, see the certificate's expiry date, days left and issuer, or why the check failed.
- `/tools/domain-expiry-checker`: enter a domain, see its expiry date, days left and registrar, or why the lookup failed; the same RDAP-then-WHOIS lookup a domain monitor uses (ADR domain-lookups).
- `/tools/dns-lookup` (added 2026-10-31): enter a hostname, see its A, AAAA, CNAME, MX (with preference), TXT and NS records at once through the resolver DNS monitors use, each type's own failure reason, or why the lookup failed.
- No sign-up needed. Under each result, a link to sign up and monitor it automatically.
- Every page prerendered, with its own title, description and `<h1>`, listed in `sitemap.xml` and linked from the site footer.
- `/tools/domain-expiry-checker` (added 2026-11-03) answers the questions searchers ask around expiry (how to check, what happens after, why a lapsed domain shows next year's date, why some TLDs show none), marked up as FAQPage JSON-LD, since it draws most of the site's search impressions.
- `/tools/cron-expression` (added 2026-11-21): enter a 5-field cron expression, see what it means in plain English and its next five run times in UTC, which a cron monitor uses, and in the visitor's own time zone, or why it isn't valid. The result follows the input as the visitor types, read by the API with the cron monitor's own schedule parser. Common expressions fill the input in one click: `*/5 * * * *`, `0 * * * *`, `0 0 * * *`, `0 9 * * 1-5`, `0 0 * * 0`, `0 0 1 * *`. It leads to cron monitoring, the product's core, where the other tools lead to SSL, domain and DNS monitors.

## Non-goals

- Carrying the checked host or expression into the sign-up flow or a prefilled monitor form.
- A port checker (Inbox "Free tools without sign-up").
- Building a cron expression from dropdowns; 6-field (seconds) or Quartz syntax; a `CRON_TZ=` prefix; `@hourly`-style shortcuts; and `every 1h` intervals, which a cron monitor takes but no searcher types.
- Choosing a time zone other than UTC and the visitor's own.
- Checking several hosts at once, history, or a result URL to share.
- Caching results.

## Anti-goals

- The cron explainer accepting an expression the cron monitor's schedule parser can't read, or showing run times other than those the monitor would expect: the tool would then teach a schedule the product gets wrong.

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
  → test: apps/api/internal/handler/tools_test.go "TestCheckDNS_RefusesWhatIsNotAHostname"
- A hostname, or a URL with scheme or path, gets all six record types in a fixed order (A, AAAA, CNAME, MX, TXT, NS), a type with no records shown as none and a failed type with its reason.
  → test: apps/api/internal/handler/tools_test.go "TestCheckDNS_ListsEveryRecordTypeOfTheHostInAURL"
  → test: apps/api/internal/worker/worker_dns_test.go "TestLookupAllDNS/lists all six record types, in order, with the answers each has"
  → test: apps/web/src/components/DNSLookupChecker.test.ts "lists each record type with its values, with an offer to monitor them"
- A name with no records of any type, or whose lookups all failed, gets `422` with the reason.
  → test: apps/api/internal/handler/tools_test.go "TestCheckDNS_AFailedLookupGivesItsReason"
  → test: apps/web/src/components/DNSLookupChecker.test.ts "shows why a lookup failed instead of a result"
- One IP gets at most 5 checks a minute per tool, and all visitors together at most 30 domain lookups a minute; past either, `429` with `Retry-After`, and the page says to try again in a minute.
  → test: apps/api/internal/server/server_test.go "TestRoutes_FreeToolsAreRateLimited"
  → test: apps/web/src/lib/toolCheck.test.ts "says to try again in a minute when rate limited"
- A failed check shows its reason on the page, not a blank result.
  → test: apps/web/src/components/ToolChecker.test.ts "shows why a check failed instead of a result"
  → test: apps/api/internal/handler/tools_test.go "TestCheckSSL_AFailedCheckGivesItsReason"
- Each result shows a sign-up link offering to monitor it.
  → test: apps/web/src/components/ToolChecker.test.ts "shows the expiry, days left and registrar, with an offer to monitor it"
- Fetching any tool page without JavaScript returns its `<h1>`, title, description and Open Graph tags; all are in `sitemap.xml`.
  → test: apps/web/scripts/prerender.test.ts "serves every free checker with their <h1>, title, description and Open Graph tags"
  → test: apps/web/scripts/generate-sitemap.test.ts "lists every free checker"
- The domain expiry checker shows its questions and answers on the page and as FAQPage JSON-LD with the same text.
  → test: apps/web/src/views/DomainExpiryCheckerView.test.ts "emits FAQPage JSON-LD with the same questions and answers"
- Every tool page is linked from the footer of every marketing page.
  → test: apps/web/src/layouts/LandingLayout.test.ts "links the free checkers from the footer"
- Five whitespace-separated fields that the cron monitor's schedule parser reads get a plain-English description and the next five times the schedule fires after the request, with no grace period; `*/5 * * * *` reads "Every 5 minutes", `0 9 * * 1-5` "At 09:00, Monday through Friday", and each common expression has a description of its own.
  → test: apps/api/internal/handler/tools_test.go "TestExplainCron_DescribesTheExpressionAndListsItsNextFiveRuns"
  → test: apps/api/internal/handler/tools_test.go "TestExplainCron_DescribesEachCommonExpression"
  → test: apps/api/internal/handler/cronexpr_test.go "TestDescribeCron"
- Anything else (not 5 fields, a value out of range, an unknown name, an `every` interval, a `CRON_TZ=` prefix) gets `400` with the reason, naming the field where there is one ("Minute: 60 is above the maximum, 59"), and no run times.
  → test: apps/api/internal/handler/tools_test.go "TestExplainCron_RefusesWhatTheScheduleParserCantRead"
  → test: apps/web/src/components/CronExpressionChecker.test.ts "shows why an expression is not valid instead of a result"
- For the same expression and start time, the first run time plus the monitor's grace period is when the cron monitor expects its next ping.
  → test: apps/api/internal/handler/tools_test.go "TestExplainCron_TheFirstRunIsWhenACronMonitorExpectsItsPing"
- Each run time shows in UTC and in the browser's time zone.
  → test: apps/web/src/components/CronExpressionChecker.test.ts "explains the expression a moment after typing stops, with its runs in UTC and an offer to monitor it"
- One IP gets at most 60 explanations a minute; past it, `429` with `Retry-After`, and the page says to try again in a minute.
  → test: apps/api/internal/server/server_test.go "TestRoutes_FreeToolsAreRateLimited/sixty cron explanations a minute per IP"
  → test: apps/web/src/lib/toolCheck.test.ts "says to try again in a minute when rate limited"
- Clicking a common expression puts it in the input and shows its result.
  → test: apps/web/src/components/CronExpressionChecker.test.ts "puts a common expression in the input and explains it at once"
- The page names each of the five fields with its allowed values, gives an example of each of `*`, `,`, `-` and `/`, and answers: what a cron expression is, which time zone cron uses, how to run a job every N minutes, why day-of-month and day-of-week together mean either, and how to know a cron job stopped running; as FAQPage JSON-LD with the same text.
  → test: apps/web/src/views/CronExpressionView.test.ts "names each field with its allowed values and gives an example of each special character"
  → test: apps/web/src/views/CronExpressionView.test.ts "answers every cron expression question on the page"
  → test: apps/web/src/views/CronExpressionView.test.ts "emits FAQPage JSON-LD with the same questions and answers"

## Open questions

- An expired or untrusted certificate shows the TLS error, not its expiry date and issuer, since the check verifies the chain as monitors do. Settled when it's judged fine, or the tool reads the certificate without verifying it.
