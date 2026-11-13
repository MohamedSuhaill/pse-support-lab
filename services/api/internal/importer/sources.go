package importer

import (
	"sync"
	"time"

	"invoicely/api/internal/pint"
	"invoicely/api/internal/sheet"
)

// Source is an uploaded file together with what was read from it.
type Source struct {
	Filename string
	Raw      []byte
	Workbook *sheet.Workbook
	Invoices []*pint.Invoice
	StoredAt time.Time
}

// SourceCache keeps uploaded files in memory so the batch page can show the
// source rows and offer the original file for download.
type SourceCache struct {
	mu      sync.RWMutex
	entries map[string]*Source
}

func NewSourceCache() *SourceCache {
	return &SourceCache{entries: map[string]*Source{}}
}

func (c *SourceCache) Put(importID string, s *Source) {
	s.StoredAt = time.Now()
	c.mu.Lock()
	c.entries[importID] = s
	c.mu.Unlock()
}

func (c *SourceCache) Get(importID string) (*Source, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.entries[importID]
	return s, ok
}
