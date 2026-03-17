package cel

import (
	"testing"
	"time"

	"github.com/google/cel-go/common/operators"
)

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}
	if engine == nil {
		t.Fatal("NewEngine() returned nil engine")
	}
	if engine.env == nil {
		t.Fatal("Engine env is nil")
	}
	if engine.sandbox == nil {
		t.Fatal("Engine sandbox is nil")
	}
}

func TestNewEngineWithConfig(t *testing.T) {
	config := &EngineConfig{
		Timeout:           2000,
		MaxExpressionSize: 8192,
	}
	engine, err := NewEngineWithConfig(config)
	if err != nil {
		t.Fatalf("NewEngineWithConfig() error = %v", err)
	}
	if engine.config.Timeout != 2000 {
		t.Errorf("Engine config timeout = %d, want 2000", engine.config.Timeout)
	}
}

func TestEngine_Evaluate_BasicExpressions(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantMatch  bool
		wantErr    bool
	}{
		{
			name:       "简单布尔值 true",
			expression: "true",
			ctx:        &MatchContext{},
			wantMatch:  true,
			wantErr:    false,
		},
		{
			name:       "简单布尔值 false",
			expression: "false",
			ctx:        &MatchContext{},
			wantMatch:  false,
			wantErr:    false,
		},
		{
			name:       "字符串相等比较",
			expression: "'hello' == 'hello'",
			ctx:        &MatchContext{},
			wantMatch:  true,
			wantErr:    false,
		},
		{
			name:       "字符串不相等比较",
			expression: "'hello' == 'world'",
			ctx:        &MatchContext{},
			wantMatch:  false,
			wantErr:    false,
		},
		{
			name:       "数字比较",
			expression: "10 > 5",
			ctx:        &MatchContext{},
			wantMatch:  true,
			wantErr:    false,
		},
		{
			name:       "逻辑与操作",
			expression: "true && false",
			ctx:        &MatchContext{},
			wantMatch:  false,
			wantErr:    false,
		},
		{
			name:       "逻辑或操作",
			expression: "true || false",
			ctx:        &MatchContext{},
			wantMatch:  true,
			wantErr:    false,
		},
		{
			name:       "逻辑非操作",
			expression: "!false",
			ctx:        &MatchContext{},
			wantMatch:  true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.wantMatch {
				t.Errorf("Evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestEngine_Evaluate_PathParams(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantMatch  bool
		wantErr    bool
	}{
		{
			name:       "检查路径参数是否存在",
			expression: "'tenant' in path",
			ctx: &MatchContext{
				PathParams: map[string]string{"tenant": "acme"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "检查路径参数值相等",
			expression: "path.tenant == 'acme'",
			ctx: &MatchContext{
				PathParams: map[string]string{"tenant": "acme"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "检查路径参数值不匹配",
			expression: "path.tenant == 'other'",
			ctx: &MatchContext{
				PathParams: map[string]string{"tenant": "acme"},
			},
			wantMatch: false,
			wantErr:   false,
		},
		{
			name:       "多条件组合 - 路径参数和逻辑运算",
			expression: "path.tenant == 'acme' && path.version == 'v1'",
			ctx: &MatchContext{
				PathParams: map[string]string{
					"tenant":  "acme",
					"version": "v1",
				},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.wantMatch {
				t.Errorf("Evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestEngine_Evaluate_URLParams(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantMatch  bool
		wantErr    bool
	}{
		{
			name:       "检查URL参数值",
			expression: "params.action == 'create'",
			ctx: &MatchContext{
				URLParams: map[string]string{"action": "create"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "URL参数多条件",
			expression: "params.action == 'create' || params.action == 'update'",
			ctx: &MatchContext{
				URLParams: map[string]string{"action": "update"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "URL参数不存在检查",
			expression: "!('missing' in params)",
			ctx: &MatchContext{
				URLParams: map[string]string{"action": "create"},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.wantMatch {
				t.Errorf("Evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestEngine_Evaluate_Headers(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantMatch  bool
		wantErr    bool
	}{
		{
			name:       "检查Header值",
			expression: "headers['Content-Type'] == 'application/json'",
			ctx: &MatchContext{
				Headers: map[string]string{"Content-Type": "application/json"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "检查Header存在性",
			expression: "'Authorization' in headers",
			ctx: &MatchContext{
				Headers: map[string]string{"Authorization": "Bearer token"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "Header前缀匹配",
			expression: "headers['Authorization'].startsWith('Bearer')",
			ctx: &MatchContext{
				Headers: map[string]string{"Authorization": "Bearer token123"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "Header后缀匹配",
			expression: "headers['Content-Type'].endsWith('json')",
			ctx: &MatchContext{
				Headers: map[string]string{"Content-Type": "application/json"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "Header包含匹配",
			expression: "headers['Content-Type'].contains('application')",
			ctx: &MatchContext{
				Headers: map[string]string{"Content-Type": "application/json"},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.wantMatch {
				t.Errorf("Evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestEngine_Evaluate_Body(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantMatch  bool
		wantErr    bool
	}{
		{
			name:       "检查Body字段值",
			expression: "body['type'] == 'user'",
			ctx: &MatchContext{
				Body: map[string]any{"type": "user"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "检查Body嵌套字段",
			expression: "body['user']['role'] == 'admin'",
			ctx: &MatchContext{
				Body: map[string]any{
					"user": map[string]any{
						"role": "admin",
					},
				},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:       "检查Body字段存在",
			expression: "'email' in body",
			ctx: &MatchContext{
				Body: map[string]any{"email": "test@example.com"},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.wantMatch {
				t.Errorf("Evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestEngine_Evaluate_ComplexExpressions(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "组合条件 - 路径参数和Header",
			expression: "path.tenant == 'acme' && headers['Content-Type'] == 'application/json'",
			ctx: &MatchContext{
				PathParams: map[string]string{"tenant": "acme"},
				Headers:    map[string]string{"Content-Type": "application/json"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "复杂组合 - 多种参数类型",
			expression: "(path.tenant == 'acme' || path.tenant == 'beta') && params.action == 'create' && 'Authorization' in headers",
			ctx: &MatchContext{
				PathParams: map[string]string{"tenant": "beta"},
				URLParams:  map[string]string{"action": "create"},
				Headers:    map[string]string{"Authorization": "Bearer token"},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "三元表达式",
			expression: "path.tenant == 'acme' ? true : false",
			ctx: &MatchContext{
				PathParams: map[string]string{"tenant": "acme"},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if matched != tt.wantMatch {
				t.Errorf("Evaluate() matched = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

func TestEngine_Evaluate_ErrorCases(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		ctx        *MatchContext
		wantErr    bool
		errContain string
	}{
		{
			name:       "语法错误表达式",
			expression: "path.tenant ==",
			ctx:        &MatchContext{},
			wantErr:    true,
			errContain: "compile error",
		},
		{
			name:       "表达式太长",
			expression: string(make([]byte, 5000)),
			ctx:        &MatchContext{},
			wantErr:    true,
			errContain: "too large",
		},
		{
			name:       "包含禁止模式",
			expression: "import('something')",
			ctx:        &MatchContext{},
			wantErr:    true,
			errContain: "disallowed pattern",
		},
		{
			name:       "包含禁止模式 - exec",
			expression: "exec('cmd')",
			ctx:        &MatchContext{},
			wantErr:    true,
			errContain: "disallowed pattern",
		},
		{
			name:       "返回非布尔值",
			expression: "'hello'",
			ctx:        &MatchContext{},
			wantErr:    true,
			errContain: "boolean",
		},
		{
			name:       "返回数字",
			expression: "42",
			ctx:        &MatchContext{},
			wantErr:    true,
			errContain: "boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := engine.Evaluate(tt.expression, tt.ctx)
			if err == nil {
				t.Errorf("Evaluate() expected error, got nil")
				return
			}
			if tt.errContain != "" {
				if !containsSubstring(err.Error(), tt.errContain) {
					t.Errorf("Evaluate() error = %v, should contain %v", err, tt.errContain)
				}
			}
		})
	}
}

func TestEngine_ValidateExpression(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{
			name:       "有效表达式",
			expression: "path.tenant == 'acme'",
			wantErr:    false,
		},
		{
			name:       "语法错误",
			expression: "path.tenant ==",
			wantErr:    true,
		},
		{
			name:       "禁止模式",
			expression: "__import__('os')",
			wantErr:    true,
		},
		{
			name:       "表达式太长",
			expression: string(make([]byte, 5000)),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateExpression(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_EvaluateWithTimeout(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	// 测试正常执行
	matched, _, err := engine.EvaluateWithTimeout("true", &MatchContext{}, 1*time.Second)
	if err != nil {
		t.Errorf("EvaluateWithTimeout() error = %v", err)
	}
	if !matched {
		t.Error("EvaluateWithTimeout() should return true")
	}
}

func TestEngine_Caching(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	expr := "path.tenant == 'acme'"
	ctx := &MatchContext{
		PathParams: map[string]string{"tenant": "acme"},
	}

	// 第一次执行，应该编译并缓存
	matched1, _, err := engine.Evaluate(expr, ctx)
	if err != nil {
		t.Fatalf("First Evaluate() error = %v", err)
	}

	// 检查缓存
	prg, ok := engine.getFromCache(expr)
	if !ok {
		t.Error("Expression should be cached")
	}
	if prg == nil {
		t.Error("Cached program should not be nil")
	}

	// 第二次执行，应该使用缓存
	matched2, _, err := engine.Evaluate(expr, ctx)
	if err != nil {
		t.Fatalf("Second Evaluate() error = %v", err)
	}

	if matched1 != matched2 {
		t.Errorf("Results should be equal: %v vs %v", matched1, matched2)
	}
}

func TestSandbox_ValidateExpression(t *testing.T) {
	sandbox := NewSandbox()

	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{
			name:       "正常表达式",
			expression: "path.tenant == 'acme'",
			wantErr:    false,
		},
		{
			name:       "包含双下划线",
			expression: "__import__('os')",
			wantErr:    true,
		},
		{
			name:       "包含import",
			expression: "import('module')",
			wantErr:    true,
		},
		{
			name:       "包含exec",
			expression: "exec('command')",
			wantErr:    true,
		},
		{
			name:       "包含system",
			expression: "system('cmd')",
			wantErr:    true,
		},
		{
			name:       "包含runtime",
			expression: "runtime.version()",
			wantErr:    true,
		},
		{
			name:       "包含os.",
			expression: "os.getenv()",
			wantErr:    true,
		},
		{
			name:       "表达式太长",
			expression: string(make([]byte, 5000)),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sandbox.ValidateExpression(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSandbox_IsFunctionAllowed(t *testing.T) {
	sandbox := NewSandbox()

	tests := []struct {
		name      string
		funcName  string
		allowed   bool
	}{
		{"允许 equals", operators.Equals, true},
		{"允许 startsWith", "startsWith", true},
		{"允许 endsWith", "endsWith", true},
		{"允许 contains", "contains", true},
		{"允许 has", "has", true},
		{"允许逻辑与", operators.LogicalAnd, true},
		{"禁止未知函数", "unknownFunc", false},
		{"禁止危险模式", "exec", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sandbox.IsFunctionAllowed(tt.funcName)
			if result != tt.allowed {
				t.Errorf("IsFunctionAllowed(%s) = %v, want %v", tt.funcName, result, tt.allowed)
			}
		})
	}
}

func TestDefaultEngineConfig(t *testing.T) {
	config := DefaultEngineConfig()
	if config.Timeout != 1000 {
		t.Errorf("Default timeout = %d, want 1000", config.Timeout)
	}
	if config.MaxExpressionSize != 4096 {
		t.Errorf("Default MaxExpressionSize = %d, want 4096", config.MaxExpressionSize)
	}
	if config.EnableExtendedFunctions {
		t.Error("Default EnableExtendedFunctions should be false")
	}
}

func TestEngine_NilContext(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	// 使用 nil 上下文
	matched, _, err := engine.Evaluate("true", nil)
	if err != nil {
		t.Errorf("Evaluate with nil context error = %v", err)
	}
	if !matched {
		t.Error("Evaluate with nil context should return true")
	}
}

func TestEngine_EmptyMaps(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	// 使用空的上下文（nil maps）
	ctx := &MatchContext{}
	matched, _, err := engine.Evaluate("true", ctx)
	if err != nil {
		t.Errorf("Evaluate with empty context error = %v", err)
	}
	if !matched {
		t.Error("Evaluate with empty context should return true")
	}
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}