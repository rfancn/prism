// Package parser provides request identifier extraction from various sources.
package parser

// JSONBodyParser extracts identifier from JSON request body.
// Supports dot notation for nested fields: "user.tenant_id"
type JSONBodyParser struct {
	fieldName string
}

// NewJSONBodyParser creates a new JSON body parser.
func NewJSONBodyParser(fieldName string) *JSONBodyParser {
	return &JSONBodyParser{
		fieldName: fieldName,
	}
}

// Extract extracts the identifier from JSON body.
func (p *JSONBodyParser) Extract(data []byte) (string, error) {
	return extractJSONField(data, p.fieldName)
}