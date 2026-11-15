package web

import (
	"strings"
	"testing"
	"time"

	"tock/internal/apptest"
)

// The invoice dialog shows the total as the CLI and the PDF write it.
func TestInvoiceTotalText(t *testing.T) {
	f := newFixture(t)
	f.withIssuer(t)
	if code, _, b := f.do("POST", "/api/library/project", testToken, map[string]any{"op": "set", "name": "tock", "rate": 60}); code != 200 {
		t.Fatalf("set rate: %d %s", code, b)
	}
	day := time.UnixMilli(f.library().NowMS).In(apptest.Zone).Format(dateLayout)
	q := "/api/invoice?project=tock&preview=1&from=" + day + "&to=" + day
	if code, _, b := f.do("GET", q, testToken, nil); code != 200 || !strings.Contains(string(b), `"totalText":"100.00"`) {
		t.Fatalf("default preview: %d %s", code, b)
	}

	f.e.WriteConfig(t, "[invoice]\nname = \"Ada Lovelace\"\ncurrency = \"€\"\ncurrency_position = \"before\"\nnumbers = \"1.234,5\"\n")
	if _, err := f.a.Reload(); err != nil {
		t.Fatal(err)
	}
	if code, _, b := f.do("GET", q, testToken, nil); code != 200 || !strings.Contains(string(b), `"totalText":"€100,00"`) {
		t.Fatalf("preview: %d %s", code, b)
	}
	body := map[string]any{"project": "tock", "from": day, "to": day}
	if code, _, b := f.do("POST", "/api/invoice/issue", testToken, body); code != 200 || !strings.Contains(string(b), `"totalText":"€100,00"`) {
		t.Fatalf("issue: %d %s", code, b)
	}
}
