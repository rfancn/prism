// Package middleware provides Gin middlewares for Prism.
package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
)

// Middleware names
const (
	NameLogger    = "logger"
	NameWhitelist = "whitelist"
	NameAuth      = "auth"
	NameRateLimit = "ratelimit"
	NameProxy     = "proxy"
)

// MiddlewareFunc is the factory function type for creating middlewares.
type MiddlewareFunc func() (gin.HandlerFunc, error)

var registeredMiddlewares = make(map[string]MiddlewareFunc)

// Register registers a middleware factory function.
func Register(name string, fn MiddlewareFunc) {
	registeredMiddlewares[name] = fn
}

// Get retrieves a middleware by name.
func Get(name string) (gin.HandlerFunc, error) {
	fn, exists := registeredMiddlewares[name]
	if !exists {
		return nil, fmt.Errorf("middleware not found: %s", name)
	}
	return fn()
}

// Config holds middleware configuration.
type Config struct {
	// Rate limiting defaults
	DefaultRPS   int
	DefaultBurst int
}

// Initialize registers all middlewares and sets up global state.
// This should be called once at application startup.
func Initialize(cfg *Config) {
	// Register all middlewares
	Register(NameLogger, NewLoggerMiddleware)
	Register(NameWhitelist, NewWhitelistMiddleware)
	Register(NameAuth, NewAuthMiddleware)
	Register(NameRateLimit, NewRateLimitMiddleware)

	// Initialize rate limiter with default values
	if cfg != nil {
		rateLimiter := NewRateLimiter(cfg.DefaultRPS, cfg.DefaultBurst)
		SetGlobalRateLimiter(rateLimiter)
		sdk.Logger().Debug("rate limiter initialized",
			"rps", cfg.DefaultRPS,
			"burst", cfg.DefaultBurst,
		)
	} else {
		// Use default values
		rateLimiter := NewRateLimiter(100, 200)
		SetGlobalRateLimiter(rateLimiter)
	}
}