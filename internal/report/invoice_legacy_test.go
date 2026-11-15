package report

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"tock/internal/datefmt"
	"tock/internal/numfmt"
	"tock/internal/pdf"
)

// TestInvoicePDFUnchanged pins the bytes of an invoice and a draft in the
// layout invoices had before the date, numbers and currency_position
// settings: the zero values of those fields must keep producing them, so
// a new invoice with the default settings looks like every earlier one.
func TestInvoicePDFUnchanged(t *testing.T) {
	sum := func(inv Invoice) string {
		var b bytes.Buffer
		if err := WriteInvoicePDF(&b, inv); err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b.Bytes())
		return hex.EncodeToString(h[:])
	}
	d := testInvoice()
	d.Draft, d.Number = true, ""
	for _, c := range []struct {
		name, got, want string
	}{
		{"issued", sum(testInvoice()), "9b2f8c890a8f858e10fa736c614ffecc09836e270457efc09b8e9591a6196d3d"},
		{"draft", sum(d), "4d1a2ffb51c3a9f0c50d52959554acf2c47c1a597338870ddd1def60a8492504"},
	} {
		if c.got != c.want {
			t.Errorf("%s pdf sha256 = %s, want %s", c.name, c.got, c.want)
		}
	}
}

// An invoice issued with a date, numbers and currency placement prints
// every date, amount, hour figure and percentage in them.
func TestInvoicePDFLayout(t *testing.T) {
	inv := testInvoice()
	inv.Currency, inv.CurrencyBefore = "€", true
	inv.Dates, inv.DatesSet, inv.Nums = datefmt.DMY, true, numfmt.Dot
	inv.RateMinor, inv.SubtotalMinor, inv.TotalMinor = 123450, 1234567, 1481480
	inv.Lines[0].AmountMinor = 1234567
	inv.TaxBP = 550
	var b bytes.Buffer
	if err := WriteInvoicePDF(&b, inv); err != nil {
		t.Fatal(err)
	}
	s := b.String()
	for _, want := range []string{"(1 Nov 2026) Tj", "(5 Oct 2026 – 31 Oct 2026) Tj", "(Mon 5 Oct 2026) Tj", "(3,33) Tj",
		"(1.234,50) Tj", "(€1.234,50 / hour) Tj", "(€12.345,67) Tj", "(VAT 5,5%) Tj", "(€14.814,80) Tj"} {
		if !strings.Contains(s, string(pdf.Encode(want))) {
			t.Errorf("pdf lacks %q", want)
		}
	}

	jpy := testInvoice()
	jpy.Currency, jpy.MinorUnits, jpy.Nums, jpy.Dates, jpy.DatesSet = "JPY", 0, numfmt.Comma, datefmt.MDY, true
	jpy.RateMinor, jpy.Lines[0].AmountMinor, jpy.TotalMinor = 8000, 26667, 1234567
	b.Reset()
	if err := WriteInvoicePDF(&b, jpy); err != nil {
		t.Fatal(err)
	}
	s = b.String()
	for _, want := range []string{"(Nov 1, 2026) Tj", "(Mon Oct 5, 2026) Tj", "(8,000) Tj", "(26,667 JPY) Tj", "(1,234,567 JPY) Tj"} {
		if !strings.Contains(s, want) {
			t.Errorf("JPY pdf lacks %q", want)
		}
	}
}
