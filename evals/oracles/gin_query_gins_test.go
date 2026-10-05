package ginS

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOracleQUERY(t *testing.T) {
	QUERY("/oracle-query", func(c *gin.Context) {
		c.String(http.StatusOK, "query")
	})

	req := httptest.NewRequest(gin.MethodQuery, "/oracle-query", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "query", w.Body.String())
}
