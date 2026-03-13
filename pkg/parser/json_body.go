package parser

import (
	"encoding/json"
	"errors"
	"strings"
)

// extractJSONField extracts a field value from JSON data.
// Supports dot notation for nested fields: "user.tenant_id"
func extractJSONField(data []byte, fieldName string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty JSON body")
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return "", err
	}

	// Handle nested fields with dot notation
	parts := strings.Split(fieldName, ".")
	current := body

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return "", errors.New("field not found in JSON body")
		}

		// If this is the last part, return the value as string
		if i == len(parts)-1 {
			switch v := value.(type) {
			case string:
				return v, nil
			case float64:
				return formatFloat(v), nil
			case bool:
				return formatBool(v), nil
			default:
				return "", errors.New("field value is not a primitive type")
			}
		}

		// Otherwise, continue navigating
		next, ok := value.(map[string]interface{})
		if !ok {
			return "", errors.New("intermediate field is not an object")
		}
		current = next
	}

	return "", errors.New("field not found in JSON body")
}

func formatFloat(v float64) string {
	// Check if it's an integer
	if v == float64(int64(v)) {
		return string(rune(int64(v)))
	}
	return string(rune(v))
}

func formatBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}