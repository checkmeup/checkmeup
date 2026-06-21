package billing

// Unit tests for plans.go. No DB, no network — every function here is a
// pure function of (plan, current) over the package-level planLimits map,
// so these are plain table-driven tests, no ADR-010 integration setup
// needed.
//
// package billing (not billing_test): the "-1 = unlimited" sentinel
// documented on Limits.MonitorTotal/StatusPages isn't wired to any real
// plan today (Hobby/Solo/Startup/Enterprise all have finite limits per
// ADR-019's pricing table), so there's no db.Plan value that exercises it.
// Being in-package lets the unlimited-sentinel tests temporarily add a
// synthetic plan to planLimits (restored via t.Cleanup) to verify that
// branch actually works, rather than leaving it permanently untested until
// an "unlimited" tier ships.

import (
	"testing"

	"github.com/checkmeup/checkmeup/internal/db"
)

// withPlan temporarily adds (or overrides) a plan in planLimits for the
// duration of the test, restoring the original map afterward.
func withPlan(t *testing.T, plan db.Plan, limits Limits) {
	t.Helper()
	original, hadOriginal := planLimits[plan]
	planLimits[plan] = limits
	t.Cleanup(func() {
		if hadOriginal {
			planLimits[plan] = original
		} else {
			delete(planLimits, plan)
		}
	})
}

func TestGetLimits(t *testing.T) {
	cases := []struct {
		plan db.Plan
		want Limits
	}{
		{db.PlanHobby, Limits{MonitorTotal: 10, StatusPages: 1, MinIntervalMins: 5}},
		{db.PlanSolo, Limits{MonitorTotal: 30, StatusPages: 3, MinIntervalMins: 1}},
		{db.PlanStartup, Limits{MonitorTotal: 100, StatusPages: 10, MinIntervalMins: 1}},
		{db.PlanEnterprise, Limits{MonitorTotal: 1000, StatusPages: 100, MinIntervalMins: 1}},
	}
	for _, tc := range cases {
		t.Run(string(tc.plan), func(t *testing.T) {
			if got := GetLimits(tc.plan); got != tc.want {
				t.Fatalf("GetLimits(%q) = %+v, want %+v", tc.plan, got, tc.want)
			}
		})
	}

	t.Run("unknown plan falls back to Hobby", func(t *testing.T) {
		want := planLimits[db.PlanHobby]
		if got := GetLimits(db.Plan("not-a-real-plan")); got != want {
			t.Fatalf("GetLimits(unknown) = %+v, want Hobby's limits %+v", got, want)
		}
	})

	t.Run("empty plan falls back to Hobby", func(t *testing.T) {
		want := planLimits[db.PlanHobby]
		if got := GetLimits(db.Plan("")); got != want {
			t.Fatalf("GetLimits(\"\") = %+v, want Hobby's limits %+v", got, want)
		}
	})
}

func TestCheckMonitorLimit(t *testing.T) {
	cases := []struct {
		name    string
		plan    db.Plan
		current int
		wantErr error
	}{
		{"Hobby under limit", db.PlanHobby, 9, nil},
		{"Hobby at limit", db.PlanHobby, 10, ErrMonitorLimit},
		{"Hobby over limit", db.PlanHobby, 11, ErrMonitorLimit},
		{"Solo under limit", db.PlanSolo, 29, nil},
		{"Solo at limit", db.PlanSolo, 30, ErrMonitorLimit},
		{"Startup under limit", db.PlanStartup, 99, nil},
		{"Startup at limit", db.PlanStartup, 100, ErrMonitorLimit},
		{"Enterprise under limit", db.PlanEnterprise, 999, nil},
		{"Enterprise at limit", db.PlanEnterprise, 1000, ErrMonitorLimit},
		{"zero current is always under any real plan's limit", db.PlanHobby, 0, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckMonitorLimit(tc.plan, tc.current); err != tc.wantErr {
				t.Fatalf("CheckMonitorLimit(%q, %d) = %v, want %v", tc.plan, tc.current, err, tc.wantErr)
			}
		})
	}

	t.Run("MonitorTotal -1 means unlimited", func(t *testing.T) {
		withPlan(t, db.Plan("unlimited-test-plan"), Limits{MonitorTotal: -1})
		if err := CheckMonitorLimit(db.Plan("unlimited-test-plan"), 1_000_000); err != nil {
			t.Fatalf("want no error for an unlimited plan regardless of count, got %v", err)
		}
	})
}

func TestCheckStatusPageLimit(t *testing.T) {
	cases := []struct {
		name    string
		plan    db.Plan
		current int
		wantErr error
	}{
		{"Hobby under limit", db.PlanHobby, 0, nil},
		{"Hobby at limit", db.PlanHobby, 1, ErrStatusPageLimit},
		{"Solo under limit", db.PlanSolo, 2, nil},
		{"Solo at limit", db.PlanSolo, 3, ErrStatusPageLimit},
		{"Startup at limit", db.PlanStartup, 10, ErrStatusPageLimit},
		{"Enterprise under limit", db.PlanEnterprise, 99, nil},
		{"Enterprise at limit", db.PlanEnterprise, 100, ErrStatusPageLimit},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckStatusPageLimit(tc.plan, tc.current); err != tc.wantErr {
				t.Fatalf("CheckStatusPageLimit(%q, %d) = %v, want %v", tc.plan, tc.current, err, tc.wantErr)
			}
		})
	}

	t.Run("StatusPages -1 means unlimited", func(t *testing.T) {
		withPlan(t, db.Plan("unlimited-test-plan"), Limits{StatusPages: -1})
		if err := CheckStatusPageLimit(db.Plan("unlimited-test-plan"), 1_000_000); err != nil {
			t.Fatalf("want no error for an unlimited plan regardless of count, got %v", err)
		}
	})
}

func TestClampInterval(t *testing.T) {
	cases := []struct {
		name          string
		plan          db.Plan
		requestedMins int
		wantMins      int
		wantErr       bool
	}{
		{"Hobby below minimum is rejected, not silently clamped", db.PlanHobby, 1, 5, true},
		{"Hobby at minimum is accepted unchanged", db.PlanHobby, 5, 5, false},
		{"Hobby above minimum is accepted unchanged", db.PlanHobby, 30, 30, false},
		{"Solo at its 1-minute minimum is accepted", db.PlanSolo, 1, 1, false},
		{"Solo below its minimum is rejected", db.PlanSolo, 0, 1, true},
		{"Enterprise above minimum is accepted unchanged", db.PlanEnterprise, 60, 60, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ClampInterval(tc.plan, tc.requestedMins)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ClampInterval(%q, %d) error = %v, wantErr %v", tc.plan, tc.requestedMins, err, tc.wantErr)
			}
			if got != tc.wantMins {
				t.Fatalf("ClampInterval(%q, %d) = %d, want %d", tc.plan, tc.requestedMins, got, tc.wantMins)
			}
		})
	}
}
