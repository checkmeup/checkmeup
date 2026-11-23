// Questions people search for around a domain's expiry, answered on
// /tools/domain-expiry-checker and marked up as FAQPage JSON-LD there.
// The TLDs with no published expiry match ADR domain-lookups.
export const domainExpiryFaqs: { q: string; a: string }[] = [
  {
    q: 'How do I check when a domain expires?',
    a: "Enter the domain above. The checker asks the registry that runs its top-level domain, over RDAP or WHOIS, and shows the registration's expiry date, the days left, the registrar and, where the registry lists the domain's status, its stage. It works for any domain, not only your own: for most TLDs the expiry date is public.",
  },
  {
    q: 'How do I know if a domain has expired?',
    a: "If the expiry date the checker shows is in the past, the registration has lapsed and the owner hasn't renewed it. A date in the future doesn't always mean it was renewed: for up to 45 days after a lapse, many registries show next year's date anyway (see below). Then the site and email often stop working, and the stage the checker shows, where the registry lists the domain's status, reads Grace period, Redemption or Pending delete instead of Active.",
  },
  {
    q: 'How do I check if a domain is in its redemption period?',
    a: "Check it above: where the registry lists the domain's status, the checker shows its stage, and Redemption means the registrar deleted the expired domain. The owner then has about 30 days to restore it through the registrar, for a fee on top of the renewal. After that the stage turns to Pending delete, and about 5 days later anyone can register the name.",
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
