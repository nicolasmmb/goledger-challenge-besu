package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"backend/internal/infrastructure/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func Logging(log *slog.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(sw, r)
			dur := time.Since(start)

			route := r.URL.Path
			status := strconv.Itoa(sw.code)
			m.ObserveHTTP(route, r.Method, status, dur)
			log.Info("http request",
				"request_id", GetRequestID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.code,
				"duration_ms", dur.Milliseconds(),
			)
		})
	}
}
