package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type listResponse struct {
	Items []Product `json:"items"`
}

// ListHandler serves GET /products with offset and limit query parameters.
func ListHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 20
		}
		all := s.All()
		if offset > len(all) {
			offset = len(all)
		}
		end := offset + limit
		if end > len(all) {
			end = len(all)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listResponse{Items: all[offset:end]})
	})
}
