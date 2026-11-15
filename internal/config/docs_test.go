package config

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The guide documents config.toml: its reference block under Files lists
// the sections and keys, and other sections show a few more. Every key the
// guide shows must be one Parse reads, and every key Parse reads must be
// in the guide.

const guide = "../../docs/guide.md"

// codeKeys is every key fileTOML reads, as "section.key", with tables in an
// array as "section.table.key" ([[invoice.client]] name).
func codeKeys() map[string]bool {
	keys := map[string]bool{}
	var walk func(prefix string, t reflect.Type)
	walk = func(prefix string, t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := f.Tag.Get("toml")
			ft := f.Type
			for ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				walk(prefix+name+".", ft)
				continue
			}
			keys[prefix+name] = true
		}
	}
	walk("", reflect.TypeOf(fileTOML{}))
	return keys
}

// tomlBlock is one ```toml block of the guide.
type tomlBlock struct {
	line int // of the fence
	text string
}

func tomlBlocks(t *testing.T) (all []tomlBlock, ref tomlBlock) {
	t.Helper()
	b, err := os.ReadFile(guide)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	afterRef := false
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "`config.toml` sections") {
			afterRef = true
		}
		if strings.TrimSpace(lines[i]) != "```toml" {
			continue
		}
		blk := tomlBlock{line: i + 1}
		var body []string
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "```"; i++ {
			body = append(body, lines[i])
		}
		blk.text = strings.Join(body, "\n")
		all = append(all, blk)
		if afterRef && ref.text == "" {
			ref = blk
		}
	}
	if ref.text == "" {
		t.Fatal("guide.md: no ```toml block after \"`config.toml` sections\"")
	}
	return all, ref
}

var (
	tableLine = regexp.MustCompile(`^\s*\[\[?([a-z_.]+)\]\]?`)
	keyLine   = regexp.MustCompile(`^\s*([a-z_]+)\s*=`)
	word      = regexp.MustCompile(`[a-z_]+`)
)

// rulesTables are rules.toml's tables: a block with one of them is not
// config.toml.
var rulesTables = map[string]bool{"category": true, "project": true, "goal": true, "rule": true, "calls": true, "renamed": true}

// blockKeys returns the keys a block sets, as "section.key" with the line
// they are on, and every word of its comments by section ("section.word").
func blockKeys(b tomlBlock) (set map[string]int, named map[string]bool, rules bool) {
	set, named = map[string]int{}, map[string]bool{}
	sect := ""
	for i, l := range strings.Split(b.text, "\n") {
		code, comment, _ := strings.Cut(l, "#")
		if m := tableLine.FindStringSubmatch(code); m != nil {
			sect = m[1]
			rules = rules || rulesTables[sect]
		} else if m := keyLine.FindStringSubmatch(code); m != nil {
			set[sect+"."+m[1]] = b.line + 1 + i
			named[sect+"."+m[1]] = true
		}
		for _, w := range word.FindAllString(comment, -1) {
			named[sect+"."+w] = true
		}
	}
	return set, named, rules
}

func TestGuideConfigKeys(t *testing.T) {
	code := codeKeys()
	all, ref := tomlBlocks(t)
	named := map[string]bool{}
	for _, b := range all {
		set, words, rules := blockKeys(b)
		if rules {
			continue // rules.toml, not config.toml
		}
		for k, line := range set {
			if !code[k] {
				t.Errorf("guide.md:%d: %s is not a config.toml key", line, k)
			}
		}
		for k := range words {
			named[k] = true
		}
	}
	// Every key in the reference block or, failing that, in the section
	// that explains it.
	_, refNamed, _ := blockKeys(ref)
	var missing []string
	for k := range code {
		if !refNamed[k] && !named[k] {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	for _, k := range missing {
		sect, key, _ := strings.Cut(k, ".")
		t.Errorf("config.toml key %s (under [%s]) is in no ```toml block of docs/guide.md; add it to the reference block under Files (line %d)", key, sect, ref.line)
	}
}
