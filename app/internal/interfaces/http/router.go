package httpapi

import (
	"net/http"

	"backend/internal/interfaces/http/handlers"
)

func NewRouter(
	health *handlers.HealthHandler,
	value *handlers.ValueHandler,
	tx *handlers.TransactionHandler,
	metricsHandler http.Handler,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Healthz)
	mux.HandleFunc("GET /readyz", health.Readyz)
	mux.HandleFunc("POST /v1/value", value.SetValue)
	mux.HandleFunc("GET /v1/value", value.GetValue)
	mux.HandleFunc("POST /v1/sync", value.Sync)
	mux.HandleFunc("GET /v1/check", value.Check)
	mux.HandleFunc("GET /v1/transactions/{txHash}", tx.GetStatus)
	mux.Handle("GET /metrics", metricsHandler)
	return mux
}

func Chain(h http.Handler, wrappers ...func(http.Handler) http.Handler) http.Handler {
	for i := len(wrappers) - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	return h
}
