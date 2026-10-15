# Domain lookups: RDAP, then WHOIS for registries without it

Date: 2026-10-15

A domain monitor reads its domain's registrar and expiry date once a day (PRD domain-expiry-monitoring). RDAP is the source wherever a registry runs it: structured JSON, one format, reached through rdap.org's bootstrap redirector. Several ccTLD registries don't run RDAP, among them .il, .io, .co, .me, .it, .ru, .jp, .cn and .se, and rdap.org answers 404 for their domains just as for an unregistered one; monitors under them could only ever show an error.

For those, the domain monitor asks the registry's WHOIS server over port 43 instead (internal/whois). WHOIS is free text in a different shape per registry, so each supported TLD has a fixed server and its own parser, written against that registry's real answers, which are kept as test fixtures. A TLD gets added only that way: there is no generic parsing of answers nobody has looked at, since a parser that guesses wrong could report a live domain as expiring or an expiring one as fine.

Some registries without RDAP don't publish an expiry date over WHOIS either (.de, .eu, .ch, .li, .at) or don't answer WHOIS from outside an allow list (.es). Those domains can't be monitored, and the API refuses them with that reason rather than accepting a monitor that will only ever be in error.

A daily query per monitored domain is well within what registries' WHOIS terms allow for a domain holder checking their own registration.

## Considered options

- A paid WHOIS/RDAP aggregation API: one format for every TLD, but another vendor, a per-lookup cost, and every monitored domain sent to a third party.
- Parsing any WHOIS answer with generic patterns ("expir", "paid-till"…): wider coverage, but formats differ in date order and field meaning, and a wrong date is worse than an honest error.
- Leaving these TLDs unsupported with a clearer error: no work, but .io and .co are common for exactly the developers and agencies Checkmeup is for.

## Consequences

- Each new WHOIS TLD needs a real answer to write its parser against, and a registry changing its format breaks only that TLD, into an error, not a wrong date.
- The rdap.org 404 message says the domain may be unregistered or its registry may not publish RDAP, since rdap.org can't tell the two apart.
