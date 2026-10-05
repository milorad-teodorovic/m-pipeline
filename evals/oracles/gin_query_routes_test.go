package gin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOracleRouteQuery(t *testing.T) {
	type search struct {
		Term string `json:"term"`
	}

	router := New()
	router.QUERY("/search", func(c *Context) {
		var body search
		require.NoError(t, c.ShouldBindJSON(&body))
		c.String(http.StatusOK, body.Term)
	})

	req := httptest.NewRequest(MethodQuery, "/search", strings.NewReader(`{"term":"gin"}`))
	req.Header.Set("Content-Type", MIMEJSON)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gin", w.Body.String())
}

func TestOracleRouteQueryNotRegisteredByAny(t *testing.T) {
	router := New()
	router.Any("/test", func(c *Context) {
		c.Status(http.StatusOK)
	})

	w := PerformRequest(router, MethodQuery, "/test")

	assert.Equal(t, http.StatusNotFound, w.Code)
}
