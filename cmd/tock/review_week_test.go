package main

import (
	"testing"
	"time"
)

// tock review shows this week on its last two days, else last week.
func TestReviewWeekStart(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 10, d, 12, 0, 0, 0, time.UTC) } // 4 Oct 2026 is a Sunday
	cases := []struct {
		first time.Weekday
		day   int
		this  bool
	}{
		{time.Monday, 9, false}, {time.Monday, 10, true}, {time.Monday, 11, true}, {time.Monday, 12, false},
		{time.Sunday, 8, false}, {time.Sunday, 9, true}, {time.Sunday, 10, true}, {time.Sunday, 11, false},
		{time.Saturday, 7, false}, {time.Saturday, 8, true}, {time.Saturday, 9, true}, {time.Saturday, 10, false},
	}
	for _, c := range cases {
		got := reviewWeek(day(c.day), c.first)
		if this := got.Equal(day(c.day)); this != c.this {
			t.Errorf("%v weeks, %d Oct: this week = %v, want %v", c.first, c.day, this, c.this)
		}
	}
}
