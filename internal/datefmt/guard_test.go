package datefmt

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// layoutRE finds a hard-coded day-and-month Go layout: "2 Jan", "Jan 2",
// "_2 Jan" and the like.
var layoutRE = regexp.MustCompile(`"[^"]*(\b_?2 Jan|Jan _?2\b)[^"]*"`)

// keepLayout are the files allowed to write such a layout themselves.
var keepLayout = map[string]string{
	"internal/datefmt/datefmt.go": "the formatter itself",
}

// TestNoHardCodedDate fails when a user-facing file writes a date in a
// fixed order instead of through datefmt, so [display] date is not
// skipped by accident.
func TestNoHardCodedDate(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, dir := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(p string, e fs.DirEntry, err error) error {
			if err != nil || e.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			if _, ok := keepLayout[rel]; ok {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for i, line := range strings.Split(string(b), "\n") {
				if layoutRE.MatchString(line) && !strings.HasPrefix(strings.TrimSpace(line), "//") {
					t.Errorf("%s:%d writes a date in a fixed order; use datefmt (or list the file in keepLayout): %s", rel, i+1, strings.TrimSpace(line))
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
