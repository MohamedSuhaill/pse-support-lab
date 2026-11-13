// Package validate implements the portal validation rules (AE-BR-xx).
package validate

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strconv"

	"invoicely/api/internal/metrics"
	"invoicely/api/internal/sheet"
)

// Error is one validation finding, shown to the customer.
type Error struct {
	Rule    string `json:"rule"`
	Field   string `json:"field"`
	Message string `json:"message"`
	Row     int    `json:"row"`
}

// Directory answers whether a Peppol participant is registered.
type Directory interface {
	IsRegistered(ctx context.Context, participantID string) (bool, error)
}

// Invoices looks up invoices that already exist for a seller.
type Invoices interface {
	ExistingStatus(ctx context.Context, sellerID int64, invoiceNumber, excludeImportID string) (string, bool, error)
}

// Seller is the company the file was uploaded for.
type Seller struct {
	ID  int64
	TRN string
}

type Validator struct {
	dir      Directory
	invoices Invoices
	log      *slog.Logger
}

func New(dir Directory, invoices Invoices, log *slog.Logger) *Validator {
	return &Validator{dir: dir, invoices: invoices, log: log}
}

var trnPattern = regexp.MustCompile(`^1\d{12}03$`)

// Rules lists the portal rule ids.
var Rules = []string{
	"AE-BR-01", "AE-BR-02", "AE-BR-03", "AE-BR-04", "AE-BR-05", "AE-BR-06",
	"AE-BR-07", "AE-BR-08", "AE-BR-09", "AE-BR-10", "AE-BR-11", "AE-BR-12",
}

func init() {
	for _, rule := range Rules {
		metrics.ValidationFailures.WithLabelValues(rule)
	}
}

var vatRates = map[string]float64{"S": 5, "Z": 0, "E": 0, "O": 0}

// Validate checks every row of an invoice, then the invoice as a whole.
func (v *Validator) Validate(ctx context.Context, seller Seller, importID string, g sheet.Group) []Error {
	var errs []Error
	for _, row := range g.Rows {
		errs = append(errs, v.validateRow(ctx, row)...)
	}

	first := g.Rows[0]
	if g.InvoiceNumber != "" {
		status, exists, err := v.invoices.ExistingStatus(ctx, seller.ID, g.InvoiceNumber, importID)
		if err != nil {
			v.log.Error("duplicate check failed", "invoice_number", g.InvoiceNumber, "error", err)
		} else if exists {
			errs = append(errs, Error{"AE-BR-09", sheet.ColInvoiceNumber,
				fmt.Sprintf("Document number '%s' already exists for this seller with status '%s'.", g.InvoiceNumber, status), first.Number})
		}
	}
	if trn := first.Get(sheet.ColSellerTRN); trn != "" && trn != seller.TRN {
		errs = append(errs, Error{"AE-BR-10", sheet.ColSellerTRN,
			"Seller TRN in the file does not match the selected company.", first.Number})
	}

	for _, e := range errs {
		metrics.ValidationFailures.WithLabelValues(e.Rule).Inc()
	}
	return errs
}

func (v *Validator) validateRow(ctx context.Context, r sheet.Row) []Error {
	var errs []Error
	add := func(rule, field, msg string) {
		errs = append(errs, Error{Rule: rule, Field: field, Message: msg, Row: r.Number})
	}

	for _, col := range sheet.Columns {
		if sheet.Mandatory(col) && r.Get(col) == "" {
			add("AE-BR-01", col, fmt.Sprintf("[%s] is mandatory and must not be empty.", col))
		}
	}

	if trn := r.Get(sheet.ColSellerTRN); trn != "" && !trnPattern.MatchString(trn) {
		add("AE-BR-02", sheet.ColSellerTRN, "Seller TRN must be 15 digits, start with 1 and end with 03.")
	}
	if trn := r.Get(sheet.ColBuyerTRN); trn != "" && !trnPattern.MatchString(trn) {
		add("AE-BR-03", sheet.ColBuyerTRN, "Buyer TRN must be 15 digits, start with 1 and end with 03.")
	}

	for _, col := range []string{sheet.ColIssueDate, sheet.ColDueDate} {
		if val := r.Get(col); val != "" {
			if _, ok := sheet.ParseDate(val); !ok {
				add("AE-BR-04", col, fmt.Sprintf("[%s] must be a date in YYYY-MM-DD format (got '%s').", col, val))
			}
		}
	}

	if cur := r.Get(sheet.ColCurrency); cur != "" && cur != "AED" {
		add("AE-BR-05", sheet.ColCurrency, "Invoice currency code must be AED.")
	}

	category := r.Get(sheet.ColVATCategory)
	expectedRate, knownCategory := vatRates[category]
	if category != "" && !knownCategory {
		add("AE-BR-06", sheet.ColVATCategory, fmt.Sprintf("VAT category code must be one of S, Z, E, O (got '%s').", category))
	}

	numbers := map[string]float64{}
	for _, col := range []string{sheet.ColQuantity, sheet.ColNetPrice, sheet.ColLineNetAmount, sheet.ColVATRate} {
		val := r.Get(col)
		if val == "" {
			continue
		}
		n, ok := sheet.ParseNumber(val)
		if !ok {
			add("AE-BR-01", col, fmt.Sprintf("[%s] must be a number (got '%s').", col, val))
			continue
		}
		numbers[col] = n
	}

	if rate, ok := numbers[sheet.ColVATRate]; ok && knownCategory && rate != expectedRate {
		add("AE-BR-07", sheet.ColVATRate, fmt.Sprintf("VAT rate %s is not valid for category %s: S must be 5, Z/E/O must be 0.",
			strconv.FormatFloat(rate, 'f', -1, 64), category))
	}

	qty, okQ := numbers[sheet.ColQuantity]
	price, okP := numbers[sheet.ColNetPrice]
	net, okN := numbers[sheet.ColLineNetAmount]
	if okQ && okP && okN {
		expected := qty * price
		if math.Abs(expected-net) > 0.01+1e-9 {
			add("AE-BR-08", sheet.ColLineNetAmount, fmt.Sprintf("Invoice line net amount %.2f must equal quantity × net price (%s × %s = %.2f).",
				net, strconv.FormatFloat(qty, 'f', -1, 64), strconv.FormatFloat(price, 'f', -1, 64), expected))
		}
	}

	if r.Get(sheet.ColPaymentMeans) == "30" && r.Get(sheet.ColPaymentAccount) == "" {
		add("AE-BR-12", sheet.ColPaymentAccount, "Payment account identifier is mandatory when payment means type code is 30 (credit transfer).")
	}

	if buyer := r.Get(sheet.ColBuyerPeppolID); buyer != "" {
		registered, err := v.dir.IsRegistered(ctx, buyer)
		switch {
		case err != nil:
			v.log.Warn("participant lookup failed", "participant_id", buyer, "row", r.Number, "error", err)
		case !registered:
			add("AE-BR-11", sheet.ColBuyerPeppolID, fmt.Sprintf("Buyer Peppol participant '%s' is not registered on the Peppol network.", buyer))
		}
	}
	return errs
}
