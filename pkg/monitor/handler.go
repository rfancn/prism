package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupMetricsHandler sets up the Prometheus metrics endpoint.
func SetupMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsHandler returns a Gin handler for the metrics endpoint.
func MetricsHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthHandler returns a Gin handler for the health check endpoint.
func HealthHandler(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		// TODO: Add actual health checks
		c.JSON(http.StatusOK, HealthResponse{
			Status:  "healthy",
			Version: version,
			Details: map[string]string{
				"database": "connected",
			},
		})
	}
}

// ReadyHandler returns a Gin handler for the readiness endpoint.
func ReadyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Add actual readiness checks
		c.JSON(http.StatusOK, gin.H{
			"ready": true,
		})
	}
}