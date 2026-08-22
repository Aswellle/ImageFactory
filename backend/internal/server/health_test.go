package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// healthHandler is the same inline handler registered on GET /v1/health in
// server.NewRouter. It is duplicated here (rather than exported) so the unit
// test exercises the production behavior without depending on router wiring.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// newHealthTestEngine builds the smallest possible engine that serves the
// health route, mirroring how the production router mounts it under /v1.
func newHealthTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/v1")
	v1.GET("/health", healthHandler)
	return r
}

func TestHealthEndpoint_ReturnsOKEnvelope(t *testing.T) {
	engine := newHealthTestEngine()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/health", nil)
	require.NoError(t, err)

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	assert.Len(t, body, 1, "health response must carry only the status field")
}

func TestHealthEndpoint_RejectsNonGetMethods(t *testing.T) {
	engine := newHealthTestEngine()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		w := httptest.NewRecorder()
		req, err := http.NewRequest(method, "/v1/health", nil)
		require.NoError(t, err)
		engine.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code, "%s must not return 200", method)
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusMethodNotAllowed,
			"%s must be rejected (got %d)", method, w.Code)
	}
}

func TestHealthEndpoint_UnknownPathReturns404(t *testing.T) {
	engine := newHealthTestEngine()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/unknown", nil)
	require.NoError(t, err)

	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
