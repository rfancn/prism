package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
)

// LoggerMiddleware logs request details.
func NewLoggerMiddleware() (gin.HandlerFunc, error) {
	return func(c *gin.Context) {

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger := sdk.Logger()
		if query != "" {
			path = path + "?" + query
		}

		// Log based on status code
		if status >= 500 {
			logger.Error("request",
				"status", status,
				"method", c.Request.Method,
				"path", path,
				"latency", latency.String(),
				"client_ip", c.ClientIP(),
			)
		} else if status >= 400 {
			logger.Warn("request",
				"status", status,
				"method", c.Request.Method,
				"path", path,
				"latency", latency.String(),
				"client_ip", c.ClientIP(),
			)
		} else {
			logger.Debug("request",
				"status", status,
				"method", c.Request.Method,
				"path", path,
				"latency", latency.String(),
				"client_ip", c.ClientIP(),
			)
		}
	}, nil
}
