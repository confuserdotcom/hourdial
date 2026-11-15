package config

import (
	"strings"
	"testing"

	"tock/internal/datefmt"
	"tock/internal/numfmt"
)

func TestInvoiceLayout(t *testing.T) {
	c, err := Load(write(t, ""))
	if err != nil {
		t.Fatal(err)
	}
	l := c.InvoiceLayout()
	if l.DatesSet || l.Nums != numfmt.Plain || l.CurrencyBefore {
		t.Fatalf("default layout = %+v, want the layout invoices always had", l)
	}

	// [display] is the fallback; [invoice] wins over it.
	c, err = Load(write(t, "[display]\ndate = \"mdy\"\nnumbers = \"1,234.5\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if l := c.InvoiceLayout(); !l.DatesSet || l.Dates != datefmt.MDY || l.Nums != numfmt.Comma || l.CurrencyBefore {
		t.Fatalf("from [display]: %+v", l)
	}
	c, err = Load(write(t, "[display]\ndate = \"mdy\"\nnumbers = \"1,234.5\"\n[invoice]\ndate = \"dmy\"\nnumbers = \"1 234,5\"\ncurrency_position = \"before\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if l := c.InvoiceLayout(); !l.DatesSet || l.Dates != datefmt.DMY || l.Nums != numfmt.Space || !l.CurrencyBefore {
		t.Fatalf("[invoice] over [display]: %+v", l)
	}
	if c.Invoice.Date != "dmy" || c.Invoice.Numbers != "1 234,5" || c.Invoice.CurrencyPosition != "before" {
		t.Fatalf("invoice = %+v", c.Invoice)
	}

	for _, bad := range []string{"date = \"us\"", "numbers = \"1'234\"", "currency_position = \"left\""} {
		key, _, _ := strings.Cut(bad, " ")
		if _, err := Load(write(t, "[invoice]\n"+bad+"\n")); err == nil || !strings.Contains(err.Error(), key) {
			t.Errorf("%s: err = %v", bad, err)
		}
	}
}
