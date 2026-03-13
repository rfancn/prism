// Package monitor provides Prometheus metrics and health check endpoints.
package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "prism"

var (
	// RequestsTotal counts total requests.
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_total",
			Help:      "Total number of requests.",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration records request latency.
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// ActiveConnections tracks active connections.
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_connections",
			Help:      "Number of active connections.",
		},
	)

	// ProxyRequestsTotal counts proxied requests.
	ProxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "proxy_requests_total",
			Help:      "Total number of proxied requests.",
		},
		[]string{"target", "status"},
	)

	// RateLimitHits counts rate limit violations.
	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_hits_total",
			Help:      "Total number of rate limit hits.",
		},
		[]string{"user_id"},
	)

	// AuthFailures counts authentication failures.
	AuthFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_failures_total",
			Help:      "Total number of authentication failures.",
		},
	)

	// RoutesTotal tracks total configured routes.
	RoutesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "routes_total",
			Help:      "Number of configured routes.",
		},
	)

	// APIKeysTotal tracks total API keys.
	APIKeysTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "api_keys_total",
			Help:      "Number of API keys.",
		},
	)
)