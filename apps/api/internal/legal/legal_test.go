package legal

import (
	"os"
	"regexp"
	"testing"
	"time"
)

func TestCurrentVersion_IsAValidDate(t *testing.T) {
	if _, err := time.Parse(time.DateOnly, CurrentVersion); err != nil {
		t.Fatalf("CurrentVersion %q is not a valid YYYY-MM-DD date: %v", CurrentVersion, err)
	}
}

// CurrentVersion gates re-prompting users to accept the Terms/Privacy pages
// (see the package doc comment), so it must always match the effective date
// rendered on both pages — a mismatch means a doc was updated without
// bumping the version, or vice versa.
func TestCurrentVersion_MatchesFrontendEffectiveDates(t *testing.T) {
	for _, path := range []string{
		"../../../web/src/views/TermsView.vue",
		"../../../web/src/views/PrivacyView.vue",
	} {
		t.Run(path, func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			match := regexp.MustCompile(`effectiveDate\s*=\s*'([^']+)'`).FindSubmatch(src)
			if match == nil {
				t.Fatalf("no `effectiveDate = '...'` assignment found in %s", path)
			}

			if got := string(match[1]); got != CurrentVersion {
				t.Fatalf("%s effectiveDate is %q, want it to match legal.CurrentVersion %q", path, got, CurrentVersion)
			}
		})
	}
}
