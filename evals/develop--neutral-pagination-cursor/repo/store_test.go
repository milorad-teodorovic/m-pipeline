package catalog

import (
	"testing"
	"time"
)

func TestStoreAddDelete(t *testing.T) {
	s := NewStore()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Add(Product{ID: 1, Name: "a", CreatedAt: base})
	s.Add(Product{ID: 2, Name: "b", CreatedAt: base.Add(time.Minute)})
	s.Delete(1)
	all := s.All()
	if len(all) != 1 || all[0].ID != 2 {
		t.Fatalf("All = %+v", all)
	}
}
