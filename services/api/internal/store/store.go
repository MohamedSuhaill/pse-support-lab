// Package store persists sellers, imports and invoices in Postgres.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Invoice statuses.
const (
	StatusParsed    = "Parsed"
	StatusValidated = "Validated"
	StatusSubmitted = "Submitted"
	StatusAccepted  = "Accepted"
	StatusRejected  = "Rejected"
	StatusFailed    = "Failed"
)

// Import statuses.
const (
	ImportUploaded   = "Uploaded"
	ImportProcessing = "Processing"
	ImportCompleted  = "Completed"
	ImportFailed     = "Failed"
)

var ErrNotFound = errors.New("not found")

type Store struct{ db *pgxpool.Pool }

func Open(ctx context.Context, url string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Ping(ctx context.Context) error { return s.db.Ping(ctx) }

type Seller struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	TRN      string `json:"trn"`
	PeppolID string `json:"peppolId"`
	Emirate  string `json:"emirate"`
}

func (s *Store) Sellers(ctx context.Context) ([]Seller, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name, trn, peppol_id, emirate FROM sellers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Seller, error) {
		var v Seller
		err := r.Scan(&v.ID, &v.Name, &v.TRN, &v.PeppolID, &v.Emirate)
		return v, err
	})
}

func (s *Store) Seller(ctx context.Context, id int64) (Seller, error) {
	var v Seller
	err := s.db.QueryRow(ctx, `SELECT id, name, trn, peppol_id, emirate FROM sellers WHERE id = $1`, id).
		Scan(&v.ID, &v.Name, &v.TRN, &v.PeppolID, &v.Emirate)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}

// ---------- imports ----------

type Import struct {
	ID           string           `json:"id"`
	SellerID     int64            `json:"sellerId"`
	SellerName   string           `json:"sellerName"`
	Filename     string           `json:"filename"`
	Status       string           `json:"status"`
	Error        string           `json:"error"`
	RowCount     int              `json:"rowCount"`
	InvoiceCount int              `json:"invoiceCount"`
	DurationMs   *int64           `json:"durationMs"`
	CreatedAt    time.Time        `json:"createdAt"`
	CompletedAt  *time.Time       `json:"completedAt"`
	StatusCounts map[string]int   `json:"statusCounts"`
	Invoices     []InvoiceSummary `json:"invoices,omitempty"`
}

func (s *Store) CreateImport(ctx context.Context, id string, sellerID int64, filename string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO imports (id, seller_id, filename, status) VALUES ($1, $2, $3, $4)`,
		id, sellerID, filename, ImportUploaded)
	return err
}

func (s *Store) SetImportStatus(ctx context.Context, id, status string) error {
	_, err := s.db.Exec(ctx, `UPDATE imports SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (s *Store) SetImportCounts(ctx context.Context, id string, rows, invoices int) error {
	_, err := s.db.Exec(ctx, `UPDATE imports SET row_count = $2, invoice_count = $3 WHERE id = $1`, id, rows, invoices)
	return err
}

func (s *Store) FinishImport(ctx context.Context, id, status, errMsg string, duration time.Duration) error {
	_, err := s.db.Exec(ctx, `UPDATE imports SET status = $2, error = $3, duration_ms = $4, completed_at = now() WHERE id = $1`,
		id, status, errMsg, duration.Milliseconds())
	return err
}

const importColumns = `
	i.id::text, i.seller_id, s.name, i.filename, i.status, i.error, i.row_count, i.invoice_count,
	i.duration_ms, i.created_at, i.completed_at,
	COALESCE((SELECT jsonb_object_agg(status, n) FROM (
		SELECT status, count(*) AS n FROM invoices WHERE import_id = i.id GROUP BY status) c), '{}'::jsonb)`

func scanImport(row pgx.Row) (Import, error) {
	var v Import
	var counts []byte
	err := row.Scan(&v.ID, &v.SellerID, &v.SellerName, &v.Filename, &v.Status, &v.Error, &v.RowCount,
		&v.InvoiceCount, &v.DurationMs, &v.CreatedAt, &v.CompletedAt, &counts)
	if err != nil {
		return v, err
	}
	v.StatusCounts = map[string]int{StatusParsed: 0, StatusValidated: 0, StatusSubmitted: 0, StatusAccepted: 0, StatusRejected: 0, StatusFailed: 0}
	var got map[string]int
	if err := json.Unmarshal(counts, &got); err != nil {
		return v, err
	}
	for k, n := range got {
		v.StatusCounts[k] = n
	}
	return v, nil
}

func (s *Store) ListImports(ctx context.Context, limit int) ([]Import, error) {
	rows, err := s.db.Query(ctx, `SELECT `+importColumns+`
		FROM imports i JOIN sellers s ON s.id = i.seller_id
		ORDER BY i.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Import{}
	for rows.Next() {
		v, err := scanImport(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) GetImport(ctx context.Context, id string) (Import, error) {
	v, err := scanImport(s.db.QueryRow(ctx, `SELECT `+importColumns+`
		FROM imports i JOIN sellers s ON s.id = i.seller_id WHERE i.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	if err != nil {
		return v, err
	}
	v.Invoices, err = s.invoiceSummaries(ctx, `WHERE v.import_id = $1 ORDER BY v.created_at, v.invoice_number`, id)
	return v, err
}

// ---------- invoices ----------

type InvoiceSummary struct {
	ID            string    `json:"id"`
	InvoiceNumber string    `json:"invoiceNumber"`
	BuyerName     string    `json:"buyerName"`
	IssueDate     string    `json:"issueDate"`
	TotalInclVAT  float64   `json:"totalInclVat"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	RejectedBy    string    `json:"rejectedBy"`
	ErrorCount    int       `json:"errorCount"`
	ASPMessageID  string    `json:"aspMessageId"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Event struct {
	Status string    `json:"status"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}

type InvoiceDetail struct {
	InvoiceSummary
	ImportID    string          `json:"importId"`
	SellerName  string          `json:"sellerName"`
	Payload     json.RawMessage `json:"payload"`
	Errors      json.RawMessage `json:"errors"`
	ASPResponse json.RawMessage `json:"aspResponse"`
	Events      []Event         `json:"events"`
}

const summaryColumns = `v.id::text, v.invoice_number, v.buyer_name, v.issue_date, v.total_incl_vat::float8, v.currency,
	v.status, v.rejected_by, jsonb_array_length(v.errors), v.asp_message_id, v.updated_at`

func scanSummary(row pgx.Row, extra ...any) (InvoiceSummary, error) {
	var v InvoiceSummary
	dest := append([]any{&v.ID, &v.InvoiceNumber, &v.BuyerName, &v.IssueDate, &v.TotalInclVAT, &v.Currency,
		&v.Status, &v.RejectedBy, &v.ErrorCount, &v.ASPMessageID, &v.UpdatedAt}, extra...)
	return v, row.Scan(dest...)
}

func (s *Store) invoiceSummaries(ctx context.Context, where string, args ...any) ([]InvoiceSummary, error) {
	rows, err := s.db.Query(ctx, `SELECT `+summaryColumns+` FROM invoices v `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InvoiceSummary{}
	for rows.Next() {
		v, err := scanSummary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) GetInvoice(ctx context.Context, id string) (InvoiceDetail, error) {
	var d InvoiceDetail
	var aspResponse []byte
	summary, err := scanSummary(s.db.QueryRow(ctx, `SELECT `+summaryColumns+`, v.import_id::text, s.name, v.payload, v.errors, v.asp_response
		FROM invoices v JOIN sellers s ON s.id = v.seller_id WHERE v.id = $1`, id),
		&d.ImportID, &d.SellerName, &d.Payload, &d.Errors, &aspResponse)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, ErrNotFound
	}
	if err != nil {
		return d, err
	}
	d.InvoiceSummary = summary
	d.ASPResponse = json.RawMessage("null")
	if aspResponse != nil {
		d.ASPResponse = aspResponse
	}
	rows, err := s.db.Query(ctx, `SELECT status, detail, created_at FROM invoice_events WHERE invoice_id = $1 ORDER BY id`, id)
	if err != nil {
		return d, err
	}
	d.Events, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Event, error) {
		var e Event
		err := r.Scan(&e.Status, &e.Detail, &e.At)
		return e, err
	})
	return d, err
}

// NewInvoice is an invoice created from an uploaded file.
type NewInvoice struct {
	ID            string
	ImportID      string
	SellerID      int64
	InvoiceNumber string
	BuyerName     string
	IssueDate     string
	Currency      string
	TotalExclVAT  float64
	TotalVAT      float64
	TotalInclVAT  float64
	Payload       []byte
}

// InsertInvoices stores freshly parsed invoices with status Parsed.
func (s *Store) InsertInvoices(ctx context.Context, invoices []NewInvoice) error {
	batch := &pgx.Batch{}
	for _, v := range invoices {
		batch.Queue(`INSERT INTO invoices (id, import_id, seller_id, invoice_number, buyer_name, issue_date, currency,
				total_excl_vat, total_vat, total_incl_vat, status, payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			v.ID, v.ImportID, v.SellerID, v.InvoiceNumber, v.BuyerName, v.IssueDate, v.Currency,
			v.TotalExclVAT, v.TotalVAT, v.TotalInclVAT, StatusParsed, v.Payload)
		batch.Queue(`INSERT INTO invoice_events (invoice_id, status, detail) VALUES ($1, $2, $3)`,
			v.ID, StatusParsed, "Created from uploaded file")
	}
	return s.db.SendBatch(ctx, batch).Close()
}

// ExistingStatus returns the status of a non-rejected invoice with the same number
// for the seller, uploaded in another import.
func (s *Store) ExistingStatus(ctx context.Context, sellerID int64, invoiceNumber, excludeImportID string) (string, bool, error) {
	var status string
	err := s.db.QueryRow(ctx, `SELECT status FROM invoices
		WHERE seller_id = $1 AND invoice_number = $2 AND import_id <> $3 AND status <> $4
		ORDER BY created_at LIMIT 1`, sellerID, invoiceNumber, excludeImportID, StatusRejected).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return status, err == nil, err
}

// SetStatus moves an invoice to a new status and records an event.
func (s *Store) SetStatus(ctx context.Context, id, status, detail string) error {
	return s.update(ctx, id, status, detail, `UPDATE invoices SET status = $2, updated_at = now() WHERE id = $1`, id, status)
}

// MarkRejected stores the rejection errors and who rejected the invoice.
func (s *Store) MarkRejected(ctx context.Context, id, rejectedBy string, errs any, aspResponse []byte, detail string) error {
	b, err := json.Marshal(errs)
	if err != nil {
		return err
	}
	return s.update(ctx, id, StatusRejected, detail,
		`UPDATE invoices SET status = $2, rejected_by = $3, errors = $4, asp_response = $5, updated_at = now() WHERE id = $1`,
		id, StatusRejected, rejectedBy, b, aspResponse)
}

// MarkAccepted stores the ASP acknowledgement.
func (s *Store) MarkAccepted(ctx context.Context, id, messageID string, aspResponse []byte) error {
	return s.update(ctx, id, StatusAccepted, "Accepted by ASP (message id "+messageID+")",
		`UPDATE invoices SET status = $2, asp_message_id = $3, asp_response = $4, updated_at = now() WHERE id = $1`,
		id, StatusAccepted, messageID, aspResponse)
}

func (s *Store) update(ctx context.Context, id, status, detail, query string, args ...any) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO invoice_events (invoice_id, status, detail) VALUES ($1, $2, $3)`, id, status, detail); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Submission is what the submitter needs to send an invoice.
type Submission struct {
	ID             string
	InvoiceNumber  string
	Status         string
	SellerPeppolID string
	Payload        []byte
}

func (s *Store) GetSubmission(ctx context.Context, id string) (Submission, error) {
	var v Submission
	err := s.db.QueryRow(ctx, `SELECT v.id::text, v.invoice_number, v.status, s.peppol_id, v.payload
		FROM invoices v JOIN sellers s ON s.id = v.seller_id WHERE v.id = $1`, id).
		Scan(&v.ID, &v.InvoiceNumber, &v.Status, &v.SellerPeppolID, &v.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}

// Retryable returns invoices of an import that may be submitted again.
func (s *Store) Retryable(ctx context.Context, importID string) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT id::text FROM invoices WHERE import_id = $1 AND status IN ($2, $3) ORDER BY created_at`,
		importID, StatusValidated, StatusFailed)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}
