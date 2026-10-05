package httpapi

import (
	"net/http"

	"fixture/tenant"
)

// Authenticate resolves the X-API-Key header to a tenant and stores it in the
// request context. It answers 401 for a missing or unknown key.
func Authenticate(keys map[string]string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := keys[r.Header.Get("X-API-Key")]
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(tenant.WithTenant(r.Context(), id)))
	})
}
