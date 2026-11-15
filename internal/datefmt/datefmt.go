// Package datefmt writes calendar dates for people, as the [display] date
// setting says: day first "Tue 3 Nov" (the default), month first
// "Tue Nov 3", or ISO "Tue 2026-11-03". Day and month names stay English.
// It is a leaf package so pure packages can take a Style in their inputs.
//
// Exports, --json, logs, file names and the text sent to a language model
// keep their own fixed layouts and do not use it.
package datefmt

import (
	"strings"
	"time"
)

// Style is the order dates are written in. The zero value is DMY.
type Style uint8

const (
	DMY Style = iota // "Tue 3 Nov", "3 Nov 2026"
	MDY              // "Tue Nov 3", "Nov 3, 2026"
	ISO              // "Tue 2026-11-03", "2026-11-03"
)

// Parse reads a config value: "dmy", "mdy" or "iso", any case.
func Parse(s string) (Style, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dmy":
		return DMY, true
	case "mdy":
		return MDY, true
	case "iso":
		return ISO, true
	}
	return DMY, false
}

// String is the config value.
func (s Style) String() string {
	switch s {
	case MDY:
		return "mdy"
	case ISO:
		return "iso"
	}
	return "dmy"
}

// Day is a weekday and date without the year: "Tue 3 Nov", "Tue Nov 3".
// ISO keeps the year: "Tue 2026-11-03".
func (s Style) Day(t time.Time) string {
	if s == ISO {
		return s.DayYear(t)
	}
	return t.Format("Mon ") + s.DayMonth(t)
}

// DayYear is Day with the year: "Tue 3 Nov 2026", "Tue Nov 3, 2026",
// "Tue 2026-11-03".
func (s Style) DayYear(t time.Time) string {
	return t.Format("Mon ") + s.Date(t)
}

// Date is a date with the year and no weekday: "3 Nov 2026",
// "Nov 3, 2026", "2026-11-03".
func (s Style) Date(t time.Time) string {
	switch s {
	case MDY:
		return t.Format("Jan 2, 2006")
	case ISO:
		return t.Format("2006-01-02")
	}
	return t.Format("2 Jan 2006")
}

// DayMonth is the day and month only: "3 Nov", "Nov 3", "11-03" (ISO
// without the year; Day keeps it).
func (s Style) DayMonth(t time.Time) string {
	switch s {
	case MDY:
		return t.Format("Jan 2")
	case ISO:
		return t.Format("01-02")
	}
	return t.Format("2 Jan")
}
