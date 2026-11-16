package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tock/internal/app"
	"tock/internal/report"
	"tock/internal/store"
)

type invoiceRowJSON struct {
	Number     string `json:"number"`
	Status     string `json:"status"` // "issued" or "void"
	Issued     string `json:"issued"` // dates as the invoice's own document writes them
	IssuedMS   int64  `json:"issuedMS"`
	Project    string `json:"project"`
	Client     string `json:"client"`
	From       string `json:"from"`
	To         string `json:"to"`
	Seconds    int64  `json:"seconds"`
	Total      string `json:"total"` // as the PDF and the CLI write it
	VoidReason string `json:"voidReason,omitempty"`
	Supersedes string `json:"supersedes,omitempty"`
	ReplacedBy string `json:"replacedBy,omitempty"`
}

type invoicesJSON struct {
	Invoices []invoiceRowJSON `json:"invoices"`
}

// invoices is tock invoice list: every issued invoice, newest first, or
// with ?year= that year's. Each one's dates are in its own zone and layout,
// as its PDF prints them.
func (s *Server) invoices(r *http.Request) (any, error) {
	year := 0
	if y := r.URL.Query().Get("year"); y != "" {
		n, err := strconv.Atoi(y)
		if err != nil || n < 1 {
			return nil, badRequest("year must be a number")
		}
		year = n
	}
	list, err := s.a.Invoices(year)
	if err != nil {
		return nil, err
	}
	here := s.now().Location()
	replacedBy := map[string]string{}
	for _, i := range list {
		if i.Supersedes != "" {
			replacedBy[i.Supersedes] = i.Number
		}
	}
	out := invoicesJSON{Invoices: make([]invoiceRowJSON, 0, len(list))}
	for n := len(list) - 1; n >= 0; n-- {
		i := list[n]
		loc := app.InvoiceLocation(i, here)
		day := func(ms int64) string { return app.InvoiceDate(i, time.UnixMilli(ms).In(loc)) }
		out.Invoices = append(out.Invoices, invoiceRowJSON{
			Number: i.Number, Status: i.Status, Issued: day(i.IssuedMS), IssuedMS: i.IssuedMS,
			Project: report.Safe(i.Project), Client: report.Safe(i.ClientName),
			From: day(i.PeriodFromMS), To: day(i.PeriodToMS - 1), Seconds: i.SecondsBilled,
			Total: app.InvoiceMoney(i, i.TotalMinor), VoidReason: report.Safe(i.VoidReason),
			Supersedes: i.Supersedes, ReplacedBy: replacedBy[i.Number],
		})
	}
	return out, nil
}

// invoicePDF is tock invoice show NUMBER --pdf: the stored bytes of
// /api/invoices/NUMBER.pdf, never a new rendering.
func (s *Server) invoicePDF(r *http.Request) (any, error) {
	number, ok := strings.CutSuffix(r.PathValue("file"), ".pdf")
	if !ok || number == "" {
		return nil, httpError{http.StatusNotFound, "ask for NUMBER.pdf"}
	}
	inv, err := s.a.InvoiceByNumber(number)
	if errors.Is(err, store.ErrNoInvoice) {
		return nil, httpError{http.StatusNotFound, "no invoice " + report.Safe(number)}
	}
	if err != nil {
		return nil, err
	}
	return file{report.Slug(inv.Number) + ".pdf", "application/pdf", inv.PDF}, nil
}
