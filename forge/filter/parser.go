package filter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Parser parses HTTP query parameters into filter values
type Parser struct {
	security *SecurityConfig
}

// NewParser creates a new query parameter parser
func NewParser(security *SecurityConfig) *Parser {
	return &Parser{
		security: security,
	}
}

// ParseQueryParams parses query parameters from an HTTP request
func (p *Parser) ParseQueryParams(r *http.Request, schema interface{}) (map[string]interface{}, error) {
	query := r.URL.Query()
	filters := make(map[string]interface{})

	for param, values := range query {
		if len(values) == 0 {
			continue
		}

		// Skip pagination and ordering params
		if isReservedParam(param) {
			continue
		}

		// Parse the parameter (may contain lookup suffix like "username__contains")
		fieldPath, lookup, err := p.parseParamName(param)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter name '%s': %w", param, err)
		}

		// Security: Validate field path if security config is provided
		if p.security != nil {
			if err := p.validateFieldAccess(fieldPath, schema); err != nil {
				return nil, fmt.Errorf("field access denied for '%s': %w", fieldPath, err)
			}

			if err := p.validateLookup(fieldPath, lookup); err != nil {
				return nil, fmt.Errorf("lookup not allowed for '%s': %w", fieldPath, err)
			}
		}

		// Parse the value based on lookup type
		value, err := p.parseValue(values[0], lookup)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value for '%s': %w", param, err)
		}

		filters[param] = value
	}

	return filters, nil
}

// parseParamName parses a parameter name like "username__contains" into field path and lookup
func (p *Parser) parseParamName(param string) (fieldPath, lookup string, err error) {
	parts := strings.Split(param, "__")
	if len(parts) == 1 {
		// Simple field name, default to "exact" lookup
		return parts[0], "exact", nil
	}

	// Last part is the lookup, rest is the field path
	lookup = parts[len(parts)-1]
	fieldPath = strings.Join(parts[:len(parts)-1], "__")

	// Validate lookup
	if !isValidLookup(lookup) {
		return "", "", fmt.Errorf("invalid lookup: %s", lookup)
	}

	return fieldPath, lookup, nil
}

// parseValue parses a string value based on the lookup type
func (p *Parser) parseValue(valueStr, lookup string) (interface{}, error) {
	if valueStr == "" {
		return nil, nil
	}

	switch lookup {
	case "exact", "iexact", "contains", "icontains", "startswith", "istartswith", "endswith", "iendswith":
		return valueStr, nil

	case "in":
		// Comma-separated values
		values := strings.Split(valueStr, ",")
		result := make([]string, 0, len(values))
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				result = append(result, v)
			}
		}
		return result, nil

	case "range":
		// Comma-separated range: "min,max"
		parts := strings.Split(valueStr, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("range requires exactly 2 values, got %d", len(parts))
		}
		return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}, nil

	case "gt", "gte", "lt", "lte":
		// Numeric comparison
		return strconv.ParseFloat(valueStr, 64)

	case "isnull":
		// Boolean: "true" or "1" means IS NULL
		return parseBool(valueStr), nil

	case "isnotnull":
		// Boolean: "true" or "1" means IS NOT NULL
		return !parseBool(valueStr), nil

	case "year", "month", "day":
		// Integer for date extraction
		return strconv.Atoi(valueStr)

	default:
		// Default to string
		return valueStr, nil
	}
}

// parseBool parses a boolean string value
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// isValidLookup checks if a lookup type is valid
func isValidLookup(lookup string) bool {
	validLookups := map[string]bool{
		"exact":      true,
		"iexact":     true,
		"contains":   true,
		"icontains":  true,
		"startswith": true,
		"istartswith": true,
		"endswith":   true,
		"iendswith":  true,
		"in":         true,
		"range":      true,
		"gt":         true,
		"gte":        true,
		"lt":         true,
		"lte":        true,
		"isnull":     true,
		"isnotnull":  true,
		"year":       true,
		"month":      true,
		"day":        true,
	}
	return validLookups[lookup]
}

// isReservedParam checks if a parameter name is reserved
func isReservedParam(param string) bool {
	reserved := map[string]bool{
		"page":      true,
		"page_size": true,
		"ordering":  true,
		"order_by":  true,
		"search":    true,
		"format":    true,
	}
	return reserved[param]
}

// validateFieldAccess validates that a field path is allowed
func (p *Parser) validateFieldAccess(fieldPath string, schema interface{}) error {
	if p.security == nil {
		return nil
	}

	// Check if field is in whitelist
	// This would need schema integration to check actual allowed fields
	// For now, just check if security config has restrictions
	if len(p.security.AllowedFields) > 0 {
		// Would need to check against schema's GetAllowedFields
		// This is a placeholder
	}

	return nil
}

// validateLookup validates that a lookup is allowed for a field
func (p *Parser) validateLookup(fieldPath, lookup string) error {
	if p.security == nil {
		return nil
	}

	// Check if lookup is in whitelist for this field
	if lookups, ok := p.security.AllowedLookups[fieldPath]; ok {
		allowed := false
		for _, l := range lookups {
			if l == lookup {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("lookup '%s' not allowed for field '%s'", lookup, fieldPath)
		}
	}

	return nil
}

// ParseFilterNode parses query parameters into a FilterNode AST
func (p *Parser) ParseFilterNode(r *http.Request, schema interface{}) (*FilterNode, error) {
	filters, err := p.ParseQueryParams(r, schema)
	if err != nil {
		return nil, err
	}

	if len(filters) == 0 {
		return nil, nil
	}

	// Convert filters to AST nodes (all ANDed together by default)
	var nodes []*FilterNode
	for param, value := range filters {
		fieldPath, lookup, err := p.parseParamName(param)
		if err != nil {
			return nil, err
		}

		node := NewFieldNode(fieldPath, lookup, value)
		nodes = append(nodes, node)
	}

	if len(nodes) == 1 {
		return nodes[0], nil
	}

	// Combine all nodes with AND
	return NewAndNode(nodes...), nil
}

