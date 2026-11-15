package web

import "testing"

// /api/status says how the dashboard writes dates and decimals.
func TestDateNumbersStatus(t *testing.T) {
	f := newFixture(t)
	var st statusJSON
	f.get("/api/status", &st)
	if st.Date != "dmy" || st.Numbers != "1234.5" {
		t.Fatalf("defaults = %q %q, want dmy 1234.5", st.Date, st.Numbers)
	}
	f.e.WriteConfig(t, "[display]\ndate = \"mdy\"\nnumbers = \"1.234,5\"\n")
	st = statusJSON{}
	f.get("/api/status", &st)
	if st.Date != "mdy" || st.Numbers != "1.234,5" {
		t.Fatalf("status = %q %q, want mdy 1.234,5", st.Date, st.Numbers)
	}
}
