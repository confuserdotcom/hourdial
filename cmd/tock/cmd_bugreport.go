package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"tock/internal/config"
	"tock/internal/crashlog"
)

// bugReportLogLines is how much of each log goes in.
const bugReportLogLines = 500

// bugReportRunDays is how far back trackd's runs go in.
const bugReportRunDays = 14

// privateLine matches log lines that could carry a window title or a URL;
// they are left out whole.
var privateLine = regexp.MustCompile(`(?i)https?://|\btitle\b|\burl\b|\bwindow\b`)

// cmdBugReport writes a .tar.gz a user can look through and attach to an
// issue: doctor --json, config.toml with every text value removed, the end
// of the logs without lines that could hold a title or URL, and trackd's
// runs. It sends nothing.
func cmdBugReport(env Env, args []string) error {
	fs := newFlags("bug-report")
	out := fs.String("o", "", "write the archive here (default ~/.local/state/tock/bug-report-<time>.tar.gz)")
	withRules := fs.Bool("include-rules", false, "add rules.toml (your app, site and project names)")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return usagef("bug-report takes only flags")
	}
	path := *out
	if path == "" {
		path = filepath.Join(env.StateDir, "bug-report-"+env.Now().Format("20060102-150405")+".tar.gz")
	}
	scrub := scrubber(env)

	type member struct {
		name string
		body []byte
	}
	var members []member
	add := func(name string, b []byte) {
		members = append(members, member{name, []byte(scrubSecrets(scrub(string(b))))})
	}

	var doc bytes.Buffer
	denv := env
	denv.Stdout, denv.Stderr, denv.Open = &doc, io.Discard, nil
	_ = cmdDoctor(denv, []string{"--json"}) // problems are what the report is for
	add("doctor.json", doc.Bytes())

	if b, err := os.ReadFile(env.configPath()); err == nil {
		add("config.toml", []byte(redactTOML(string(b))))
	}
	for _, l := range []string{"trackd.log", "web.log", "tockbar.log"} {
		if b, err := logTail(filepath.Join(env.StateDir, l)); err == nil {
			add(l, b)
		}
	}
	for _, prog := range []string{"trackd", "tock", "tockbar"} {
		p := crashlog.Path(env.StateDir, prog)
		for _, f := range []string{p, p + ".1"} {
			if b, err := os.ReadFile(f); err == nil && len(bytes.TrimSpace(b)) > 0 {
				add(filepath.Base(f), []byte(redactCrash(string(b))))
			}
		}
	}
	if b, err := bugReportRuns(env); err == nil {
		add("trackd_runs.json", b)
	}
	if *withRules {
		if b, err := os.ReadFile(env.rulesPath()); err == nil {
			add("rules.toml", b)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("bug-report: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("bug-report: %w", err)
	}
	zw := gzip.NewWriter(f)
	tw := tar.NewWriter(zw)
	for _, m := range members {
		h := &tar.Header{Name: m.name, Mode: 0o600, Size: int64(len(m.body)), ModTime: env.Now()}
		if err := tw.WriteHeader(h); err != nil {
			f.Close()
			return fmt.Errorf("bug-report: %w", err)
		}
		if _, err := tw.Write(m.body); err != nil {
			f.Close()
			return fmt.Errorf("bug-report: %w", err)
		}
	}
	err = tw.Close()
	if zerr := zw.Close(); err == nil {
		err = zerr
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("bug-report: %w", err)
	}

	fmt.Fprintf(env.Stdout, "wrote %s with:\n", path)
	for _, m := range members {
		fmt.Fprintf(env.Stdout, "  %-18s %s\n", m.name, sizeKB(len(m.body)))
	}
	fmt.Fprintf(env.Stdout, "Nothing was sent. Look through it first (tar xzf %s), then attach it to an issue.\n", filepath.Base(path))
	if !*withRules {
		fmt.Fprintln(env.Stdout, "Your rules are not in it; --include-rules adds rules.toml if the problem is about labels.")
	}
	return nil
}

func sizeKB(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f KB", float64(n)/1024)
}

// tomlString matches a TOML string literal.
var tomlString = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"|'([^']*)'`)

// scrubber replaces, in every member, each text value of config.toml with
// <removed> (doctor quotes them in its errors), the sync folder with
// <sync> and the home folder with ~: paths say who you are and where you
// sync.
func scrubber(env Env) func(string) string {
	var olds []string
	repl := map[string]string{}
	if c, err := config.Load(env.configPath()); err == nil && c.Sync.Dir != "" {
		olds, repl[c.Sync.Dir] = append(olds, c.Sync.Dir), "<sync>"
	}
	if src, err := os.ReadFile(env.configPath()); err == nil {
		for _, m := range tomlString.FindAllStringSubmatch(string(src), -1) {
			v := m[1] + m[2]
			if len(v) < 3 { // "", "on", short words: nothing private, and too common to replace
				continue
			}
			if _, dup := repl[v]; !dup {
				olds, repl[v] = append(olds, v), "<removed>"
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil && len(home) > 1 {
		olds, repl[home] = append(olds, home), "~"
	}
	// Longest first, so a value that contains another is replaced whole.
	sort.SliceStable(olds, func(a, b int) bool { return len(olds[a]) > len(olds[b]) })
	var pairs []string
	for _, o := range olds {
		pairs = append(pairs, o, repl[o])
	}
	return strings.NewReplacer(pairs...).Replace
}

// redactTOML keeps config.toml's tables, keys, numbers and booleans and
// replaces every string and array with "<removed>": names, addresses, the
// sync folder, hidden apps and sites. Comments go.
func redactTOML(src string) string {
	var b strings.Builder
	depth := 0 // inside a multi-line array
	for _, line := range strings.Split(src, "\n") {
		t := strings.TrimSpace(line)
		if depth > 0 {
			depth += strings.Count(t, "[") - strings.Count(t, "]")
			continue
		}
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if strings.HasPrefix(t, "[") {
			b.WriteString(t + "\n")
			continue
		}
		key, val, ok := strings.Cut(t, "=")
		if !ok {
			continue
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		switch {
		case strings.HasPrefix(val, "["):
			depth = strings.Count(val, "[") - strings.Count(val, "]")
			val = `"<removed>"`
		case strings.HasPrefix(val, `"`) || strings.HasPrefix(val, "'") || strings.HasPrefix(val, "{"):
			val = `"<removed>"`
		default:
			if i := strings.Index(val, "#"); i >= 0 {
				val = strings.TrimSpace(val[:i])
			}
		}
		b.WriteString(key + " = " + val + "\n")
	}
	return b.String()
}

// logTail is the last bugReportLogLines lines of a log, without the lines
// privateLine matches.
func logTail(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if privateLine.MatchString(sc.Text()) {
			continue
		}
		lines = append(lines, sc.Text())
		if len(lines) > bugReportLogLines {
			lines = lines[1:]
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

// secretPatterns are replaced in every member: other people's home folders
// (a binary built elsewhere names its builder's), Claude and Slack tokens,
// and the dashboard's per-run token.
var secretPatterns = []struct {
	re   *regexp.Regexp
	with string
}{
	{regexp.MustCompile(`/(?:Users|home)/[^/\s"'` + "`" + `]+`), "~"},
	{regexp.MustCompile(`(?i)\b[a-z]:\\Users\\[^\\\s"']+`), "~"},
	{regexp.MustCompile(`sk-ant-[A-Za-z0-9_\-]+`), "<token>"},
	{regexp.MustCompile(`xox[a-z]-[A-Za-z0-9\-]+`), "<token>"},
	{regexp.MustCompile(`(?i)(bearer\s+)\S+`), "${1}<token>"},
	{regexp.MustCompile(`#[0-9a-f]{32,}`), "#<token>"},
}

func scrubSecrets(s string) string {
	for _, p := range secretPatterns {
		s = p.re.ReplaceAllString(s, p.with)
	}
	return s
}

// Crash log lines are kept only when they have the shape of a Go trace,
// which holds code (function names, files, addresses), never data.
var crashKeep = []*regexp.Regexp{
	regexp.MustCompile(`^--- \S+ \S+ \S+ pid \d+: panic in [\w .-]+$`),                // crashlog's note
	regexp.MustCompile(`^goroutine \d+ .*\]:$`),                                       // goroutine 20 [running]:
	regexp.MustCompile(`^\S+\((?:0x[0-9a-f]+\??|\.\.\.|[{}, ])*\)$`),                  // pkg.F({0x1, 0x2})
	regexp.MustCompile(`^\t\S+:\d+(?: \+0x[0-9a-f]+)?(?: [a-z]+=0x[0-9a-f]+)*$`),      // \tfile.go:12 +0x54 [fp=0x…]
	regexp.MustCompile(`^created by \S+(?: in goroutine \d+)?$`),                      // created by …
	regexp.MustCompile(`^\[(?:signal [^\]]*|originating from goroutine \d+)\]:?$`),    // [signal SIGSEGV …]
	regexp.MustCompile(`^SIG[A-Z]+: [a-z ]+$`),                                        // SIGQUIT: quit
	regexp.MustCompile(`^PC=0x[0-9a-f]+ m=\d+ sigcode=\d+(?: addr=0x[0-9a-f]+)?$`),    // PC=0x… m=0 sigcode=0
	regexp.MustCompile(`^signal arrived during [a-z ]+$`),                             // … cgo execution
	regexp.MustCompile(`^[a-z0-9]+ +0x[0-9a-f]+$`),                                    // registers
	regexp.MustCompile(`^(?:runtime stack:|\.\.\.additional frames elided\.\.\.)$`),   // runtime's own
	regexp.MustCompile(`^goroutine \d+ gp=0x[0-9a-f]+ m=\S+ mp=0x[0-9a-f]+ \[.*\]:$`), // with GOTRACEBACK=system
}

// crashMessage is a panic or fatal error line; its message is kept only
// when the runtime wrote it (runtime errors and fatal errors carry
// indexes, sizes and types, never data), else it is <removed>.
var crashMessage = regexp.MustCompile(`^(\t?(?:panic|fatal error): )(.*?)((?: \[recovered(?:, repanicked)?\])?)$`)

var runtimeMessage = regexp.MustCompile(`^(?:runtime error: [\w .:\[\]-]+|interface conversion: [\w .*,{}()\[\]-]+|` +
	`concurrent map [a-z ]+|all goroutines are asleep - deadlock!|out of memory|stack overflow|` +
	`unexpected signal during runtime execution|sync: [a-z ]+|send on closed channel|close of (?:closed|nil) channel)$`)

// redactCrash keeps the shape of a crash log and the runtime's own
// messages, and replaces every other line (a panic's message can be any
// value, a window title included) with <removed>, one per run of them.
func redactCrash(s string) string {
	var out []string
	removed := false
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if m := crashMessage.FindStringSubmatch(line); m != nil {
			msg := m[2]
			if !runtimeMessage.MatchString(msg) {
				msg = "<removed>"
			}
			out, removed = append(out, m[1]+msg+m[3]), false
			continue
		}
		keep := strings.TrimSpace(line) == ""
		for _, re := range crashKeep {
			if keep {
				break
			}
			keep = re.MatchString(line)
		}
		switch {
		case keep:
			out, removed = append(out, line), false
		case !removed:
			out, removed = append(out, "<removed>"), true
		}
	}
	if len(out) > bugReportLogLines {
		out = out[len(out)-bugReportLogLines:]
	}
	return strings.Join(out, "\n")
}

// bugReportRuns is trackd's runs of the last two weeks: when it ran,
// stopped and why, never what it recorded.
func bugReportRuns(env Env) ([]byte, error) {
	st, err := openStore(env)
	if err != nil {
		return nil, err
	}
	defer st.Close()
	now := env.Now()
	runs, err := st.Runs(now.Add(-bugReportRunDays*24*time.Hour).UnixMilli(), now.UnixMilli())
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(runs, "", "  ")
}
