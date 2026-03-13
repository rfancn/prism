package server

import (
	"testing"
)

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
			result := convertPattern(tt.input)
			if result != tt.expected {
				t.Errorf("convertPattern(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}