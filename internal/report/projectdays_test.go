package report

import (
	"reflect"
	"testing"
	"time"

	"tock/internal/classify"
	"tock/internal/span"
)

func TestProjectDays(t *testing.T) {
	loc := time.FixedZone("x", 2*3600)
	day := func(d, h, m int) int64 { return time.Date(2026, 11, d, h, m, 0, 0, loc).UnixMilli() }
	from, to := time.Date(2026, 11, 3, 0, 0, 0, 0, loc), time.Date(2026, 11, 6, 0, 0, 0, 0, loc)
	views := []classify.SpanView{
		{Start: day(3, 10, 0), End: day(3, 11, 0), BundleID: "a"},
		{Start: day(3, 23, 30), End: day(4, 0, 45), BundleID: "a"}, // crosses midnight
		{Start: day(5, 9, 0), End: day(5, 9, 30), BundleID: "a"},
		{Start: day(5, 10, 0), End: day(5, 12, 0), BundleID: span.IdleBundleID},
		{Start: day(5, 13, 0), End: day(5, 13, 20), BundleID: "a"},
		{Start: day(2, 23, 0), End: day(3, 0, 10), BundleID: "a"}, // starts before the range
	}
	labels := []classify.Label{{Project: "Atlas"}, {Project: "Atlas"}, {Project: "tock"}, {Project: "tock"}, {}, {Project: "tock"}}
	const min = 60_000
	got := ProjectDays(from, to, views, labels)
	want := []ProjectDayRow{
		{Name: "Atlas", MS: []int64{90 * min, 45 * min, 0}},
		{Name: "tock", MS: []int64{10 * min, 0, 30 * min}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
	if got := ProjectDays(from, from, views, labels); got != nil {
		t.Fatalf("empty range: %+v", got)
	}
}
