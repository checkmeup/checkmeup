package worker

import (
	"testing"
	"time"
)

func TestNextPing(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	t.Run("interval schedule adds the interval plus grace", func(t *testing.T) {
		got := NextPing("every 1h", now, 5)
		want := now.Add(time.Hour + 5*time.Minute)
		if !got.Equal(want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("cron expression uses the next scheduled fire time plus grace", func(t *testing.T) {
		// "0 13 * * *" = 13:00 daily; from 12:00 the next fire is 13:00 the same day.
		got := NextPing("0 13 * * *", now, 10)
		want := time.Date(2026, 6, 19, 13, 10, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("unparseable schedule falls back to 1 hour plus grace", func(t *testing.T) {
		got := NextPing("garbage schedule", now, 5)
		want := now.Add(time.Hour + 5*time.Minute)
		if !got.Equal(want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestParseEveryDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2hours", 2 * time.Hour},
		{"45mins", 45 * time.Minute},
		{"3hr", 3 * time.Hour},
		{"", 0},
		{"abc", 0},
		{"0h", 0},
		{"-5m", 0},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := ParseEveryDuration(tc.in); got != tc.want {
				t.Fatalf("ParseEveryDuration(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
