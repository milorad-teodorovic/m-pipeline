package catalog

import (
	"sort"
	"sync"
	"time"
)

// Product is one catalog item.
type Product struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Store keeps products in memory.
type Store struct {
	mu       sync.Mutex
	products map[int64]Product
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{products: map[int64]Product{}}
}

// Add stores p, replacing any product with the same ID.
func (s *Store) Add(p Product) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.products[p.ID] = p
}

// Delete removes the product with the given ID.
func (s *Store) Delete(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.products, id)
}

// All returns every product ordered by CreatedAt.
func (s *Store) All() []Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Product, 0, len(s.products))
	for _, p := range s.products {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
