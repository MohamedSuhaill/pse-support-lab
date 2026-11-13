// Package metrics holds the Prometheus metrics exposed on /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "invoicely_http_request_duration_seconds",
		Help:    "HTTP request latency by route and status code.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	}, []string{"method", "route", "code"})

	ImportsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "invoicely_imports_total",
		Help: "Imports finished, by final status.",
	}, []string{"status"})

	ImportDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "invoicely_import_duration_seconds",
		Help:    "Time to parse and validate an uploaded file.",
		Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300, 600},
	})

	ImportRows = promauto.NewCounter(prometheus.CounterOpts{
		Name: "invoicely_import_rows_processed_total",
		Help: "Spreadsheet rows read from uploaded files.",
	})

	InvoicesParsed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "invoicely_invoices_parsed_total",
		Help: "Invoices created from uploaded files.",
	})

	ValidationFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "invoicely_validation_failures_total",
		Help: "Portal validation errors, by rule id.",
	}, []string{"rule"})

	ParticipantLookups = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "invoicely_participant_lookup_duration_seconds",
		Help:    "Peppol participant lookups made against the ASP directory.",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"result"})

	Submissions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "invoicely_submissions_total",
		Help: "Invoice submissions to the ASP, by outcome.",
	}, []string{"outcome"})

	ASPResponses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "invoicely_asp_responses_total",
		Help: "HTTP responses received from the ASP, by status code.",
	}, []string{"code"})

	QueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "invoicely_submission_queue_depth",
		Help: "Invoices waiting to be submitted to the ASP.",
	})
)

func init() {
	for _, status := range []string{"completed", "failed"} {
		ImportsTotal.WithLabelValues(status)
	}
	for _, outcome := range []string{"accepted", "rejected", "failed"} {
		Submissions.WithLabelValues(outcome)
	}
	for _, result := range []string{"registered", "not_registered", "error"} {
		ParticipantLookups.WithLabelValues(result)
	}
}
