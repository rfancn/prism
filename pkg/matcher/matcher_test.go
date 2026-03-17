package matcher

import (
	"bytes"
	"database/sql"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// 测试辅助函数
func setupTestContext(method, path string, body io.Reader, contentType string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.Request = req
	return c
}

func newTestCelEngine(t *testing.T) *cel.Engine {
	engine, err := cel.NewEngine()
	if err != nil {
		t.Fatalf("创建CEL引擎失败: %v", err)
	}
	return engine
}

// ============== matcher.go 测试 ==============

func TestNewFactory(t *testing.T) {
	celEngine := newTestCelEngine(t)
	factory := NewFactory(celEngine, nil)

	if factory == nil {
		t.Fatal("工厂不应为空")
	}
	if factory.celEngine == nil {
		t.Fatal("CEL引擎不应为空")
	}
}

func TestFactoryCreate(t *testing.T) {
	celEngine := newTestCelEngine(t)
	factory := NewFactory(celEngine, nil)

	tests := []struct {
		matchType string
		wantNil   bool
	}{
		{MatchTypeParamPath, false},
		{MatchTypeURLParam, false},
		{MatchTypeRequestBody, false},
		{MatchTypeRequestForm, false},
		{MatchTypePlugin, false},
		{"unknown", true},
	}

	for _, tt := range tests {
		matcher := factory.Create(tt.matchType)
		if (matcher == nil) != tt.wantNil {
			t.Errorf("Create(%s) = %v, want nil: %v", tt.matchType, matcher, tt.wantNil)
		}
	}
}

func TestIsValidMatchType(t *testing.T) {
	tests := []struct {
		matchType string
		want      bool
	}{
		{MatchTypeParamPath, true},
		{MatchTypeURLParam, true},
		{MatchTypeRequestBody, true},
		{MatchTypeRequestForm, true},
		{MatchTypePlugin, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsValidMatchType(tt.matchType)
		if got != tt.want {
			t.Errorf("IsValidMatchType(%s) = %v, want %v", tt.matchType, got, tt.want)
		}
	}
}

// ============== helper.go 测试 ==============

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    map[string]string
	}{
		{
			path:    "/users/123/orders/456",
			pattern: "/users/{id}/orders/{orderId}",
			want:    map[string]string{"id": "123", "orderId": "456"},
		},
		{
			path:    "/api/v1/resource",
			pattern: "/api/{version}/resource",
			want:    map[string]string{"version": "v1"},
		},
		{
			path:    "/users/123",
			pattern: "/users/{id}",
			want:    map[string]string{"id": "123"},
		},
		{
			path:    "/static/path",
			pattern: "/static/path",
			want:    map[string]string{},
		},
		{
			path:    "/a/b/c",
			pattern: "/x/y/z",
			want:    map[string]string{},
		},
	}

	for _, tt := range tests {
		got := extractPathParams(tt.path, tt.pattern)
		if len(got) != len(tt.want) {
			t.Errorf("extractPathParams(%s, %s) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			continue
		}
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("extractPathParams(%s, %s)[%s] = %s, want %s", tt.path, tt.pattern, k, got[k], v)
			}
		}
	}
}

func TestMatchPathPattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"/users/123/orders/456", "/users/{id}/orders/{orderId}", true},
		{"/api/v1/resource", "/api/{version}/resource", true},
		{"/users/123", "/users/{id}", true},
		{"/static/path", "/static/path", true},
		{"/a/b/c", "/x/y/z", false},
		{"/users/123/orders", "/users/{id}/orders/{orderId}", false},
		{"/users/123/orders/456/extra", "/users/{id}/orders/{orderId}", false},
	}

	for _, tt := range tests {
		got := matchPathPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPathPattern(%s, %s) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestGetURLParams(t *testing.T) {
	c := setupTestContext("GET", "/test?foo=bar&baz=qux", nil, "")

	params := getURLParams(c)

	if params["foo"] != "bar" {
		t.Errorf("params[foo] = %s, want bar", params["foo"])
	}
	if params["baz"] != "qux" {
		t.Errorf("params[baz] = %s, want qux", params["baz"])
	}
}

func TestGetHeaders(t *testing.T) {
	c := setupTestContext("GET", "/test", nil, "")
	c.Request.Header.Set("X-Custom-Header", "custom-value")
	c.Request.Header.Set("Authorization", "Bearer token123")

	headers := getHeaders(c)

	if headers["X-Custom-Header"] != "custom-value" {
		t.Errorf("headers[X-Custom-Header] = %s, want custom-value", headers["X-Custom-Header"])
	}
	if headers["Authorization"] != "Bearer token123" {
		t.Errorf("headers[Authorization] = %s, want Bearer token123", headers["Authorization"])
	}
}

func TestGetBody(t *testing.T) {
	body := `{"name": "test", "value": 123}`
	c := setupTestContext("POST", "/test", bytes.NewBufferString(body), "application/json")

	result := getBody(c)

	if result["name"] != "test" {
		t.Errorf("body[name] = %v, want test", result["name"])
	}
}

func TestMergeParams(t *testing.T) {
	params1 := map[string]string{"a": "1", "b": "2"}
	params2 := map[string]string{"c": "3", "d": "4"}
	params3 := map[string]string{"e": "5"}

	result := mergeParams(params1, params2, params3)

	if len(result) != 5 {
		t.Errorf("mergeParams returned %d params, want 5", len(result))
	}
	if result["a"] != "1" || result["c"] != "3" || result["e"] != "5" {
		t.Errorf("mergeParams = %v, want combined params", result)
	}
}

func TestConvertPatternToGin(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"/users/{id}", "/users/:id"},
		{"/users/{id}/orders/{orderId}", "/users/:id/orders/:orderId"},
		{"/api/{version}", "/api/:version"},
		{"/static/path", "/static/path"},
	}

	for _, tt := range tests {
		got := convertPatternToGin(tt.input)
		if got != tt.output {
			t.Errorf("convertPatternToGin(%s) = %s, want %s", tt.input, got, tt.output)
		}
	}
}

func TestConvertPatternFromGin(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"/users/:id", "/users/{id}"},
		{"/users/:id/orders/:orderId", "/users/{id}/orders/{orderId}"},
		{"/api/:version", "/api/{version}"},
		{"/static/path", "/static/path"},
	}

	for _, tt := range tests {
		got := convertPatternFromGin(tt.input)
		if got != tt.output {
			t.Errorf("convertPatternFromGin(%s) = %s, want %s", tt.input, got, tt.output)
		}
	}
}

// ============== param_path.go 测试 ==============

func TestParamPathMatcher_Match(t *testing.T) {
	celEngine := newTestCelEngine(t)
	matcher := NewParamPathMatcher(celEngine)

	tests := []struct {
		name     string
		path     string
		rule     *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "nil rule",
			path: "/users/123",
			rule: nil,
			wantMatch: false,
			wantErr: true,
		},
		{
			name: "empty path pattern",
			path: "/users/123",
			rule: &db.RouteRule{
				PathPattern: sql.NullString{Valid: false},
			},
			wantMatch: false,
			wantErr: true,
		},
		{
			name: "path not match",
			path: "/users/123/orders/456",
			rule: &db.RouteRule{
				PathPattern: sql.NullString{String: "/products/{id}", Valid: true},
			},
			wantMatch: false,
			wantErr: false,
		},
		{
			name: "path match without CEL",
			path: "/users/123",
			rule: &db.RouteRule{
				PathPattern: sql.NullString{String: "/users/{id}", Valid: true},
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch: true,
			wantErr: false,
		},
		{
			name: "path match with CEL true",
			path: "/users/123",
			rule: &db.RouteRule{
				PathPattern: sql.NullString{String: "/users/{id}", Valid: true},
				CelExpression: sql.NullString{String: "true", Valid: true},
			},
			wantMatch: true,
			wantErr: false,
		},
		{
			name: "path match with CEL false",
			path: "/users/123",
			rule: &db.RouteRule{
				PathPattern: sql.NullString{String: "/users/{id}", Valid: true},
				CelExpression: sql.NullString{String: "false", Valid: true},
			},
			wantMatch: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("GET", tt.path, nil, "")
			result := matcher.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== url_param.go 测试 ==============

func TestURLParamMatcher_Match(t *testing.T) {
	celEngine := newTestCelEngine(t)
	matcher := NewURLParamMatcher(celEngine)

	tests := []struct {
		name      string
		path      string
		rule      *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "nil rule",
			path: "/test?foo=bar",
			rule: nil,
			wantMatch: false,
			wantErr: true,
		},
		{
			name: "empty CEL expression",
			path: "/test?foo=bar",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch: false,
			wantErr: true,
		},
		{
			name: "CEL expression true",
			path: "/test?foo=bar",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "true", Valid: true},
			},
			wantMatch: true,
			wantErr: false,
		},
		{
			name: "CEL expression false",
			path: "/test?foo=bar",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "false", Valid: true},
			},
			wantMatch: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("GET", tt.path, nil, "")
			result := matcher.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== request_body.go 测试 ==============

func TestRequestBodyMatcher_Match(t *testing.T) {
	celEngine := newTestCelEngine(t)
	matcher := NewRequestBodyMatcher(celEngine)

	tests := []struct {
		name      string
		body      string
		rule      *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "nil rule",
			body: `{"name": "test"}`,
			rule: nil,
			wantMatch: false,
			wantErr: true,
		},
		{
			name: "empty CEL expression",
			body: `{"name": "test"}`,
			rule: &db.RouteRule{
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch: false,
			wantErr: true,
		},
		{
			name: "CEL expression true",
			body: `{"name": "test"}`,
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "true", Valid: true},
			},
			wantMatch: true,
			wantErr: false,
		},
		{
			name: "CEL expression false",
			body: `{"name": "test"}`,
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "false", Valid: true},
			},
			wantMatch: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("POST", "/test", bytes.NewBufferString(tt.body), "application/json")
			result := matcher.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== request_form.go 测试 ==============

func TestRequestFormMatcher_Match(t *testing.T) {
	celEngine := newTestCelEngine(t)
	matcher := NewRequestFormMatcher(celEngine)

	tests := []struct {
		name      string
		formData  string
		rule      *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name:     "nil rule",
			formData: "name=test&value=123",
			rule:     nil,
			wantMatch: false,
			wantErr: true,
		},
		{
			name:     "empty CEL expression",
			formData: "name=test&value=123",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch: false,
			wantErr: true,
		},
		{
			name:     "CEL expression true",
			formData: "name=test&value=123",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "true", Valid: true},
			},
			wantMatch: true,
			wantErr: false,
		},
		{
			name:     "CEL expression false",
			formData: "name=test&value=123",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "false", Valid: true},
			},
			wantMatch: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("POST", "/test", bytes.NewBufferString(tt.formData), "application/x-www-form-urlencoded")
			result := matcher.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== plugin.go 测试 ==============

func TestPluginMatcher_Match_NilRule(t *testing.T) {
	matcher := NewPluginMatcher(nil)
	c := setupTestContext("GET", "/test", nil, "")

	result := matcher.Match(c, nil)

	if result.Matched {
		t.Error("Matched should be false for nil rule")
	}
	if result.Error != ErrNilRule {
		t.Errorf("Error should be ErrNilRule, got %v", result.Error)
	}
}

func TestPluginMatcher_Match_NilPluginManager(t *testing.T) {
	matcher := NewPluginMatcher(nil)
	c := setupTestContext("GET", "/test", nil, "")
	rule := &db.RouteRule{
		PluginName: sql.NullString{String: "test-plugin", Valid: true},
	}

	result := matcher.Match(c, rule)

	if result.Matched {
		t.Error("Matched should be false for nil plugin manager")
	}
	if result.Error != ErrNilPluginManager {
		t.Errorf("Error should be ErrNilPluginManager, got %v", result.Error)
	}
}

func TestPluginMatcher_Match_EmptyPluginName(t *testing.T) {
	// 这个测试需要mock plugin.Manager，暂时跳过
	// 实际实现中可以使用mock库
}

// ============== errors.go 测试 ==============

func TestErrorMessages(t *testing.T) {
	errors := []error{
		ErrNilRule,
		ErrEmptyPathPattern,
		ErrEmptyCelExpression,
		ErrEmptyPluginName,
		ErrNilPluginManager,
		ErrPluginNotFound,
		ErrNilCelEngine,
	}

	for _, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error message should not be empty: %v", err)
		}
	}
}