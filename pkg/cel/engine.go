package cel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Engine provides a CEL expression engine with sandbox restrictions.
type Engine struct {
	env     *cel.Env
	sandbox *Sandbox
	config  *EngineConfig
	cache   sync.Map // expression cache
	cacheMu sync.RWMutex
}

// NewEngine creates a new CEL engine with default configuration.
func NewEngine() (*Engine, error) {
	return NewEngineWithConfig(DefaultEngineConfig())
}

// NewEngineWithConfig creates a new CEL engine with custom configuration.
func NewEngineWithConfig(config *EngineConfig) (*Engine, error) {
	if config == nil {
		config = DefaultEngineConfig()
	}

	// Create sandbox
	sandbox := NewSandbox()

	// Build CEL environment options
	opts := []cel.EnvOption{
		// Request context variables - each Variable returns an EnvOption
		cel.Variable("path", cel.MapType(cel.StringType, cel.StringType)),
		cel.Variable("params", cel.MapType(cel.StringType, cel.StringType)),
		cel.Variable("headers", cel.MapType(cel.StringType, cel.StringType)),
		cel.Variable("body", cel.MapType(cel.StringType, cel.AnyType)),
		// Convenience variables
		cel.Variable("method", cel.StringType),
		cel.Variable("host", cel.StringType),
		cel.Variable("pathStr", cel.StringType),
		// Enable standard library
		cel.StdLib(),
	}

	// Create environment
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	return &Engine{
		env:     env,
		sandbox: sandbox,
		config:  config,
	}, nil
}

// Evaluate executes a CEL expression with the given context data.
// Returns: matched (whether the expression evaluated to true),
// params (extracted parameters), error (any error that occurred).
func (e *Engine) Evaluate(expression string, ctx *MatchContext) (bool, map[string]string, error) {
	// Validate expression against sandbox rules
	if err := e.sandbox.ValidateExpression(expression); err != nil {
		return false, nil, fmt.Errorf("expression validation failed: %w", err)
	}

	// Check cache for compiled expression
	prg, ok := e.getFromCache(expression)
	if !ok {
		// Compile expression
		ast, iss := e.env.Compile(expression)
		if iss != nil && iss.Err() != nil {
			return false, nil, fmt.Errorf("compile error: %w", iss.Err())
		}

		// Check AST safety
		if err := e.sandbox.CheckASTSafety(ast); err != nil {
			return false, nil, fmt.Errorf("expression safety check failed: %w", err)
		}

		// Create program with cost limit
		var err error
		prg, err = e.env.Program(ast, cel.CostLimit(1000000)) // 1M cost limit
		if err != nil {
			return false, nil, fmt.Errorf("program creation error: %w", err)
		}

		// Cache the compiled program
		e.storeInCache(expression, prg)
	}

	// Prepare input data
	data := e.prepareInputData(ctx)

	// Create context with timeout
	timeout := time.Duration(e.config.Timeout) * time.Millisecond
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Evaluate with context
	out, _, err := prg.Eval(data)
	if err != nil {
		return false, nil, fmt.Errorf("evaluation error: %w", err)
	}

	// Check if context was canceled
	select {
	case <-ctxWithTimeout.Done():
		return false, nil, fmt.Errorf("expression evaluation timed out")
	default:
	}

	// Parse result
	return e.parseResult(out)
}

// EvaluateWithTimeout executes a CEL expression with a custom timeout.
func (e *Engine) EvaluateWithTimeout(expression string, ctx *MatchContext, timeout time.Duration) (bool, map[string]string, error) {
	originalTimeout := e.config.Timeout
	e.config.Timeout = int(timeout.Milliseconds())
	defer func() { e.config.Timeout = originalTimeout }()

	return e.Evaluate(expression, ctx)
}

// parseResult converts the CEL evaluation result to boolean and extracted params.
func (e *Engine) parseResult(out ref.Val) (bool, map[string]string, error) {
	if out == nil {
		return false, nil, fmt.Errorf("nil result from expression")
	}

	// Convert to boolean
	boolVal, ok := out.(types.Bool)
	if !ok {
		return false, nil, fmt.Errorf("expression must evaluate to boolean, got %T", out)
	}

	return bool(boolVal), make(map[string]string), nil
}

// prepareInputData converts MatchContext to a map for CEL evaluation.
func (e *Engine) prepareInputData(ctx *MatchContext) map[string]any {
	if ctx == nil {
		return map[string]any{
			"path":    map[string]string{},
			"params":  map[string]string{},
			"headers": map[string]string{},
			"body":    map[string]any{},
			"method":  "",
			"host":    "",
			"pathStr": "",
		}
	}

	// Ensure maps are not nil
	pathParams := ctx.PathParams
	if pathParams == nil {
		pathParams = make(map[string]string)
	}
	urlParams := ctx.URLParams
	if urlParams == nil {
		urlParams = make(map[string]string)
	}
	headers := ctx.Headers
	if headers == nil {
		headers = make(map[string]string)
	}
	body := ctx.Body
	if body == nil {
		body = make(map[string]any)
	}

	return map[string]any{
		"path":    pathParams,
		"params":  urlParams,
		"headers": headers,
		"body":    body,
		"method":  "",
		"host":    "",
		"pathStr": "",
	}
}

// getFromCache retrieves a compiled program from the cache.
func (e *Engine) getFromCache(expression string) (cel.Program, bool) {
	e.cacheMu.RLock()
	defer e.cacheMu.RUnlock()
	if val, ok := e.cache.Load(expression); ok {
		return val.(cel.Program), true
	}
	return nil, false
}

// storeInCache stores a compiled program in the cache.
func (e *Engine) storeInCache(expression string, prg cel.Program) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()
	e.cache.Store(expression, prg)
}

// ValidateExpression validates an expression without executing it.
// Returns an error if the expression is invalid or violates security rules.
func (e *Engine) ValidateExpression(expression string) error {
	// Validate against sandbox rules
	if err := e.sandbox.ValidateExpression(expression); err != nil {
		return err
	}

	// Try to compile
	ast, iss := e.env.Compile(expression)
	if iss != nil && iss.Err() != nil {
		return fmt.Errorf("compile error: %w", iss.Err())
	}

	// Check AST safety
	return e.sandbox.CheckASTSafety(ast)
}