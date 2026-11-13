// Package workbook writes .xlsx files in the Invoicely UAE template layout.
package workbook

import (
	"fmt"
	"math"

	"github.com/xuri/excelize/v2"
)

// Headers of the "Invoices" sheet, in template order.
var Headers = []string{
	"Source Name*", "Invoice type*", "Invoice number*", "Invoice issue date*", "Payment due date*", "Invoice currency code*",
	"Seller name*", "Seller TRN identifier*", "Seller Peppol Id*", "Seller address line 1*", "Seller city*", "Seller country subdivision*", "Seller country code*",
	"Buyer name*", "Buyer TRN identifier*", "Buyer Peppol Id*", "Buyer address line 1*", "Buyer city*", "Buyer country subdivision*", "Buyer country code*",
	"Invoice line identifier*", "Item name*", "Itemquantity*", "Unit of measurement*", "Item net price*", "Invoice line net amount*",
	"Invoiced item VAT category code*", "Invoiced item VAT rate*", "VAT exemption reason code", "Item type*",
	"Payment means type code*", "Payment account identifier",
}

type Party struct {
	Name, TRN, Address, City, Subdivision string
}

// PeppolID derives the participant id used in the samples (scheme 0235 + first 10 TRN digits).
func (p Party) PeppolID() string { return "0235:" + p.TRN[:10] }

type Line struct {
	Item      string
	Qty       float64
	Unit      string
	Price     float64
	Category  string
	Exemption string
	ItemType  string
}

type Invoice struct {
	Number    string
	IssueDate string
	DueDate   string
	Buyer     Party
	Lines     []Line
}

type File struct {
	Source       string
	Seller       Party
	PaymentMeans string
	Account      string
	Invoices     []Invoice
}

// LineNet is quantity × price rounded half-up to 2 decimals, as Excel's ROUND does.
func LineNet(qty, price float64) float64 {
	thousandths := math.Round(qty * price * 1000)
	cents := math.Floor((thousandths+5)/10)
	return cents / 100
}

func rate(category string) float64 {
	if category == "S" {
		return 5
	}
	return 0
}

// Build renders the workbook.
func Build(f File) ([]byte, error) {
	x := excelize.NewFile()
	defer x.Close()

	if err := x.SetSheetName("Sheet1", "Instructions"); err != nil {
		return nil, err
	}
	instructions := []string{
		"Invoicely UAE invoice template (PINT AE)",
		"",
		"1. Fill one row per invoice line on the 'Invoices' sheet. Repeat the invoice details on every line of the same invoice.",
		"2. Columns marked * are mandatory.",
		"3. Dates must be entered as YYYY-MM-DD (for example 2026-11-13) or as Excel date cells.",
		"4. Amounts are in AED. Invoice totals are calculated by the portal from the lines.",
		"5. Allowed values for coded columns are listed on the 'Lists' sheet.",
	}
	for i, s := range instructions {
		if err := x.SetCellStr("Instructions", fmt.Sprintf("A%d", i+1), s); err != nil {
			return nil, err
		}
	}

	if _, err := x.NewSheet("Invoices"); err != nil {
		return nil, err
	}
	sw, err := x.NewStreamWriter("Invoices")
	if err != nil {
		return nil, err
	}
	header := make([]any, len(Headers))
	for i, h := range Headers {
		header[i] = h
	}
	if err := sw.SetRow("A1", header); err != nil {
		return nil, err
	}
	rowNum := 2
	for _, inv := range f.Invoices {
		for li, l := range inv.Lines {
			row := []any{
				f.Source, 380, inv.Number, inv.IssueDate, inv.DueDate, "AED",
				f.Seller.Name, f.Seller.TRN, f.Seller.PeppolID(), f.Seller.Address, f.Seller.City, f.Seller.Subdivision, "AE",
				inv.Buyer.Name, inv.Buyer.TRN, inv.Buyer.PeppolID(), inv.Buyer.Address, inv.Buyer.City, inv.Buyer.Subdivision, "AE",
				li + 1, l.Item, l.Qty, l.Unit, l.Price, LineNet(l.Qty, l.Price),
				l.Category, rate(l.Category), l.Exemption, l.ItemType,
				f.PaymentMeans, f.Account,
			}
			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			if err := sw.SetRow(cell, row); err != nil {
				return nil, err
			}
			rowNum++
		}
	}
	if err := sw.Flush(); err != nil {
		return nil, err
	}

	if err := writeLists(x); err != nil {
		return nil, err
	}
	idx, _ := x.GetSheetIndex("Invoices")
	x.SetActiveSheet(idx)

	buf, err := x.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var units = [][2]string{
	{"EA", "Each"}, {"LTR", "Litre"}, {"KGM", "Kilogram"}, {"TNE", "Tonne"}, {"MTR", "Metre"}, {"MTK", "Square metre"},
	{"MTQ", "Cubic metre"}, {"HUR", "Hour"}, {"DAY", "Day"}, {"MON", "Month"}, {"BX", "Box"}, {"PK", "Pack"},
	{"BG", "Bag"}, {"SET", "Set"}, {"PR", "Pair"}, {"RO", "Roll"}, {"CT", "Carton"}, {"DZN", "Dozen"},
}

var countries = []string{
	"AE", "SA", "OM", "QA", "KW", "BH", "EG", "JO", "LB", "IN", "PK", "BD", "LK", "PH", "CN", "JP", "KR", "SG", "MY", "ID",
	"TH", "VN", "AU", "NZ", "GB", "IE", "FR", "DE", "NL", "BE", "IT", "ES", "PT", "CH", "AT", "SE", "NO", "DK", "FI", "PL",
	"TR", "US", "CA", "MX", "BR", "AR", "ZA", "KE", "NG", "MA",
}

// writeLists adds the reference lists used by the template's drop-downs.
func writeLists(x *excelize.File) error {
	if _, err := x.NewSheet("Lists"); err != nil {
		return err
	}
	sw, err := x.NewStreamWriter("Lists")
	if err != nil {
		return err
	}
	if err := sw.SetRow("A1", []any{"Unit code", "Unit name", "Country code", "Emirate", "VAT category", "Item type",
		"Payment means", "Exemption reason", "Commodity code", "Commodity description"}); err != nil {
		return err
	}
	emirates := []string{"AUH", "DXB", "SHJ", "UAQ", "FUJ", "AJM", "RAK"}
	vat := []string{"S", "Z", "E", "O"}
	itemTypes := []string{"Goods", "Services", "Both"}
	means := []string{"10 Cash", "30 Credit transfer", "48 Bank card", "49 Direct debit"}
	exemptions := []string{"VATEX-AE-EXEMPT-FIN", "VATEX-AE-EXEMPT-RES", "VATEX-AE-EXEMPT-TRANS", "VATEX-AE-EXEMPT-LAND"}
	const commodityRows = 4000
	for i := range commodityRows {
		row := make([]any, 10)
		if i < len(units) {
			row[0], row[1] = units[i][0], units[i][1]
		}
		if i < len(countries) {
			row[2] = countries[i]
		}
		if i < len(emirates) {
			row[3] = emirates[i]
		}
		if i < len(vat) {
			row[4] = vat[i]
		}
		if i < len(itemTypes) {
			row[5] = itemTypes[i]
		}
		if i < len(means) {
			row[6] = means[i]
		}
		if i < len(exemptions) {
			row[7] = exemptions[i]
		}
		chapter := 1 + i%97
		heading := 1 + (i/97)%99
		row[8] = fmt.Sprintf("%02d%02d.%02d", chapter, heading, i%90+10)
		row[9] = fmt.Sprintf("Chapter %02d / heading %02d%02d — goods classified under subheading %02d", chapter, chapter, heading, i%90+10)
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := sw.SetRow(cell, row); err != nil {
			return err
		}
	}
	return sw.Flush()
}
