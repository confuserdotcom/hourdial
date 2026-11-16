package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicRejectsPaths(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"../x.json", "a/b.json", `a\b.json`, "..", "", "."} {
		if err := writeFileAtomic(filepath.Join(dir, "inv"), name, []byte("x")); err == nil {
			t.Errorf("writeFileAtomic(%q) = nil, want error", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "x.json")); !os.IsNotExist(err) {
		t.Fatalf("file escaped the invoices dir: %v", err)
	}
	if err := writeFileAtomic(dir, "INV-2026-001.json", []byte("x")); err != nil {
		t.Fatalf("plain name: %v", err)
	}
}
