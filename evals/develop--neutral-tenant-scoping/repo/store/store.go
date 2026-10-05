package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"fixture/tenant"
)

// PageSize is the number of invoices in one dashboard page.
const PageSize = 50

// ErrNoTenant reports a customer-facing read without a tenant in the context.
var ErrNoTenant = errors.New("no tenant in context")

// Invoice is a customer invoice.
type Invoice struct {
	ID        string
	TenantID  string
	Status    string
	Cents     int64
	CreatedAt time.Time
	Deleted   bool
}

// Filter selects invoices for Query. An empty field matches every value.
type Filter struct {
	TenantID string
	Status   string
}

// Store holds invoices in memory.
type Store struct {
	mu       sync.Mutex
	invoices []Invoice
}

// New returns a Store that holds the given invoices.
func New(invoices ...Invoice) *Store {
	return &Store{invoices: append([]Invoice(nil), invoices...)}
}

// List returns the first dashboard page of the context tenant's invoices,
// newest first.
func (s *Store) List(ctx context.Context) ([]Invoice, error) {
	id, ok := tenant.FromContext(ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Invoice
	for _, inv := range s.invoices {
		if inv.TenantID == id && !inv.Deleted {
			out = append(out, inv)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if len(out) > PageSize {
		out = out[:PageSize]
	}
	return out, nil
}

// Query returns every invoice that matches f, in insertion order.
func (s *Store) Query(f Filter) []Invoice {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Invoice
	for _, inv := range s.invoices {
		if (f.TenantID == "" || inv.TenantID == f.TenantID) && (f.Status == "" || inv.Status == f.Status) {
			out = append(out, inv)
		}
	}
	return out
}
