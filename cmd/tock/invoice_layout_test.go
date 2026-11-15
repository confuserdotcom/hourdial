package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A new invoice takes the date, numbers and currency layout from config;
// one issued before the change shows and prints exactly as it did.
func TestInvoiceLayoutCLI(t *testing.T) {
	te := invoiceEnv(t)
	if code, _ := te.run(t, "", append([]string{"invoice", "issue"}, day...)...); code != 0 {
		t.Fatal(te.errb.String())
	}
	_, oldShow := te.run(t, "", "invoice", "show", "2026-0001")
	_, oldList := te.run(t, "", "invoice", "list")
	oldPDF := te.mustShowPDF(t)

	cfg := "[display]\ndate = \"mdy\"\nnumbers = \"1,234.5\"\n" +
		"[invoice]\nname = \"Ada Lovelace\"\ncurrency = \"€\"\ncurrency_position = \"before\"\nnumbers = \"1.234,5\"\n"
	if err := os.WriteFile(filepath.Join(te.ConfigDir, "config.toml"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "draft.pdf")
	if code, out := te.run(t, "", append(append([]string{"invoice"}, day...), "-o", path)...); code != 0 ||
		!strings.Contains(out, "tock, Oct 10, 2026 to Oct 10, 2026: 1 days, 1,67 h, total €100,00\n") {
		t.Fatalf("preview: %d %q %s", code, out, te.errb)
	}
	code, out := te.run(t, "", append([]string{"invoice", "issue", "--allow-overlap"}, day...)...)
	if code != 0 || !strings.Contains(out, "issued 2026-0002: €100,00\n") {
		t.Fatalf("issue: %d %q %s", code, out, te.errb)
	}
	code, out = te.run(t, "", "invoice", "show", "2026-0002")
	for _, want := range []string{"tock, Oct 10, 2026 to Oct 10, 2026", "Oct 10, 2026  1,67 h  €100,00", "rate €60,00/h", "total €100,00"} {
		if code != 0 || !strings.Contains(out, want) {
			t.Errorf("show lacks %q:\n%s", want, out)
		}
	}
	pdf, err := os.ReadFile(filepath.Join(filepath.Dir(te.dbPath()), "invoices", "2026-0002.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"(Sat Oct 10, 2026) Tj", "(1,67) Tj", "(60,00) Tj", "(\x80100,00) Tj"} {
		if !strings.Contains(string(pdf), want) {
			t.Errorf("pdf lacks %q", want)
		}
	}

	// The first invoice, issued in the old layout, has not moved.
	if _, again := te.run(t, "", "invoice", "show", "2026-0001"); again != oldShow {
		t.Errorf("show of the old invoice changed:\n%s\nwas\n%s", again, oldShow)
	}
	if te.mustShowPDF(t) != oldPDF {
		t.Error("pdf of the old invoice changed")
	}
	_, list := te.run(t, "", "invoice", "list")
	if !strings.HasPrefix(list, strings.TrimSuffix(oldList, "\n")) || !strings.Contains(list, "Oct 10, 2026 to Oct 10, 2026  €100,00") {
		t.Errorf("list:\n%s\nwas\n%s", list, oldList)
	}
	if code, out := te.run(t, "", "invoice", "verify"); code != 0 {
		t.Errorf("verify: %d %s %s", code, out, te.errb)
	}
}
