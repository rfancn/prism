// Package middleware provides Gin middlewares for Prism.
package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Middleware names
const (
	NameLogger    = "logger"
	NameWhitelist = "whitelist"
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

// Initialize registers all middlewares and sets up global state.
// This should be called once at application startup.
func Initialize() {
	// Register all middlewares
	Register(NameLogger, NewLoggerMiddleware)
	Register(NameWhitelist, NewWhitelistMiddleware)
}