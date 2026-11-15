package config

import (
	"fmt"
	"strings"
	"time"
)

// Display is the [display] section: how dates are laid out.
type Display struct {
	// WeekStart is the first day of a week as config.toml names it:
	// "monday" (also the zero value), "sunday" or "saturday".
	WeekStart string
}

type displayTOML struct {
	WeekStart *string `toml:"week_start"`
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

func decodeDisplay(f displayTOML, d *Display) error {
	if f.WeekStart != nil {
		w := strings.ToLower(strings.TrimSpace(*f.WeekStart))
		if _, ok := weekStarts[w]; !ok {
			return fmt.Errorf(`[display] week_start = %q, want "monday", "sunday" or "saturday"`, *f.WeekStart)
		}
		d.WeekStart = w
	}
	return nil
}
