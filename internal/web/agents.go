package web

import (
	"net/http"
	"time"

	"tock/internal/classify"
)

type agentRowJSON struct {
	Project  string `json:"project"`
	AgentMS  int64  `json:"agentMS"`
	YouMS    int64  `json:"youMS"`
	Prompts  int    `json:"prompts"`
	Sessions int    `json:"sessions"`
}

type agentsJSON struct {
	From    string         `json:"from"`
	To      string         `json:"to"`
	Rows    []agentRowJSON `json:"rows"`
	AgentMS int64          `json:"agentMS"`
	WallMS  int64          `json:"wallMS"`
	YouMS   int64          `json:"youMS"`
	Peak    int            `json:"peak"`
}

// agents is the agent ledger of a day (date=YYYY-MM-DD, default today), of
// the week or month holding it (range=week, range=month), or of the days
// from= to to= inclusive, a year at most.
func (s *Server) agents(r *http.Request) (any, error) {
	day, err := s.date(r)
	if err != nil {
		return nil, err
	}
	q := r.URL.Query()
	from, to := day, day.AddDate(0, 0, 1)
	if q.Has("from") || q.Has("to") {
		loc := s.now().Location()
		f, err1 := time.ParseInLocation(dateLayout, q.Get("from"), loc)
		t, err2 := time.ParseInLocation(dateLayout, q.Get("to"), loc)
		if err1 != nil || err2 != nil {
			return nil, badRequest("from and to must both be YYYY-MM-DD")
		}
		if t.Before(f) || t.After(f.AddDate(1, 0, 0)) {
			return nil, badRequest("to must be on or after from, within a year of it")
		}
		from, to = f, t.AddDate(0, 0, 1)
	} else {
		switch q.Get("range") {
		case "", "day":
		case "week":
			from = s.a.WeekOf(day)
			to = from.AddDate(0, 0, 7)
		case "month":
			from = time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
			to = from.AddDate(0, 1, 0)
		default:
			return nil, badRequest("range must be day, week or month")
		}
	}
	ag, err := s.a.Agents(from, to)
	if err != nil {
		return nil, err
	}
	out := agentsJSON{From: from.Format(dateLayout), To: to.AddDate(0, 0, -1).Format(dateLayout), Rows: []agentRowJSON{},
		AgentMS: ag.AgentMS, WallMS: ag.WallMS, YouMS: ag.YouMS, Peak: ag.Peak}
	for _, row := range ag.Rows {
		out.Rows = append(out.Rows, agentRowJSON{row.Project, row.AgentMS, row.YouMS, row.Prompts, row.Sessions})
	}
	return out, nil
}

type budgetJSON struct {
	Project   string `json:"project"`
	BudgetMS  int64  `json:"budgetMS"`
	Budget    string `json:"budget"` // as rules.toml writes it, "40h"
	Since     string `json:"since,omitempty"`
	SpentMS   int64  `json:"spentMS"`
	PerDayMS  int64  `json:"perDayMS"`
	RunsOutMS int64  `json:"runsOutMS,omitempty"`
	Pct       int    `json:"pct"`
	Over      bool   `json:"over"`
}

// budgets is every project budget, in rules order.
func (s *Server) budgets(*http.Request) (any, error) {
	rows, err := s.a.Budgets()
	if err != nil {
		return nil, err
	}
	since := map[string]string{}
	for _, p := range s.a.Rules().Projects {
		since[p.Name] = p.BudgetSince
	}
	out := []budgetJSON{}
	for _, b := range rows {
		out = append(out, budgetJSON{Project: b.Project, BudgetMS: b.BudgetMS, Budget: classify.FormatDur(b.BudgetMS),
			Since: since[b.Project], SpentMS: b.SpentMS, PerDayMS: b.PerDayMS, RunsOutMS: b.RunsOutMS, Pct: b.Pct(), Over: b.Over()})
	}
	return out, nil
}
