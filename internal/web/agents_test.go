package web

import (
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tock/internal/app"
	"tock/internal/apptest"
)

func TestAgentsAndBudgets(t *testing.T) {
	e := apptest.New(t, strings.Replace(apptest.Rules, "name = \"tock\"\n", "name = \"tock\"\nrepos = [\"/r/tock\"]\nbudget = \"10h\"\n", 1))
	e.Seed(t, apptest.DaySpans()...)
	logs := t.TempDir()
	var b strings.Builder
	for ms := apptest.At(t, "09:00", 0); ms <= apptest.At(t, "09:30", 0); ms += 60_000 {
		fmt.Fprintf(&b, `{"type":"assistant","timestamp":%q,"cwd":"/r/tock","sessionId":"s"}`+"\n", time.UnixMilli(ms).UTC().Format(time.RFC3339Nano))
	}
	os.MkdirAll(filepath.Join(logs, "p"), 0o700)
	os.WriteFile(filepath.Join(logs, "p", "s.jsonl"), []byte(b.String()), 0o600)
	a, err := app.Open(app.Paths{DB: e.DB, Rules: e.Rules, Learned: e.Learned, Config: e.Config, Agents: logs}, e.NowFunc())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Close() })
	srv := httptest.NewUnstartedServer(nil)
	srv.Config.Handler = New(a, testToken, srv.Listener.Addr().(*net.TCPAddr).Port, e.NowFunc()).Handler()
	srv.Start()
	t.Cleanup(srv.Close)
	f := &fixture{t, srv, e, a}

	var ag agentsJSON
	f.get("/api/agents", &ag)
	if len(ag.Rows) != 1 || ag.Rows[0].Project != "tock" || ag.Rows[0].AgentMS != 30*60_000 || ag.From != "2026-10-10" {
		t.Fatalf("agents = %+v", ag)
	}
	f.get("/api/agents?range=week", &ag)
	if ag.From != "2026-10-05" || ag.To != "2026-10-11" || len(ag.Rows) != 1 {
		t.Fatalf("week = %+v", ag)
	}
	f.get("/api/agents?range=month", &ag)
	if ag.From != "2026-10-01" || ag.To != "2026-10-31" || len(ag.Rows) != 1 {
		t.Fatalf("month = %+v", ag)
	}
	f.get("/api/agents?from=2026-10-09&to=2026-10-10", &ag)
	if ag.From != "2026-10-09" || ag.To != "2026-10-10" || len(ag.Rows) != 1 || ag.Rows[0].AgentMS != 30*60_000 {
		t.Fatalf("from-to = %+v", ag)
	}
	f.get("/api/agents?from=2026-10-08&to=2026-10-09", &ag)
	if len(ag.Rows) != 0 {
		t.Fatalf("days before = %+v", ag)
	}
	for _, q := range []string{"range=year", "from=2026-10-09", "from=2026-10-10&to=2026-10-09", "from=x&to=2026-10-09", "from=2025-01-01&to=2026-10-09"} {
		if code, _, _ := f.do("GET", "/api/agents?"+q, testToken, nil); code != 400 {
			t.Fatalf("%s: %d", q, code)
		}
	}
	var bs []budgetJSON
	f.get("/api/budgets", &bs)
	if len(bs) != 1 || bs[0].Project != "tock" || bs[0].Budget != "10h" || bs[0].SpentMS != 100*60_000 || bs[0].Pct != 16 || bs[0].Over {
		t.Fatalf("budgets = %+v", bs)
	}
}
