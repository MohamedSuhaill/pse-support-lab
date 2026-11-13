// Command asp-mock simulates the Accredited Service Provider (ASP) access point
// that Invoicely Lite submits PINT AE invoices to.
//
// Endpoints:
//
//	POST /v1/documents               submit one invoice (header X-Sender-Id)
//	GET  /v1/documents?sender=<id>   list documents received from a sender
//	GET  /v1/documents/{messageId}   one received document
//	GET  /v1/participants/{peppolId} Peppol participant (SMP) lookup
//	GET  /healthz, /metrics
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	customizationID = "urn:peppol:pint:billing-1@ae-1"

	// Fair-use policy per sender: bursts of up to 8 documents, then 2 per second.
	senderBurst      = 8
	senderRatePerSec = 2.0

	// Typical SMP lookup round trip.
	lookupLatency = 200 * time.Millisecond
)

var (
	trnPattern         = regexp.MustCompile(`^1\d{12}03$`)
	participantPattern = regexp.MustCompile(`^0235:1\d{9}$`)
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "asp_requests_total",
		Help: "Requests handled by the ASP, by endpoint and HTTP status code.",
	}, []string{"endpoint", "code"})
	documentsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "asp_documents_total",
		Help: "Documents processed by the ASP, by outcome.",
	}, []string{"outcome"})
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "asp_request_duration_seconds",
		Help:    "ASP request latency.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"endpoint"})
)

// ---------- payload (subset of the PINT AE JSON the portal sends) ----------

type party struct {
	Name     string `json:"name"`
	TRN      string `json:"trn"`
	PeppolID string `json:"peppolId"`
}

type line struct {
	ID          string  `json:"id"`
	NetAmount   float64 `json:"netAmount"`
	VATCategory string  `json:"vatCategory"`
	VATRate     float64 `json:"vatRate"`
	VATAmount   float64 `json:"vatAmount"`
}

type taxSubtotal struct {
	Category      string  `json:"category"`
	Rate          float64 `json:"rate"`
	TaxableAmount float64 `json:"taxableAmount"`
	TaxAmount     float64 `json:"taxAmount"`
}

type totals struct {
	LineExtensionAmount float64 `json:"lineExtensionAmount"`
	TaxExclusiveAmount  float64 `json:"taxExclusiveAmount"`
	TaxAmount           float64 `json:"taxAmount"`
	TaxInclusiveAmount  float64 `json:"taxInclusiveAmount"`
	PayableAmount       float64 `json:"payableAmount"`
}

type document struct {
	CustomizationID string        `json:"customizationId"`
	InvoiceNumber   string        `json:"invoiceNumber"`
	IssueDate       string        `json:"issueDate"`
	Currency        string        `json:"currency"`
	Seller          party         `json:"seller"`
	Buyer           party         `json:"buyer"`
	Lines           []line        `json:"lines"`
	TaxSubtotals    []taxSubtotal `json:"taxSubtotals"`
	Totals          totals        `json:"totals"`
}

type ruleError struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type received struct {
	MessageID     string      `json:"messageId"`
	InvoiceNumber string      `json:"invoiceNumber"`
	Sender        string      `json:"sender"`
	Receiver      string      `json:"receiver"`
	Status        string      `json:"status"`
	Errors        []ruleError `json:"errors,omitempty"`
	ReceivedAt    time.Time   `json:"receivedAt"`
}

// ---------- rate limiting ----------

type bucket struct {
	tokens float64
	last   time.Time
}

type limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

// allow reports whether the sender may send now; otherwise how long to wait.
func (l *limiter) allow(sender string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[sender]
	if !ok {
		b = &bucket{tokens: senderBurst, last: now}
		l.buckets[sender] = b
	}
	b.tokens = math.Min(senderBurst, b.tokens+now.Sub(b.last).Seconds()*senderRatePerSec)
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	wait := time.Duration((1 - b.tokens) / senderRatePerSec * float64(time.Second))
	return false, wait
}

// ---------- store ----------

type store struct {
	mu   sync.RWMutex
	docs map[string]*received
}

func (s *store) add(r *received) {
	s.mu.Lock()
	s.docs[r.MessageID] = r
	s.mu.Unlock()
}

func (s *store) get(id string) (*received, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.docs[id]
	return r, ok
}

func (s *store) bySender(sender string) []*received {
	s.mu.RLock()
	out := make([]*received, 0)
	for _, r := range s.docs {
		if sender == "" || r.Sender == sender {
			out = append(out, r)
		}
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].ReceivedAt.Before(out[j].ReceivedAt) })
	return out
}

// ---------- validation ----------

func cents(v float64) int64 { return int64(math.Round(v * 100)) }

func money(c int64) string { return fmt.Sprintf("%.2f", float64(c)/100) }

func validate(d *document) []ruleError {
	var errs []ruleError
	if d.CustomizationID != customizationID {
		errs = append(errs, ruleError{"PEPPOL-R001", fmt.Sprintf("Specification identifier must be '%s'.", customizationID)})
	}
	if !trnPattern.MatchString(d.Seller.TRN) {
		errs = append(errs, ruleError{"IBR-132-AE", "Seller VAT identifier (IBT-031) must be a TRN: 15 digits, starting with 1 and ending with 03."})
	}
	if !trnPattern.MatchString(d.Buyer.TRN) {
		errs = append(errs, ruleError{"IBR-132-AE", "Buyer VAT identifier (IBT-048) must be a TRN: 15 digits, starting with 1 and ending with 03."})
	}
	if len(d.Lines) == 0 {
		errs = append(errs, ruleError{"BR-16", "An invoice must have at least one invoice line (IBG-25)."})
	}
	var lineSum int64
	for _, l := range d.Lines {
		lineSum += cents(l.NetAmount)
	}
	t := d.Totals
	if lineSum != cents(t.LineExtensionAmount) {
		errs = append(errs, ruleError{"BR-CO-10", fmt.Sprintf(
			"Sum of Invoice line net amount (IBT-106) must equal the sum of all Invoice line net amounts (IBT-131). Got %s, expected %s.",
			money(cents(t.LineExtensionAmount)), money(lineSum))})
	}
	var subtotalTax int64
	for _, s := range d.TaxSubtotals {
		subtotalTax += cents(s.TaxAmount)
	}
	if subtotalTax != cents(t.TaxAmount) {
		errs = append(errs, ruleError{"BR-CO-14", fmt.Sprintf(
			"Invoice total VAT amount (IBT-110) must equal the sum of VAT category tax amounts (IBT-117). Got %s, expected %s.",
			money(cents(t.TaxAmount)), money(subtotalTax))})
	}
	expectedInclusive := cents(t.TaxExclusiveAmount) + cents(t.TaxAmount)
	if cents(t.TaxInclusiveAmount) != expectedInclusive {
		errs = append(errs, ruleError{"BR-CO-15", fmt.Sprintf(
			"Invoice total amount with VAT (IBT-112) must equal Invoice total amount without VAT (IBT-109) + Invoice total VAT amount (IBT-110). Got %s, expected %s.",
			money(cents(t.TaxInclusiveAmount)), money(expectedInclusive))})
	}
	return errs
}

// ---------- http ----------

type server struct {
	log     *slog.Logger
	limiter *limiter
	store   *store
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *server) instrument(endpoint string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		h(rec, r)
		requestsTotal.WithLabelValues(endpoint, strconv.Itoa(rec.code)).Inc()
		requestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds())
	}
}

func (s *server) submit(w http.ResponseWriter, r *http.Request) {
	sender := r.Header.Get("X-Sender-Id")
	if sender == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing X-Sender-Id header"})
		return
	}
	if ok, wait := s.limiter.allow(sender, time.Now()); !ok {
		retryAfter := int(math.Ceil(wait.Seconds()))
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		s.log.Warn("rate limit exceeded", "sender", sender, "retry_after_s", retryAfter)
		documentsTotal.WithLabelValues("throttled").Inc()
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error":   "rate_limited",
			"message": fmt.Sprintf("Too many documents from sender %s. Retry after %ds.", sender, retryAfter),
		})
		return
	}

	var d document
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 5<<20)).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	rec := &received{
		MessageID:     uuid.NewString(),
		InvoiceNumber: d.InvoiceNumber,
		Sender:        sender,
		Receiver:      d.Buyer.PeppolID,
		ReceivedAt:    time.Now().UTC(),
	}
	if errs := validate(&d); len(errs) > 0 {
		rec.Status = "rejected"
		rec.Errors = errs
		s.store.add(rec)
		documentsTotal.WithLabelValues("rejected").Inc()
		s.log.Info("document rejected", "message_id", rec.MessageID, "sender", sender,
			"invoice_number", d.InvoiceNumber, "rule", errs[0].Rule, "error", errs[0].Message)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"status": "rejected", "messageId": rec.MessageID, "errors": errs,
		})
		return
	}
	rec.Status = "accepted"
	s.store.add(rec)
	documentsTotal.WithLabelValues("accepted").Inc()
	s.log.Info("document accepted", "message_id", rec.MessageID, "sender", sender,
		"receiver", d.Buyer.PeppolID, "invoice_number", d.InvoiceNumber)
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "accepted", "messageId": rec.MessageID, "receivedAt": rec.ReceivedAt,
	})
}

func (s *server) listDocuments(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.bySender(r.URL.Query().Get("sender")))
}

func (s *server) getDocument(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.store.get(r.PathValue("id"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *server) participant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	time.Sleep(lookupLatency)
	if !participantPattern.MatchString(id) {
		writeJSON(w, http.StatusNotFound, map[string]any{"participantId": id, "registered": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"participantId": id,
		"registered":    true,
		"documentTypes": []string{customizationID},
	})
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	for _, outcome := range []string{"accepted", "rejected", "throttled"} {
		documentsTotal.WithLabelValues(outcome)
	}
	s := &server{
		log:     log,
		limiter: &limiter{buckets: map[string]*bucket{}},
		store:   &store{docs: map[string]*received{}},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/documents", s.instrument("submit", s.submit))
	mux.HandleFunc("GET /v1/documents", s.instrument("list", s.listDocuments))
	mux.HandleFunc("GET /v1/documents/{id}", s.instrument("get", s.getDocument))
	mux.HandleFunc("GET /v1/participants/{id}", s.instrument("participant", s.participant))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("GET /metrics", promhttp.Handler())

	addr := ":8090"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Info("asp-mock listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
