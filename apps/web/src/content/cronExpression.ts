// The reference text of the free cron explainer at /tools/cron-expression
// (PRD free-tools). The ranges and names are those of the parser a cron
// monitor uses (apps/api/internal/handler/cronexpr.go).

// One click puts each in the explainer's input.
export const commonCronExpressions = [
  '*/5 * * * *',
  '0 * * * *',
  '0 0 * * *',
  '0 9 * * 1-5',
  '0 0 * * 0',
  '0 0 1 * *',
]

export const cronFields: { name: string; values: string }[] = [
  { name: 'Minute', values: '0–59' },
  { name: 'Hour', values: '0–23' },
  { name: 'Day of month', values: '1–31' },
  { name: 'Month', values: '1–12 or JAN–DEC' },
  { name: 'Day of week', values: '0–6 or SUN–SAT (0 is Sunday)' },
]

export const cronCharacters: { char: string; meaning: string; example: string; reads: string }[] = [
  { char: '*', meaning: 'every value', example: '* * * * *', reads: 'every minute' },
  { char: ',', meaning: 'a list of values', example: '0 9,17 * * *', reads: 'at 09:00 and 17:00' },
  {
    char: '-',
    meaning: 'a range of values',
    example: '0 9 * * 1-5',
    reads: 'at 09:00, Monday through Friday',
  },
  {
    char: '/',
    meaning: 'a step through a range',
    example: '*/15 * * * *',
    reads: 'every 15 minutes',
  },
]

// Answered on the page and marked up as FAQPage JSON-LD there.
export const cronExpressionFaqs: { q: string; a: string }[] = [
  {
    q: 'What is a cron expression?',
    a: 'A cron expression is the schedule line of a cron job: five fields separated by spaces, for the minute, hour, day of month, month and day of week. The job runs at every minute that matches all five. "30 2 * * *" runs at 02:30 every day.',
  },
  {
    q: 'Which time zone does cron use?',
    a: "Cron reads the schedule in the time zone of the machine it runs on, usually UTC on servers and in containers. GitHub Actions and most hosted schedulers use UTC too. That's why the explainer shows each run in UTC and in your own time zone.",
  },
  {
    q: 'How do I run a cron job every 5 minutes?',
    a: 'Write "*/5 * * * *": the step "*/5" in the minute field means every fifth minute, starting at minute 0. "*/10" runs every 10 minutes, and "0 */2 * * *" runs every 2 hours, on the hour.',
  },
  {
    q: 'What happens when I set both the day of month and the day of week?',
    a: 'The job runs on a day that matches either of them, not both. "0 0 1 * 1" runs at midnight on the 1st of each month and on every Monday. When either field is "*", only the other one limits the days.',
  },
  {
    q: 'How do I know a cron job stopped running?',
    a: "Cron says nothing when a job fails to start, crashes, or its server goes away. A cron monitor waits for the job to ping a URL after each run, and alerts you when a ping doesn't arrive on schedule. Checkmeup reads the same five-field expression, so the times above are when it expects each ping.",
  },
]
