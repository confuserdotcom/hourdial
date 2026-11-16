//go:build demo

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tock/internal/app"
	"tock/internal/apptest"
)

// showcaseRules is a freelancer's rule set: two clients billed by the hour,
// tock itself, and the usual leaks.
const showcaseRules = `
[[category]]
name = "Code"
kind = "deep"

[[category]]
name = "Design"
kind = "deep"

[[category]]
name = "Writing"
kind = "deep"

[[category]]
name = "Meetings"
kind = "shallow"

[[category]]
name = "Communication"
kind = "shallow"

[[category]]
name = "Research"
kind = "browse"

[[category]]
name = "Social"
kind = "distract"

[[category]]
name = "Video"
kind = "distract"

[[project]]
name = "Atlas"
aliases = ["atlas-ios"]
repos = ["/Users/sam/code/atlas-ios"]
rate = 95
budget = "120h"

[[project]]
name = "Northwind"
repos = ["/Users/sam/code/northwind-web"]
rate = 120
budget = "60h"

[[project]]
name = "tock"
repos = ["/Users/sam/code/tock"]
category = "Code"

[[project]]
name = "Blog"
category = "Writing"

[[project]]
name = "North-wind"
category = "Design"

[[project]]
name = "Pitch deck"
category = "Design"

[[project]]
name = "Old site"
archived = true

[[rule]]
title = 'Atlas'
project = "Atlas"

[[rule]]
title = 'Northwind'
project = "Northwind"

[[rule]]
title = '^tock'
project = "tock"

[[rule]]
title = 'blog'
project = "Blog"

[[rule]]
bundle = ["com.apple.dt.Xcode", "term"]
category = "Code"

[[rule]]
bundle = ["com.figma.Desktop"]
category = "Design"

[[rule]]
bundle = ["notion.id"]
category = "Writing"

[[rule]]
bundle = ["us.zoom.xos"]
category = "Meetings"

[[rule]]
bundle = ["com.tinyspeck.slackmacgap", "mail"]
category = "Communication"

[[rule]]
url = '^developer\.apple\.com|^stackoverflow\.com|^github\.com'
category = "Research"

[[rule]]
url = '^(x|twitter)\.com|^reddit\.com'
category = "Social"

[[rule]]
url = '^(www\.)?youtube\.com'
category = "Video"

[[goal]]
category = "Code"
at_least = "3h"

[[goal]]
category = "Design"
at_least = "8h"
per = "week"

[[goal]]
category = "Social"
at_most = "30m"
`

// block is one stretch of a showcase day: from and to as minutes after
// midnight, what was in front.
type block struct {
	from, to    int
	bundle, app string
	title, url  string
}

func hm(h, m int) int { return h*60 + m }

// showcaseToday is the day the video shows, a Thursday up to 17:20.
func showcaseToday() []block {
	return []block{
		{hm(8, 52), hm(9, 6), "mail", "Mail", "Inbox (12)", ""},
		{hm(9, 6), hm(9, 18), "com.tinyspeck.slackmacgap", "Slack", "#atlas", ""},
		{hm(9, 18), hm(10, 2), "com.apple.dt.Xcode", "Xcode", "Atlas — PlayerView.swift", ""},
		{hm(10, 2), hm(10, 5), "com.tinyspeck.slackmacgap", "Slack", "#atlas", ""},
		{hm(10, 5), hm(11, 4), "com.apple.dt.Xcode", "Xcode", "Atlas — Playback.swift", ""},
		{hm(11, 4), hm(11, 13), "arc", "Arc", "Video", "www.youtube.com/watch"},
		{hm(11, 13), hm(11, 30), "arc", "Arc", "AVPlayer | Apple Developer", "developer.apple.com/documentation/avfoundation"},
		{hm(11, 30), hm(12, 2), "us.zoom.xos", "zoom.us", "Atlas weekly", ""},
		{hm(12, 2), hm(12, 51), "tock.idle", "Idle", "", ""},
		{hm(12, 51), hm(13, 4), "arc", "Arc", "Home / X", "x.com/home"},
		{hm(13, 4), hm(14, 41), "com.figma.Desktop", "Figma", "Northwind — Checkout flow", ""},
		{hm(14, 41), hm(14, 53), "arc", "Arc", "Spacing scale for dense UI", "chatgpt.com/c/7"},
		{hm(14, 53), hm(15, 9), "arc", "Arc", "Dribbble - checkout", "dribbble.com/shots/2214"},
		{hm(15, 9), hm(16, 31), "term", "WezTerm", "tock │ Resize the notch", ""},
		{hm(16, 31), hm(16, 44), "com.tinyspeck.slackmacgap", "Slack", "#northwind", ""},
		{hm(16, 44), -1, "notion.id", "Notion", "Northwind — Spec v3", ""},
	}
}

// showcaseDay makes up a past day: mornings in Xcode or Figma, a meeting,
// a lunch away, an afternoon on a second project, and some leaks.
func showcaseDay(r *rand.Rand, wd time.Weekday) []block {
	if wd == time.Sunday || (wd == time.Saturday && r.Intn(3) > 0) {
		if r.Intn(2) == 0 {
			return nil
		}
		s := hm(10+r.Intn(2), r.Intn(50))
		return []block{
			{s, s + 70 + r.Intn(60), "term", "WezTerm", "tock │ nvim", ""},
			{s + 140, s + 175, "arc", "Arc", "Video", "www.youtube.com/watch"},
		}
	}
	j := func(n int) int { return r.Intn(n) }
	t := hm(8, 40+j(35))
	var out []block
	add := func(d int, bundle, app, title, url string) {
		out = append(out, block{t, t + d, bundle, app, title, url})
		t += d
	}
	add(8+j(10), "mail", "Mail", "Inbox", "")
	add(6+j(12), "com.tinyspeck.slackmacgap", "Slack", "#atlas", "")
	if r.Intn(5) < 3 {
		add(60+j(70), "com.apple.dt.Xcode", "Xcode", "Atlas — Player.swift", "")
	} else {
		add(60+j(70), "com.figma.Desktop", "Figma", "Northwind — Components", "")
	}
	if r.Intn(3) == 0 {
		add(5+j(20), "arc", "Arc", "Video", "www.youtube.com/watch")
	}
	add(10+j(25), "arc", "Arc", "Docs", "developer.apple.com/documentation")
	if wd == time.Monday || wd == time.Thursday {
		add(30, "us.zoom.xos", "zoom.us", "Atlas weekly", "")
	}
	add(25+j(50), "com.apple.dt.Xcode", "Xcode", "Atlas — Tests", "")
	add(40+j(25), "tock.idle", "Idle", "", "")
	add(5+j(20), "arc", "Arc", "Home / X", "x.com/home")
	if r.Intn(2) == 0 {
		add(60+j(60), "com.figma.Desktop", "Figma", "Northwind — Checkout flow", "")
	} else {
		add(50+j(50), "notion.id", "Notion", "Northwind — Spec", "")
	}
	add(6+j(10), "com.tinyspeck.slackmacgap", "Slack", "#northwind", "")
	add(50+j(70), "term", "WezTerm", "tock │ Fix the flaky store test", "")
	if r.Intn(3) == 0 {
		add(10+j(25), "arc", "Arc", "r/macapps", "reddit.com/r/macapps")
	}
	if r.Intn(4) == 0 {
		add(30+j(30), "notion.id", "Notion", "blog: what I learned", "")
	}
	return out
}

// agentRun is one coding-agent session: minutes of the day it worked in a
// repo, and how many prompts started it off.
type agentRun struct {
	repo     string
	from, to int
	prompts  int
	title    string // the session's title, as its terminal shows it
}

// showcaseAgentsToday runs beside showcaseToday: tock's agent works through
// the Atlas morning, Atlas and Northwind get one each, and tock's second
// one is the terminal afternoon.
func showcaseAgentsToday() []agentRun {
	return []agentRun{
		{"tock", hm(9, 24), hm(10, 52), 4, "Notch on every display"},
		{"atlas-ios", hm(10, 8), hm(11, 2), 6, "Playback tests"},
		{"northwind-web", hm(13, 10), hm(14, 22), 5, "Checkout form states"},
		{"tock", hm(15, 12), hm(16, 28), 9, "Resize the notch"},
	}
}

// showcaseAgentsDay makes up a past day's sessions: client repos in the
// morning, tock in the afternoon.
func showcaseAgentsDay(r *rand.Rand, wd time.Weekday) []agentRun {
	if wd == time.Saturday || wd == time.Sunday {
		return nil
	}
	var out []agentRun
	if r.Intn(4) > 0 {
		s := hm(9, 20+r.Intn(40))
		out = append(out, agentRun{[]string{"atlas-ios", "atlas-ios", "northwind-web"}[r.Intn(3)], s, s + 35 + r.Intn(70), 2 + r.Intn(6), "Morning run"})
	}
	if r.Intn(3) > 0 {
		s := hm(14, 30+r.Intn(60))
		out = append(out, agentRun{"tock", s, s + 30 + r.Intn(80), 2 + r.Intn(8), "Fix the flaky store test"})
	}
	return out
}

// writeAgentLogs writes the runs as Claude Code session logs under dir: a
// human prompt now and then, and an assistant line for each minute worked.
func writeAgentLogs(t *testing.T, dir string, day time.Time, n int, runs []agentRun) {
	t.Helper()
	for i, run := range runs {
		var b strings.Builder
		id := fmt.Sprintf("%s-%d-%d", day.Format("0102"), n, i)
		line := func(typ string, min int, extra string) {
			ts := day.Add(time.Duration(min) * time.Minute).UTC().Format(time.RFC3339Nano)
			fmt.Fprintf(&b, `{"type":%q,"timestamp":%q,"cwd":%q,"sessionId":%q%s}`+"\n", typ, ts, "/Users/sam/code/"+run.repo, id, extra)
		}
		fmt.Fprintf(&b, `{"type":"ai-title","aiTitle":%q,"sessionId":%q}`+"\n", run.title, id)
		every := max(1, (run.to-run.from)/max(1, run.prompts))
		for m := run.from; m <= run.to; m++ {
			if (m-run.from)%every == 0 && (m-run.from)/every < run.prompts {
				line("user", m, `,"origin":{"kind":"human"}`)
			}
			line("assistant", m, "")
		}
		p := filepath.Join(dir, run.repo)
		if err := os.MkdirAll(p, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, id+".jsonl"), []byte(b.String()), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// TestShowcase serves a month of made-up freelancer data for the README
// videos and screenshots:
//
//	TOCK_DEMO_PORT=7799 go test -tags demo -run TestShowcase -timeout 0 ./internal/web
func TestShowcase(t *testing.T) {
	a, e := showcase(t)
	port := os.Getenv("TOCK_DEMO_PORT")
	if port == "" {
		port = "7799"
	}
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		t.Fatal(err)
	}
	token := fmt.Sprintf("%064d", 0)
	fmt.Printf("showcase: http://127.0.0.1:%d/#t=%s\n", ln.Addr().(*net.TCPAddr).Port, token)
	t.Fatal(http.Serve(ln, New(a, token, ln.Addr().(*net.TCPAddr).Port, e.NowFunc()).Handler()))
}

// showcase is the app over the made-up month, its clock at 17:20 on the
// showcase day.
func showcase(t *testing.T) (*app.App, *apptest.Env) {
	e := apptest.New(t, showcaseRules)
	e.WriteConfig(t, "[invoice]\nname = \"Sam Rivera\"\ncurrency = \"EUR\"\n")
	now := time.Date(2026, 11, 5, 17, 20, 0, 0, apptest.Zone)
	e.Clock = now
	db := e.RawDB(t)
	defer db.Close()
	id := int64(0)
	put := func(day time.Time, bs []block) {
		for _, b := range bs {
			id++
			start := day.Add(time.Duration(b.from) * time.Minute).UnixMilli()
			var end any
			if b.to >= 0 {
				end = day.Add(time.Duration(b.to) * time.Minute).UnixMilli()
			}
			var title, url any
			if b.title != "" {
				title = b.title
			}
			if b.url != "" {
				url = b.url
			}
			if _, err := db.Exec(`INSERT INTO span(id, start, "end", bundle_id, app, title, url, private) VALUES(?,?,?,?,?,?,?,0)`,
				id, start, end, b.bundle, b.app, title, url); err != nil {
				t.Fatal(err)
			}
		}
	}
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, apptest.Zone)
	r := rand.New(rand.NewSource(7))
	for d := 42; d >= 1; d-- {
		day := midnight.AddDate(0, 0, -d)
		put(day, showcaseDay(r, day.Weekday()))
	}
	put(midnight, showcaseToday())
	logs := t.TempDir()
	ar := rand.New(rand.NewSource(11))
	for d := 42; d >= 1; d-- {
		day := midnight.AddDate(0, 0, -d)
		writeAgentLogs(t, logs, day, d, showcaseAgentsDay(ar, day.Weekday()))
	}
	writeAgentLogs(t, logs, midnight, 0, showcaseAgentsToday())
	at := func(h, m int) int64 { return midnight.Add(time.Duration(hm(h, m)) * time.Minute).UnixMilli() }
	if _, err := db.Exec(`INSERT INTO call(start_ms, end_ms) VALUES(?, ?)`, at(11, 30), at(12, 2)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO event(start_ms, end_ms, title, calendar, join_url, mine) VALUES(?, ?, 'Atlas weekly', 'Work', '', 1)`, at(11, 30), at(12, 0)); err != nil {
		t.Fatal(err)
	}
	day := now.Format("2006-01-02")
	e.Summary(t, day, 0, "A long morning in Atlas playback, the Northwind checkout flow after lunch, then tock's notch until five.")
	e.Heartbeat(t, now.UnixMilli())
	a, err := app.Open(app.Paths{DB: e.DB, Rules: e.Rules, Learned: e.Learned, Config: e.Config, ShellLog: e.ShellLog, Agents: logs}, e.NowFunc())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Close() })
	return a, e
}

// TestSampleData writes the sample a first run shows (web/public/sample.json):
// the API's answers on the showcase day, with one invoice issued so the list
// has a row.
//
//	TOCK_SAMPLE_OUT=$PWD/web/public/sample.json go test -tags demo -run TestSampleData ./internal/web
func TestSampleData(t *testing.T) {
	out := os.Getenv("TOCK_SAMPLE_OUT")
	if out == "" {
		t.Skip("TOCK_SAMPLE_OUT names the file to write")
	}
	a, e := showcase(t)
	oct := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, apptest.Zone) }
	if _, err := a.IssueInvoice(app.InvoiceSpec{Project: "Atlas", From: oct(1), To: oct(31)}); err != nil {
		t.Fatal(err)
	}
	h := New(a, testToken, 0, e.NowFunc()).Handler()
	const d = "2026-11-05"
	paths := []string{"/api/day?date=" + d, "/api/week?date=" + d, "/api/hours", "/api/month?month=" + d[:7],
		"/api/budgets", "/api/library", "/api/inbox?days=7", "/api/labels", "/api/invoices"}
	for _, r := range []string{"day", "week", "month"} {
		ag := "/api/agents?date=" + d
		if r != "day" {
			ag += "&range=" + r
		}
		top := "/api/top?date=" + d
		if r != "day" {
			top += "&range=" + r
		}
		paths = append(paths, ag, top, top+"&series=project")
		if r == "day" {
			paths = append(paths, top+"&range=day", top+"&range=day&series=project")
		}
	}
	paths = append(paths, "/api/top?range=today", "/api/top?range=week", "/api/top?range=month")
	answers := map[string]json.RawMessage{}
	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		req.Header.Set("X-Tock-Token", testToken)
		req.Host = "127.0.0.1:0"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("GET %s: %d %s", p, w.Code, w.Body)
		}
		answers[p] = json.RawMessage(bytes.TrimSpace(w.Body.Bytes()))
	}
	b, err := json.Marshal(struct {
		Day     string                     `json:"day"`
		Answers map[string]json.RawMessage `json:"answers"`
	}{d, answers})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
