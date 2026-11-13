// Package directory checks Peppol participant registration through the ASP.
package directory

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"invoicely/api/internal/metrics"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

// IsRegistered reports whether a Peppol participant id can receive invoices.
func (c *Client) IsRegistered(ctx context.Context, participantID string) (bool, error) {
	start := time.Now()
	result := "error"
	defer func() {
		metrics.ParticipantLookups.WithLabelValues(result).Observe(time.Since(start).Seconds())
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v1/participants/"+url.PathEscape(participantID), nil)
	if err != nil {
		return false, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		result = "registered"
		return true, nil
	case http.StatusNotFound:
		result = "not_registered"
		return false, nil
	default:
		return false, fmt.Errorf("participant lookup returned HTTP %d", resp.StatusCode)
	}
}
