package errors

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ErrorCodeDocumentation generates documentation for all error codes
type ErrorCodeDocumentation struct {
	Version string
	Codes   []*ErrorCode
}

// GenerateMarkdown generates markdown documentation for error codes
func GenerateMarkdown(version string) string {
	codes := GetErrorCodesByVersion(version)
	if len(codes) == 0 {
		codes = GetAllErrorCodes()
	}

	// Sort by type, then by code
	sort.Slice(codes, func(i, j int) bool {
		if codes[i].Type != codes[j].Type {
			return codes[i].Type < codes[j].Type
		}
		return codes[i].Code < codes[j].Code
	})

	var builder strings.Builder
	builder.WriteString("# Error Codes Reference\n\n")
	builder.WriteString("This document lists all error codes that may be returned by the API.\n\n")

	// Group by type
	currentType := ErrorType("")
	for _, code := range codes {
		if code.Type != currentType {
			if currentType != "" {
				builder.WriteString("\n")
			}
			currentType = code.Type
			builder.WriteString(fmt.Sprintf("## %s\n\n", formatTypeTitle(string(currentType))))
			builder.WriteString("| Code | Description | HTTP Status |\n")
			builder.WriteString("|------|-------------|-------------|\n")
		}
		builder.WriteString(fmt.Sprintf("| `%s` | %s | %d |\n", code.Code, code.Description, code.HTTPStatus))
	}

	return builder.String()
}

// GenerateJSON generates JSON documentation for error codes
func GenerateJSON(version string) ([]byte, error) {
	codes := GetErrorCodesByVersion(version)
	if len(codes) == 0 {
		codes = GetAllErrorCodes()
	}

	// Sort by code
	sort.Slice(codes, func(i, j int) bool {
		return codes[i].Code < codes[j].Code
	})

	type CodeDoc struct {
		Code        string    `json:"code"`
		Description string    `json:"description"`
		HTTPStatus  int       `json:"http_status"`
		Type        ErrorType `json:"type"`
		Version     string    `json:"version"`
	}

	docs := make([]CodeDoc, len(codes))
	for i, code := range codes {
		docs[i] = CodeDoc{
			Code:        code.Code,
			Description: code.Description,
			HTTPStatus:  code.HTTPStatus,
			Type:        code.Type,
			Version:     code.Version,
		}
	}

	return json.MarshalIndent(docs, "", "  ")
}

// formatTypeTitle formats an error type for display
func formatTypeTitle(t string) string {
	// Convert "validation-error" to "Validation Error"
	parts := strings.Split(t, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// GetErrorCodeByCode returns error code information by code string
func GetErrorCodeByCode(code string) (*ErrorCode, bool) {
	return GetErrorCode(code)
}

// ListErrorCodesByType returns all error codes of a specific type
func ListErrorCodesByType(errType ErrorType) []*ErrorCode {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	codes := make([]*ErrorCode, 0)
	for _, code := range globalRegistry.codes {
		if code.Type == errType {
			codes = append(codes, code)
		}
	}

	sort.Slice(codes, func(i, j int) bool {
		return codes[i].Code < codes[j].Code
	})

	return codes
}
