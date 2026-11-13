// Package importer turns an uploaded workbook into invoices:
// parse → map to PINT AE → validate → hand over for submission.
package importer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"invoicely/api/internal/metrics"
	"invoicely/api/internal/pint"
	"invoicely/api/internal/sheet"
	"invoicely/api/internal/store"
	"invoicely/api/internal/submit"
	"invoicely/api/internal/validate"
)

// processingBudget bounds how long one file may take end to end.
const processingBudget = 10 * time.Minute

type Importer struct {
	store     *store.Store
	validator *validate.Validator
	submitter *submit.Submitter
	sources   *SourceCache
	log       *slog.Logger
}

func New(st *store.Store, v *validate.Validator, sub *submit.Submitter, sources *SourceCache, log *slog.Logger) *Importer {
	return &Importer{store: st, validator: v, submitter: sub, sources: sources, log: log}
}

// Import processes one uploaded file and returns the resulting import.
func (im *Importer) Import(seller store.Seller, filename string, data []byte) (store.Import, error) {
	ctx, cancel := context.WithTimeout(context.Background(), processingBudget)
	defer cancel()

	start := time.Now()
	id := uuid.NewString()
	log := im.log.With("import_id", id, "seller_id", seller.ID, "filename", filename)

	if err := im.store.CreateImport(ctx, id, seller.ID, filename); err != nil {
		return store.Import{}, err
	}
	log.Info("import received", "bytes", len(data))
	if err := im.store.SetImportStatus(ctx, id, store.ImportProcessing); err != nil {
		return store.Import{}, err
	}

	wb, err := sheet.Read(data)
	if err != nil {
		log.Warn("import failed", "error", err)
		metrics.ImportsTotal.WithLabelValues("failed").Inc()
		if err := im.store.FinishImport(ctx, id, store.ImportFailed, err.Error(), time.Since(start)); err != nil {
			return store.Import{}, err
		}
		return im.store.GetImport(ctx, id)
	}

	groups := sheet.GroupByInvoice(wb.Rows)
	invoices := make([]*pint.Invoice, len(groups))
	records := make([]store.NewInvoice, len(groups))
	for i, g := range groups {
		inv := pint.Map(g)
		payload, err := json.Marshal(inv)
		if err != nil {
			return store.Import{}, err
		}
		invoices[i] = inv
		records[i] = store.NewInvoice{
			ID:            uuid.NewString(),
			ImportID:      id,
			SellerID:      seller.ID,
			InvoiceNumber: g.InvoiceNumber,
			BuyerName:     inv.Buyer.Name,
			IssueDate:     inv.IssueDate,
			Currency:      inv.Currency,
			TotalExclVAT:  inv.Totals.TaxExclusiveAmount,
			TotalVAT:      inv.Totals.TaxAmount,
			TotalInclVAT:  inv.Totals.TaxInclusiveAmount,
			Payload:       payload,
		}
	}
	if err := im.store.InsertInvoices(ctx, records); err != nil {
		return store.Import{}, err
	}
	if err := im.store.SetImportCounts(ctx, id, len(wb.Rows), len(groups)); err != nil {
		return store.Import{}, err
	}
	metrics.ImportRows.Add(float64(len(wb.Rows)))
	metrics.InvoicesParsed.Add(float64(len(groups)))
	im.sources.Put(id, &Source{Filename: filename, Raw: data, Workbook: wb, Invoices: invoices})
	log.Info("import parsed", "rows", len(wb.Rows), "invoices", len(groups))

	ready := make([]string, 0, len(groups))
	rejected := 0
	for i, g := range groups {
		errs := im.validator.Validate(ctx, validate.Seller{ID: seller.ID, TRN: seller.TRN}, id, g)
		if len(errs) > 0 {
			rejected++
			if err := im.store.MarkRejected(ctx, records[i].ID, "portal", errs, nil,
				"Failed portal validation: "+errs[0].Rule+" "+errs[0].Message); err != nil {
				return store.Import{}, err
			}
			continue
		}
		if err := im.store.SetStatus(ctx, records[i].ID, store.StatusValidated, "Passed portal validation"); err != nil {
			return store.Import{}, err
		}
		ready = append(ready, records[i].ID)
	}

	elapsed := time.Since(start)
	if err := im.store.FinishImport(ctx, id, store.ImportCompleted, "", elapsed); err != nil {
		return store.Import{}, err
	}
	metrics.ImportsTotal.WithLabelValues("completed").Inc()
	metrics.ImportDuration.Observe(elapsed.Seconds())
	log.Info("import completed", "rows", len(wb.Rows), "invoices", len(groups),
		"validated", len(ready), "rejected", rejected, "duration_ms", elapsed.Milliseconds())

	im.submitter.Enqueue(ready...)
	return im.store.GetImport(ctx, id)
}
