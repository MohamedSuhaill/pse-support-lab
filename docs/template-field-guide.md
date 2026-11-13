# Template field guide — Invoicely UAE invoice template

Customers upload an Excel workbook (`.xlsx`) with a sheet named **`Invoices`**. Row 1 holds the column headers exactly as below. Columns marked `*` are mandatory. The workbook may contain other sheets (the official template has `Instructions` and `Lists`); they are ignored.

## How rows become invoices

- One row = one **invoice line**.
- Invoice-level columns (number, dates, seller, buyer, payment) are repeated on every line of the same invoice.
- Rows are grouped by `Invoice number*`. All rows with the same number form one invoice, in file order.
- Invoice **totals are not in the file**. The portal calculates them from the lines: total without VAT, total VAT (per VAT category), total with VAT and amount due.

## Columns

| Column | Mandatory | Format | Example | Notes |
|---|---|---|---|---|
| `Source Name*` | Yes | Text | `Desert Rose Back Office` | Name of the system that produced the file. |
| `Invoice type*` | Yes | Code | `380` | 380 = tax invoice. |
| `Invoice number*` | Yes | Text | `DRT-2611-001` | Unique per seller. Re-using a number gives AE-BR-09. |
| `Invoice issue date*` | Yes | Date `YYYY-MM-DD` | `2026-11-13` | See [Dates](#dates). |
| `Payment due date*` | Yes | Date `YYYY-MM-DD` | `2026-12-13` | See [Dates](#dates). |
| `Invoice currency code*` | Yes | ISO 4217 | `AED` | Only AED is supported. |
| `Seller name*` | Yes | Text | `Desert Rose Trading LLC` | |
| `Seller TRN identifier*` | Yes | 15 digits | `100234567800003` | Starts with 1, ends with 03. Must match the company chosen at upload. |
| `Seller Peppol Id*` | Yes | `0235:` + 10 digits | `0235:1002345678` | |
| `Seller address line 1*` | Yes | Text | `Office 1204, Al Fahidi Tower` | |
| `Seller city*` | Yes | Text | `Dubai` | |
| `Seller country subdivision*` | Yes | Emirate code | `DXB` | AUH, DXB, SHJ, UAQ, FUJ, AJM, RAK. |
| `Seller country code*` | Yes | ISO 3166 | `AE` | |
| `Buyer name*` | Yes | Text | `Al Waha Hypermarkets LLC` | |
| `Buyer TRN identifier*` | Yes | 15 digits | `100129384700003` | Starts with 1, ends with 03. |
| `Buyer Peppol Id*` | Yes | `0235:` + 10 digits | `0235:1001293847` | Must be registered on Peppol (AE-BR-11). |
| `Buyer address line 1*` | Yes | Text | `Al Wahda Street` | |
| `Buyer city*` | Yes | Text | `Sharjah` | |
| `Buyer country subdivision*` | Yes | Emirate code | `SHJ` | |
| `Buyer country code*` | Yes | ISO 3166 | `AE` | |
| `Invoice line identifier*` | Yes | Text / number | `1` | Unique within the invoice. |
| `Item name*` | Yes | Text | `Arabic coffee beans 1 kg` | |
| `Itemquantity*` | Yes | Number | `40` | The header is spelt this way in the official template. |
| `Unit of measurement*` | Yes | UN/ECE code | `KGM` | EA, LTR, KGM, TNE, MTR, HUR, BX… (see `Lists`). |
| `Item net price*` | Yes | Number, up to 3 decimals | `85` or `2.835` | Price per unit, excluding VAT. Fuel and other quantity-priced goods often use 3 decimals. |
| `Invoice line net amount*` | Yes | Number, 2 decimals | `3400.00` | Quantity × net price, rounded to 2 decimals (AE-BR-08). |
| `Invoiced item VAT category code*` | Yes | Code | `S` | S = standard, Z = zero-rated, E = exempt, O = out of scope. |
| `Invoiced item VAT rate*` | Yes | Number | `5` | S = 5, Z/E/O = 0. |
| `VAT exemption reason code` | No | Code | `VATEX-AE-EXEMPT-FIN` | For exempt (E) lines. |
| `Item type*` | Yes | Text | `Goods` | `Goods`, `Services` or `Both` (exact spelling). |
| `Payment means type code*` | Yes | Code | `30` | 30 = credit transfer, 10 = cash. |
| `Payment account identifier` | When code is 30 | IBAN | `AE070331234567890123456` | AE-BR-12. |

## Dates

Dates must be **`YYYY-MM-DD`** text (for example `2026-11-13`) or real Excel date cells.

Formats such as `13/11/2026` or `11/13/2026` are **rejected with AE-BR-04**. Written as text, they are ambiguous: `05/11/2026` is 5 November in the UAE and 11 May in the United States, and the portal must never guess the date on a tax invoice.

How to fix it in Excel:

- Select the date column → **Format Cells** → **Custom** → type `yyyy-mm-dd`. This works when the cells are real dates (they align to the right by default).
- If the dates were typed as text, convert them, for example with a helper column `=TEXT(DATE(RIGHT(A2,4),MID(A2,4,2),LEFT(A2,2)),"yyyy-mm-dd")`, then paste the results as values.

## Amounts

- All amounts are in AED.
- `Item net price*` may have up to 3 decimals; `Invoice line net amount*` has 2 decimals.
- Line VAT, VAT per category and invoice totals are calculated by the portal.

## Portal validation rules

These run when the file is uploaded. An invoice that fails any of them is **Rejected (by portal)** and is not sent to the ASP. The row number in the error points to the row in Excel.

| Rule | Message |
|---|---|
| AE-BR-01 | `[{field}] is mandatory and must not be empty.` (also used for a non-numeric value in a number column) |
| AE-BR-02 | `Seller TRN must be 15 digits, start with 1 and end with 03.` |
| AE-BR-03 | `Buyer TRN must be 15 digits, start with 1 and end with 03.` |
| AE-BR-04 | `[{field}] must be a date in YYYY-MM-DD format (got '{value}').` |
| AE-BR-05 | `Invoice currency code must be AED.` |
| AE-BR-06 | `VAT category code must be one of S, Z, E, O (got '{value}').` |
| AE-BR-07 | `VAT rate {rate} is not valid for category {category}: S must be 5, Z/E/O must be 0.` |
| AE-BR-08 | `Invoice line net amount {net} must equal quantity × net price ({qty} × {price} = {expected}).` (tolerance 0.01) |
| AE-BR-09 | `Document number '{number}' already exists for this seller with status '{status}'.` |
| AE-BR-10 | `Seller TRN in the file does not match the selected company.` |
| AE-BR-11 | `Buyer Peppol participant '{id}' is not registered on the Peppol network.` |
| AE-BR-12 | `Payment account identifier is mandatory when payment means type code is 30 (credit transfer).` |

## ASP rules

The ASP runs its own checks when it receives an invoice. An invoice that fails them is **Rejected (by ASP)**.

| Rule | Meaning |
|---|---|
| BR-CO-10 | Sum of invoice line net amounts must equal the invoice total of line net amounts. |
| BR-CO-14 | Invoice total VAT amount must equal the sum of the VAT breakdown (per category) amounts. |
| BR-CO-15 | `Invoice total amount with VAT (IBT-112) must equal Invoice total amount without VAT (IBT-109) + Invoice total VAT amount (IBT-110).` |
| IBR-132-AE | Seller / buyer VAT identifier must be a valid TRN. |
| PEPPOL-R001 | Specification identifier must be `urn:peppol:pint:billing-1@ae-1`. |
