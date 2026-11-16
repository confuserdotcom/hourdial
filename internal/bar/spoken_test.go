package bar

import (
	"testing"

	"tock/internal/app"
	"tock/internal/classify"
	"tock/internal/focus"
)

func TestSay(t *testing.T) {
	for in, want := range map[string]string{
		"3h12":                              "3 hours 12 minutes",
		"1h00 today":                        "1 hour today",
		"40h budget":                        "40 hours budget",
		"45m":                               "45 minutes",
		"1m":                                "1 minute",
		"30s":                               "30 seconds",
		"Deep focus 3h12, +24m":             "Deep focus 3 hours 12 minutes, plus 24 minutes",
		"Distracting 14m, −9m":              "Distracting 14 minutes, minus 9 minutes",
		"Score 81, +6 vs usual Thu":         "Score 81, plus 6 vs usual Thu",
		"Score 70, −5":                      "Score 70, minus 5",
		"Rate ★4":                           "Rate 4 stars",
		"Last session · 50m · Q78 · ★1":     "Last session, 50 minutes, quality 78, 1 star",
		"15:00–15:50 · 78% deep":            "15:00 to 15:50, 78% deep",
		"focus 64% ↗":                       "focus 64%, rising",
		"Next: 18:00 Read (in 40 min)":      "Next: 18:00 Read (in 40 min)",
		"5:25pm":                            "5:25pm",
		"Start 25 min focus":                "Start 25 min focus",
		"Q78":                               "quality 78",
		"FAQ78":                             "FAQ78",
		"hourdial · 19:00 left · focus 64%": "hourdial, 19 minutes left, focus 64%",
		"1:05:00 left":                      "65 minutes left",
		"Break · until 17:28":               "Break, until 17:28",
		"x2m":                               "x2m",
		"Delta ±0":                          "Delta no change",
		"Refile → tock":                     "Refile to tock",
		"Budget 85%":                        "Budget 85%",
		"Inbox · 48m":                       "Inbox, 48 minutes",
		"Weekly planning with th… (in 4m)":  "Weekly planning with th… (in 4 minutes)",
	} {
		if got := Say(in); got != want {
			t.Errorf("Say(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSnapshotSpoken(t *testing.T) {
	now := t1720
	for _, c := range []struct {
		s    Snapshot
		want string
	}{
		{NotReady(classify.ErrNoRules), "hourdial, not set up yet"},
		{NotReady(app.ErrNoDB), "hourdial, waiting for trackd"},
		{NotReady(nil), "hourdial, could not open"},
		{Snapshot{OK: true, TrackdUp: true, TrackedMS: 174 * 60_000}, "hourdial, 2 hours 54 minutes tracked today"},
		{Snapshot{OK: true, TrackedMS: 60 * 60_000}, "hourdial, 1 hour tracked today, trackd is not running"},
		{Snapshot{OK: true, TrackdUp: true, Paused: true}, "hourdial, 0 seconds tracked today, not recording"},
		{Snapshot{OK: true, Session: &focus.Session{PlannedMS: now + 19*60_000 - 500}}, "hourdial, focus session, 19 minutes left"},
		{Snapshot{OK: true, Session: &focus.Session{Goal: "TD 3", PlannedMS: now + 30_000}}, "hourdial, TD 3, under a minute left"},
		{Snapshot{OK: true, Session: &focus.Session{PlannedMS: now - 1}}, "hourdial, focus session, time is up"},
	} {
		if got := c.s.Spoken(now); got != c.want {
			t.Errorf("Spoken = %q, want %q", got, c.want)
		}
	}
}

// The colour strip, which VoiceOver would read as a row of block glyphs,
// says how the day splits instead.
func TestStripItemSpoken(t *testing.T) {
	it := Item{Strip: []string{"deep", "deep", "", "shallow", "distract", "", "deep", "deep"}}
	if got := it.Spoken(); got != "Today: deep 50%, shallow 13%, distracting 13%, no label 25%" {
		t.Fatalf("strip = %q", got)
	}
	if got := (Item{Title: "Open dashboard"}).Spoken(); got != "" {
		t.Fatalf("plain title spoken = %q", got)
	}
}

// The start buttons say what they do; the quiet face names its score.
func TestIslandSpoken(t *testing.T) {
	s := Snapshot{OK: true, TrackedMS: 4*3600_000 + 12*60_000, Score: 78, Active: true, UnlabelledMS: 85 * 60_000, Presets: []int{25, 50}, Loc: islandLoc}
	h := s.IslandBase(t1720).withSpoken().Hover
	if h.Face.Spoken != "4 hours 12 minutes today, score 78, 1 hour 25 minutes to label" {
		t.Fatalf("face = %q", h.Face.Spoken)
	}
	if h.Buttons[0].Spoken != "Start a 25 minute focus session" || h.Buttons[1].Spoken != "Start a 50 minute focus session" ||
		h.Buttons[3].Spoken != "Label 1 hour 25 minutes" {
		t.Fatalf("buttons = %+v", h.Buttons)
	}
	if s.IslandBase(t1720).Hover.Buttons[0].Spoken != "" {
		t.Fatal("withSpoken changed the island it was called on")
	}
}
