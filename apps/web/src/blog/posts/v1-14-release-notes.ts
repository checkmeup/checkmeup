import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "This release started from an unglamorous place: checking Checkmeup's site against a payment processor's domain-verification checklist. That surfaced two real gaps — the Terms never named who actually operates Checkmeup, and the refund policy was buried as one sentence instead of its own document. Pulling that thread turned into a full legal read-through, which found a few more things worth fixing while the page was open.",
  },
  {
    type: 'h3',
    text: 'Terms of Service, properly',
  },
  {
    type: 'p',
    text: 'The Terms grew from 12 sections to 15. Checkmeup now explicitly states who\'s behind it: Andrew Molyuk, a sole proprietor (עוסק פטור) based in Israel — previously the Terms only ever said "we"/"us". New sections: Eligibility (you must be of legal age and have authority to bind your organization), Indemnification (you\'re responsible for claims arising from your own misuse — most relevantly, monitoring a system you don\'t have the right to monitor, which the Acceptable Use section already prohibited but had no teeth behind), and a proper Governing Law and Disputes section: any dispute goes exclusively to Israeli courts, no mandatory arbitration. A closing Miscellaneous section adds the standard boilerplate that was missing entirely — entire agreement, severability, no waiver, assignment on a sale or merger.',
  },
  {
    type: 'h3',
    text: 'A standalone Refund Policy',
  },
  {
    type: 'p',
    text: "The 30-day, no-questions-asked refund policy existed, but only as one sentence inside the Terms' billing section and an FAQ answer — not its own document, unlike Terms and Privacy. It's now a dedicated page at /refund, linked from the footer alongside Terms and Privacy, and cross-referenced from the Terms' billing section instead of just stated inline.",
  },
  {
    type: 'h3',
    text: "Privacy Policy: controller identity and children's privacy",
  },
  {
    type: 'p',
    text: "Two fixes here, mirroring the Terms work. The policy now names Andrew Molyuk as the data controller — GDPR Art. 13 specifically requires disclosing the controller's identity, not just a product name. And there's a new Children's Privacy section stating the Service isn't directed at anyone under the Terms' eligibility age and that we don't knowingly collect data from children.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'p',
    text: "The production Docker image now carries the standard OpenContainers labels — source, title, description, licenses, vendor, plus version/revision/build-date populated at build time from git — so anyone (including GitHub itself) can tell what's actually running without guessing. GHCR also now has actual cleanup: every deploy left the previous image sitting in the registry forever with nothing pruning it, so there's now a `make ghcr-clean` step that runs automatically after every deploy and keeps only the 5 most recent image versions.",
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Microsoft Teams alerts are still next on the board. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this — including, now, the why behind the Terms of Service.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
