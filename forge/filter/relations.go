package filter

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/orm"
)

// RelatedFilter filters across relationships
type RelatedFilter[T any, TRelated any] struct {
	*BaseFilter[T]
	relationPath string
}

// NewRelatedFilter creates a new related filter
func NewRelatedFilter[T any, TRelated any](fieldPath, relationPath string) *RelatedFilter[T, TRelated] {
	return &RelatedFilter[T, TRelated]{
		BaseFilter:   NewBaseFilter[T](fieldPath, "exact"),
		relationPath: relationPath,
	}
}

// DeepRelationFilter handles deep relation filtering (e.g., author__company__country)
type DeepRelationFilter struct {
	fieldPath string
	maxDepth  int
}

// NewDeepRelationFilter creates a new deep relation filter
func NewDeepRelationFilter(fieldPath string, maxDepth int) *DeepRelationFilter {
	return &DeepRelationFilter{
		fieldPath: fieldPath,
		maxDepth:  maxDepth,
	}
}

// ValidateDepth validates that a relation path doesn't exceed max depth
func ValidateDepth(schema *orm.ModelSchema, path string, maxDepth int) error {
	depth, err := schema.GetRelationDepth(path)
	if err != nil {
		return err
	}

	if depth > maxDepth {
		return fmt.Errorf("relation depth %d exceeds maximum allowed depth %d for path: %s", depth, maxDepth, path)
	}

	return nil
}

// GetRelationPath extracts the relation path from a field path
func GetRelationPath(fieldPath string) (string, string, error) {
	parts := strings.Split(fieldPath, "__")
	if len(parts) < 2 {
		return "", fieldPath, nil // No relation, just a field
	}

	// Last part is the field, rest is the relation path
	relationPath := strings.Join(parts[:len(parts)-1], "__")
	fieldName := parts[len(parts)-1]

	return relationPath, fieldName, nil
}

// OptimizeRelationQuery determines the best query strategy for a relation
func OptimizeRelationQuery(schema *orm.ModelSchema, fieldPath string, strategy string) (string, error) {
	// This would analyze the relation and choose:
	// - JOIN for small relations
	// - EXISTS for large relations
	// - Subquery for aggregates

	relationPath, _, err := GetRelationPath(fieldPath)
	if err != nil {
		return "", err
	}

	if relationPath == "" {
		// No relation, use direct field access
		return "direct", nil
	}

	// Count relation depth
	depth := strings.Count(relationPath, "__") + 1

	// Choose strategy based on depth and provided strategy hint
	switch strategy {
	case "join":
		if depth <= 2 {
			return "join", nil
		}
		return "exists", nil
	case "exists":
		return "exists", nil
	case "subquery":
		return "subquery", nil
	default:
		// Auto-choose based on depth
		if depth <= 2 {
			return "join", nil
		}
		return "exists", nil
	}
}

// BuildRelationJoin builds JOIN clauses for relation paths
func BuildRelationJoin(schema *orm.ModelSchema, fieldPath string) ([]string, error) {
	relationPath, _, err := GetRelationPath(fieldPath)
	if err != nil {
		return nil, err
	}

	if relationPath == "" {
		return nil, nil
	}

	parts := strings.Split(relationPath, "__")
	joins := make([]string, 0, len(parts))

	currentSchema := schema

	for _, part := range parts {
		rel := currentSchema.GetRelation(part)
		if rel == nil {
			return nil, fmt.Errorf("relation '%s' not found", part)
		}

		// Build JOIN clause
		// This would need actual SQL generation
		// For now, return placeholder
		join := fmt.Sprintf("JOIN %s ON ...", rel.TargetModel)
		joins = append(joins, join)

		// Move to next relation
		targetSchema, err := orm.GetModelSchemaByName(rel.TargetModel)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve relation '%s': %w", part, err)
		}
		currentSchema = targetSchema
	}

	return joins, nil
}

// DeduplicateJoins removes duplicate JOIN clauses
func DeduplicateJoins(joins []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(joins))

	for _, join := range joins {
		if !seen[join] {
			seen[join] = true
			result = append(result, join)
		}
	}

	return result
}
