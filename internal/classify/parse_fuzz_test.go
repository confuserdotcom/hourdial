package classify

import "testing"

// Any rules file, however broken, is an error or a ruleset that labels a
// span; it never panics or hangs.
func FuzzParse(f *testing.F) {
	f.Add([]byte(baseRules), []byte(nil))
	f.Add([]byte(callRules), []byte(nil))
	f.Add([]byte(baseRules), []byte("[[rule]]\napp = \"x\"\ncategory = \"Work\"\n"))
	f.Add([]byte("[[rule]]\ntitle_regex = \"(\"\n"), []byte(nil))
	f.Add([]byte{0xff, '[', '['}, []byte{'='})
	f.Fuzz(func(t *testing.T, rules, learned []byte) {
		rs, err := Parse(rules, learned)
		if err != nil {
			return
		}
		title, url := "main.go - tock", "https://example.com/a"
		v := SpanView{ID: 1, Start: 0, End: 60_000, BundleID: "com.example", App: "Example", Title: &title, URL: &url}
		rs.LabelAll([]SpanView{v}, nil)
	})
}
