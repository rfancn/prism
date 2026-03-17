package cel

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Sandbox provides security constraints for CEL expression evaluation.
// It limits available operations and functions to prevent abuse.
type Sandbox struct {
	// allowedFunctions is a set of function names that are permitted
	allowedFunctions map[string]bool
	// disallowedPatterns contains patterns that should be rejected
	disallowedPatterns []string
}

// NewSandbox creates a new sandbox with default security settings.
func NewSandbox() *Sandbox {
	s := &Sandbox{
		allowedFunctions: make(map[string]bool),
		disallowedPatterns: []string{
			"__",
			"import",
			"module",
			"exec",
			"system",
			"runtime",
			"os.",
			"io.",
			"net/",
			"file:",
			"http:",
			"https:",
		},
	}

	// Allow safe functions
	s.addAllowedFunctions(
		// Comparison operators
		operators.Equals,
		operators.NotEquals,
		operators.Less,
		operators.LessEquals,
		operators.Greater,
		operators.GreaterEquals,
		// Logical operators
		operators.LogicalAnd,
		operators.LogicalOr,
		operators.LogicalNot,
		// Conditional (ternary)
		operators.Conditional,
		// Membership operators
		operators.In,
		operators.OldIn,
		// String functions
		"startsWith",
		"endsWith",
		"contains",
		"matches",
		"size",
		"lower",
		"upper",
		"trim",
		// Type functions
		"type",
		// Container operations
		"has",
		// Map/list operations
		"get",
		// Index operators
		operators.Index,
		// Arithmetic (safe for integers)
		operators.Add,
		operators.Subtract,
		operators.Multiply,
		operators.Divide,
		operators.Modulo,
		// Negation
		operators.Negate,
	)

	return s
}

// addAllowedFunctions adds multiple functions to the allowed list.
func (s *Sandbox) addAllowedFunctions(funcs ...string) {
	for _, f := range funcs {
		s.allowedFunctions[f] = true
	}
}

// IsFunctionAllowed checks if a function is permitted in the sandbox.
func (s *Sandbox) IsFunctionAllowed(name string) bool {
	// Check against disallowed patterns first
	for _, pattern := range s.disallowedPatterns {
		if strings.Contains(name, pattern) {
			return false
		}
	}
	return s.allowedFunctions[name]
}

// ValidateExpression validates an expression for security concerns.
// Returns an error if the expression contains disallowed content.
func (s *Sandbox) ValidateExpression(expr string) error {
	// Check expression size
	if len(expr) > 4096 {
		return fmt.Errorf("expression too large: %d bytes (max 4096)", len(expr))
	}

	// Check for disallowed patterns
	for _, pattern := range s.disallowedPatterns {
		if strings.Contains(expr, pattern) {
			return fmt.Errorf("expression contains disallowed pattern: %s", pattern)
		}
	}

	return nil
}

// safeRegexMatch performs a safe regex match with ReDoS protection.
func safeRegexMatch(str, pattern string) (bool, error) {
	// Validate pattern length
	if len(pattern) > 256 {
		return false, fmt.Errorf("regex pattern too long: %d bytes (max 256)", len(pattern))
	}

	// Check for potentially dangerous regex patterns (ReDoS)
	dangerousPatterns := []string{
		"(.*)*",
		"(.+)+",
		"(.?)+",
		"(.+)*",
		"(.*)+",
		"\\d+\\d+",
		"\\w+\\w+",
	}
	for _, dp := range dangerousPatterns {
		if strings.Contains(pattern, dp) {
			return false, fmt.Errorf("potentially dangerous regex pattern detected")
		}
	}

	// Compile and match
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return re.MatchString(str), nil
}

// CheckASTSafety performs additional safety checks on the parsed AST.
func (s *Sandbox) CheckASTSafety(ast *cel.Ast) error {
	// Get the checked expression
	checked := ast.Expr()
	if checked == nil {
		return nil
	}

	// Recursively check the expression tree
	return s.checkExprSafety(checked)
}

// checkExprSafety recursively checks an expression for safety.
func (s *Sandbox) checkExprSafety(expr *exprpb.Expr) error {
	if expr == nil {
		return nil
	}

	switch e := expr.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		// Check function calls
		funcName := e.CallExpr.Function
		if !s.IsFunctionAllowed(funcName) {
			return fmt.Errorf("function not allowed: %s", funcName)
		}
		// Check arguments
		for _, arg := range e.CallExpr.Args {
			if err := s.checkExprSafety(arg); err != nil {
				return err
			}
		}

	case *exprpb.Expr_ComprehensionExpr:
		// Comprehensions are allowed but need size limits
		// (handled by timeout in the engine)
		if e.ComprehensionExpr != nil {
			if err := s.checkExprSafety(e.ComprehensionExpr.IterRange); err != nil {
				return err
			}
			if err := s.checkExprSafety(e.ComprehensionExpr.AccuInit); err != nil {
				return err
			}
			if err := s.checkExprSafety(e.ComprehensionExpr.LoopCondition); err != nil {
				return err
			}
			if err := s.checkExprSafety(e.ComprehensionExpr.LoopStep); err != nil {
				return err
			}
		}

	case *exprpb.Expr_SelectExpr:
		// Check field selection
		if e.SelectExpr.TestOnly {
			// has() macro is safe
			return nil
		}
		return s.checkExprSafety(e.SelectExpr.Operand)

	case *exprpb.Expr_ListExpr:
		// Check list elements
		for _, elem := range e.ListExpr.Elements {
			if err := s.checkExprSafety(elem); err != nil {
				return err
			}
		}

	case *exprpb.Expr_StructExpr:
		// Check struct fields (entries)
		for _, entry := range e.StructExpr.Entries {
			if err := s.checkExprSafety(entry.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// celSafeRegexMatch provides a CEL-compatible safe regex match function.
func celSafeRegexMatch(args ...ref.Val) ref.Val {
	if len(args) != 2 {
		return types.NewErr("matches requires 2 arguments")
	}

	str, ok := args[0].(types.String)
	if !ok {
		return types.NewErr("first argument must be a string")
	}

	pattern, ok := args[1].(types.String)
	if !ok {
		return types.NewErr("second argument must be a string")
	}

	matched, err := safeRegexMatch(string(str), string(pattern))
	if err != nil {
		return types.NewErr("regex match error: %v", err)
	}

	return types.Bool(matched)
}