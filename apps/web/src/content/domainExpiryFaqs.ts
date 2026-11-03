// Questions people search for around a domain's expiry, answered on
// /tools/domain-expiry-checker and marked up as FAQPage JSON-LD there.
// The TLDs with no published expiry match ADR domain-lookups.
export const domainExpiryFaqs: { q: string; a: string }[] = [
  {
    q: 'How do I check when a domain expires?',
    a: "Enter the domain above. The checker asks the registry that runs its top-level domain, over RDAP or WHOIS, and shows the registration's expiry date, the days left and the registrar. It works for any domain, not only your own: for most TLDs the expiry date is public.",
  },
  {
    q: 'What happens when a domain expires?',
    a: "It doesn't disappear at once. For .com and most other generic TLDs, the registrar usually holds it for a grace period of up to 45 days, when the owner can still renew at the normal price, though the site and email often stop working. Then comes a 30-day redemption period, when restoring it costs an extra fee, and about 5 days pending deletion before anyone can register it. Country-code TLDs set their own rules.",
  },
  {
    q: 'Why does an expired domain show next year as its expiry date?',
    a: "Many generic TLD registries renew a domain for a year the moment it expires, and refund the registrar if it's deleted during the grace period. So for up to 45 days after a lapse, the registry can show next year's date though nobody renewed it. The registrar's dashboard shows whether it was really renewed.",
  },
  {
    q: 'Why is there no expiry date for .de or .eu domains?',
    a: "Their registries, like those of .at, .ch, .li and .es, don't publish one over RDAP or WHOIS. Only the registrar and the owner know the date, so no public lookup can show it.",
  },
]
