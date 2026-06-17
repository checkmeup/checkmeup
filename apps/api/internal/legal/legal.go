// Package legal tracks the current version of the Terms of Service and
// Privacy Policy. They're accepted together as a single version (EP-21) —
// bump CurrentVersion whenever either document changes materially, which
// re-prompts every user to accept again on next sign-in.
package legal

// CurrentVersion must match the effective date shown on the /terms and
// /privacy pages (apps/web/src/views/TermsView.vue, PrivacyView.vue).
const CurrentVersion = "2026-06-17"
