// Package pint maps template rows to a simplified PINT AE invoice payload.
package pint

import (
	"math"
	"sort"

	"invoicely/api/internal/sheet"
)

const (
	CustomizationID = "urn:peppol:pint:billing-1@ae-1"
	ProfileID       = "urn:peppol:bis:billing"
)

type Address struct {
	Line1       string `json:"line1"`
	City        string `json:"city"`
	Subdivision string `json:"subdivision"`
	Country     string `json:"country"`
}

type Party struct {
	Name     string  `json:"name"`
	TRN      string  `json:"trn"`
	PeppolID string  `json:"peppolId"`
	Address  Address `json:"address"`
}

type Line struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Quantity        float64 `json:"quantity"`
	UnitCode        string  `json:"unitCode"`
	UnitPrice       float64 `json:"unitPrice"`
	NetAmount       float64 `json:"netAmount"`
	VATCategory     string  `json:"vatCategory"`
	VATRate         float64 `json:"vatRate"`
	VATAmount       float64 `json:"vatAmount"`
	ExemptionReason string  `json:"exemptionReason,omitempty"`
	ItemType        string  `json:"itemType"`
}

type TaxSubtotal struct {
	Category      string  `json:"category"`
	Rate          float64 `json:"rate"`
	TaxableAmount float64 `json:"taxableAmount"`
	TaxAmount     float64 `json:"taxAmount"`
}

type Totals struct {
	LineExtensionAmount float64 `json:"lineExtensionAmount"`
	TaxExclusiveAmount  float64 `json:"taxExclusiveAmount"`
	TaxAmount           float64 `json:"taxAmount"`
	TaxInclusiveAmount  float64 `json:"taxInclusiveAmount"`
	PayableAmount       float64 `json:"payableAmount"`
}

type Payment struct {
	MeansCode string `json:"meansCode"`
	AccountID string `json:"accountId,omitempty"`
}

type Invoice struct {
	CustomizationID string        `json:"customizationId"`
	ProfileID       string        `json:"profileId"`
	SourceName      string        `json:"sourceName"`
	InvoiceNumber   string        `json:"invoiceNumber"`
	InvoiceTypeCode string        `json:"invoiceTypeCode"`
	IssueDate       string        `json:"issueDate"`
	DueDate         string        `json:"dueDate"`
	Currency        string        `json:"currency"`
	Seller          Party         `json:"seller"`
	Buyer           Party         `json:"buyer"`
	Lines           []Line        `json:"lines"`
	TaxSubtotals    []TaxSubtotal `json:"taxSubtotals"`
	Totals          Totals        `json:"totals"`
	Payment         Payment       `json:"payment"`
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func num(r sheet.Row, col string) float64 {
	v, _ := sheet.ParseNumber(r.Get(col))
	return v
}

func date(r sheet.Row, col string) string {
	if d, ok := sheet.ParseDate(r.Get(col)); ok {
		return d
	}
	return r.Get(col)
}

// Map builds the payload for one invoice. Invoice-level values are taken from the
// first row of the group; every row becomes an invoice line.
func Map(g sheet.Group) *Invoice {
	first := g.Rows[0]
	inv := &Invoice{
		CustomizationID: CustomizationID,
		ProfileID:       ProfileID,
		SourceName:      first.Get(sheet.ColSourceName),
		InvoiceNumber:   g.InvoiceNumber,
		InvoiceTypeCode: first.Get(sheet.ColInvoiceType),
		IssueDate:       date(first, sheet.ColIssueDate),
		DueDate:         date(first, sheet.ColDueDate),
		Currency:        first.Get(sheet.ColCurrency),
		Seller: Party{
			Name:     first.Get(sheet.ColSellerName),
			TRN:      first.Get(sheet.ColSellerTRN),
			PeppolID: first.Get(sheet.ColSellerPeppolID),
			Address: Address{
				Line1:       first.Get(sheet.ColSellerAddress),
				City:        first.Get(sheet.ColSellerCity),
				Subdivision: first.Get(sheet.ColSellerSubdivision),
				Country:     first.Get(sheet.ColSellerCountry),
			},
		},
		Buyer: Party{
			Name:     first.Get(sheet.ColBuyerName),
			TRN:      first.Get(sheet.ColBuyerTRN),
			PeppolID: first.Get(sheet.ColBuyerPeppolID),
			Address: Address{
				Line1:       first.Get(sheet.ColBuyerAddress),
				City:        first.Get(sheet.ColBuyerCity),
				Subdivision: first.Get(sheet.ColBuyerSubdivision),
				Country:     first.Get(sheet.ColBuyerCountry),
			},
		},
		Payment: Payment{
			MeansCode: first.Get(sheet.ColPaymentMeans),
			AccountID: first.Get(sheet.ColPaymentAccount),
		},
	}
	for _, r := range g.Rows {
		net := num(r, sheet.ColLineNetAmount)
		rate := num(r, sheet.ColVATRate)
		inv.Lines = append(inv.Lines, Line{
			ID:              r.Get(sheet.ColLineID),
			Name:            r.Get(sheet.ColItemName),
			Quantity:        num(r, sheet.ColQuantity),
			UnitCode:        r.Get(sheet.ColUnit),
			UnitPrice:       num(r, sheet.ColNetPrice),
			NetAmount:       net,
			VATCategory:     r.Get(sheet.ColVATCategory),
			VATRate:         rate,
			VATAmount:       round2(net * rate / 100),
			ExemptionReason: r.Get(sheet.ColExemptionReason),
			ItemType:        r.Get(sheet.ColItemType),
		})
	}
	inv.TaxSubtotals, inv.Totals = computeTotals(inv.Lines)
	return inv
}

// computeTotals builds the VAT breakdown and document totals.
func computeTotals(lines []Line) ([]TaxSubtotal, Totals) {
	byCategory := map[string]*TaxSubtotal{}
	var net, vat, gross float64
	for _, l := range lines {
		s, ok := byCategory[l.VATCategory]
		if !ok {
			s = &TaxSubtotal{Category: l.VATCategory, Rate: l.VATRate}
			byCategory[l.VATCategory] = s
		}
		s.TaxableAmount += l.NetAmount
		s.TaxAmount += l.VATAmount

		net += l.NetAmount
		vat += l.VATAmount
		gross += l.Quantity * l.UnitPrice * (1 + l.VATRate/100)
	}
	subtotals := make([]TaxSubtotal, 0, len(byCategory))
	for _, s := range byCategory {
		s.TaxableAmount = round2(s.TaxableAmount)
		s.TaxAmount = round2(s.TaxAmount)
		subtotals = append(subtotals, *s)
	}
	sort.Slice(subtotals, func(i, j int) bool { return subtotals[i].Category < subtotals[j].Category })
	return subtotals, Totals{
		LineExtensionAmount: round2(net),
		TaxExclusiveAmount:  round2(net),
		TaxAmount:           round2(vat),
		TaxInclusiveAmount:  round2(gross),
		PayableAmount:       round2(gross),
	}
}
