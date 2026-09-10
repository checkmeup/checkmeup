import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Another quiet stretch of versions, and almost none of it is visible from the outside: four releases that mostly moved code around rather than adding anything you can click. That's the boring half. The interesting half is that moving the code around is what finally exposed a bug that had been sitting in the uptime monitor's edit form for months, quietly serving people their own unsaved edits.",
  },
  {
    type: 'h2',
    text: 'Fourteen copies of the same form',
  },
  {
    type: 'p',
    text: 'Every monitor type in checkmeup has a create screen and an edit screen, and until this cycle each of those 14 screens carried its own full copy of the form — the same fields, the same validation, the same page chrome, written out twice per monitor type. Nothing was broken about that, which is exactly why it survived so long: every file sat comfortably under the size limit my own architecture audit checks for, so the audit reported clean for months while the same uptime form was being maintained in two places. Every field I added and every label I reworded, I did twice.',
  },
  {
    type: 'p',
    text: "The audit now looks for near-duplicate sibling files rather than just big ones, and it immediately flagged five real pairs. Those are now five shared form components, one shared page shell, and five plain-TypeScript modules holding the form logic. The views went from 3,821 lines to 1,206. What you get out of that is not a feature — it's that the next uptime-form change lands in one place instead of two, and stops drifting apart in between.",
  },
  {
    type: 'p',
    text: "Two deliberate non-decisions in there. I did not merge each create/edit pair into one view behind an isEdit flag — that trades duplicated markup for a conditional maze, and the create/edit split is the route structure anyway. And I left the SSL and domain forms alone: they look similar on a line-count metric, but the edit screens carry alert settings the create screens don't and render the hostname read-only, so a shared component there would have been an abstraction over two genuinely different screens. Duplication is cheaper than the wrong abstraction.",
  },
  {
    type: 'h2',
    text: 'The bug that fell out of it',
  },
  {
    type: 'p',
    text: "With the form logic finally sitting in plain modules instead of inside 14 components, I could run mutation testing against it — a tool that deliberately breaks your code line by line and checks whether any test notices. Line coverage tells you a line ran. A surviving mutant tells you nothing would have complained if that line were wrong. Two survivors pointed at the same blind spot: nothing verified that an existing monitor's JSON assertions and notification channels load into the edit form at all.",
  },
  {
    type: 'p',
    text: 'Reading that gap turned up a real bug. The form copied the list of JSON assertions but not the assertion objects inside it, and each field binds directly to those objects — so editing a JSON assertion was writing straight into the cached copy of the monitor the app holds in memory. Open an uptime monitor, change an assertion, decide against it and navigate away without saving, come back: the app would show you your abandoned edit as if the server had it. A reload cleared it. Fixed now, along with seven new tests covering that path and the validation edges nothing was asserting on.',
  },
  {
    type: 'p',
    text: "Worth being clear that the refactor didn't cause this — both aliasing bugs predate it, and the extraction copied them across faithfully. It just made them findable. The refactor did introduce three regressions of its own, and the 683-test suite caught all three before they left my machine, which is the entire argument for not editing tests to accommodate a refactor.",
  },
  {
    type: 'h2',
    text: 'Also this cycle',
  },
  {
    type: 'ul',
    items: [
      '"Why Monitoring Matters: Finding Out Before Users Do" — a new post on the blog, less about checkmeup and more about the failure modes that produce no signal at all, and why a green dashboard proves less than most people assume.',
      'A site-verification key file, so search engines can confirm I own checkmeup.net — dull, necessary, one step of a slow SEO grind.',
      'Mutation testing is now a repeatable tool in this repo rather than a one-off experiment, pointed at the pure-logic modules and deliberately kept out of CI, where it would be far too slow to be worth it.',
      'Housekeeping: the running hours log split into one file per month after it hit 146 rows, a .gitignore rule that had only ever matched a file literally named ".log", and Codacy\'s generated config files no longer dirtying the tree on every local analysis run.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "Two cycles in a row now that were mostly foundation rather than features, and the next one should swing back the other way — the roadmap's Now column is still where the user-facing work lives. Full commit history and every architecture decision record are on GitHub if you want the unfiltered version.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
