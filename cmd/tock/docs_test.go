package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The user guide shows commands in code blocks. Each one must name a
// command tock has, with flags that command reads: a renamed or dropped
// flag otherwise lives on in the docs.
var guidePages = []string{"../../README.md", "../../CONTRIBUTING.md", "../../docs/guide.md", "../../docs/accessibility.md", "../../docs/stability.md", "../../docs/security-review.md", "../../docs/wezterm-title.md", "../../docs/architecture.md", "../../docs/index.md", "../../docs/verify.md", "../../docs/shell-hook.md", "../../docs/privacy.md"}

// flagCalls are the flag.FlagSet methods that declare a flag, and which
// argument holds its name.
var flagCalls = map[string]int{
	"String": 0, "Bool": 0, "Int": 0, "Int64": 0, "Uint": 0, "Float64": 0, "Duration": 0, "Func": 0, "BoolFunc": 0,
	"StringVar": 1, "BoolVar": 1, "IntVar": 1, "Int64Var": 1, "UintVar": 1, "Float64Var": 1, "DurationVar": 1, "Var": 1, "TextVar": 1,
}

var flagLit = regexp.MustCompile(`^--?[a-z][a-z0-9-]*$`)

// funcInfo is what one top-level function of package main declares and
// names: the flags it reads (FlagSet calls, or "--x" compared by hand) and
// every identifier it mentions, to follow into the functions it uses.
type funcInfo struct {
	flags map[string]bool
	names map[string]bool
}

// commandFlags maps each command in commands to every flag the code it
// reaches reads, without the leading dashes.
func commandFlags(t *testing.T) map[string]map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	funcs := map[string]*funcInfo{}
	var table *ast.CompositeLit
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil || d.Body == nil {
					continue
				}
				funcs[d.Name.Name] = scanFunc(d.Body)
			case *ast.GenDecl:
				for _, s := range d.Specs {
					vs, ok := s.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, n := range vs.Names {
						if i >= len(vs.Values) {
							continue
						}
						if n.Name == "commands" {
							// Not followed: every command would reach every flag.
							table, _ = vs.Values[i].(*ast.CompositeLit)
							continue
						}
						// A package-level value (a table of subcommands)
						// counts as a function that names what it holds.
						funcs[n.Name] = scanFunc(vs.Values[i])
					}
				}
			}
		}
	}
	if table == nil {
		t.Fatal("no commands table in package main")
	}
	out := map[string]map[string]bool{}
	for _, e := range table.Elts {
		kv := e.(*ast.KeyValueExpr)
		name, _ := strconv.Unquote(kv.Key.(*ast.BasicLit).Value)
		fn := kv.Value.(*ast.Ident).Name
		flags := map[string]bool{}
		seen := map[string]bool{}
		todo := []string{fn}
		for len(todo) > 0 {
			n := todo[len(todo)-1]
			todo = todo[:len(todo)-1]
			fi, ok := funcs[n]
			if !ok || seen[n] {
				continue
			}
			seen[n] = true
			for f := range fi.flags {
				flags[f] = true
			}
			for m := range fi.names {
				todo = append(todo, m)
			}
		}
		out[name] = flags
	}
	return out
}

func scanFunc(n ast.Node) *funcInfo {
	fi := &funcInfo{flags: map[string]bool{}, names: map[string]bool{}}
	lit := func(e ast.Expr) {
		if b, ok := e.(*ast.BasicLit); ok && b.Kind == token.STRING {
			if s, err := strconv.Unquote(b.Value); err == nil && flagLit.MatchString(s) {
				fi.flags[strings.TrimLeft(s, "-")] = true
			}
		}
	}
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Ident:
			fi.names[n.Name] = true
		case *ast.CallExpr:
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				break
			}
			if i, ok := flagCalls[sel.Sel.Name]; ok && i < len(n.Args) {
				if b, ok := n.Args[i].(*ast.BasicLit); ok && b.Kind == token.STRING {
					s, _ := strconv.Unquote(b.Value)
					fi.flags[s] = true
				}
			}
		case *ast.BinaryExpr:
			if n.Op == token.EQL || n.Op == token.NEQ {
				lit(n.X)
				lit(n.Y)
			}
		case *ast.CaseClause:
			for _, e := range n.List {
				lit(e)
			}
		}
		return true
	})
	return fi
}

var (
	fence   = regexp.MustCompile("^\\s*```")
	tockCmd = regexp.MustCompile(`(?:^\s*|\$ |\(|\| |; )tock ([a-z][a-z-]*)`)
	// inlineCmd is a command in prose: `tock away --skip 12:10`.
	inlineCmd = regexp.MustCompile("`tock ([a-z][a-z-]*)([^`]*)`")
	flagWord  = regexp.MustCompile(`(?:^|[\s\[(|])(--?[a-zA-Z][a-zA-Z0-9-]*)`)
)

// docCommand is one tock command line in a code block of a page.
type docCommand struct {
	where, name string
	flags       []string
}

// docCommands reads every "tock <command> ... --flag" in the code blocks of
// page, and in code spans of its prose. A line that names no tock command but starts with a space carries on
// the command above it (the wrapped lines of the usage list).
func docCommands(t *testing.T, page string) []docCommand {
	t.Helper()
	b, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	var out []docCommand
	in := false
	last := -1
	for i, line := range strings.Split(string(b), "\n") {
		if fence.MatchString(line) {
			in = !in
			last = -1
			continue
		}
		where := filepath.Base(page) + ":" + strconv.Itoa(i+1)
		if !in {
			for _, m := range inlineCmd.FindAllStringSubmatch(line, -1) {
				c := docCommand{where: where, name: m[1]}
				for _, f := range flagWord.FindAllStringSubmatch(m[2], -1) {
					c.flags = append(c.flags, f[1])
				}
				out = append(out, c)
			}
			continue
		}
		// A comment after the command is prose.
		code := line
		if j := strings.Index(code, " #"); j >= 0 {
			code = code[:j]
		}
		locs := tockCmd.FindAllStringSubmatchIndex(code, -1)
		if len(locs) == 0 {
			if last >= 0 && strings.HasPrefix(line, " ") {
				for _, m := range flagWord.FindAllStringSubmatch(code, -1) {
					out[last].flags = append(out[last].flags, m[1])
				}
			} else {
				last = -1
			}
			continue
		}
		for k, loc := range locs {
			end := len(code)
			if k+1 < len(locs) {
				end = locs[k+1][0]
			}
			c := docCommand{where: where, name: code[loc[2]:loc[3]]}
			for _, m := range flagWord.FindAllStringSubmatch(code[loc[3]:end], -1) {
				c.flags = append(c.flags, m[1])
			}
			out = append(out, c)
			last = len(out) - 1
		}
	}
	return out
}

func TestGuideCommandsExist(t *testing.T) {
	known := commandFlags(t)
	builtin := map[string]bool{"help": true, "version": true}
	n := 0
	for _, page := range guidePages {
		for _, c := range docCommands(t, page) {
			n++
			flags, ok := known[c.name]
			if !ok {
				if !builtin[c.name] {
					t.Errorf("%s: tock %s: no such command", c.where, c.name)
				}
				continue
			}
			for _, f := range c.flags {
				name := strings.TrimLeft(f, "-")
				if name == "h" || name == "help" {
					continue
				}
				if !flags[name] {
					have := make([]string, 0, len(flags))
					for k := range flags {
						have = append(have, k)
					}
					sort.Strings(have)
					t.Errorf("%s: tock %s has no flag %s (it reads: %s)", c.where, c.name, f, strings.Join(have, " "))
				}
			}
		}
	}
	if n < 100 {
		t.Fatalf("found %d tock commands in the guide, want 100 or more: is the parser still reading the code blocks?", n)
	}
}

// The guide's Commands section shows tock help all; it must be what tock
// help all prints, so a new or changed command reaches the guide.
func TestGuideCommandList(t *testing.T) {
	b, err := os.ReadFile("../../docs/guide.md")
	if err != nil {
		t.Fatal(err)
	}
	_, sec, ok := strings.Cut(string(b), "\n## Commands\n")
	if !ok {
		t.Fatal("guide.md has no ## Commands section")
	}
	_, block, ok := strings.Cut(sec, "```\n")
	if !ok {
		t.Fatal("guide.md: no code block under ## Commands")
	}
	block, _, _ = strings.Cut(block, "```\n")
	// usage without its first line and the notes after the list, two spaces
	// less indented.
	_, body, _ := strings.Cut(usage, "\n\n")
	body, _, _ = strings.Cut(body, "\nNames match")
	var want strings.Builder
	for _, l := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		want.WriteString(strings.TrimPrefix(l, "  ") + "\n")
	}
	if block != want.String() {
		gl, wl := strings.Split(block, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(gl) || i < len(wl); i++ {
			var g, w string
			if i < len(gl) {
				g = gl[i]
			}
			if i < len(wl) {
				w = wl[i]
			}
			if g != w {
				t.Fatalf("the Commands block in docs/guide.md differs from tock help all at its line %d:\nguide: %q\nhelp:  %q\npaste tock help all into it, two spaces less indented", i+1, g, w)
			}
		}
	}
}
