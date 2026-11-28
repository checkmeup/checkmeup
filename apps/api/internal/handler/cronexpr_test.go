package handler

import "testing"

func TestDescribeCron_ReadsEachShapeOfExpressionInPlainEnglish(t *testing.T) {
	cases := map[string]string{
		"* * * * *":        "Every minute",
		"*/1 * * * *":      "Every minute",
		"30 * * * *":       "At 30 minutes past every hour",
		"1 * * * *":        "At 1 minute past every hour",
		"0 */2 * * *":      "Every 2 hours",
		"0 9,17 * * *":     "At 09:00 and 17:00, every day",
		"*/15 9-17 * * *":  "Every 15 minutes, between 09:00 and 17:59",
		"0,30 * * * *":     "At minutes 0 and 30",
		"0-10 * * * *":     "At minutes 0 through 10",
		"*/10 */3 * * *":   "Every 10 minutes, every 3 hours",
		"0 0 1,15 * *":     "At 00:00, on days 1 and 15 of the month",
		"0 0 */2 * *":      "At 00:00, every 2 days",
		"0 0 1 * 1":        "At 00:00, on day 1 of the month or on Monday",
		"0 0 ? * SUN":      "At 00:00, on Sunday",
		"0 6 * * 1,3,5":    "At 06:00, on Monday, Wednesday and Friday",
		"0 0 1 1 *":        "At 00:00, on day 1 of the month, in January",
		"0 0 1 */3 *":      "At 00:00, on day 1 of the month, every 3 months",
		"0 12 * JAN-MAR *": "At 12:00, in January through March",
		"5 4 * * sun":      "At 04:05, on Sunday",
		"*/5 * * * *":      "Every 5 minutes",
		"10-50/20 * * * *": "Every 20 minutes from 10 through 50",
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			_, fields, err := parseCronExpression(expr)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := describeCron(fields); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestDescribeSchedule_WordsBothFormsOfACronMonitorsSchedule(t *testing.T) {
	cases := map[string]string{
		"0 0 * * *":   "At 00:00, every day",
		"every 1d":    "Every day",
		"every 30m":   "Every 30 minutes",
		"every 2 hrs": "Every 2 hours",
		"Every 1W":    "Every week",
		"every 36h":   "Every 36 hours",
		"every soon":  "",
	}
	for schedule, want := range cases {
		if got := describeSchedule(schedule); got != want {
			t.Errorf("describeSchedule(%q) = %q, want %q", schedule, got, want)
		}
	}
}
