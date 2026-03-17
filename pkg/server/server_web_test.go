package server

import (
	"regexp"
	"testing"
)

// convertPatternToGin 将路径模式转换为Gin路由格式
// 这个函数是为了测试而保留的本地副本
// 实际实现已移到 pkg/matcher/helper.go
func convertPatternToGin(pattern string) string {
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	return re.ReplaceAllString(pattern, ":$1")
}

func TestConvertPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/api/{tenant}/users",
			expected: "/api/:tenant/users",
		},
		{
			input:    "/api/{tenant}",
			expected: "/api/:tenant",
		},
		{
			input:    "/{id}",
			expected: "/:id",
		},
		{
			input:    "/api/users",
			expected: "/api/users",
		},
		{
			input:    "/api/{org}/{user}/posts/{post_id}",
			expected: "/api/:org/:user/posts/:post_id",
		},
		{
			input:    "/",
			expected: "/",
		},
		{
			input:    "/api/{tenant_id}/items/{item_id}",
			expected: "/api/:tenant_id/items/:item_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertPatternToGin(tt.input)
			if result != tt.expected {
				t.Errorf("convertPatternToGin(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}