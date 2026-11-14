package classify

import "testing"

// TestRoleTemplatesOnStarter applies each role template to the starter rules,
// as a fresh tock setup does, and checks the result labels what it promises.
func TestRoleTemplatesOnStarter(t *testing.T) {
	str := func(s string) *string { return &s }
	chrome := func(url string) SpanView {
		return SpanView{BundleID: "com.google.Chrome", App: "Google Chrome", URL: str(url)}
	}
	cases := []struct {
		tpl  string
		v    SpanView
		want string
	}{
		{"developer", SpanView{BundleID: "com.jetbrains.goland", App: "GoLand"}, "Code"},
		{"developer", SpanView{BundleID: "dev.warp.Warp-Stable", App: "Warp"}, "Code"},
		{"developer", SpanView{App: "kitty"}, "Code"},
		{"developer", chrome("localhost:5173/app"), "Code"},
		{"developer", chrome("docs.rs/serde"), "Code"},
		{"developer", chrome("go.readthedocs.io/en/latest"), "Research"},
		{"designer", SpanView{BundleID: "com.figma.Desktop", App: "Figma"}, "Design"},
		{"designer", SpanView{App: "Inkscape"}, "Design"},
		{"designer", chrome("www.figma.com/design/abc/Home"), "Design"},
		{"designer", chrome("dribbble.com/shots/1"), "Research"},
		{"writer", SpanView{BundleID: "com.ulyssesapp.mac", App: "Ulysses"}, "Writing"},
		{"writer", SpanView{App: "ghostwriter"}, "Writing"},
		{"writer", chrome("wordpress.com/post/me.blog/12"), "Writing"},
		{"writer", SpanView{BundleID: "org.zotero.zotero", App: "Zotero"}, "Research"},
	}
	for _, c := range cases {
		rs, err := Parse(Starter, nil)
		if err != nil {
			t.Fatal(err)
		}
		p, err := PlanTemplates(rs, Starter, mustTemplates(t, c.tpl), 0, true)
		if err != nil {
			t.Fatalf("%s: %v", c.tpl, err)
		}
		got, err := Parse(p.Text, nil)
		if err != nil {
			t.Fatalf("%s: applied rules do not parse: %v", c.tpl, err)
		}
		if l := got.Label(c.v, nil); l.Category != c.want {
			t.Errorf("%s: %s %q %v = %q, want %q", c.tpl, c.v.BundleID, c.v.App, c.v.URL, l.Category, c.want)
		}
	}
}
