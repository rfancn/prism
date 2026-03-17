// Package cel provides a CEL (Common Expression Language) expression engine
// for request matching and parameter extraction.
package cel

// EvaluationResult represents the result of CEL expression evaluation.
type EvaluationResult struct {
	// Matched indicates whether the expression evaluated to true
	Matched bool
	// Params contains extracted parameters from the expression
	Params map[string]string
	// Error contains any error that occurred during evaluation
	Error error
}

// MatchContext provides context data for CEL expression matching.
type MatchContext struct {
	// PathParams contains path parameters extracted from the URL
	PathParams map[string]string
	// URLParams contains query parameters from the URL
	URLParams map[string]string
	// Headers contains HTTP request headers
	Headers map[string]string
	// Body contains parsed request body (for JSON requests)
	Body map[string]any
}

// EngineConfig contains configuration options for the CEL engine.
type EngineConfig struct {
	// Timeout is the maximum time for expression evaluation
	Timeout int // in milliseconds
	// MaxExpressionSize is the maximum allowed expression size in bytes
	MaxExpressionSize int
	// EnableExtendedFunctions enables extended function library
	EnableExtendedFunctions bool
}

// DefaultEngineConfig returns the default engine configuration.
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		Timeout:                 1000,  // 1 second
		MaxExpressionSize:       4096,  // 4KB
		EnableExtendedFunctions: false, // disable extended functions for security
	}
}