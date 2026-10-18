# FAQ page

## Problem

The same questions (what counts as a monitor, how plans change, how long data is kept) came in by email, and the only answers lived in a billing list on the pricing page.

## Users

Visitors deciding whether to sign up, and users looking for a quick answer.

## Success signal

Common questions are answered on one page that other pages can link into by section.

## Scope

- A public `/faq` page grouped into categories (Getting started, Billing & plans, Monitors & alerts, Status pages, Privacy & security), each with an anchor.
- The billing questions kept in one place and shown on the pricing page from the same source.
- Links from the footer, the nav and the sign-up page.

## Non-goals

- Search, a help center, or a ticket system (CLAUDE.md: no public feedback board).

## Acceptance criteria

- The page lists every category in a sidebar and shows each category's questions and answers.
  → test: apps/web/src/views/FaqView.test.ts "renders a sidebar nav link for every category"
  → test: apps/web/src/views/FaqView.test.ts "renders every category heading and its questions and answers"
- The pricing page shows the billing questions from the same entries as `/faq`, not a copy.
  → test: apps/web/src/views/PricingView.test.ts "renders the billing FAQ entries from findFaqCategory"
- Each category has an anchor (such as `/faq#billing`) that opens the page at that section.
  → test: apps/web/src/views/FaqView.test.ts "gives the billing category the #billing anchor its section opens at"
  → test: apps/web/src/router/index.test.ts "opens a URL with a hash, such as /faq#billing, at the element with that id"
- The sign-up page links to the "Getting started" section.
  → test: apps/web/src/views/SignUpView.test.ts "links to the Getting started section of the FAQ"
