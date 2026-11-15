package config

import (
	"fmt"
	"strings"
	"time"

	"tock/internal/clockfmt"
	"tock/internal/datefmt"
	"tock/internal/numfmt"
)

// Display is the [display] section: how dates and times are laid out.
type Display struct {
	// WeekStart is the first day of a week as config.toml names it:
	// "monday" (also the zero value), "sunday" or "saturday".
	WeekStart string
	// Clock is how times of day are shown to people: "24h" (also the zero
	// value) or "12h".
	Clock string
	// Date is the order dates are written in: "dmy" (also the zero
	// value), "mdy" or "iso". "" means config.toml does not set it.
	Date string
	// Numbers is how decimals are written, as 1234.5 would be: "1234.5"
	// (also the zero value), "1,234.5", "1.234,5", "1 234,5" or "1234,5".
	// "" means config.toml does not set it.
	Numbers string
}

type displayTOML struct {
	WeekStart *string `toml:"week_start"`
	Clock     *string `toml:"clock"`
	Date      *string `toml:"date"`
	Numbers   *string `toml:"numbers"`
}

// weekStarts are the accepted week_start values.
var weekStarts = map[string]time.Weekday{"monday": time.Monday, "sunday": time.Sunday, "saturday": time.Saturday}

// FirstDay is the weekday weeks start on.
func (d Display) FirstDay() time.Weekday {
	if w, ok := weekStarts[d.WeekStart]; ok {
		return w
	}
	return time.Monday
}

// WeekStartName is FirstDay as written in config.toml: "monday", "sunday"
// or "saturday".
func (d Display) WeekStartName() string {
	return strings.ToLower(d.FirstDay().String())
}

// ClockFormat is the clock setting for clockfmt; 24-hour unless set to 12h.
func (d Display) ClockFormat() clockfmt.Clock {
	c, _ := clockfmt.Parse(d.Clock)
	return c
}

// DateStyle is the date setting for datefmt; day first unless set.
func (d Display) DateStyle() datefmt.Style {
	s, _ := datefmt.Parse(d.Date)
	return s
}

// NumberStyle is the numbers setting for numfmt; "1234.5" unless set.
func (d Display) NumberStyle() numfmt.Style {
	s, _ := numfmt.Parse(d.Numbers)
	return s
}

// numbersWant is the accepted numbers values, for error messages.
func numbersWant() string {
	n := numfmt.Names()
	return `"` + strings.Join(n[:len(n)-1], `", "`) + `" or "` + n[len(n)-1] + `"`
}

func decodeDisplay(f displayTOML, d *Display) error {
	if f.WeekStart != nil {
		w := strings.ToLower(strings.TrimSpace(*f.WeekStart))
		if _, ok := weekStarts[w]; !ok {
			return fmt.Errorf(`[display] week_start = %q, want "monday", "sunday" or "saturday"`, *f.WeekStart)
		}
		d.WeekStart = w
	}
	if f.Clock != nil {
		c, ok := clockfmt.Parse(*f.Clock)
		if !ok {
			return fmt.Errorf(`[display] clock = %q, want "24h" or "12h"`, *f.Clock)
		}
		d.Clock = c.String()
	}
	if f.Date != nil {
		s, ok := datefmt.Parse(*f.Date)
		if !ok {
			return fmt.Errorf(`[display] date = %q, want "dmy" (3 Nov), "mdy" (Nov 3) or "iso" (2026-11-03)`, *f.Date)
		}
		d.Date = s.String()
	}
	if f.Numbers != nil {
		s, ok := numfmt.Parse(*f.Numbers)
		if !ok {
			return fmt.Errorf(`[display] numbers = %q, want %s`, *f.Numbers, numbersWant())
		}
		d.Numbers = s.String()
	}
	return nil
}
