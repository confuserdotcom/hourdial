package numfmt

import (
	"strconv"
	"testing"
)

func TestParse(t *testing.T) {
	for in, want := range map[string]Style{"1234.5": Plain, "1,234.5": Comma, "1.234,5": Dot, "1 234,5": Space, "1234,5": PlainComma, " 1.234,5 ": Dot} {
		got, ok := Parse(in)
		if !ok || got != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", ".", ",", "1'234.5", "1.234.5", "en", "system"} {
		if _, ok := Parse(bad); ok {
			t.Errorf("Parse(%q) accepted", bad)
		}
	}
	for _, s := range []Style{Plain, Comma, Dot, Space, PlainComma} {
		if back, ok := Parse(s.String()); !ok || back != s {
			t.Errorf("%v does not round-trip through String", s)
		}
	}
	var zero Style
	if zero != Plain {
		t.Fatal("the zero Style must be Plain, today's output")
	}
}

func TestDecimal(t *testing.T) {
	cases := []struct {
		in                                string
		plain, comma, dot, space, plainCm string
	}{
		{"0", "0", "0", "0", "0", "0"},
		{"5.5", "5.5", "5.5", "5,5", "5,5", "5,5"},
		{"999.99", "999.99", "999.99", "999,99", "999,99", "999,99"},
		{"1000", "1000", "1,000", "1.000", "1 000", "1000"},
		{"1234.50", "1234.50", "1,234.50", "1.234,50", "1 234,50", "1234,50"},
		{"-1234.50", "-1234.50", "-1,234.50", "-1.234,50", "-1 234,50", "-1234,50"},
		{"1234567.125", "1234567.125", "1,234,567.125", "1.234.567,125", "1 234 567,125", "1234567,125"},
		{"123456", "123456", "123,456", "123.456", "123 456", "123456"},
	}
	for _, c := range cases {
		for s, want := range map[Style]string{Plain: c.plain, Comma: c.comma, Dot: c.dot, Space: c.space, PlainComma: c.plainCm} {
			if got := s.Decimal(c.in); got != want {
				t.Errorf("%v.Decimal(%q) = %q, want %q", s, c.in, got, want)
			}
		}
	}
}

func TestPlainIsToday(t *testing.T) {
	for _, f := range []float64{0, 0.05, 1.5, 12.25, 1234.5, 98765.4321} {
		for prec := 0; prec <= 2; prec++ {
			if got, want := Plain.Fixed(f, prec), strconv.FormatFloat(f, 'f', prec, 64); got != want {
				t.Errorf("Plain.Fixed(%v, %d) = %q, want %q", f, prec, got, want)
			}
		}
	}
}

func TestFixed(t *testing.T) {
	if got := Dot.Fixed(2.5, 1); got != "2,5" {
		t.Errorf("Dot.Fixed = %q", got)
	}
	if got := Space.Fixed(12345.678, 2); got != "12 345,68" {
		t.Errorf("Space.Fixed = %q", got)
	}
}

func TestMinor(t *testing.T) {
	cases := []struct {
		v     int64
		units int
		s     Style
		want  string
	}{
		{31833, 2, Plain, "318.33"},
		{5, 2, Plain, "0.05"},
		{-5, 2, Plain, "-0.05"},
		{123450, 2, Dot, "1.234,50"},
		{123450, 2, Space, "1 234,50"},
		{1234567, 0, Comma, "1,234,567"}, // JPY: no decimals
		{1234567, 0, Plain, "1234567"},
		{1234567, 3, Comma, "1,234.567"}, // KWD: three
		{7, 3, Dot, "0,007"},
	}
	for _, c := range cases {
		if got := c.s.Minor(c.v, c.units); got != c.want {
			t.Errorf("%v.Minor(%d, %d) = %q, want %q", c.s, c.v, c.units, got, c.want)
		}
	}
}

func TestMoney(t *testing.T) {
	cases := []struct {
		v      int64
		units  int
		s      Style
		mark   string
		before bool
		want   string
	}{
		{123450, 2, Plain, "EUR", false, "1234.50 EUR"}, // what invoices printed so far
		{123450, 2, Plain, "", false, "1234.50"},
		{123450, 2, Plain, "", true, "1234.50"},
		{123450, 2, Dot, "€", true, "€1.234,50"},
		{123450, 2, Comma, "USD", false, "1,234.50 USD"},
		{123450, 2, Space, "MAD", true, "MAD 1 234,50"},
		{123450, 2, Space, "€", false, "1 234,50 €"},
		{123450, 2, Comma, "US$", true, "US$1,234.50"},
		{150000, 0, Comma, "¥", true, "¥150,000"},
		{1234567, 3, Comma, "KWD", true, "KWD 1,234.567"},
		{-500, 2, Dot, "€", true, "-€5,00"},
	}
	for _, c := range cases {
		if got := c.s.Money(c.v, c.units, c.mark, c.before); got != c.want {
			t.Errorf("%v.Money(%d, %q, before %v) = %q, want %q", c.s, c.v, c.mark, c.before, got, c.want)
		}
	}
}
