// Package submit delivers validated invoices to the ASP.
package submit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"invoicely/api/internal/metrics"
	"invoicely/api/internal/store"
)

// ASPClient sends documents to the ASP.
type ASPClient struct {
	baseURL string
	http    *http.Client
}

func NewASPClient(baseURL string) *ASPClient {
	return &ASPClient{baseURL: baseURL, http: &http.Client{Timeout: 15 * time.Second}}
}

type Response struct {
	StatusCode int
	Body       []byte
}

// Send posts one invoice payload on behalf of a sender.
func (c *ASPClient) Send(ctx context.Context, sender string, payload []byte) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/documents", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sender-Id", sender)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	metrics.ASPResponses.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
	return &Response{StatusCode: resp.StatusCode, Body: body}, nil
}

// Submitter runs a pool of workers that take invoice ids from a queue.
type Submitter struct {
	store   *store.Store
	asp     *ASPClient
	log     *slog.Logger
	queue   chan string
	workers int
}

func New(st *store.Store, asp *ASPClient, log *slog.Logger, workers int) *Submitter {
	return &Submitter{store: st, asp: asp, log: log, queue: make(chan string, 100000), workers: workers}
}

// Enqueue schedules invoices for submission.
func (s *Submitter) Enqueue(ids ...string) {
	for _, id := range ids {
		metrics.QueueDepth.Inc()
		s.queue <- id
	}
}

// Start launches the workers.
func (s *Submitter) Start(ctx context.Context) {
	for range s.workers {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case id := <-s.queue:
					metrics.QueueDepth.Dec()
					s.submit(ctx, id)
				}
			}
		}()
	}
	s.log.Info("submission workers started", "workers", s.workers)
}

type aspError struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type aspReply struct {
	Status    string     `json:"status"`
	MessageID string     `json:"messageId"`
	Errors    []aspError `json:"errors"`
}

func (s *Submitter) submit(ctx context.Context, id string) {
	inv, err := s.store.GetSubmission(ctx, id)
	if err != nil {
		s.log.Error("load invoice for submission", "invoice_id", id, "error", err)
		return
	}
	if inv.Status != store.StatusValidated && inv.Status != store.StatusFailed {
		return
	}
	if err := s.store.SetStatus(ctx, id, store.StatusSubmitted, "Sent to ASP"); err != nil {
		s.log.Error("update status", "invoice_id", id, "error", err)
		return
	}

	start := time.Now()
	resp, err := s.asp.Send(ctx, inv.SellerPeppolID, inv.Payload)
	if err != nil {
		s.log.Error("asp request failed", "invoice_id", id, "invoice_number", inv.InvoiceNumber, "error", err)
		metrics.Submissions.WithLabelValues("failed").Inc()
		_ = s.store.SetStatus(ctx, id, store.StatusFailed, "Could not reach ASP: "+err.Error())
		return
	}
	s.log.Info("asp response", "invoice_id", id, "invoice_number", inv.InvoiceNumber,
		"status_code", resp.StatusCode, "duration_ms", time.Since(start).Milliseconds())

	var reply aspReply
	_ = json.Unmarshal(resp.Body, &reply)

	switch resp.StatusCode {
	case http.StatusOK:
		metrics.Submissions.WithLabelValues("accepted").Inc()
		if err := s.store.MarkAccepted(ctx, id, reply.MessageID, resp.Body); err != nil {
			s.log.Error("update status", "invoice_id", id, "error", err)
		}
	case http.StatusUnprocessableEntity:
		metrics.Submissions.WithLabelValues("rejected").Inc()
		detail := "Rejected by ASP"
		if len(reply.Errors) > 0 {
			detail = fmt.Sprintf("Rejected by ASP: %s %s", reply.Errors[0].Rule, reply.Errors[0].Message)
		}
		errs := make([]map[string]any, 0, len(reply.Errors))
		for _, e := range reply.Errors {
			errs = append(errs, map[string]any{"rule": e.Rule, "field": "", "message": e.Message, "row": 0})
		}
		if err := s.store.MarkRejected(ctx, id, "asp", errs, resp.Body, detail); err != nil {
			s.log.Error("update status", "invoice_id", id, "error", err)
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		metrics.Submissions.WithLabelValues("failed").Inc()
		_ = s.store.SetStatus(ctx, id, store.StatusFailed, fmt.Sprintf("ASP returned HTTP %d", resp.StatusCode))
	}
}
