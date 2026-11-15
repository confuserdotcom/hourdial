package report

import (
	"strconv"
	"time"

	"tock/internal/classify"
)

// StreakDays is how many days before today streaks look back.
const StreakDays = 365

// GoalDay is one day as streaks see it.
type GoalDay struct {
	Date   time.Time // local midnight
	Active bool      // the day has active time
	Off    bool      // a day off in the work schedule
	MS     []int64   // whole-day time per goal, indexed like the goals
}

func (d GoalDay) skip() bool { return !d.Active || d.Off }

// Streak is goals[gi]'s current and best streak: consecutive held days for
// a day goal, held weeks for a week goal. past is the days before today,
// most recent first; closed reports that today's work hours are over.
//
// Skipped days (no activity, or a day off) neither count nor break a day
// goal's streak, and a week with no active day is skipped the same way.
// Today counts once an at_least goal is met, or once it is closed and the
// goal held; a limit already over breaks the streak; otherwise today is
// left out until it is decided.
func Streak(gi int, g classify.Goal, past []GoalDay, today GoalDay, closed bool) (cur, best int) {
	return StreakFrom(gi, g, past, today, closed, time.Monday)
}

// StreakFrom is Streak with weeks starting on first.
func StreakFrom(gi int, g classify.Goal, past []GoalDay, today GoalDay, closed bool, first time.Weekday) (cur, best int) {
	var seq []bool // oldest first
	if g.Week {
		seq = weekStreakSeq(gi, g, past, today, first)
	} else {
		for i := len(past) - 1; i >= 0; i-- {
			if d := past[i]; !d.skip() {
				seq = append(seq, held(goalState(g, d.MS[gi])))
			}
		}
		if !today.skip() {
			if h, ok := decided(g, goalState(g, today.MS[gi]), closed); ok {
				seq = append(seq, h)
			}
		}
	}
	run := 0
	for _, h := range seq {
		if h {
			run++
			best = max(best, run)
		} else {
			run = 0
		}
	}
	return run, best
}

func held(state string) bool { return state == GoalMet || state == GoalOK }

// decided reports whether an unfinished period already settles the goal,
// and how.
func decided(g classify.Goal, state string, closed bool) (held, ok bool) {
	switch {
	case closed:
		return state == GoalMet || state == GoalOK, true
	case state == GoalOver:
		return false, true
	case state == GoalMet:
		return true, true
	}
	return false, false
}

// weekStreakSeq is the held state of each week starting on first, oldest
// first. The oldest week is dropped when the lookback starts after its
// first day.
func weekStreakSeq(gi int, g classify.Goal, past []GoalDay, today GoalDay, first time.Weekday) []bool {
	type week struct {
		ms     int64
		active bool
	}
	thisMon := WeekOf(today.Date, first)
	var weeks []week // most recent first, index 0 is this week
	add := func(d GoalDay) {
		i := int(thisMon.Sub(WeekOf(d.Date, first)).Hours()+12) / (7 * 24)
		for len(weeks) <= i {
			weeks = append(weeks, week{})
		}
		weeks[i].ms += d.MS[gi]
		weeks[i].active = weeks[i].active || d.Active
	}
	add(today)
	for _, d := range past {
		add(d)
	}
	if n := len(past); n > 0 && len(weeks) > 1 && !past[n-1].Date.Equal(WeekOf(past[n-1].Date, first)) {
		weeks = weeks[:len(weeks)-1]
	}
	var seq []bool
	for i := len(weeks) - 1; i >= 1; i-- {
		if weeks[i].active {
			seq = append(seq, held(goalState(g, weeks[i].ms)))
		}
	}
	if weeks[0].active {
		if h, ok := decided(g, goalState(g, weeks[0].ms), false); ok {
			seq = append(seq, h)
		}
	}
	return seq
}

// streakText is " · 5d" or " · 3w" for a streak of at least 2.
func streakText(g classify.Goal, n int) string {
	if n < 2 {
		return ""
	}
	return " · " + streakUnit(g, n)
}

func streakUnit(g classify.Goal, n int) string {
	if g.Week {
		return strconv.Itoa(n) + "w"
	}
	return strconv.Itoa(n) + "d"
}
