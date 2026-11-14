package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tock/internal/testfs"
)

const sentinel = "ZZSENTINELZZ"

// No title, URL, name, address, sync folder or hidden site from the
// user's data reaches the archive.
func TestBugReportLeaksNothing(t *testing.T) {
	te := newTestEnv(t, testRules)
	te.seed(t,
		seedSpan{1, "09:00", "10:00", "arc", "Arc", sentinel + " secret page", "https://" + sentinel + ".example/x", false},
		seedSpan{2, "10:00", "11:00", "term", "WezTerm", "vim " + sentinel, "", false},
	)
	syncDir := filepath.Join(t.TempDir(), sentinel+"-sync")
	cfg := `# my notes ` + sentinel + `
[privacy]
keep_titles = 30
hide_sites = [
  "` + sentinel + `.example",
]

[sync]
dir = "` + syncDir + `"

[invoice]
name = "` + sentinel + ` Ltd"
address = ["1 ` + sentinel + ` Street", "Town"]
`
	if err := os.WriteFile(filepath.Join(te.ConfigDir, "config.toml"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(te.StateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	trackdLog := "level=INFO msg=\"trackd starting\"\nlevel=WARN msg=x title=\"" + sentinel + "\"\nlevel=INFO msg=y url=https://" + sentinel + ".example\nlevel=INFO msg=\"kept line\"\n"
	if err := os.WriteFile(filepath.Join(te.StateDir, "trackd.log"), []byte(trackdLog), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(te.StateDir, "web.log"), []byte("GET /?u=https://"+sentinel+".example\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "report.tar.gz")
	code, stdout := te.run(t, "", "bug-report", "-o", out)
	if code != 0 {
		t.Fatalf("exit %d: %s %s", code, stdout, te.errb)
	}
	if !strings.Contains(stdout, "doctor.json") || !strings.Contains(stdout, "Nothing was sent") {
		t.Fatalf("stdout:\n%s", stdout)
	}
	if fi, err := os.Stat(out); err != nil || !testfs.Perm(fi, 0o600) {
		t.Fatalf("archive mode: %v %v", fi, err)
	}
	members := readTarGz(t, out)
	for _, want := range []string{"doctor.json", "config.toml", "trackd.log", "trackd_runs.json"} {
		if _, ok := members[want]; !ok {
			t.Errorf("archive lacks %s; has %v", want, keys(members))
		}
	}
	if _, ok := members["rules.toml"]; ok {
		t.Error("rules.toml without --include-rules")
	}
	for name, body := range members {
		if strings.Contains(body, sentinel) {
			t.Errorf("%s leaks the sentinel:\n%s", name, body)
		}
	}
	if !strings.Contains(members["trackd.log"], "kept line") {
		t.Errorf("trackd.log lost an ordinary line:\n%s", members["trackd.log"])
	}
	if !strings.Contains(members["config.toml"], "keep_titles = 30") || !strings.Contains(members["config.toml"], "[sync]") {
		t.Errorf("config.toml lost its shape:\n%s", members["config.toml"])
	}

	code, _ = te.run(t, "", "bug-report", "-o", out, "--include-rules")
	if _, ok := readTarGz(t, out)["rules.toml"]; code != 0 || !ok {
		t.Fatal("--include-rules did not add rules.toml")
	}
}

// A crash log as the runtime writes it, with a user's title, URL, token and
// home folder in the panic message.
const crashLog = `
--- 2026-11-14T03:12:09Z trackd v1.0.0 pid 4242: panic in sync loop
panic: bad span "` + sentinel + ` - Inbox" at https://mail.` + sentinel + `.example/u/0 [recovered, repanicked]
	panic: second line of the message ` + sentinel + `
token sk-ant-api03-` + sentinel + `abc and xoxp-1234-` + sentinel + `

goroutine 20 [running]:
tock/internal/crashlog.Recover({0x1030e5814, 0x9})
	/Users/` + sentinel + `/src/tock/internal/crashlog/crashlog.go:105 +0x54
panic({0x10323d370?, 0xaeb8851a310?})
	/home/` + sentinel + `/go/src/runtime/panic.go:859 +0x120
main.(*syncLoop).start.func1()
	tock/cmd/trackd/sync.go:88 +0x60
created by main.(*syncLoop).start in goroutine 1
	tock/cmd/trackd/sync.go:86 +0x1f4

--- 2026-11-14T04:00:00Z trackd v1.0.0 pid 4300: panic in pipeline
panic: runtime error: index out of range [3] with length 3 [recovered, repanicked]

goroutine 7 [running]:
main.(*pipeline).run(0x14000123000)
	tock/cmd/trackd/pipeline.go:40 +0x1c
`

// Crash logs go in, with the trace kept and everything that could be the
// user's (the panic message, paths with their name, tokens) taken out.
func TestBugReportCrashLogs(t *testing.T) {
	te := newTestEnv(t, testRules)
	if err := os.MkdirAll(te.StateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(te.StateDir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("trackd-crash.log", crashLog)
	sigquit := "SIGQUIT: quit\nPC=0x183e37c34 m=0 sigcode=0\nsignal arrived during cgo execution\n\n" +
		"goroutine 1 gp=0xa2e4a5e81e0 m=0 mp=0x104d28a60 [syscall, locked to thread]:\n" +
		"runtime.gopark(0x0?, 0x0?)\n\truntime/proc.go:480 +0xc4 fp=0x1 sp=0x2 pc=0x3\nr0      0x0\n"
	write("tock-crash.log.1", "fatal error: concurrent map writes\n\ngoroutine 1 [running]:\nmain.main()\n\ttock/cmd/tock/main.go:9 +0x10\n"+sigquit)
	write("tockbar-crash.log", "") // nothing crashed

	out := filepath.Join(t.TempDir(), "report.tar.gz")
	if code, stdout := te.run(t, "", "bug-report", "-o", out); code != 0 {
		t.Fatalf("exit %d: %s %s", code, stdout, te.errb)
	}
	members := readTarGz(t, out)
	got, ok := members["trackd-crash.log"]
	if !ok {
		t.Fatalf("archive lacks trackd-crash.log; has %v", keys(members))
	}
	if _, ok := members["tockbar-crash.log"]; ok {
		t.Error("an empty crash log went in")
	}
	if got := members["tock-crash.log.1"]; !strings.Contains(got, "fatal error: concurrent map writes") ||
		!strings.Contains(got, strings.TrimSpace(sigquit)) || strings.Contains(got, "<removed>") {
		t.Errorf("tock-crash.log.1 lost part of the trace:\n%s", got)
	}
	for name, body := range members {
		for _, leak := range []string{sentinel, "sk-ant-", "xoxp-", "https://"} {
			if strings.Contains(body, leak) {
				t.Errorf("%s leaks %q:\n%s", name, leak, body)
			}
		}
	}
	for _, want := range []string{
		"--- 2026-11-14T03:12:09Z trackd v1.0.0 pid 4242: panic in sync loop",
		"panic: <removed> [recovered, repanicked]",
		"panic: runtime error: index out of range [3] with length 3 [recovered, repanicked]",
		"goroutine 20 [running]:",
		"tock/internal/crashlog.Recover({0x1030e5814, 0x9})",
		"main.(*syncLoop).start.func1()",
		"\ttock/cmd/trackd/sync.go:88 +0x60",
		"crashlog/crashlog.go:105 +0x54",
		"created by main.(*syncLoop).start in goroutine 1",
		"main.(*pipeline).run(0x14000123000)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("trackd-crash.log lost %q:\n%s", want, got)
		}
	}
}

func readTarGz(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(zr)
	out := map[string]string{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return out
		}
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		out[h.Name] = string(b)
	}
}

func keys(m map[string]string) []string {
	var k []string
	for n := range m {
		k = append(k, n)
	}
	return k
}
