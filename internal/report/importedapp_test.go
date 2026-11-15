package report

import (
	"testing"
	"time"

	"tock/internal/classify"
)

// imported is a view of ActivityWatch history: a remote id and the import
// machine's prefix on its app.
func imported(from, to int, app string) classify.SpanView {
	return classify.SpanView{ID: -(classify.RemoteIDBase + int64(from) + 1), Start: int64(from) * 60_000, End: int64(to) * 60_000,
		BundleID: app, App: classify.ImportMachine + " · " + app}
}

// Imported Xcode and local Xcode are one app: the reports that rank apps
// list it once, under its own name.
func TestImportedAppMergesWithLocal(t *testing.T) {
	code := classify.Label{Category: "Code", Kind: classify.KindDeep}
	vs := []classify.SpanView{bv(0, 30, "Xcode"), imported(30, 50, "Xcode"), bv(50, 60, "term")}
	ls := []classify.Label{code, code, code}

	from := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	a := BuildTop(from, from.AddDate(0, 0, 7), vs, ls).Sections[TopApp]
	if len(a) != 2 || a[0].Name != "Xcode" || a[0].MS != 50*60_000 {
		t.Fatalf("top apps = %+v, want Xcode 50m once", a)
	}

	bs := Blocks(vs, ls)
	if len(bs) != 1 || len(bs[0].Top) != 2 || bs[0].Top[0] != (NameMS{"Xcode", 50 * 60_000}) {
		t.Fatalf("block apps = %+v, want Xcode 50m once", bs)
	}

	if got := AppName(imported(0, 1, "Xcode")); got != "Xcode" {
		t.Fatalf("week app name = %q, want Xcode", got)
	}
	// Another machine's prefix stays: it says where the time came from.
	ryuk := bv(0, 1, "kitty")
	ryuk.App = "ryuk · kitty"
	if got := AppName(ryuk); got != "ryuk · kitty" {
		t.Fatalf("remote app name = %q", got)
	}
}
