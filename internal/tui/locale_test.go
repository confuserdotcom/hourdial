package tui

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

// The score panel with [display] date and numbers set: the title and the
// header carry the date in that order, and the breakdown's decimals use
// that separator. TestGoldenScore is the default.
func scoreWith(t *testing.T, display string) string {
	t.Helper()
	m, e := seeded(t)
	e.WriteConfig(t, "[display]\n"+display)
	e.Heartbeat(t, e.Clock.UnixMilli())
	m = step(t, m, tickMsg{})
	m = press(t, m, "r", "i")
	return screen(m)
}

func TestGoldenScoreMDY(t *testing.T) {
	golden.RequireEqual(t, scoreWith(t, "date = \"mdy\"\nnumbers = \"1,234.5\"\n"))
}

func TestGoldenScoreISO(t *testing.T) {
	golden.RequireEqual(t, scoreWith(t, "date = \"iso\"\nnumbers = \"1.234,5\"\n"))
}

func TestGoldenScoreSpace(t *testing.T) {
	golden.RequireEqual(t, scoreWith(t, "date = \"dmy\"\nnumbers = \"1 234,5\"\n"))
}

// Setting the defaults by hand changes nothing.
func TestScoreDefaultsUnchanged(t *testing.T) {
	m, e := seeded(t)
	e.Heartbeat(t, e.Clock.UnixMilli())
	m = press(t, m, "r", "i")
	if got, want := scoreWith(t, "date = \"dmy\"\nnumbers = \"1234.5\"\n"), screen(m); got != want {
		t.Fatalf("explicit defaults moved the score panel:\n%s\nwant\n%s", got, want)
	}
}
