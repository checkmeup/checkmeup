# Who Checkmeup is for, and its license

Date: 2026-11-20

The primary customer is the freelance web developer or solo agency operator who builds and maintains sites for small local businesses: technical enough to set up monitors alone, and hurt in reputation when a client's SSL or domain expires or the site goes down unnoticed. Copy speaks to "give your clients a professional status page", not to white-label. A broad developer audience (Hacker News, Indie Hackers, r/webdev, r/selfhosted) is a second, free-tier channel for volume and feedback, not the paid path.

The source is public under the Business Source License 1.1: any use is allowed, including production use and monitoring clients' systems, except offering the software as a hosted monitoring service that competes with checkmeup.net. On 2030-07-01 it converts to Apache 2.0 by itself. No CLA; contributions come in under the same terms.

If checkmeup.net ever shuts down, every account holder gets an email at least 90 days before the service stops, a refund of any prepaid time past the shutdown date (annual plans), and the source stays public under the same license so they can run it themselves. The Terms of Service (section 11) bind this; the homepage and the About page promise it, as the answer to a one-person product's obvious risk. Adding it to the Terms didn't bump `legal.CurrentVersion`: it only gives users more, so it isn't a material change that needs re-acceptance.

## Considered options

- Non-technical small businesses directly: they wouldn't set monitors up themselves.
- Formal multi-person agencies: need custom-domain white-label pages first. Revisit when custom domains ship.
- MIT (until 2026-07-01): allows a verbatim competing hosted clone.
- GPL v3: copyleft doesn't restrict hosting a competing service.
- Closed source: breaks the project's openness, including publishing its decisions.
