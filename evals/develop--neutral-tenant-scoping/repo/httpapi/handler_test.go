package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fixture/store"
)

func TestListInvoices(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := store.New(
		store.Invoice{ID: "a1", TenantID: "acme", Status: "paid", Cents: 100, CreatedAt: now},
		store.Invoice{ID: "b1", TenantID: "globex", Status: "paid", Cents: 200, CreatedAt: now},
	)
	h := NewHandler(s, map[string]string{"key-acme": "acme"})
	req := httptest.NewRequest(http.MethodGet, "/invoices", nil)
	req.Header.Set("X-API-Key", "key-acme")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var got []invoiceJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got) != 1 || got[0].ID != "a1" {
		t.Fatalf("GET /invoices = %d %s", rec.Code, rec.Body.String())
	}
}
