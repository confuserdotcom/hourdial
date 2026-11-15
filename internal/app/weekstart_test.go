package app

import (
	"testing"
	"time"

	"tock/internal/apptest"
	"tock/internal/report"
)

// week_start moves the week the app reports, the plan key and the week
// of a week goal; the default stays on Monday.
func TestWeekStartSetting(t *testing.T) {
	// apptest.Now is Sat 10 Oct 2026. Seed Sun 4 Oct and Mon 5 Oct.
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, apptest.Zone) }
	cases := []struct {
		config string
		first  time.Time
		holds  []int // seeded days (offsets from Now) inside the week
	}{
		{"", day(5), []int{-5}},
		{"[display]\nweek_start = \"monday\"\n", day(5), []int{-5}},
		{"[display]\nweek_start = \"sunday\"\n", day(4), []int{-6, -5}},
		{"[display]\nweek_start = \"saturday\"\n", day(10), nil},
	}
	for _, c := range cases {
		e := apptest.New(t, apptest.Rules)
		if c.config != "" {
			e.WriteConfig(t, c.config)
		}
		e.Seed(t,
			apptest.Span{ID: 1, Start: "09:00", End: "10:00", Bundle: "term", App: "WezTerm", Day: -6},
			apptest.Span{ID: 2, Start: "09:00", End: "11:00", Bundle: "term", App: "WezTerm", Day: -5})
		a := open(t, e)
		if got := a.WeekOf(e.Clock); !got.Equal(c.first) {
			t.Fatalf("%q: WeekOf = %v, want %v", c.config, got, c.first)
		}
		w, err := a.Week(e.Clock)
		if err != nil {
			t.Fatal(err)
		}
		if !w.Monday.Equal(c.first) || !w.Days[0].Date.Equal(c.first) || !w.Days[6].Date.Equal(c.first.AddDate(0, 0, 6)) {
			t.Fatalf("%q: week %v, days %v..%v", c.config, w.Monday, w.Days[0].Date, w.Days[6].Date)
		}
		var want int64
		for _, d := range c.holds {
			want += map[int]int64{-6: 3_600_000, -5: 7_200_000}[d]
		}
		if w.TrackedMS != want {
			t.Fatalf("%q: tracked %d, want %d", c.config, w.TrackedMS, want)
		}
		if err := a.SetPlan(e.Clock, "ship it"); err != nil {
			t.Fatal(err)
		}
		if w, _ = a.Week(c.first.AddDate(0, 0, 3)); w.Plan != "ship it" {
			t.Fatalf("%q: plan %q not found by the week's first day", c.config, w.Plan)
		}
		if got := report.WeekOf(c.first.AddDate(0, 0, 6), a.Config().Display.FirstDay()); !got.Equal(c.first) {
			t.Fatalf("%q: last day of week maps to %v", c.config, got)
		}
	}
}
