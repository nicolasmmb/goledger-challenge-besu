package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	requestsTotal     *prometheus.CounterVec
	requestLatency    *prometheus.HistogramVec
	rpcErrorsTotal    *prometheus.CounterVec
	statusTransitions *prometheus.CounterVec
}

func New(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		}, []string{"route", "method", "status"}),
		requestLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method"}),
		rpcErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "blockchain_rpc_errors_total",
			Help: "Total blockchain RPC errors",
		}, []string{"operation"}),
		statusTransitions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tx_status_transitions_total",
			Help: "Total transaction status transitions",
		}, []string{"from", "to"}),
	}

	registry.MustRegister(
		m.requestsTotal,
		m.requestLatency,
		m.rpcErrorsTotal,
		m.statusTransitions,
	)
	return m
}

func Handler(registry *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func (m *Metrics) ObserveHTTP(route string, method string, status string, d time.Duration) {
	m.requestsTotal.WithLabelValues(route, method, status).Inc()
	m.requestLatency.WithLabelValues(route, method).Observe(d.Seconds())
}

func (m *Metrics) IncRPCError(operation string) {
	m.rpcErrorsTotal.WithLabelValues(operation).Inc()
}

func (m *Metrics) IncTransition(from string, to string) {
	m.statusTransitions.WithLabelValues(from, to).Inc()
}
