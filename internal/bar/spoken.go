package bar

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"tock/internal/app"
	"tock/internal/classify"
	"tock/internal/report"
)

// What VoiceOver says. The menu and the notch print short forms (3h12,
// ★4, "·") that read badly aloud, and the notch draws strips, rings and
// bars that VoiceOver cannot see at all. These are the words the Cocoa side
// hands it instead, as accessibility labels and values.

// durRE matches the durations report.Dur and classify.FormatDur print, with
// an optional sign: 3h12, 40h, 45m, 30s, +24m, −1h05.
var durRE = regexp.MustCompile(`(^|[^\pL\pN:.])([+−]?)(\d+)(?:h(\d\d)?|(m)|(s))(\b|$)`)

var spokenSigns = strings.NewReplacer(
	" · ", ", ",
	"–", " to ",
	"→", "to",
	" ↗", ", rising",
	" ↘", ", falling",
	"±0", "no change",
)

var starsRE = regexp.MustCompile(`★(\d)`)

var signRE = regexp.MustCompile(`(^|\s)([+−])(\d)`)

// Say is s with tock's short forms written out for speech: "3h12" is
// "3 hours 12 minutes", "+6" is "plus 6", "★4" is "4 stars", "Q78" is
// "quality 78" and " · " a comma. Clock times are left alone: VoiceOver reads those well.
func Say(s string) string {
	s = durRE.ReplaceAllStringFunc(s, func(m string) string {
		g := durRE.FindStringSubmatch(m)
		n, _ := strconv.Atoi(g[3])
		var w string
		switch {
		case g[5] != "":
			w = plural(n, "minute", "minutes")
		case g[6] != "":
			w = plural(n, "second", "seconds")
		default:
			w = plural(n, "hour", "hours")
			if mm, _ := strconv.Atoi(g[4]); mm > 0 {
				w += " " + plural(mm, "minute", "minutes")
			}
		}
		switch g[2] {
		case "+":
			w = "plus " + w
		case "−":
			w = "minus " + w
		}
		return g[1] + w + g[7]
	})
	s = starsRE.ReplaceAllStringFunc(s, func(m string) string {
		n, _ := strconv.Atoi(strings.TrimPrefix(m, "★"))
		return plural(n, "star", "stars")
	})
	s = signRE.ReplaceAllStringFunc(s, func(m string) string {
		g := signRE.FindStringSubmatch(m)
		if g[2] == "+" {
			return g[1] + "plus " + g[3]
		}
		return g[1] + "minus " + g[3]
	})
	s = qualityRE.ReplaceAllString(s, "quality $1")
	s = leftRE.ReplaceAllStringFunc(s, func(m string) string {
		g := leftRE.FindStringSubmatch(m)
		a, _ := strconv.Atoi(g[1])
		b, _ := strconv.Atoi(g[2])
		secs := a*60 + b // M:SS
		if g[3] != "" {  // H:MM:SS
			c, _ := strconv.Atoi(g[3])
			secs = a*3600 + b*60 + c
		}
		return spokenLeft(int64(secs) * 1000)
	})
	return spokenSigns.Replace(s)
}

// leftRE is a countdown as report.Countdown prints it before "left":
// 19:00 left, 1:05:00 left. Read as a clock time it would mislead.
var leftRE = regexp.MustCompile(`\b(\d+):(\d\d)(?::(\d\d))? left\b`)

// qualityRE is a session's quality score as the menu prints it: Q78.
var qualityRE = regexp.MustCompile(`\bQ(\d+)\b`)

// spokenLeft is the time to a countdown's end, to the minute.
func spokenLeft(ms int64) string {
	switch {
	case ms <= 0:
		return "time is up"
	case ms < 60_000:
		return "under a minute left"
	}
	return plural(int((ms+59_999)/60_000), "minute", "minutes") + " left"
}

// Spoken is what VoiceOver says for the menu bar item at nowMS: tock, then
// the state the title shows as a glyph and a short figure.
func (s Snapshot) Spoken(nowMS int64) string {
	switch {
	case !s.OK && errors.Is(s.Why, classify.ErrNoRules):
		return "tock, not set up yet"
	case !s.OK && errors.Is(s.Why, app.ErrNoDB):
		return "tock, waiting for trackd"
	case !s.OK:
		return "tock, could not open"
	case s.Session != nil:
		goal := s.Session.Goal
		if goal == "" {
			goal = "focus session"
		}
		return "tock, " + goal + ", " + spokenLeft(s.Session.PlannedMS-nowMS)
	}
	out := "tock, " + Say(report.Dur(s.TrackedMS)) + " tracked today"
	switch {
	case s.TooOld || !s.TrackdUp:
		out += ", trackd is not running"
	case s.Paused:
		out += ", not recording"
	}
	return out
}

// Spoken is what VoiceOver says for a menu row, "" when its title reads
// well as it is. A strip row says how today splits by kind.
func (it Item) Spoken() string {
	if len(it.Strip) > 0 {
		return "Today: " + kindShares(it.Strip, "no label")
	}
	if w := Say(it.Title); w != it.Title {
		return w
	}
	return ""
}

// kindName is a kind as speech; none names the "" and "none" cells.
func kindName(k, none string) string {
	switch k {
	case "", "none":
		return none
	case "away":
		return "away"
	case DotBreak:
		return "spent"
	}
	if n, ok := kindNames[k]; ok {
		return strings.ToLower(n)
	}
	return k
}

// kindShares is a strip of kind cells as shares: "deep 40%, no label 60%".
func kindShares(cells []string, none string) string {
	n := map[string]int{}
	var order []string
	for _, k := range cells {
		if k == "" {
			k = "none"
		}
		if n[k] == 0 {
			order = append(order, k)
		}
		n[k]++
	}
	rank := func(k string) int {
		if i := slices.Index(kindOrder, k); i >= 0 && k != "" {
			return i
		}
		return len(kindOrder)
	}
	slices.SortStableFunc(order, func(a, b string) int { return rank(a) - rank(b) })
	var parts []string
	for _, k := range order {
		parts = append(parts, fmt.Sprintf("%s %d%%", kindName(k, none), (n[k]*100+len(cells)/2)/len(cells)))
	}
	return strings.Join(parts, ", ")
}

// partsSpoken is a bar's parts: "deep 3 hours, no label 1 hour 15 minutes".
func partsSpoken(ps []Part) string {
	var out []string
	for _, p := range ps {
		out = append(out, kindName(p.Kind, "no label")+" "+Say(report.Dur(p.MS)))
	}
	return strings.Join(out, ", ")
}

// withSpoken is i with the words VoiceOver reads filled in. The Cocoa side
// gets the island through JSON, which calls this; i itself is not changed.
func (i Island) withSpoken() Island {
	i.Face = faceSpoken(i.Face, i.State)
	if i.Pending && i.Hover != nil {
		i.Face.Spoken = "Missed: " + faceSpoken(i.Hover.Face, "").Spoken
	}
	i.Panel = panelSpoken(i.Panel, i.State)
	i.Hover = panelSpoken(i.Hover, i.State)
	return i
}

// faceSpoken fills f.Spoken: the face's words in reading order. A countdown
// (EndMS) is left out: the Cocoa side reads it live as the value.
func faceSpoken(f Face, state string) Face {
	if f.Spoken != "" {
		return f
	}
	var parts []string
	add := func(s ...string) {
		if t := strings.TrimSpace(strings.Join(s, " ")); t != "" {
			parts = append(parts, Say(t))
		}
	}
	switch state {
	case IslandQuiet:
		add(f.Left, f.LeftDim)
		if f.Right != "" {
			add("score", f.Right)
		}
		add(f.RightDim)
	case IslandFocus:
		if f.Left == "Focus" {
			add("Focus session")
		} else {
			add("Focus session:", f.Left)
		}
	case IslandBudget:
		add(f.Left)
		if f.Right != "" {
			add(f.Right, "of its budget")
		}
	default:
		add(f.Left, f.LeftDim)
		if f.EndMS == 0 {
			if f.RightDim == "in" {
				add(f.RightDim, f.Right)
			} else {
				add(f.Right, f.RightDim)
			}
		}
	}
	f.Spoken = strings.Join(parts, ", ")
	return f
}

// panelSpoken is a copy of p with its words for VoiceOver filled in.
func panelSpoken(p *Panel, state string) *Panel {
	if p == nil {
		return nil
	}
	c := *p
	c.Face = faceSpoken(c.Face, state)
	c.Cells = cellsSpoken(c.Cells)
	c.Buttons = slices.Clone(c.Buttons)
	for k, b := range c.Buttons {
		c.Buttons[k].Spoken = buttonSpoken(b)
	}
	if len(c.Strip) > 0 && c.StripSpoken == "" {
		if state == IslandBudget {
			c.StripSpoken = budgetSpoken(c.Cells)
		} else {
			c.StripSpoken = "The day by kind: " + kindShares(c.Strip, "no label or away")
		}
	}
	c.RowsSpoken = rowsSpoken(c.Rows)
	c.Entries = slices.Clone(c.Entries)
	for k := range c.Entries {
		c.Entries[k] = entrySpoken(c.Entries[k])
	}
	if len(c.Pages) > 0 {
		c.Pages = slices.Clone(c.Pages)
		for k := range c.Pages {
			c.Pages[k] = pageSpoken(c.Pages[k])
		}
	}
	return &c
}

// budgetSpoken is the budget gauge as words, from the panel's cells.
func budgetSpoken(cells []Cell) string {
	var spent, budget string
	for _, c := range cells {
		switch c.K {
		case "Spent":
			spent = c.V
		case "Budget":
			budget = c.V
		}
	}
	return Say(spent + " spent of " + budget)
}

func cellsSpoken(cs []Cell) []Cell {
	cs = slices.Clone(cs)
	for k, c := range cs {
		w := c.K + " " + c.V
		if c.C != "" {
			w += ", " + c.C
		}
		cs[k].Spoken = Say(w)
	}
	return cs
}

func buttonSpoken(b NButton) string {
	if m, ok := startMinutes(b.ID); ok {
		return "Start a " + strconv.Itoa(m) + " minute focus session"
	}
	return Say(b.Title)
}

func rowsSpoken(rows [][2]string) []string {
	var out []string
	for _, r := range rows {
		out = append(out, r[0]+", "+Say(r[1]))
	}
	return out
}

func entrySpoken(e Entry) Entry {
	w := e.App + ", " + e.Title
	if e.Sub != "" {
		w += ", " + e.Sub
	}
	e.Spoken = w + ", " + Say(e.Dur)
	e.Opts = slices.Clone(e.Opts)
	for k, o := range e.Opts {
		w := o.Label
		if o.Pct > 0 {
			w += fmt.Sprintf(", %d%% likely", o.Pct)
		}
		e.Opts[k].Spoken = w
	}
	return e
}

func pageSpoken(p Page) Page {
	p.Cells = cellsSpoken(p.Cells)
	p.Rings = slices.Clone(p.Rings)
	for k, r := range p.Rings {
		w := r.Label + ": " + Say(strings.Replace(r.Value, " / ", " of ", 1))
		if p.ID == "goals" {
			switch r.Kind {
			case "distract":
				w += ", over"
			case "shallow":
				w += ", short"
			}
		}
		if r.Sub != "" {
			w += ", " + r.Sub
		}
		p.Rings[k].Spoken = w
	}
	p.Bars = slices.Clone(p.Bars)
	for k, b := range p.Bars {
		p.Bars[k].Spoken = barSpoken(b, p.ID == "week")
	}
	if len(p.Strip) > 0 && p.StripSpoken == "" {
		p.StripSpoken = "By kind: " + kindShares(p.Strip, "no label")
	}
	p.RowsSpoken = rowsSpoken(p.Rows)
	if p.Mid != nil {
		m := pageSpoken(*p.Mid)
		p.Mid = &m
	}
	if p.Big != nil {
		b := pageSpoken(*p.Big)
		p.Big = &b
	}
	return p
}

// barSpoken is one column of a bar chart; a week's columns are days.
func barSpoken(b BarCol, week bool) string {
	w := b.Label
	if week && b.Hi {
		w += ", today"
	}
	var total int64
	for _, p := range b.Parts {
		total += p.MS
	}
	switch {
	case total == 0:
		w += ": nothing"
	case len(b.Parts) == 1:
		w += ": " + partsSpoken(b.Parts)
	default:
		w += ": " + Say(report.Dur(total)) + "; " + partsSpoken(b.Parts)
	}
	if b.Sub != "" && week {
		w += "; score " + b.Sub
	}
	return w
}
