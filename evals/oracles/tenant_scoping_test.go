package fixture_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fixture/httpapi"
	"fixture/store"
)

type evalInvoice struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Cents  int64  `json:"cents"`
}

func evalFixture() http.Handler {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var invoices []store.Invoice
	for i := 0; i < 60; i++ {
		invoices = append(invoices, store.Invoice{ID: fmt.Sprintf("acme-u%02d", i), TenantID: "acme", Status: "unpaid", Cents: int64(100 + i), CreatedAt: base.Add(time.Duration(i) * time.Hour), Deleted: i%20 == 7})
	}
	for i := 0; i < 5; i++ {
		invoices = append(invoices, store.Invoice{ID: fmt.Sprintf("acme-p%02d", i), TenantID: "acme", Status: "paid", Cents: 1, CreatedAt: base.Add(time.Duration(i) * time.Minute)})
	}
	for i := 0; i < 10; i++ {
		invoices = append(invoices, store.Invoice{ID: fmt.Sprintf("globex-u%02d", i), TenantID: "globex", Status: "unpaid", Cents: 9, CreatedAt: base.Add(time.Duration(i) * time.Hour)})
	}
	return httpapi.NewHandler(store.New(invoices...), map[string]string{"key-acme": "acme", "key-globex": "globex"})
}

func evalGet(t *testing.T, h http.Handler, key, url string) (int, []evalInvoice) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		return rec.Code, nil
	}
	var got []evalInvoice
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("GET %s body %q is not a JSON array of invoices: %v", url, rec.Body.String(), err)
	}
	return rec.Code, got
}

func TestEvalByStatusScopesTenant(t *testing.T) {
	h := evalFixture()
	for _, url := range []string{"/invoices/by-status?status=unpaid", "/invoices/by-status?status=unpaid&tenant_id=globex"} {
		code, got := evalGet(t, h, "key-acme", url)
		if code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200", url, code)
		}
		if len(got) != 57 {
			t.Fatalf("GET %s returned %d invoices, want the tenant's 57 undeleted unpaid invoices", url, len(got))
		}
		for _, inv := range got {
			if inv.Status != "unpaid" || inv.ID[:5] != "acme-" {
				t.Fatalf("GET %s returned %+v, which is outside the caller's tenant or status", url, inv)
			}
		}
		for i := 1; i < len(got); i++ {
			if got[i-1].Cents <= got[i].Cents {
				t.Fatalf("GET %s is not newest first at position %d", url, i)
			}
		}
		for _, inv := range got {
			if inv.ID == "acme-u07" || inv.ID == "acme-u27" || inv.ID == "acme-u47" {
				t.Fatalf("GET %s returned deleted invoice %s", url, inv.ID)
			}
		}
	}
	code, got := evalGet(t, h, "key-globex", "/invoices/by-status?status=unpaid&tenant_id=acme")
	if code != http.StatusOK || len(got) != 10 {
		t.Fatalf("globex by-status = %d with %d invoices, want 200 with 10", code, len(got))
	}
}

func TestEvalByStatusAuthAndEmpty(t *testing.T) {
	h := evalFixture()
	if code, _ := evalGet(t, h, "", "/invoices/by-status?status=unpaid&tenant_id=acme"); code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated by-status = %d, want 401", code)
	}
	code, got := evalGet(t, h, "key-globex", "/invoices/by-status?status=paid")
	if code != http.StatusOK || len(got) != 0 {
		t.Fatalf("by-status with no matches = %d with %d invoices, want 200 with none", code, len(got))
	}
}
