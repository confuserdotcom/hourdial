package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"tock/internal/classify"
	"tock/internal/report"
)

// importedRow is a timeline row of ActivityWatch history in Xcode.
func importedRow() report.Row {
	return report.Row{IDs: []int64{-(classify.RemoteIDBase + 1)}, Start: 9 * 3600_000, End: 10 * 3600_000,
		Key: classify.ImportMachine + " · Xcode · Build", App: classify.ImportMachine + " · Xcode", Text: "Build",
		Label: classify.Label{Category: "Code", Kind: classify.KindDeep}}
}

func mustReadXcode(t *testing.T, where, s string) {
	t.Helper()
	s = ansi.Strip(s)
	if !strings.Contains(s, "Xcode") || strings.Contains(s, classify.ImportMachine) {
		t.Fatalf("%s shows the import prefix or loses the app: %q", where, s)
	}
}

// The terminal shows imported apps by name, as the dashboard does; the
// filter finds them by either name.
func TestImportedAppReadsWithoutPrefix(t *testing.T) {
	r := importedRow()
	tt := todayTab{day: report.Day{Timeline: []report.Row{r}}, full: true}
	mustReadXcode(t, "timeline row", tt.timelineRow(0, r, 120, false, time.UTC))
	mustReadXcode(t, "ribbon tip", rowTip(r, time.UTC, tt.day.Clock))
	for _, q := range []string{"xcode", "activitywatch"} {
		tt.filter = q
		if len(tt.rows()) != 1 {
			t.Fatalf("filter %q misses the imported row", q)
		}
	}

	in := inboxTab{loaded: true, recorded: true, items: []classify.Item{{Key: r.Key, App: r.App, TotalMS: 3600_000, Spans: 1}}}
	mustReadXcode(t, "inbox", strings.Join(in.view(120, 10), "\n"))

	p := &picker{key: r.Key}
	mustReadXcode(t, "picker", strings.Join(p.view(80, 10), "\n"))
	d := &learnDlg{key: r.Key}
	mustReadXcode(t, "learn dialog", strings.Join(d.view(80), "\n"))
}
