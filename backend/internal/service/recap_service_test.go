package service

import (
	"testing"
	"time"

	"github.com/blueship581/mindgarden/backend/internal/model"
)

func TestWeeklyBounds(t *testing.T) {
	// 2026-09-23 is a Wednesday, 14:00 Beijing time.
	now := time.Date(2026, 9, 23, 14, 0, 0, 0, beijingZone)
	from, to, start, end := weeklyBounds(now)
	if got := start.Format("2006-01-02"); got != "2026-09-21" {
		t.Fatalf("start = %s, want 2026-09-21 (Monday)", got)
	}
	if got := end.Format("2006-01-02"); got != "2026-09-23" {
		t.Fatalf("end = %s, want 2026-09-23 (today)", got)
	}
	if !from.Equal(start) || !to.Equal(end.AddDate(0, 0, 1)) {
		t.Fatalf("bounds mismatch: from=%v to=%v", from, to)
	}

	// Sunday must roll back to the Monday six days earlier.
	sun := time.Date(2026, 9, 27, 23, 30, 0, 0, beijingZone)
	if _, _, s, _ := weeklyBounds(sun); s.Format("2006-01-02") != "2026-09-21" {
		t.Fatalf("sunday start = %s, want 2026-09-21", s.Format("2006-01-02"))
	}

	// 02:00 Beijing on Monday is still Sunday in UTC; the week must already
	// be computed from Beijing time.
	early := time.Date(2026, 9, 20, 18, 0, 0, 0, time.UTC) // Mon 02:00 +08:00
	if _, _, s, e := weeklyBounds(early); s.Format("2006-01-02") != "2026-09-21" || e.Format("2006-01-02") != "2026-09-21" {
		t.Fatalf("early monday bounds = %s..%s", s.Format("2006-01-02"), e.Format("2006-01-02"))
	}
}

func TestTopTag(t *testing.T) {
	records := func(rows ...struct {
		date time.Time
		tags string
	}) []model.Mood {
		out := make([]model.Mood, 0, len(rows))
		for _, r := range rows {
			out = append(out, model.Mood{RecordDate: r.date, MoodTags: r.tags})
		}
		return out
	}
	if got := topTag(nil); got != "" {
		t.Fatalf("empty records top tag = %q, want empty", got)
	}
	// Newest record first: happy and calm both appear twice; calm belongs to
	// the most recent record and must win.
	rs := records(
		struct {
			date time.Time
			tags string
		}{time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC), `["calm"]`},
		struct {
			date time.Time
			tags string
		}{time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC), `["happy","calm"]`},
		struct {
			date time.Time
			tags string
		}{time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC), `["happy"]`},
	)
	if got := topTag(rs); got != "calm" {
		t.Fatalf("tie top tag = %q, want calm", got)
	}
}

func TestRoundOneDecimal(t *testing.T) {
	cases := map[float64]float64{3.0: 3, 3.24: 3.2, 3.25: 3.3, 2.666: 2.7}
	for in, want := range cases {
		if got := roundOneDecimal(in); got != want {
			t.Fatalf("roundOneDecimal(%v) = %v, want %v", in, got, want)
		}
	}
}
