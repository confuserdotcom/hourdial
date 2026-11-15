package store

import (
	"bytes"
	"testing"
	"time"
)

// An invoice issued in the old layout has no "layout" in its canonical
// JSON, so it hashes as before; one issued with a layout keeps it, and
// every read returns it, also after recovery from the invoice files.
func TestInvoiceLayoutFrozen(t *testing.T) {
	path := dbPath(t)
	s := openAt(t, path)
	at := time.Date(2026, 11, 3, 10, 0, 0, 0, paris)

	plain := issue(t, s, IssueSpec{Draft: draftInvoice(at)})
	if bytes.Contains(plain.CanonicalJSON, []byte(`"layout"`)) {
		t.Fatalf("old-layout invoice carries a layout:\n%s", plain.CanonicalJSON)
	}

	d := draftInvoice(at.Add(time.Hour))
	d.DateStyle, d.NumberStyle, d.CurrencyPosition = "dmy", "1.234,5", "before"
	laid := issue(t, s, IssueSpec{Draft: d})
	if !bytes.Contains(laid.CanonicalJSON, []byte(`"layout":{"date":"dmy","numbers":"1.234,5","currency_position":"before"}`)) {
		t.Fatalf("canonical json:\n%s", laid.CanonicalJSON)
	}

	check := func(where string, inv Invoice, date, nums, pos string) {
		t.Helper()
		if inv.DateStyle != date || inv.NumberStyle != nums || inv.CurrencyPosition != pos {
			t.Errorf("%s %s: layout %q %q %q, want %q %q %q", where, inv.Number, inv.DateStyle, inv.NumberStyle, inv.CurrencyPosition, date, nums, pos)
		}
	}
	for _, n := range []string{plain.Number, laid.Number} {
		got, err := s.InvoiceByNumber(n)
		if err != nil {
			t.Fatal(err)
		}
		if n == plain.Number {
			check("by number", got, "", "", "")
		} else {
			check("by number", got, "dmy", "1.234,5", "before")
		}
	}
	list, err := s.Invoices(0)
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %v %v", list, err)
	}
	check("list", list[0], "", "", "")
	check("list", list[1], "dmy", "1.234,5", "before")

	// Recovery from the files into an empty database keeps it.
	dir := InvoiceDir(path)
	for _, inv := range []Invoice{plain, laid} {
		if err := WriteInvoiceFiles(dir, inv); err != nil {
			t.Fatal(err)
		}
	}
	other := openAt(t, dbPath(t))
	if _, err := other.RecoverInvoices(dir); err != nil {
		t.Fatal(err)
	}
	got, err := other.InvoiceByNumber(laid.Number)
	if err != nil {
		t.Fatal(err)
	}
	check("recovered", got, "dmy", "1.234,5", "before")
	if checks, err := other.VerifyInvoices(dir, ""); err != nil || !okAll(checks) {
		t.Fatalf("verify after recovery: %+v %v", checks, err)
	}
}
