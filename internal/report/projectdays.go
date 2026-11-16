package report

import (
	"sort"
	"time"

	"tock/internal/classify"
	"tock/internal/span"
)

// ProjectDayRow is one project's time on each day of a range.
type ProjectDayRow struct {
	Name string
	MS   []int64
}

// ProjectDays splits the non-idle views with a project over the local days
// of [from, to), from and to being midnights: a view across midnight counts
// on both days. labels[i] labels views[i]. Most time first.
func ProjectDays(from, to time.Time, views []classify.SpanView, labels []classify.Label) []ProjectDayRow {
	var bounds []int64
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		bounds = append(bounds, d.UnixMilli())
	}
	days := len(bounds) - 1
	if days < 1 {
		return nil
	}
	rows := map[string][]int64{}
	for i, v := range views {
		p := labels[i].Project
		if p == "" || v.BundleID == span.IdleBundleID || v.DurMS() <= 0 {
			continue
		}
		start, end := v.Start, v.Start+v.DurMS()
		for d := 0; d < days; d++ {
			if st, e := max(start, bounds[d]), min(end, bounds[d+1]); e > st {
				if rows[p] == nil {
					rows[p] = make([]int64, days)
				}
				rows[p][d] += e - st
			}
		}
	}
	out := make([]ProjectDayRow, 0, len(rows))
	total := map[string]int64{}
	for n, ms := range rows {
		out = append(out, ProjectDayRow{n, ms})
		for _, x := range ms {
			total[n] += x
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if total[out[i].Name] != total[out[j].Name] {
			return total[out[i].Name] > total[out[j].Name]
		}
		return out[i].Name < out[j].Name
	})
	return out
}
