# Suggest a feature from the app

## Problem

Feature ideas came only by email or GitHub Issues, which signed-in users rarely bother with; a form where they already are lowers the effort, without building a public board (CLAUDE.md).

## Users

Signed-in users.

## Success signal

Suggestions arrive in the founder's inbox from inside the app, with no spam reaching it.

## Scope

- A suggestion box in Settings with the account email shown, and a link to it from the Docs page's "Need help?" section.
- Each suggestion stored with org, user and time, and emailed to the founder.
- Rate limits per IP and per org.

## Non-goals

- Statuses, voting, comments or a public roadmap.

## Acceptance criteria

- The Send button stays disabled until there is text; sending shows a confirmation.
  → test: apps/web/src/views/SettingsView.test.ts "disables the send button until the suggestion has content"
  → test: apps/web/src/views/SettingsView.test.ts "submits the suggestion and shows a confirmation on success"
- Empty or over-long text is refused; accepted text is trimmed and stored with the org and user.
  → test: apps/api/internal/handler/suggestions_test.go "TestSubmitSuggestion/empty text after trimming"
  → test: apps/api/internal/handler/suggestions_test.go "TestSubmitSuggestion/text over the length limit is rejected"
  → test: apps/api/internal/handler/suggestions_test.go "TestSubmitSuggestion/success trims whitespace and persists org/user/text"
- Each suggestion is emailed to the founder.
  → test: apps/api/internal/email/founder_test.go "TestSendFeatureSuggestion"
- A sixth suggestion within an hour from one IP, or a twenty-first from one org, is refused with a clear message.
  → test: apps/api/internal/server/server_test.go "TestSuggestionRateLimited"
- The Docs page's "Need help?" section links to in-app feedback, email and GitHub Issues.
  → test: apps/web/src/views/DocsView.test.ts "routes help to in-app feedback, email, and GitHub Issues rather than a feature board"
