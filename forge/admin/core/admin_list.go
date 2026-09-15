package core

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/forgego/forge/orm"
)

// Data operations

func (a *Admin[T]) ListObjects(ctx context.Context, params ListParams) (*PaginatedResponse, error) {
	qs, err := a.GetQueryset(ctx)
	if err != nil {
		return nil, err
	}

	qs = a.applySearchFilter(qs, params.Search)
	qs = a.applyLookupFilters(ctx, qs, params.Filters)
	qs = a.applyOrdering(qs, params.Ordering)

	pagedQs, pagination, err := a.paginateQueryset(ctx, qs, params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}

	return a.serializeListRows(ctx, pagedQs, pagination)
}

func (a *Admin[T]) applySearchFilter(qs orm.QuerySet[T], search string) orm.QuerySet[T] {
	if search != "" && len(a.config.SearchFields) > 0 {
		searchQueries := make([]orm.Expression, 0, len(a.config.SearchFields))
		for _, field := range a.config.SearchFields {
			// Use F() helper for dynamic field names
			searchQueries = append(searchQueries, orm.F(orm.ExtractPathFromAny(field)).IContains(search))
		}
		return qs.Filter(orm.Or(searchQueries...))
	}
	return qs
}

func (a *Admin[T]) applyLookupFilters(ctx context.Context, qs orm.QuerySet[T], filters map[string]interface{}) orm.QuerySet[T] {
	for key, value := range filters {
		// Check for manual filter overrides
		applied := false
		for _, filter := range a.config.Filters {
			if filter.Name == key && filter.Handler != nil {
				qs = filter.Handler(ctx, qs, value)
				applied = true
				break
			}
		}
		if !applied {
			field, lookup := parseLookup(key)
			qs = qs.Filter(buildLookupExpression(field, lookup, value))
		}
	}
	return qs
}

func buildLookupExpression(field, lookup string, value interface{}) orm.Expression {
	f := orm.F(field)

	switch lookup {
	case "exact":
		return f.Eq(value)
	case "ne":
		return f.Ne(value)
	case "gt":
		return f.Gt(value)
	case "gte":
		return f.Gte(value)
	case "lt":
		return f.Lt(value)
	case "lte":
		return f.Lte(value)
	case "contains":
		if s, ok := value.(string); ok {
			return f.Contains(s)
		}
		return f.Eq(value) // Fallback
	case "icontains":
		if s, ok := value.(string); ok {
			return f.IContains(s)
		}
		return f.Eq(value) // Fallback
	case "startswith":
		if s, ok := value.(string); ok {
			return f.StartsWith(s)
		}
		return f.Eq(value)
	case "endswith":
		if s, ok := value.(string); ok {
			return f.EndsWith(s)
		}
		return f.Eq(value)
	case "in":
		// value should be slice or comma-separated string
		// For now assume value is single string from query param
		if s, ok := value.(string); ok {
			parts := strings.Split(s, ",")
			args := make([]interface{}, 0, len(parts))
			for _, v := range parts {
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					args = append(args, trimmed)
				}
			}
			if len(args) > 0 {
				return f.In(args...)
			}
			return f.Eq(value)
		}
		return f.Eq(value)
	case "isnull":
		if s, ok := value.(string); ok && (s == "true" || s == "1") {
			return f.IsNull()
		}
		return f.IsNotNull()
	default:
		return f.Eq(value)
	}
}

func (a *Admin[T]) applyOrdering(qs orm.QuerySet[T], reqOrdering []string) orm.QuerySet[T] {
	// Request ordering wins; unknown fields are dropped so a bad `ordering`
	// param can never produce a SQL error.
	if validOrdering := sanitizeOrdering(reqOrdering, a.orderableFields()); len(validOrdering) > 0 {
		ordering := make([]any, len(validOrdering))
		for i, v := range validOrdering {
			ordering[i] = v
		}
		return qs.OrderBy(ordering...)
	} else if len(a.config.Ordering) > 0 {
		ordering := make([]any, len(a.config.Ordering))
		for i, v := range a.config.Ordering {
			ordering[i] = v
		}
		return qs.OrderBy(ordering...)
	}
	return qs
}

type listPagination struct {
	page       int
	limit      int
	count      int64
	totalPages int
}

func (a *Admin[T]) paginateQueryset(ctx context.Context, qs orm.QuerySet[T], requestedPage, requestedPageSize int) (orm.QuerySet[T], listPagination, error) {
	// Guard against zero/negative/huge values so direct callers can never produce
	// a negative offset or div-by-zero.
	page := requestedPage
	if page < 1 {
		page = 1
	}
	limit := requestedPageSize
	if limit <= 0 {
		limit = a.config.ListPerPage
	}
	if limit <= 0 {
		limit = 25
	}
	const maxListLimit = 1000
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := (page - 1) * limit

	count, err := qs.Count(ctx)
	if err != nil {
		return nil, listPagination{}, err
	}

	totalPages := (int(count) + limit - 1) / limit

	return qs.Limit(limit).Offset(offset), listPagination{
		page:       page,
		limit:      limit,
		count:      count,
		totalPages: totalPages,
	}, nil
}

func (a *Admin[T]) serializeListRows(ctx context.Context, qs orm.QuerySet[T], p listPagination) (*PaginatedResponse, error) {
	results, err := qs.All(ctx)
	if err != nil {
		return nil, err
	}

	return &PaginatedResponse{
		Count:      p.count,
		PageSize:   p.limit,
		Page:       p.page,
		TotalPages: p.totalPages,
		Results:    results,
	}, nil
}

// orderableFields returns the set of schema field names that may be used
// for ordering (plus the conventional "id"/"pk" aliases).
func (a *Admin[T]) orderableFields() map[string]bool {
	fields := make(map[string]bool)
	if a.schema != nil {
		for _, f := range a.schema.Fields() {
			if f.Name != "" {
				fields[f.Name] = true
			}
		}
	}
	fields["id"] = true
	fields["pk"] = true
	return fields
}

// sanitizeOrdering drops ordering keys that reference unknown fields or
// contain unsafe characters, so user input can never break the query.
func sanitizeOrdering(ordering []string, valid map[string]bool) []string {
	sanitized := make([]string, 0, len(ordering))
	for _, key := range ordering {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		desc := strings.HasPrefix(key, "-")
		name := strings.TrimPrefix(key, "-")
		if !isSafeOrderField(name) {
			continue
		}
		if !valid[name] {
			continue
		}
		if desc {
			sanitized = append(sanitized, "-"+name)
		} else {
			sanitized = append(sanitized, name)
		}
	}
	return sanitized
}

// isSafeOrderField reports whether name is a plain field identifier.
func isSafeOrderField(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' {
			continue
		}
		return false
	}
	return true
}

func (a *Admin[T]) Autocomplete(ctx context.Context, query string, limit int) ([]AutocompleteItem, error) {
	qs, err := a.GetQueryset(ctx)
	if err != nil {
		return nil, err
	}

	// Apply search
	if query != "" && len(a.config.SearchFields) > 0 {
		searchQueries := make([]orm.Expression, 0, len(a.config.SearchFields))
		for _, field := range a.config.SearchFields {
			searchQueries = append(searchQueries, orm.F(orm.ExtractPathFromAny(field)).IContains(query))
		}
		qs = qs.Filter(orm.Or(searchQueries...))
	}

	// Fetch results
	results, err := qs.Limit(limit).All(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to AutocompleteItem
	items := make([]AutocompleteItem, 0, len(results))
	for _, res := range results {
		items = append(items, AutocompleteItem{
			Value: a.getObjectID(res),
			Label: a.getObjectLabel(res),
		})
	}

	return items, nil
}

// LabelResolver is implemented by admins that can label objects by id in bulk.
type LabelResolver interface {
	ObjectLabels(ctx context.Context, ids []interface{}) (map[string]string, error)
}

// ObjectLabels resolves human-readable labels for objects by their primary keys in bulk.
func (a *Admin[T]) ObjectLabels(ctx context.Context, ids []interface{}) (map[string]string, error) {
	if len(ids) == 0 {
		return make(map[string]string), nil
	}

	if len(ids) > 1000 {
		ids = ids[:1000]
	}

	qs, err := a.GetQueryset(ctx)
	if err != nil {
		return nil, err
	}

	pkField := "id"
	if a.modelSchema != nil && a.modelSchema.PrimaryKey != "" {
		pkField = a.modelSchema.PrimaryKey
	}

	normalizedIDs := make([]interface{}, len(ids))
	for i, id := range ids {
		if f, ok := id.(float64); ok && f == math.Floor(f) && !math.IsNaN(f) && !math.IsInf(f, 0) {
			normalizedIDs[i] = int64(f)
		} else {
			normalizedIDs[i] = id
		}
	}

	qs = qs.Filter(orm.F(pkField).In(normalizedIDs...))

	results, err := qs.All(ctx)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string, len(results))
	for _, obj := range results {
		objID := a.getObjectID(obj)
		if objID != nil {
			labels[fmt.Sprint(objID)] = a.getObjectLabel(obj)
		}
	}

	return labels, nil
}

// parseLookup splits key into field path and lookup (e.g. "price__gt" -> "price", "gt")
func parseLookup(key string) (string, string) {
	parts := strings.Split(key, "__")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		// Check if last part is a known lookup
		switch last {
		case "exact", "ne", "gt", "gte", "lt", "lte", "contains", "icontains", "startswith", "endswith", "in", "isnull":
			return strings.Join(parts[:len(parts)-1], "__"), last
		}
	}
	return key, "exact"
}
