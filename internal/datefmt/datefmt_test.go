package datefmt

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	for in, want := range map[string]Style{"dmy": DMY, "mdy": MDY, "iso": ISO, " MDY ": MDY, "Iso": ISO} {
		got, ok := Parse(in)
		if !ok || got != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "ymd", "us", "dd/mm/yyyy", "system"} {
		if _, ok := Parse(bad); ok {
			t.Errorf("Parse(%q) accepted", bad)
		}
	}
	for _, s := range []Style{DMY, MDY, ISO} {
		if back, ok := Parse(s.String()); !ok || back != s {
			t.Errorf("%v does not round-trip through String", s)
		}
	}
	var zero Style
	if zero != DMY {
		t.Fatal("the zero Style must be DMY, today's output")
	}
}

func TestFormats(t *testing.T) {
	cases := []struct {
		t                  time.Time
		style              Style
		day, dayYear, date string
		dayMonth           string
	}{
		{time.Date(2026, 11, 3, 17, 20, 0, 0, time.UTC), DMY, "Tue 3 Nov", "Tue 3 Nov 2026", "3 Nov 2026", "3 Nov"},
		{time.Date(2026, 11, 3, 17, 20, 0, 0, time.UTC), MDY, "Tue Nov 3", "Tue Nov 3, 2026", "Nov 3, 2026", "Nov 3"},
		{time.Date(2026, 11, 3, 17, 20, 0, 0, time.UTC), ISO, "Tue 2026-11-03", "Tue 2026-11-03", "2026-11-03", "11-03"},
		{time.Date(2027, 1, 24, 0, 0, 0, 0, time.UTC), DMY, "Sun 24 Jan", "Sun 24 Jan 2027", "24 Jan 2027", "24 Jan"},
		{time.Date(2027, 1, 24, 0, 0, 0, 0, time.UTC), MDY, "Sun Jan 24", "Sun Jan 24, 2027", "Jan 24, 2027", "Jan 24"},
		{time.Date(2027, 1, 24, 0, 0, 0, 0, time.UTC), ISO, "Sun 2027-01-24", "Sun 2027-01-24", "2027-01-24", "01-24"},
	}
	for _, c := range cases {
		if got := c.style.Day(c.t); got != c.day {
			t.Errorf("%v.Day = %q, want %q", c.style, got, c.day)
		}
		if got := c.style.DayYear(c.t); got != c.dayYear {
			t.Errorf("%v.DayYear = %q, want %q", c.style, got, c.dayYear)
		}
		if got := c.style.Date(c.t); got != c.date {
			t.Errorf("%v.Date = %q, want %q", c.style, got, c.date)
		}
		if got := c.style.DayMonth(c.t); got != c.dayMonth {
			t.Errorf("%v.DayMonth = %q, want %q", c.style, got, c.dayMonth)
		}
	}
}

// TestDMYIsToday pins the default to the layouts tock printed before the
// setting existed, so existing output does not move.
func TestDMYIsToday(t *testing.T) {
	for d := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC); d.Year() == 2026; d = d.AddDate(0, 0, 1) {
		if DMY.Day(d) != d.Format("Mon 2 Jan") || DMY.DayYear(d) != d.Format("Mon 2 Jan 2006") || DMY.Date(d) != d.Format("2 Jan 2006") {
			t.Fatalf("DMY moved on %s", d.Format(time.DateOnly))
		}
		if ISO.Date(d) != d.Format("2006-01-02") || ISO.DayYear(d) != d.Format("Mon 2006-01-02") {
			t.Fatalf("ISO is not ISO on %s", d.Format(time.DateOnly))
		}
	}
}
