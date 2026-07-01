export interface FaqEntry {
  q: string
  a: string
}

export interface FaqCategory {
  id: string
  label: string
  entries: FaqEntry[]
}

export const faqCategories: FaqCategory[] = [
  {
    id: 'getting-started',
    label: 'Getting started',
    entries: [
      {
        q: 'How do I create my first monitor?',
        a: "Sign up, pick a monitor type from the dashboard, and fill in the form. There's no setup wizard — the whole thing takes under a minute.",
      },
      {
        q: 'What can I monitor?',
        a: "Five things: scheduled jobs that should run on a cadence (cron monitors), URLs that should always respond (uptime monitors), TLS certificates that shouldn't be allowed to quietly expire (SSL monitors), domain registrations that shouldn't be allowed to lapse (domain monitors), and raw host:port connectivity for anything that isn't HTTP — mail servers, databases, custom daemons (port monitors).",
      },
      {
        q: 'Do I need to install anything?',
        a: 'No. Cron monitors work by calling a ping URL at the end of your job (a single curl line). Uptime, SSL, domain, and port monitors just need a URL, hostname, domain, or host:port — nothing runs on your servers.',
      },
    ],
  },
  {
    id: 'billing',
    label: 'Billing & plans',
    entries: [
      {
        q: 'Do I need a credit card to start?',
        a: 'No. The Hobby plan is free forever with no credit card required.',
      },
      {
        q: 'What counts as a "monitor"?',
        a: 'Each cron job, uptime URL, SSL certificate, domain, or port you track counts as one monitor. The limit applies to the total across all types.',
      },
      {
        q: 'Can I change plans later?',
        a: 'Yes — upgrade or downgrade at any time. Billing adjusts automatically on your next cycle.',
      },
      {
        q: 'What happens if I exceed my monitor limit?',
        a: "You'll see an error when creating a new monitor and can choose to upgrade. Existing monitors keep running — we never pause them mid-cycle.",
      },
      {
        q: 'Which payment methods do you accept?',
        a: 'All major credit and debit cards via LemonSqueezy, which handles global tax compliance so you pay the right tax wherever you are.',
      },
      {
        q: 'Is there a refund policy?',
        a: "Yes. Contact us within 30 days of any charge and we'll issue a full refund, no questions asked.",
      },
    ],
  },
  {
    id: 'monitors',
    label: 'Monitors & alerts',
    entries: [
      {
        q: 'How do alerts work?',
        a: "Add notification channels in Settings — Telegram, email, Slack, or webhook — then assign them to each monitor. Alerts are capped at 3 per incident by default (configurable per monitor) so a flapping check doesn't spam you; a single alert event counts as one toward the cap regardless of how many channels it fires on. The recovery alert always sends regardless of the cap.",
      },
      {
        q: "What's the minimum check interval?",
        a: "5 minutes for uptime and port monitors on Hobby, 1 minute on paid plans. SSL certificates and domains are both checked once a day. Cron monitors alert based on your job's own schedule plus a grace period you choose.",
      },
      {
        q: 'Can a port monitor alert me if a port becomes reachable, not just unreachable?',
        a: 'Yes — set its expected state to "closed" instead of the default "open". That flips the check into a security monitor: it alerts if a port that should be firewalled off (a database bound to a public interface, an admin panel, a debug port) unexpectedly starts accepting connections, rather than alerting when a service goes down.',
      },
      {
        q: 'Can I check the response body, not just the status code?',
        a: "Yes — every uptime monitor, on every plan including Hobby, supports an optional keyword check: require text to be present, or fail if it is. Useful for catching a maintenance page served with a 200 or an error embedded in a JSON response. We search the first 512 KB of the body; it's never stored, only the pass/fail reason.",
      },
      {
        q: 'Can I mute alerts for a noisy monitor?',
        a: 'Yes — each monitor has its own alert toggle. Muted monitors keep tracking status and history; only notifications are suppressed.',
      },
      {
        q: 'Can I pause checks during planned maintenance?',
        a: "Yes — schedule a maintenance window covering any combination of monitors. While it's active, those monitors aren't checked at all: no alerts, no incidents, and uptime stats stay untouched.",
      },
    ],
  },
  {
    id: 'status-pages',
    label: 'Status pages',
    entries: [
      {
        q: 'Do visitors need an account to view my status page?',
        a: 'No. Status pages are public at checkmeup.net/status/your-slug — no login required, nothing for visitors to install.',
      },
      {
        q: 'Can I use a custom domain or subdomain?',
        a: "Not today, by design — status pages live on the same domain (a path, not a subdomain), so there's no DNS to configure before sharing a link.",
      },
      {
        q: 'What do visitors see during planned maintenance?',
        a: 'Covered monitors show "Under maintenance" with your optional message, instead of up/down — and it doesn\'t trigger the page\'s outage banner.',
      },
    ],
  },
  {
    id: 'privacy',
    label: 'Privacy & security',
    entries: [
      {
        q: 'How is my password stored?',
        a: 'As a bcrypt hash — never in plain text. See the full Privacy Policy for what else we collect and why.',
      },
      {
        q: 'Where is my data hosted?',
        a: 'Application hosting and the database run on Hetzner in Germany, within the EU.',
      },
      {
        q: 'Can I delete my account?',
        a: "Yes — email us and we'll delete your account, monitors, check history, and status pages. There's no self-service delete button yet.",
      },
    ],
  },
]

export function findFaqCategory(id: string): FaqCategory | undefined {
  return faqCategories.find((c) => c.id === id)
}
