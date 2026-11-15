package report

import (
	"testing"
	"time"

	"tock/internal/classify"
)

func mustZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Skipf("no zone %s: %v", name, err)
	}
	return loc
}

func TestWeekOf(t *testing.T) {
	paris := mustZone(t, "Europe/Paris")
	ny := mustZone(t, "America/New_York")
	day := func(loc *time.Location, y int, m time.Month, d, h int) time.Time {
		return time.Date(y, m, d, h, 0, 0, 0, loc)
	}
	cases := []struct {
		name  string
		at    time.Time
		first time.Weekday
		want  time.Time
	}{
		// Sat 10 Oct 2026, fixed zone.
		{"sat monday", day(zone, 2026, 10, 10, 15), time.Monday, day(zone, 2026, 10, 5, 0)},
		{"sat sunday", day(zone, 2026, 10, 10, 15), time.Sunday, day(zone, 2026, 10, 4, 0)},
		{"sat saturday", day(zone, 2026, 10, 10, 15), time.Saturday, day(zone, 2026, 10, 10, 0)},
		// Each start day maps to itself at midnight and at 23:59.
		{"mon 00:00 monday", day(zone, 2026, 10, 5, 0), time.Monday, day(zone, 2026, 10, 5, 0)},
		{"sun 00:00 sunday", day(zone, 2026, 10, 11, 0), time.Sunday, day(zone, 2026, 10, 11, 0)},
		{"sun 23:00 monday", day(zone, 2026, 10, 11, 23), time.Monday, day(zone, 2026, 10, 5, 0)},
		{"fri 23:00 saturday", day(zone, 2026, 10, 9, 23), time.Saturday, day(zone, 2026, 10, 3, 0)},
		// Paris springs forward on Sun 29 Mar 2026 and falls back on Sun 25 Oct.
		{"paris spring sunday", day(paris, 2026, 3, 29, 12), time.Sunday, day(paris, 2026, 3, 29, 0)},
		{"paris spring monday", day(paris, 2026, 3, 29, 12), time.Monday, day(paris, 2026, 3, 23, 0)},
		{"paris after spring saturday", day(paris, 2026, 4, 3, 22), time.Saturday, day(paris, 2026, 3, 28, 0)},
		{"paris fall sunday", day(paris, 2026, 10, 31, 23), time.Sunday, day(paris, 2026, 10, 25, 0)},
		{"paris fall saturday", day(paris, 2026, 10, 25, 2), time.Saturday, day(paris, 2026, 10, 24, 0)},
		// New York springs forward on Sun 8 Mar and falls back on Sun 1 Nov 2026.
		{"ny spring sunday", day(ny, 2026, 3, 14, 20), time.Sunday, day(ny, 2026, 3, 8, 0)},
		{"ny fall saturday", day(ny, 2026, 11, 6, 9), time.Saturday, day(ny, 2026, 10, 31, 0)},
		{"ny fall monday", day(ny, 2026, 11, 1, 23), time.Monday, day(ny, 2026, 10, 26, 0)},
	}
	for _, c := range cases {
		got := WeekOf(c.at, c.first)
		if !got.Equal(c.want) || got.Location() != c.at.Location() {
			t.Errorf("%s: WeekOf(%v, %v) = %v, want %v", c.name, c.at, c.first, got, c.want)
		}
		if got.Weekday() != c.first || got.Hour() != 0 {
			t.Errorf("%s: %v is not a %v midnight", c.name, got, c.first)
		}
		if c.first == time.Monday && !got.Equal(MondayOf(c.at)) {
			t.Errorf("%s: MondayOf disagrees", c.name)
		}
	}
}

// A Sunday-start week holding a DST change keeps seven local-midnight days
// and puts each span on its local day.
func TestBuildWeekSundayDST(t *testing.T) {
	paris := mustZone(t, "Europe/Paris")
	sun := WeekOf(time.Date(2026, 3, 31, 12, 0, 0, 0, paris), time.Sunday) // 29 Mar, 23 hours long
	at := func(d, h int) int64 { return time.Date(2026, 3, 29+d, h, 0, 0, 0, paris).UnixMilli() }
	vs := []classify.SpanView{
		{ID: 1, Start: at(0, 9), End: at(0, 11), BundleID: "term", App: "WezTerm"},  // Sunday, after the change
		{ID: 2, Start: at(6, 22), End: at(6, 23), BundleID: "term", App: "WezTerm"}, // Saturday
	}
	ls := []classify.Label{{Category: "Code", Kind: "deep", Project: "tock"}, {Category: "Code", Kind: "deep", Project: "tock"}}
	w := BuildWeek(WeekInput{Monday: sun, NowMS: at(7, 12), Views: vs, Labels: ls})
	for i, d := range w.Days {
		want := time.Date(2026, 3, 29+i, 0, 0, 0, 0, paris)
		if !d.Date.Equal(want) {
			t.Fatalf("day %d = %v, want %v", i, d.Date, want)
		}
	}
	if w.Days[0].TrackedMS != 2*hr || w.Days[6].TrackedMS != hr || w.TrackedMS != 3*hr {
		t.Fatalf("tracked by day %d %d, week %d", w.Days[0].TrackedMS, w.Days[6].TrackedMS, w.TrackedMS)
	}
}

// A week goal's streak follows the week start: a Sunday with the goal's
// time belongs to this week when weeks start on Sunday, to last week
// otherwise.
func TestStreakFromSunday(t *testing.T) {
	g := classify.Goal{Category: "Study", MS: 5 * hr, Week: true}
	past := []GoalDay{{Date: day0.AddDate(0, 0, -6), Active: true, MS: []int64{5 * hr}}} // Sun 4 Oct; day0 is Sat 10 Oct
	if cur, _ := StreakFrom(0, g, past, today(0), false, time.Sunday); cur != 1 {
		t.Errorf("sunday start: streak %d, want 1 (this week met)", cur)
	}
	if cur, _ := StreakFrom(0, g, past, today(0), false, time.Monday); cur != 0 {
		t.Errorf("monday start: streak %d, want 0 (this week open)", cur)
	}
	if a, b := Streak(0, g, past, today(0), false); a != 0 || b != 0 {
		t.Errorf("Streak = %d %d, want Monday weeks", a, b)
	}
}
