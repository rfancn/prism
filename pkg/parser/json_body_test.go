package parser

import (
	"testing"
)

func TestJSONBodyParser_Extract(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		data      string
		expected  string
		wantErr   bool
	}{
		{
			name:      "simple field",
			fieldName: "tenant_id",
			data:      `{"tenant_id": "acme"}`,
			expected:  "acme",
			wantErr:   false,
		},
		{
			name:      "nested field",
			fieldName: "user.tenant_id",
			data:      `{"user": {"tenant_id": "acme"}}`,
			expected:  "acme",
			wantErr:   false,
		},
		{
			name:      "deeply nested field",
			fieldName: "data.user.org.id",
			data:      `{"data": {"user": {"org": {"id": "org123"}}}}`,
			expected:  "org123",
			wantErr:   false,
		},
		{
			name:      "field not found",
			fieldName: "missing",
			data:      `{"tenant_id": "acme"}`,
			wantErr:   true,
		},
		{
			name:      "empty body",
			fieldName: "tenant_id",
			data:      "",
			wantErr:   true,
		},
		{
			name:      "invalid json",
			fieldName: "tenant_id",
			data:      `{invalid}`,
			wantErr:   true,
		},
		{
			name:      "intermediate not object",
			fieldName: "user.tenant_id",
			data:      `{"user": "not an object"}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewJSONBodyParser(tt.fieldName)
			result, err := parser.Extract([]byte(tt.data))

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Extract() = %q, want %q", result, tt.expected)
			}
		})
	}
}