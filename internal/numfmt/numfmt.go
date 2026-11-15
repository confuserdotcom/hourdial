// Package numfmt writes decimal numbers and money for people, as the
// [display] numbers setting says: "1234.5" (the default: a point and no
// grouping), "1,234.5", "1.234,5", "1 234,5" or "1234,5". It is a leaf
// package so pure packages can take a Style in their inputs.
//
// Exports, --json, CSV, logs and the text sent to a language model keep
// plain "1234.5" numbers and do not use it. Money is held in integer minor
// units; nothing here rounds through a float except Fixed, which is for
// figures that are not money.
package numfmt

import (
	"strconv"
	"strings"
	"unicode"
)

// Style is how a decimal number is written. The zero value is Plain.
type Style uint8

const (
	Plain      Style = iota // 1234.5
	Comma                   // 1,234.5
	Dot                     // 1.234,5
	Space                   // 1 234,5
	PlainComma              // 1234,5
)

var names = [...]string{Plain: "1234.5", Comma: "1,234.5", Dot: "1.234,5", Space: "1 234,5", PlainComma: "1234,5"}

// Parse reads a config value, written as the number 1234.5 would be:
// "1234.5", "1,234.5", "1.234,5", "1 234,5" or "1234,5".
func Parse(s string) (Style, bool) {
	s = strings.TrimSpace(s)
	for i, n := range names {
		if s == n {
			return Style(i), true
		}
	}
	return Plain, false
}

// String is the config value.
func (s Style) String() string {
	if int(s) < len(names) {
		return names[s]
	}
	return names[Plain]
}

// Names lists the accepted values, for error messages.
func Names() []string { return append([]string(nil), names[:]...) }

func (s Style) seps() (dec, group string) {
	switch s {
	case Comma:
		return ".", ","
	case Dot:
		return ",", "."
	case Space:
		return ",", " "
	case PlainComma:
		return ",", ""
	}
	return ".", ""
}

// Decimal rewrites a plain decimal string ("-1234.50", as strconv writes
// it) in the style: "-1.234,50" for Dot.
func (s Style) Decimal(n string) string {
	dec, group := s.seps()
	sign := ""
	if strings.HasPrefix(n, "-") {
		sign, n = "-", n[1:]
	}
	whole, frac, hasFrac := strings.Cut(n, ".")
	if group != "" && len(whole) > 3 {
		var b strings.Builder
		lead := len(whole) % 3
		if lead > 0 {
			b.WriteString(whole[:lead])
		}
		for i := lead; i < len(whole); i += 3 {
			if b.Len() > 0 {
				b.WriteString(group)
			}
			b.WriteString(whole[i : i+3])
		}
		whole = b.String()
	}
	if hasFrac {
		return sign + whole + dec + frac
	}
	return sign + whole
}

// Fixed is f with prec decimals, as strconv.FormatFloat(f, 'f', prec, 64)
// in the style. It is for durations in hours, rates of switches and the
// like, never for money.
func (s Style) Fixed(f float64, prec int) string {
	return s.Decimal(strconv.FormatFloat(f, 'f', prec, 64))
}

// Minor writes v minor units with units decimals (0 to 3): 123450 with 2
// is "1234.50" in Plain, "1.234,50" in Dot.
func (s Style) Minor(v int64, units int) string {
	sign := ""
	u := uint64(v)
	if v < 0 {
		sign, u = "-", uint64(-v)
	}
	d := strconv.FormatUint(u, 10)
	if units > 0 {
		if len(d) <= units {
			d = strings.Repeat("0", units-len(d)+1) + d
		}
		d = d[:len(d)-units] + "." + d[len(d)-units:]
	}
	return s.Decimal(sign + d)
}

// Money is Minor with a currency mark ("EUR", "€", "MAD"; "" = none).
// After the amount the mark follows a space: "1234.50 EUR". Before it, a
// mark that ends in a letter is followed by a space ("MAD 1 234,50") and
// a symbol is not ("€1.234,50"). A minus sign goes first: "-€5,00".
func (s Style) Money(v int64, units int, mark string, before bool) string {
	n := s.Minor(v, units)
	if mark == "" {
		return n
	}
	if !before {
		return n + " " + mark
	}
	sign := ""
	if strings.HasPrefix(n, "-") {
		sign, n = "-", n[1:]
	}
	r := []rune(mark)
	if unicode.IsLetter(r[len(r)-1]) {
		return sign + mark + " " + n
	}
	return sign + mark + n
}
