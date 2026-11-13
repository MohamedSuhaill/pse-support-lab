// Package sheet reads customer workbooks that follow the Invoicely UAE template.
package sheet

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Name is the worksheet that holds invoice rows.
const Name = "Invoices"

// Template column headers. A trailing "*" marks a mandatory column.
const (
	ColSourceName        = "Source Name*"
	ColInvoiceType       = "Invoice type*"
	ColInvoiceNumber     = "Invoice number*"
	ColIssueDate         = "Invoice issue date*"
	ColDueDate           = "Payment due date*"
	ColCurrency          = "Invoice currency code*"
	ColSellerName        = "Seller name*"
	ColSellerTRN         = "Seller TRN identifier*"
	ColSellerPeppolID    = "Seller Peppol Id*"
	ColSellerAddress     = "Seller address line 1*"
	ColSellerCity        = "Seller city*"
	ColSellerSubdivision = "Seller country subdivision*"
	ColSellerCountry     = "Seller country code*"
	ColBuyerName         = "Buyer name*"
	ColBuyerTRN          = "Buyer TRN identifier*"
	ColBuyerPeppolID     = "Buyer Peppol Id*"
	ColBuyerAddress      = "Buyer address line 1*"
	ColBuyerCity         = "Buyer city*"
	ColBuyerSubdivision  = "Buyer country subdivision*"
	ColBuyerCountry      = "Buyer country code*"
	ColLineID            = "Invoice line identifier*"
	ColItemName          = "Item name*"
	ColQuantity          = "Itemquantity*"
	ColUnit              = "Unit of measurement*"
	ColNetPrice          = "Item net price*"
	ColLineNetAmount     = "Invoice line net amount*"
	ColVATCategory       = "Invoiced item VAT category code*"
	ColVATRate           = "Invoiced item VAT rate*"
	ColExemptionReason   = "VAT exemption reason code"
	ColItemType          = "Item type*"
	ColPaymentMeans      = "Payment means type code*"
	ColPaymentAccount    = "Payment account identifier"
)

// Columns lists the template columns in template order.
var Columns = []string{
	ColSourceName, ColInvoiceType, ColInvoiceNumber, ColIssueDate, ColDueDate, ColCurrency,
	ColSellerName, ColSellerTRN, ColSellerPeppolID, ColSellerAddress, ColSellerCity, ColSellerSubdivision, ColSellerCountry,
	ColBuyerName, ColBuyerTRN, ColBuyerPeppolID, ColBuyerAddress, ColBuyerCity, ColBuyerSubdivision, ColBuyerCountry,
	ColLineID, ColItemName, ColQuantity, ColUnit, ColNetPrice, ColLineNetAmount, ColVATCategory, ColVATRate,
	ColExemptionReason, ColItemType, ColPaymentMeans, ColPaymentAccount,
}

// Mandatory reports whether a column must have a value.
func Mandatory(col string) bool { return strings.HasSuffix(col, "*") }

// Row is one data row of the Invoices sheet.
type Row struct {
	Number int // row number as shown in Excel (header is row 1)
	cells  map[string]string
}

// Get returns the trimmed raw value of a column ("" when absent).
func (r Row) Get(col string) string { return r.cells[col] }

// Workbook is a parsed customer file.
type Workbook struct {
	File    *excelize.File
	Sheet   string
	Headers []string
	Rows    []Row
}

// ErrNotTemplate is returned when the file does not look like the UAE template.
var ErrNotTemplate = errors.New("the file does not follow the Invoicely UAE template (sheet 'Invoices' with an 'Invoice number*' column)")

// Read parses an .xlsx file.
func Read(data []byte) (*Workbook, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not open the file as an Excel workbook: %w", err)
	}
	sheet := Name
	if idx, _ := f.GetSheetIndex(sheet); idx < 0 {
		sheet = f.GetSheetName(0)
	}
	raw, err := f.GetRows(sheet, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("could not read sheet %q: %w", sheet, err)
	}
	if len(raw) == 0 {
		return nil, ErrNotTemplate
	}
	headers := make([]string, len(raw[0]))
	found := false
	for i, h := range raw[0] {
		headers[i] = strings.TrimSpace(h)
		if headers[i] == ColInvoiceNumber {
			found = true
		}
	}
	if !found {
		return nil, ErrNotTemplate
	}
	wb := &Workbook{File: f, Sheet: sheet, Headers: headers}
	for i, values := range raw[1:] {
		row := Row{Number: i + 2, cells: make(map[string]string, len(headers))}
		empty := true
		for c, v := range values {
			if c >= len(headers) || headers[c] == "" {
				continue
			}
			v = strings.TrimSpace(v)
			if v != "" {
				empty = false
			}
			row.cells[headers[c]] = v
		}
		if !empty {
			wb.Rows = append(wb.Rows, row)
		}
	}
	return wb, nil
}

// DisplayRows returns the sheet as the customer sees it in Excel (formatted values).
func (w *Workbook) DisplayRows() ([][]string, error) {
	return w.File.GetRows(w.Sheet)
}

// Group is the set of rows that make up one invoice.
type Group struct {
	InvoiceNumber string
	Rows          []Row
}

// GroupByInvoice groups rows by invoice number, keeping file order.
func GroupByInvoice(rows []Row) []Group {
	var groups []Group
	index := map[string]int{}
	for _, r := range rows {
		n := r.Get(ColInvoiceNumber)
		if n == "" {
			groups = append(groups, Group{Rows: []Row{r}})
			continue
		}
		if i, ok := index[n]; ok {
			groups[i].Rows = append(groups[i].Rows, r)
			continue
		}
		index[n] = len(groups)
		groups = append(groups, Group{InvoiceNumber: n, Rows: []Row{r}})
	}
	return groups
}

var isoDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// ParseDate accepts ISO dates (YYYY-MM-DD) and Excel date cells (serial numbers)
// and returns the date as YYYY-MM-DD.
func ParseDate(v string) (string, bool) {
	if isoDate.MatchString(v) {
		if _, err := time.Parse("2006-01-02", v); err == nil {
			return v, true
		}
		return "", false
	}
	if serial, err := strconv.ParseFloat(v, 64); err == nil && serial > 0 && serial < 2958466 {
		t, err := excelize.ExcelDateToTime(serial, false)
		if err != nil {
			return "", false
		}
		return t.Format("2006-01-02"), true
	}
	return "", false
}

// ParseNumber parses a numeric cell value.
func ParseNumber(v string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", ""), 64)
	return f, err == nil
}
