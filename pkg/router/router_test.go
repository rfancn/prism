package router

import (
	"bytes"
	"database/sql"
	"io"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
	"github.com/rfancn/prism/pkg/matcher"
	"github.com/rfancn/prism/pkg/plugin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============== 测试辅助函数 ==============

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

// ============== extractSourceName 测试 ==============

func TestExtractSourceName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/weixin/orders/123", "weixin"},
		{"/kdniao/api/v1/track", "kdniao"},
		{"/source1", "source1"},
		{"/source1/", "source1"},
		{"", ""},
		{"/", ""},
		{"no-prefix", "no-prefix"},
	}

	for _, tt := range tests {
		result := extractSourceName(tt.path)
		if result != tt.expected {
			t.Errorf("extractSourceName(%s) = %s, want %s", tt.path, result, tt.expected)
		}
	}
}

// ============== 匹配器工厂测试 ==============

func TestMatcherFactory_Create(t *testing.T) {
	celEngine := newTestCelEngine(t)
	pluginMgr := plugin.NewManager("")
	factory := matcher.NewFactory(celEngine, pluginMgr)

	tests := []struct {
		matchType string
		wantNil   bool
	}{
		{matcher.MatchTypePathParam, false},
		{matcher.MatchTypeURLParam, false},
		{matcher.MatchTypeRequestBody, false},
		{matcher.MatchTypeRequestForm, false},
		{matcher.MatchTypePlugin, false},
		{"unknown_type", true},
	}

	for _, tt := range tests {
		m := factory.Create(tt.matchType)
		if (m == nil) != tt.wantNil {
			t.Errorf("Factory.Create(%s) nil = %v, want %v", tt.matchType, m == nil, tt.wantNil)
		}
	}
}

// ============== ParamPathMatcher 集成测试 ==============

func TestParamPathMatcher_Integration(t *testing.T) {
	celEngine := newTestCelEngine(t)
	m := matcher.NewParamPathMatcher(celEngine)

	tests := []struct {
		name       string
		path       string
		rule       *db.RouteRule
		wantMatch  bool
		wantParams map[string]string
		wantErr    bool
	}{
		{
			name: "简单路径匹配",
			path: "/users/123",
			rule: &db.RouteRule{
				PathPattern:   sql.NullString{String: "/users/{id}", Valid: true},
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch:  true,
			wantParams: map[string]string{"id": "123"},
			wantErr:    false,
		},
		{
			name: "多参数路径匹配",
			path: "/users/123/orders/456",
			rule: &db.RouteRule{
				PathPattern:   sql.NullString{String: "/users/{userId}/orders/{orderId}", Valid: true},
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch:  true,
			wantParams: map[string]string{"userId": "123", "orderId": "456"},
			wantErr:    false,
		},
		{
			name: "带CEL表达式的匹配",
			path: "/tenants/acme/orders",
			rule: &db.RouteRule{
				PathPattern:   sql.NullString{String: "/tenants/{tenant}/orders", Valid: true},
				CelExpression: sql.NullString{String: "path.tenant == 'acme'", Valid: true},
			},
			wantMatch:  true,
			wantParams: map[string]string{"tenant": "acme"},
			wantErr:    false,
		},
		{
			name: "CEL表达式不匹配",
			path: "/tenants/beta/orders",
			rule: &db.RouteRule{
				PathPattern:   sql.NullString{String: "/tenants/{tenant}/orders", Valid: true},
				CelExpression: sql.NullString{String: "path.tenant == 'acme'", Valid: true},
			},
			wantMatch:  false,
			wantParams: nil,
			wantErr:    false,
		},
		{
			name: "路径不匹配",
			path: "/products/123",
			rule: &db.RouteRule{
				PathPattern:   sql.NullString{String: "/users/{id}", Valid: true},
				CelExpression: sql.NullString{Valid: false},
			},
			wantMatch:  false,
			wantParams: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("GET", tt.path, nil, "")
			result := m.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}

			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}

			if tt.wantParams != nil {
				for key, expectedValue := range tt.wantParams {
					if result.Params[key] != expectedValue {
						t.Errorf("Params[%s] = %v, want %v", key, result.Params[key], expectedValue)
					}
				}
			}
		})
	}
}

// ============== URLParamMatcher 集成测试 ==============

func TestURLParamMatcher_Integration(t *testing.T) {
	celEngine := newTestCelEngine(t)
	m := matcher.NewURLParamMatcher(celEngine)

	tests := []struct {
		name      string
		path      string
		rule      *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "URL参数匹配成功",
			path: "/api?action=create&tenant=acme",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "params.action == 'create'", Valid: true},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "URL参数匹配失败",
			path: "/api?action=delete&tenant=acme",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "params.action == 'create'", Valid: true},
			},
			wantMatch: false,
			wantErr:   false,
		},
		{
			name: "多条件URL参数匹配",
			path: "/api?action=create&tenant=acme",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "params.action == 'create' && params.tenant == 'acme'", Valid: true},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("GET", tt.path, nil, "")
			result := m.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== RequestBodyMatcher 集成测试 ==============

func TestRequestBodyMatcher_Integration(t *testing.T) {
	celEngine := newTestCelEngine(t)
	m := matcher.NewRequestBodyMatcher(celEngine)

	tests := []struct {
		name      string
		body      string
		rule      *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "JSON body匹配成功",
			body: `{"user": {"role": "admin", "tenant_id": "acme"}}`,
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "body['user']['role'] == 'admin'", Valid: true},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "JSON body匹配失败",
			body: `{"user": {"role": "user", "tenant_id": "acme"}}`,
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "body['user']['role'] == 'admin'", Valid: true},
			},
			wantMatch: false,
			wantErr:   false,
		},
		{
			name: "多条件JSON body匹配",
			body: `{"user": {"role": "admin", "tenant_id": "acme"}}`,
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "body['user']['role'] == 'admin' && body['user']['tenant_id'] == 'acme'", Valid: true},
			},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("POST", "/api", bytes.NewBufferString(tt.body), "application/json")
			result := m.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== RequestFormMatcher 集成测试 ==============

func TestRequestFormMatcher_Integration(t *testing.T) {
	celEngine := newTestCelEngine(t)
	m := matcher.NewRequestFormMatcher(celEngine)

	tests := []struct {
		name      string
		formData  string
		rule      *db.RouteRule
		wantMatch bool
		wantErr   bool
	}{
		{
			name:     "表单数据匹配成功",
			formData: "action=create&tenant=acme",
			rule: &db.RouteRule{
				// 注意：表单数据通过 body 变量访问
				CelExpression: sql.NullString{String: "body['action'] == 'create'", Valid: true},
			},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:     "表单数据匹配失败",
			formData: "action=delete&tenant=acme",
			rule: &db.RouteRule{
				CelExpression: sql.NullString{String: "body['action'] == 'create'", Valid: true},
			},
			wantMatch: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext("POST", "/api", bytes.NewBufferString(tt.formData), "application/x-www-form-urlencoded")
			result := m.Match(c, tt.rule)

			if result.Matched != tt.wantMatch {
				t.Errorf("Matched = %v, want %v", result.Matched, tt.wantMatch)
			}
			if (result.Error != nil) != tt.wantErr {
				t.Errorf("Error = %v, want error: %v", result.Error, tt.wantErr)
			}
		})
	}
}

// ============== 三层路由匹配测试 ==============

func TestThreeLayerRouting_MatchFlow(t *testing.T) {
	celEngine := newTestCelEngine(t)
	factory := matcher.NewFactory(celEngine, nil)

	// 模拟三层路由配置
	sourceConfig := &SourceConfig{
		Source: &db.Source{
			ID:   "src-001",
			Name: "weixin",
		},
		Projects: []*ProjectConfig{
			{
				Project: &db.Project{
					ID:        "proj-001",
					SourceID:  "src-001",
					Name:      "order-service",
					TargetUrl: sql.NullString{String: "http://backend/api", Valid: true},
				},
				Rules: []*RouteRuleConfig{
					{
						Rule: &db.RouteRule{
							ID:          "rule-001",
							ProjectID:   "proj-001",
							Name:        "get-order",
							MatchType:   matcher.MatchTypePathParam,
							PathPattern: sql.NullString{String: "/orders/{orderId}", Valid: true},
							Priority:    sql.NullInt64{Int64: 1, Valid: true},
						},
					},
					{
						Rule: &db.RouteRule{
							ID:          "rule-002",
							ProjectID:   "proj-001",
							Name:        "list-orders",
							MatchType:   matcher.MatchTypePathParam,
							PathPattern: sql.NullString{String: "/orders", Valid: true},
							// 使用 URL 参数来区分不同的操作
							CelExpression: sql.NullString{String: "'action' in params", Valid: true},
							Priority:      sql.NullInt64{Int64: 2, Valid: true},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name        string
		method      string
		path        string
		expectMatch bool
		expectRule  string
	}{
		{
			name:        "匹配GET订单详情",
			method:      "GET",
			path:        "/orders/12345",
			expectMatch: true,
			expectRule:  "get-order", // 使用 Name 字段
		},
		{
			name:        "匹配带参数的订单列表",
			method:      "GET",
			path:        "/orders?action=list",
			expectMatch: true,
			expectRule:  "list-orders",
		},
		{
			name:        "路径不匹配",
			method:      "GET",
			path:        "/products/123",
			expectMatch: false,
			expectRule:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestContext(tt.method, tt.path, nil, "")

			// 遍历项目查找匹配的规则
			var matchedRule *db.RouteRule
			var matchedParams map[string]string

			for _, projectConfig := range sourceConfig.Projects {
				for _, ruleConfig := range projectConfig.Rules {
					rule := ruleConfig.Rule
					m := factory.Create(rule.MatchType)
					if m == nil {
						continue
					}

					result := m.Match(c, rule)
					if result.Error != nil {
						continue
					}

					if result.Matched {
						matchedRule = rule
						matchedParams = result.Params
						break
					}
				}
				if matchedRule != nil {
					break
				}
			}

			if tt.expectMatch {
				if matchedRule == nil {
					t.Errorf("期望匹配成功，但未找到匹配的规则")
					return
				}
				if matchedRule.Name != tt.expectRule {
					t.Errorf("匹配的规则 = %s, want %s", matchedRule.Name, tt.expectRule)
				}
				t.Logf("匹配成功: rule=%s, params=%v", matchedRule.Name, matchedParams)
			} else {
				if matchedRule != nil {
					t.Errorf("期望不匹配，但匹配到了规则: %s", matchedRule.Name)
				}
			}
		})
	}
}

// ============== 优先级测试 ==============

func TestRulePriority(t *testing.T) {
	celEngine := newTestCelEngine(t)
	factory := matcher.NewFactory(celEngine, nil)

	// 创建多个规则，测试优先级
	rules := []*RouteRuleConfig{
		{
			Rule: &db.RouteRule{
				ID:          "rule-low",
				Name:        "low-priority",
				MatchType:   matcher.MatchTypePathParam,
				PathPattern: sql.NullString{String: "/users/{id}", Valid: true},
				Priority:    sql.NullInt64{Int64: 10, Valid: true},
			},
		},
		{
			Rule: &db.RouteRule{
				ID:          "rule-high",
				Name:        "high-priority",
				MatchType:   matcher.MatchTypePathParam,
				PathPattern: sql.NullString{String: "/users/{id}", Valid: true},
				Priority:    sql.NullInt64{Int64: 1, Valid: true},
			},
		},
		{
			Rule: &db.RouteRule{
				ID:          "rule-mid",
				Name:        "mid-priority",
				MatchType:   matcher.MatchTypePathParam,
				PathPattern: sql.NullString{String: "/users/{id}", Valid: true},
				Priority:    sql.NullInt64{Int64: 5, Valid: true},
			},
		},
	}

	// 模拟优先级排序
	// 在实际实现中，排序在 findMatch 方法中完成
	// 这里我们手动排序来测试
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Rule.Priority.Int64 < rules[j].Rule.Priority.Int64
	})

	// 验证排序结果
	if rules[0].Rule.Name != "high-priority" {
		t.Errorf("第一个规则应该是 high-priority, got %s", rules[0].Rule.Name)
	}
	if rules[1].Rule.Name != "mid-priority" {
		t.Errorf("第二个规则应该是 mid-priority, got %s", rules[1].Rule.Name)
	}
	if rules[2].Rule.Name != "low-priority" {
		t.Errorf("第三个规则应该是 low-priority, got %s", rules[2].Rule.Name)
	}

	// 测试匹配 - 应该命中高优先级规则
	c := setupTestContext("GET", "/users/123", nil, "")

	for _, ruleConfig := range rules {
		m := factory.Create(ruleConfig.Rule.MatchType)
		result := m.Match(c, ruleConfig.Rule)
		if result.Matched {
			if ruleConfig.Rule.Name != "high-priority" {
				t.Errorf("应该首先匹配高优先级规则, got %s", ruleConfig.Rule.Name)
			}
			break
		}
	}
}

// ============== 错误处理测试 ==============

func TestMatcher_ErrorHandling(t *testing.T) {
	celEngine := newTestCelEngine(t)

	t.Run("ParamPathMatcher nil rule", func(t *testing.T) {
		m := matcher.NewParamPathMatcher(celEngine)
		c := setupTestContext("GET", "/test", nil, "")
		result := m.Match(c, nil)
		if result.Error != matcher.ErrNilRule {
			t.Errorf("Expected ErrNilRule, got %v", result.Error)
		}
	})

	t.Run("URLParamMatcher empty CEL", func(t *testing.T) {
		m := matcher.NewURLParamMatcher(celEngine)
		c := setupTestContext("GET", "/test", nil, "")
		result := m.Match(c, &db.RouteRule{})
		if result.Error != matcher.ErrEmptyCelExpression {
			t.Errorf("Expected ErrEmptyCelExpression, got %v", result.Error)
		}
	})

	t.Run("RequestBodyMatcher nil rule", func(t *testing.T) {
		m := matcher.NewRequestBodyMatcher(celEngine)
		c := setupTestContext("POST", "/test", nil, "application/json")
		result := m.Match(c, nil)
		if result.Error != matcher.ErrNilRule {
			t.Errorf("Expected ErrNilRule, got %v", result.Error)
		}
	})
}

// ============== CEL表达式安全测试 ==============

func TestCELExpression_Security(t *testing.T) {
	celEngine := newTestCelEngine(t)

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
			name:       "禁止的import模式",
			expression: "import('os')",
			wantErr:    true,
		},
		{
			name:       "禁止的exec模式",
			expression: "exec('cmd')",
			wantErr:    true,
		},
		{
			name:       "禁止的双下划线模式",
			expression: "__import__('os')",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := celEngine.ValidateExpression(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExpression(%s) error = %v, wantErr %v", tt.expression, err, tt.wantErr)
			}
		})
	}
}
