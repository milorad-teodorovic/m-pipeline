package httpapi

import (
	"encoding/json"
	"net/http"

	"fixture/store"
)

type invoiceJSON struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Cents  int64  `json:"cents"`
}

// NewHandler returns the authenticated invoice API.
func NewHandler(s *store.Store, keys map[string]string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /invoices", func(w http.ResponseWriter, r *http.Request) {
		invoices, err := s.List(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeInvoices(w, invoices)
	})
	return Authenticate(keys, mux)
}

func writeInvoices(w http.ResponseWriter, invoices []store.Invoice) {
	out := make([]invoiceJSON, 0, len(invoices))
	for _, inv := range invoices {
		out = append(out, invoiceJSON{ID: inv.ID, Status: inv.Status, Cents: inv.Cents})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
