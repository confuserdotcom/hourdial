package bar

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/golden"

	"tock/internal/app"
	"tock/internal/classify"
	"tock/internal/clockfmt"
	"tock/internal/config"
	"tock/internal/datefmt"
	"tock/internal/focus"
	"tock/internal/recommend"
	"tock/internal/report"
	"tock/internal/stats"
	"tock/internal/store"
)

// The menu bar title, the menu and the notch as the Cocoa side gets them,
// with what VoiceOver says for each,
// for a fixed set of days at a fixed now (Thursday 5 Nov 2026, 17:20 UTC).
// The notch is the JSON tockbar hands to notch_darwin.m, with the quiet or
// running panel's pages filled in as the runner does.
// Regenerate with: go test ./internal/bar -run Golden -update

// goldenNow is t1720: Thursday 5 Nov 2026, 17:20 UTC.
const goldenNow = t1720

func goldenDay() time.Time { return time.Date(2026, 11, 5, 0, 0, 0, 0, islandLoc) }

func goldenAt(hh, mm int) int64 {
	return goldenDay().Add(time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute).UnixMilli()
}

// menuText renders a menu: "--" separators, "[strip k,k,…]" strip rows,
// "Title ▸" submenu parents with their rows indented, "✓ " checked rows,
// " [id]" after clickable ones and « what VoiceOver says » when that is not
// the title.
func menuText(items []Item) string {
	var b []string
	var walk func([]Item, string)
	walk = func(items []Item, ind string) {
		for _, it := range items {
			n := len(b)
			switch {
			case it.Separator:
				b = append(b, ind+"--")
			case len(it.Strip) > 0:
				ks := make([]string, len(it.Strip))
				for i, k := range it.Strip {
					ks[i] = k
					if k == "" {
						ks[i] = "none"
					}
				}
				b = append(b, ind+"[strip "+strings.Join(ks, ",")+"]")
			case len(it.Sub) > 0:
				b = append(b, ind+it.Title+" ▸")
				walk(it.Sub, ind+"  ")
			case it.ID == "":
				b = append(b, ind+it.Title)
			default:
				c := ""
				if it.Checked {
					c = "✓ "
				}
				b = append(b, ind+c+it.Title+" ["+it.ID+"]")
			}
			if w := it.Spoken(); w != "" && len(b) > n {
				b[n] += " «" + w + "»"
			}
		}
	}
	walk(items, "")
	return strings.Join(b, "\n")
}

// pageData is what the runner reads for the notch's pages.
type pageData struct {
	d report.Day
	w report.Week
	u app.Usual
}

// islandOf is the notch at nowMS as renderNotch builds it with no moment.
func islandOf(s Snapshot, pd *pageData, nowMS int64) Island {
	i := s.IslandBase(nowMS)
	if pd != nil && i.Hover != nil && (i.State == IslandQuiet || i.State == IslandFocus) {
		h := *i.Hover
		h.Pages = TodayPages(s, pd.d, pd.w, pd.u, nowMS)
		i.Hover = &h
	}
	return i
}

func islandText(t *testing.T, i Island) string {
	t.Helper()
	b, err := json.MarshalIndent(i.withSpoken(), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// screens renders the title, the menu and the notch of s at nowMS.
func screens(t *testing.T, s Snapshot, pd *pageData, nowMS int64) string {
	t.Helper()
	return fmt.Sprintf("== title\n%s «%s»\n\n== menu\n%s\n\n== notch\n%s\n",
		s.Title(nowMS), s.Spoken(nowMS), menuText(s.Items(nowMS)), islandText(t, islandOf(s, pd, nowMS)))
}

// emptyWeek is the week of goldenDay with nothing in it.
func emptyWeek() report.Week {
	w := report.Week{Monday: goldenDay().AddDate(0, 0, -3), ByKind: map[string]int64{}}
	for i := range w.Days {
		w.Days[i].Date = w.Monday.AddDate(0, 0, i)
		w.Days[i].Future = i > 3
	}
	return w
}

// fullDay is a worked Thursday: 5h tracked, three projects, two goals.
func fullDay(clock clockfmt.Clock) pageData {
	day := goldenDay()
	var deep, shallow [24]int64
	deep[9], deep[10], shallow[14], shallow[15] = 50*60_000, 40*60_000, 30*60_000, 20*60_000
	d := report.Day{Date: day, Clock: clock, TrackedMS: 5 * 3600_000,
		Score: stats.Score{Value: 81},
		Stats: stats.Day{FocusMS: 192 * 60_000, LongestMS: 70 * 60_000, DistractMS: 25 * 60_000, ActiveMS: 5 * 3600_000},
		Hours: report.Heatmap{First: 9, Last: 15, MS: map[string][24]int64{"deep": deep, "shallow": shallow}},
		Projects: []report.Share{{Name: "tock", MS: 3 * 3600_000}, {Name: "", MS: 3600_000},
			{Name: "Licence", MS: 1800_000}, {Name: "x", MS: 1}},
		Timeline: []report.Row{
			{Start: goldenAt(9, 0), End: goldenAt(11, 0), Label: classify.Label{Kind: "deep"}},
			{Start: goldenAt(11, 0), End: goldenAt(12, 0), Idle: true},
			{Start: goldenAt(14, 0), End: goldenAt(15, 30), Label: classify.Label{Kind: "shallow"}},
			{Start: goldenAt(15, 30), End: goldenAt(16, 0), Label: classify.Label{Kind: "distract"}},
			{Start: goldenAt(16, 0), End: goldenAt(17, 0)},
		},
		Goals: []report.GoalRow{
			{Goal: classify.Goal{Category: "Study", MS: 2 * 3600_000}, DoneMS: 3600_000, State: "short", Streak: 3},
			{Goal: classify.Goal{Category: "Fun", AtMost: true, MS: 45 * 60_000}, DoneMS: 50 * 60_000, State: "over"},
		}}
	w := emptyWeek()
	w.Clock, w.TrackedMS = clock, 20*3600_000
	w.ByKind = map[string]int64{"deep": 12 * 3600_000, "shallow": 2 * 3600_000, "distract": 3600_000, "": 5 * 3600_000}
	for i := 0; i <= 3; i++ {
		w.Days[i].TrackedMS = 5 * 3600_000
		w.Days[i].ByKind = map[string]int64{"deep": 3 * 3600_000, "shallow": 30 * 60_000, "distract": 15 * 60_000, "": 75 * 60_000}
		w.Days[i].Score = stats.Score{Value: 70 + 3*i}
	}
	return pageData{d: d, w: w, u: app.Usual{Days: 3, Score: 75, FocusMS: 168 * 60_000, LongestMS: 55 * 60_000, DistractMS: 30 * 60_000}}
}

// fullSnap is the menu bar's snapshot of fullDay: an unrated session this
// afternoon, a plan item in 40 minutes, a pending suggestion, hidden apps
// and sites, and every Settings toggle.
func fullSnap(clock clockfmt.Clock) Snapshot {
	url := "news.ycombinator.com/item?id=1"
	last := report.SessionDetail{Clock: clock,
		Sess: focus.Session{ID: 41, StartMS: goldenAt(15, 0), PlannedMS: goldenAt(15, 50), EndMS: goldenAt(15, 50),
			Goal: "TD 3, exercises 4 to 7", Interruptions: 1, LongestMS: 32 * 60_000, Switches: 4},
		Q:   focus.QualityScore{Score: 78},
		Top: focus.Breakdown{Project: "Licence"}}
	return Snapshot{OK: true, TrackdUp: true, Access: app.NotifyGranted,
		TrackedMS: 5 * 3600_000, FocusMS: 192 * 60_000, Score: 81, Active: true, UnlabelledMS: 48 * 60_000,
		Strip:   []string{"deep", "deep", "deep", "deep", "", "", "shallow", "shallow", "shallow", "distract", "", ""},
		Presets: []int{25, 50, 90}, SnoozeMin: 15,
		Last: &LastSession{Detail: last, Refile: true, Loc: islandLoc},
		Slots: []focus.Slot{{Key: "plan:18:00", Kind: focus.SlotPlan, StartMS: goldenAt(18, 0), EndMS: goldenAt(19, 0),
			Title: "Read chapter 4"}},
		Loc: islandLoc, Clock: clock,
		Rec:      &recommend.Rec{ID: 3, Kind: recommend.KindFocus, AtMS: goldenNow - 5*60_000, LengthMS: 50 * 60_000, Weekday: 4, Hour: 17},
		Features: featureStates(config.Default()),
		Privacy: PrivacyState{Front: &store.Active{Bundle: "com.google.Chrome", App: "Google Chrome", URL: &url},
			Apps:  []HiddenApp{{Bundle: "com.apple.MobileSMS", Name: "Messages"}, {Bundle: "org.whispersystems.signal-desktop", Name: "Signal"}},
			Sites: []string{"mail.google.com"}}}
}

// A first launch: tock has no rules yet.
func TestGoldenBarNotSetUp(t *testing.T) {
	golden.RequireEqual(t, screens(t, NotReady(classify.ErrNoRules), nil, goldenNow))
}

// trackd has not made the database yet.
func TestGoldenBarWaitingForTrackd(t *testing.T) {
	golden.RequireEqual(t, screens(t, NotReady(app.ErrNoDB), nil, goldenNow))
}

// Set up, trackd running, nothing tracked today and no history.
func TestGoldenBarEmptyDay(t *testing.T) {
	s := Snapshot{OK: true, TrackdUp: true, Access: app.NotifyGranted, Presets: []int{25, 50, 90}, Loc: islandLoc}
	pd := &pageData{d: report.Day{Date: goldenDay()}, w: emptyWeek()}
	golden.RequireEqual(t, screens(t, s, pd, goldenNow))
}

// The first day with data: no usual day to compare with, one day in the
// week, a little unlabelled time.
func TestGoldenBarFirstDay(t *testing.T) {
	s := Snapshot{OK: true, TrackdUp: true, Access: app.NotifyGranted, TrackedMS: 80 * 60_000, FocusMS: 35 * 60_000,
		Score: 52, Active: true, UnlabelledMS: 12 * 60_000, Strip: []string{"", "deep", "deep", "shallow"},
		Presets: []int{25}, Loc: islandLoc}
	var deep [24]int64
	deep[16] = 35 * 60_000
	d := report.Day{Date: goldenDay(), TrackedMS: 80 * 60_000, Score: stats.Score{Value: 52},
		Stats:    stats.Day{FocusMS: 35 * 60_000, ActiveMS: 80 * 60_000},
		Hours:    report.Heatmap{First: 16, Last: 17, MS: map[string][24]int64{"deep": deep}},
		Projects: []report.Share{{Name: "tock", MS: 50 * 60_000}},
		Timeline: []report.Row{{Start: goldenAt(16, 0), End: goldenAt(17, 20), Label: classify.Label{Kind: "deep"}}}}
	w := emptyWeek()
	w.TrackedMS = 80 * 60_000
	w.ByKind = map[string]int64{"deep": 35 * 60_000, "": 45 * 60_000}
	w.Days[3].TrackedMS, w.Days[3].ByKind, w.Days[3].Score = 80*60_000, w.ByKind, stats.Score{Value: 52}
	golden.RequireEqual(t, screens(t, s, &pageData{d: d, w: w}, goldenNow))
}

// A worked day with everything the menu can offer, 24-hour clock.
func TestGoldenBarFullDay(t *testing.T) {
	pd := fullDay(clockfmt.H24)
	golden.RequireEqual(t, screens(t, fullSnap(clockfmt.H24), &pd, goldenNow))
}

// The same day with [display] clock = "12h".
func TestGoldenBarFullDay12h(t *testing.T) {
	pd := fullDay(clockfmt.H12)
	golden.RequireEqual(t, screens(t, fullSnap(clockfmt.H12), &pd, goldenNow))
}

// A session runs; a distraction notice came a minute ago.
func TestGoldenBarSession(t *testing.T) {
	pd := fullDay(clockfmt.H24)
	s := fullSnap(clockfmt.H24)
	s.Session = &focus.Session{ID: 42, StartMS: goldenNow - 6*60_000, PlannedMS: goldenNow + 19*60_000, Goal: "tock notch", Interruptions: 2}
	s.Live = focus.Live{Level: 64}
	s.DistractAtMS = goldenNow - 60_000
	golden.RequireEqual(t, screens(t, s, &pd, goldenNow))
}

// Names longer than the menu and the notch have room for.
func TestGoldenBarLongNames(t *testing.T) {
	pd := fullDay(clockfmt.H24)
	pd.d.Projects[0].Name = "Internationalisation of the accounting backend"
	pd.d.Goals[0].Goal.Category = "Reading for the comparative literature seminar"
	s := fullSnap(clockfmt.H24)
	s.Session = &focus.Session{ID: 42, StartMS: goldenNow - 6*60_000, PlannedMS: goldenNow + 19*60_000,
		Goal: "Write the introduction and the first two sections of the report"}
	s.Live = focus.Live{Level: -1}
	s.Slots[0].Title = "Weekly planning with the whole extended product and design team"
	s.Privacy.Apps = append(s.Privacy.Apps, HiddenApp{Bundle: "com.example.verylongname", Name: "A Very Long Application Name That Keeps Going"})
	s.Last.Detail.Sess.Note = strings.Repeat("Went off on a tangent about the parser again. ", 3)
	s.Last.Detail.Top.Project = "Internationalisation of the accounting backend"
	golden.RequireEqual(t, screens(t, s, &pd, goldenNow))
}

// trackd is down, Accessibility is not granted, notifications are off.
func TestGoldenBarTrackdDown(t *testing.T) {
	s := Snapshot{OK: true, TrackedMS: 2*3600_000 + 54*60_000, AXDenied: true, Access: app.NotifyDenied,
		Presets: []int{25, 50}, Loc: islandLoc}
	golden.RequireEqual(t, screens(t, s, nil, goldenNow))
}

// The Focus group is off: no sessions, no score.
func TestGoldenBarFocusOff(t *testing.T) {
	pd := fullDay(clockfmt.H24)
	s := fullSnap(clockfmt.H24)
	s.NoFocus, s.Presets, s.Rec = true, nil, nil
	golden.RequireEqual(t, screens(t, s, &pd, goldenNow))
}

// A break the user took runs.
func TestGoldenBarBreak(t *testing.T) {
	s := Snapshot{OK: true, TrackdUp: true, TrackedMS: 3 * 3600_000, Active: true, Score: 70, Presets: []int{25},
		BreakUntil: time.UnixMilli(goldenNow + 8*60_000), Loc: islandLoc}
	golden.RequireEqual(t, screens(t, s, nil, goldenNow))
}

// A meeting with a link starts in 4 minutes.
func TestGoldenBarMeeting(t *testing.T) {
	s := Snapshot{OK: true, TrackdUp: true, TrackedMS: 3 * 3600_000, Active: true, Score: 70, Presets: []int{25}, Loc: islandLoc,
		Slots: []focus.Slot{{Key: "event:17:24", Kind: focus.SlotEvent, StartMS: goldenNow + 4*60_000, EndMS: goldenNow + 34*60_000,
			Title: "Standup", Calendar: "Work", JoinURL: "https://meet.google.com/abc-defg-hij"}}}
	golden.RequireEqual(t, screens(t, s, nil, goldenNow))
}

// The moments the notch opens by itself, each as the JSON it is drawn from.
func TestGoldenNotchMoments(t *testing.T) {
	const h = int64(3600_000)
	pd := fullDay(clockfmt.H24)
	det := &report.SessionDetail{Sess: focus.Session{ID: 41, StartMS: goldenAt(15, 0), PlannedMS: goldenAt(15, 50),
		EndMS: goldenAt(15, 50), Goal: "TD 3, exercises 4 to 7", LongestMS: 32 * 60_000, Switches: 4, DistractMS: 3 * 60_000}}
	entries := []Entry{
		{Key: "a", App: "Google Chrome", Title: "Pull requests · tock", Sub: "github.com", Dur: "24m", MS: 24 * 60_000,
			Opts: []Option{{Label: "Code / tock", Kind: "deep", Pct: 86}, {Label: "Admin", Kind: "shallow"}}, Picked: -1},
		{Key: "b", App: "Preview", Title: "chapter4.pdf", Dur: "12m", MS: 12 * 60_000,
			Opts: []Option{{Label: "Study", Kind: "deep"}}, Picked: -1},
	}
	moments := []struct {
		name string
		i    Island
	}{
		{"budget 85%", BudgetIsland(store.Notice{ID: 7, Topic: focus.TopicBudget, Title: "tock: 85% of its 40h budget", Body: "34h spent, 6h left."},
			app.NoticeData{Project: "tock", BudgetMS: 40 * h, SpentMS: 34 * h, RunsOutMS: goldenAt(12, 0) + 15*24*h}, islandLoc, datefmt.DMY)},
		{"budget over", BudgetIsland(store.Notice{ID: 8, Topic: focus.TopicBudget, Title: "tock is over its 40h budget", Body: "44h spent."},
			app.NoticeData{Project: "tock", BudgetMS: 40 * h, SpentMS: 44 * h}, islandLoc, datefmt.DMY)},
		{"budget from an older trackd", BudgetIsland(store.Notice{ID: 9, Topic: focus.TopicBudget, Title: "tock: 80% of its budget", Body: "32h spent."},
			app.NoticeData{}, islandLoc, datefmt.DMY)},
		{"session done", DoneIsland(store.Notice{ID: 10, Topic: focus.TopicDone, Title: "Session done", Body: "50m · 78% deep"},
			app.NoticeData{Session: 41, Minutes: 50, Project: "Licence", Refile: true}, det, islandLoc, 48*60_000)},
		{"drift", DriftIsland(store.Notice{ID: 11, Topic: focus.TopicFocus, Title: "Back to tock notch?", Body: "YouTube for 2 min"},
			app.NoticeData{Session: 42, SnoozeMin: 15, Category: "Video"})},
		{"meeting", MeetIsland(store.Notice{ID: 12, Topic: focus.TopicStarting, Title: "Standup", Body: "17:24 to 17:54 · Work"},
			app.NoticeData{Join: "https://meet.google.com/abc-defg-hij"})},
		{"recap", RecapIsland(store.Notice{ID: 13, Topic: focus.TopicRecap, Title: "Thursday"}, pd.d, pd.u, 48*60_000, islandLoc)},
		{"labelling", LabelIsland(entries, 0, "")},
		{"inbox clear", LabelIsland(entries, len(entries), HapticAlignment)},
		{"missed done in the quiet notch", Missed(fullSnap(clockfmt.H24).IslandBase(goldenNow),
			DoneIsland(store.Notice{ID: 10, Topic: focus.TopicDone, Title: "Session done"}, app.NoticeData{}, nil, islandLoc, 0))},
	}
	var b strings.Builder
	for _, m := range moments {
		fmt.Fprintf(&b, "== %s\n%s\n\n", m.name, islandText(t, m.i))
	}
	golden.RequireEqual(t, b.String())
}
