package web

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"tock/internal/apptest"
)

// The dashboard lists what tock invoice list prints, and hands out the
// stored PDF of tock invoice show --pdf.
func TestInvoicesListAndPDF(t *testing.T) {
	f := newFixture(t)
	f.withIssuer(t)
	var empty invoicesJSON
	f.get("/api/invoices", &empty)
	if empty.Invoices == nil || len(empty.Invoices) != 0 {
		t.Fatalf("before any: %+v", empty)
	}
	if code, _, b := f.do("POST", "/api/library/project", testToken, map[string]any{"op": "set", "name": "tock", "rate": 60, "client": "Acme"}); code != 200 {
		t.Fatalf("set rate: %d %s", code, b)
	}
	day := time.UnixMilli(f.library().NowMS).In(apptest.Zone).Format(dateLayout)
	body := map[string]any{"project": "tock", "from": day, "to": day}
	if code, _, b := f.do("POST", "/api/invoice/issue", testToken, body); code != 200 {
		t.Fatalf("issue: %d %s", code, b)
	}
	body["allowOverlap"] = true
	if code, _, b := f.do("POST", "/api/invoice/issue", testToken, body); code != 200 {
		t.Fatalf("issue again: %d %s", code, b)
	}
	if _, err := f.a.VoidInvoice("2026-0001", "sent twice"); err != nil {
		t.Fatal(err)
	}
	var got invoicesJSON
	f.get("/api/invoices", &got)
	if len(got.Invoices) != 2 {
		t.Fatalf("list: %+v", got)
	}
	// Newest first, as the list reads on a screen.
	a, b := got.Invoices[0], got.Invoices[1]
	if a.Number != "2026-0002" || a.Status != "issued" || a.Project != "tock" || a.From != day || a.To != day || a.Issued != day ||
		a.Total != "100.00" || a.Seconds != 6000 || a.VoidReason != "" {
		t.Fatalf("first: %+v", a)
	}
	if b.Number != "2026-0001" || b.Status != "void" || b.VoidReason != "sent twice" {
		t.Fatalf("second: %+v", b)
	}
	var none invoicesJSON
	f.get("/api/invoices?year=2025", &none)
	if len(none.Invoices) != 0 {
		t.Fatalf("another year: %+v", none)
	}
	if code, _, _ := f.do("GET", "/api/invoices?year=x", testToken, nil); code != 400 {
		t.Fatalf("bad year: %d", code)
	}

	stored, err := f.a.InvoiceByNumber("2026-0002")
	if err != nil {
		t.Fatal(err)
	}
	code, h, pdf := f.do("GET", "/api/invoices/2026-0002.pdf", testToken, nil)
	if code != 200 || h.Get("Content-Type") != "application/pdf" || !bytes.Equal(pdf, stored.PDF) || !strings.Contains(h.Get("Content-Disposition"), "2026-0002.pdf") {
		t.Fatalf("pdf: %d %v, %d bytes of %d", code, h, len(pdf), len(stored.PDF))
	}
	for path, want := range map[string]int{"/api/invoices/2026-0099.pdf": 404, "/api/invoices/2026-0002": 404, "/api/invoices/2026-0002.pdf?x": 200} {
		if code, _, _ := f.do("GET", path, testToken, nil); code != want {
			t.Errorf("%s: %d, want %d", path, code, want)
		}
	}
	if code, _, _ := f.do("GET", "/api/invoices/2026-0002.pdf", "wrong", nil); code == 200 {
		t.Fatal("pdf without the token")
	}
	if code, _, _ := f.do("POST", "/api/invoices", testToken, nil); code == 200 {
		t.Fatal("POST listed")
	}
}
