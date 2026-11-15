package web

import "testing"

// The week, top and agents ranges follow week_start; monday stays the
// week's Monday and start names its first day.
func TestWeekStartAPI(t *testing.T) {
	f := newFixture(t) // the clock is Sat 10 Oct 2026
	var st statusJSON
	var w weekJSON
	f.get("/api/status", &st)
	f.get("/api/week", &w)
	if st.WeekStart != "monday" || w.Monday != "2026-10-05" || w.Start != "2026-10-05" {
		t.Fatalf("default: status %q, week monday %s start %s", st.WeekStart, w.Monday, w.Start)
	}
	for _, c := range []struct{ name, start, monday string }{
		{"sunday", "2026-10-04", "2026-10-05"},
		{"saturday", "2026-10-10", "2026-10-12"},
	} {
		f.e.WriteConfig(t, "[display]\nweek_start = \""+c.name+"\"\n")
		st, w = statusJSON{}, weekJSON{}
		f.get("/api/status", &st)
		f.get("/api/week", &w)
		if st.WeekStart != c.name || w.Start != c.start || w.Monday != c.monday || w.Days[0].Date != c.start {
			t.Fatalf("%s: status %q, week start %s monday %s first day %s", c.name, st.WeekStart, w.Start, w.Monday, w.Days[0].Date)
		}
		var top topJSON
		f.get("/api/top?range=week", &top)
		var ag agentsJSON
		f.get("/api/agents?range=week", &ag)
		if top.From != c.start || ag.From != c.start {
			t.Fatalf("%s: top from %s, agents from %s", c.name, top.From, ag.From)
		}
	}
}
