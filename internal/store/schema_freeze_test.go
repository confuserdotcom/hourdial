package store

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"strings"
	"testing"
)

// shippedMigrations pins every migration by name and a hash of its text
// (line endings ignored). A database that ran one of these must never meet
// a different version of it, so a shipped migration is never edited.
//
// Schema changes are batched: at most one new migration per release, and
// adding it means adding its line here in the same commit, on purpose.
// Hash a new file with: tr -d '\r' < FILE | shasum -a 256 | cut -c1-16
var shippedMigrations = []struct{ name, hash string }{
	{"0001_init.sql", "b2a31e57e5f338ec"},
	{"0002_classify.sql", "f4f26cc54ac41b57"},
	{"0003_session.sql", "289ff3fd930b4a0e"},
	{"0004_suggestion.sql", "47f72cdec6332504"},
	{"0005_call.sql", "7f2fc2f58c82812a"},
	{"0006_away.sql", "b4bce685224e4c67"},
	{"0007_plan.sql", "21c156b7bd05db33"},
	{"0008_event.sql", "b85851d6c38eb632"},
	{"0009_notice.sql", "c35040a591f70f64"},
	{"0010_remote.sql", "3585347214766fb8"},
	{"0011_edit.sql", "147bc4aeaf0425b9"},
	{"0012_summary.sql", "42ef0d573597386a"},
	{"0013_kept.sql", "62f38ee11ae3c72d"},
	{"0014_session_quality.sql", "4d5d2926bda585d0"},
	{"0015_session_urges.sql", "942249a3444eafe7"},
	{"0016_slot.sql", "546c6b446f882606"},
	{"0017_event_mine.sql", "43a22ba8b15cde57"},
	{"0018_recommend.sql", "cce2d076aa97b1ca"},
	{"0019_ocr.sql", "0a08c7e7fa95f754"},
	{"0020_auto_label.sql", "7fdedf9ae65d8390"},
	{"0021_derived.sql", "f79f5a2ab1306c36"},
	{"0022_wake.sql", "74d2e27d1d4849d2"},
	{"0023_session_goal.sql", "6baa4c75287fe03f"},
	{"0024_trackd_run.sql", "af3ea1ac524d1298"},
	{"0025_schema_meta.sql", "0f4d51a5dcd13102"},
	{"0026_invoices.sql", "c0b2b62e2f3b3c31"},
	{"0027_range_indexes.sql", "fe453838dc532d87"},
}

// frozenCompat is the compat level of the 1.x line. Raising CompatLevel
// locks older 1.x binaries out of the database: that is a major release.
const frozenCompat = 1

const freezeRules = `
The schema is frozen and changes are batched (CONTRIBUTING.md, docs/stability.md):
  - never edit a shipped migration; fix forward with a new one;
  - at most one new migration per release, pinned in shippedMigrations
    in internal/store/schema_freeze_test.go in the same commit;
  - never raise CompatLevel within 1.x.`

func migrationHash(t *testing.T, name string) string {
	t.Helper()
	b, err := migrationFS.ReadFile("migrations/" + name)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(strings.ReplaceAll(string(b), "\r", "")))
	return hex.EncodeToString(sum[:])[:16]
}

func TestSchemaFreeze(t *testing.T) {
	files, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	pinned := map[string]string{}
	for _, m := range shippedMigrations {
		pinned[m.name] = m.hash
	}
	for _, f := range files {
		name := strings.TrimPrefix(f, "migrations/")
		want, ok := pinned[name]
		switch {
		case !ok:
			t.Errorf("migration %s is not pinned: add {%q, %q} to shippedMigrations on purpose.%s",
				name, name, migrationHash(t, name), freezeRules)
		case migrationHash(t, name) != want:
			t.Errorf("shipped migration %s was edited (hash %s, pinned %s).%s",
				name, migrationHash(t, name), want, freezeRules)
		}
		delete(pinned, name)
	}
	for name := range pinned {
		t.Errorf("shipped migration %s is gone.%s", name, freezeRules)
	}
	if CompatLevel != frozenCompat {
		t.Errorf("CompatLevel = %d, the 1.x line is frozen at %d.%s", CompatLevel, frozenCompat, freezeRules)
	}
}
