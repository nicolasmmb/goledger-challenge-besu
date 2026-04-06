package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "backend/internal/interfaces/http"

	"github.com/stretchr/testify/require"
)

func TestHealthRoute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := httpapi.Chain(mux)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
