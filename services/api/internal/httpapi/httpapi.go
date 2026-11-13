// Package httpapi exposes the REST API used by the web portal and integrations.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"invoicely/api/internal/importer"
	"invoicely/api/internal/metrics"
	"invoicely/api/internal/store"
	"invoicely/api/internal/submit"
)

const maxUploadBytes = 20 << 20

type Server struct {
	store     *store.Store
	importer  *importer.Importer
	sources   *importer.SourceCache
	submitter *submit.Submitter
	log       *slog.Logger
}

func New(st *store.Store, im *importer.Importer, sources *importer.SourceCache, sub *submit.Submitter, log *slog.Logger) *Server {
	return &Server{store: st, importer: im, sources: sources, submitter: sub, log: log}
}

// Handler returns the HTTP handler with routes and access logging.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /v1/sellers", s.listSellers)
	mux.HandleFunc("POST /v1/imports", s.createImport)
	mux.HandleFunc("GET /v1/imports", s.listImports)
	mux.HandleFunc("GET /v1/imports/{id}", s.getImport)
	mux.HandleFunc("GET /v1/imports/{id}/rows", s.importRows)
	mux.HandleFunc("GET /v1/imports/{id}/source", s.importSource)
	mux.HandleFunc("POST /v1/imports/{id}/retry", s.retryImport)
	mux.HandleFunc("GET /v1/invoices/{id}", s.getInvoice)
	mux.HandleFunc("POST /v1/invoices/{id}/retry", s.retryInvoice)
	return s.observe(mux)
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *Server) observe(next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		if i := strings.IndexByte(route, ' '); i >= 0 {
			route = route[i+1:]
		}
		elapsed := time.Since(start)
		metrics.HTTPDuration.WithLabelValues(r.Method, route, strconv.Itoa(rec.code)).Observe(elapsed.Seconds())
		if route == "/metrics" || route == "/healthz" {
			return
		}
		s.log.Info("http request", "method", r.Method, "path", r.URL.Path, "route", route,
			"status", rec.code, "duration_ms", elapsed.Milliseconds())
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) internalError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("request failed", "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func validID(id string) bool { return uuid.Validate(id) == nil }

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listSellers(w http.ResponseWriter, r *http.Request) {
	sellers, err := s.store.Sellers(r.Context())
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, sellers)
}

func (s *Server) createImport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "expected a multipart form with 'file' and 'seller_id' (max 20 MB)")
		return
	}
	sellerID, err := strconv.ParseInt(r.FormValue("seller_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "seller_id is required")
		return
	}
	seller, err := s.store.Seller(r.Context(), sellerID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "unknown seller_id")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()
	if !strings.EqualFold(filepath.Ext(header.Filename), ".xlsx") {
		writeError(w, http.StatusBadRequest, "only .xlsx files are supported")
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read the uploaded file")
		return
	}

	imp, err := s.importer.Import(seller, filepath.Base(header.Filename), data)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, imp)
}

func (s *Server) listImports(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 500 {
		limit = v
	}
	imports, err := s.store.ListImports(r.Context(), limit)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, imports)
}

func (s *Server) getImport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		writeError(w, http.StatusNotFound, "import not found")
		return
	}
	imp, err := s.store.GetImport(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "import not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, imp)
}

func (s *Server) importRows(w http.ResponseWriter, r *http.Request) {
	src, ok := s.sources.Get(r.PathValue("id"))
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"available": false})
		return
	}
	rows, err := src.Workbook.DisplayRows()
	if err != nil || len(rows) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"available": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"available": true, "headers": rows[0], "rows": rows[1:]})
}

func (s *Server) importSource(w http.ResponseWriter, r *http.Request) {
	src, ok := s.sources.Get(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "the original file is no longer available")
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="`+src.Filename+`"`)
	_, _ = w.Write(src.Raw)
}

func (s *Server) retryImport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		writeError(w, http.StatusNotFound, "import not found")
		return
	}
	ids, err := s.store.Retryable(r.Context(), id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	s.submitter.Enqueue(ids...)
	s.log.Info("retry requested", "import_id", id, "requeued", len(ids))
	writeJSON(w, http.StatusOK, map[string]int{"requeued": len(ids)})
}

func (s *Server) getInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	inv, err := s.store.GetInvoice(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (s *Server) retryInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID(id) {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	inv, err := s.store.GetSubmission(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "invoice not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if inv.Status != store.StatusValidated && inv.Status != store.StatusFailed {
		writeError(w, http.StatusConflict, "only invoices in status Validated or Failed can be retried (current status: "+inv.Status+")")
		return
	}
	s.submitter.Enqueue(id)
	s.log.Info("retry requested", "invoice_id", id)
	writeJSON(w, http.StatusOK, map[string]int{"requeued": 1})
}
