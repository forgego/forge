package orm

import "context"

// NewUpdateBuilderFromQuerySet creates an UpdateBuilder from a QuerySet
func NewUpdateBuilderFromQuerySet[T any](qs QuerySet[T]) (*UpdateBuilder[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, err
	}

	return &UpdateBuilder[T]{
		qs:      &querySetWrapper[T]{qs: qs},
		schema:  schema,
		updates: make(map[string]interface{}),
	}, nil
}

// querySetWrapper wraps QuerySet to work with UpdateBuilder
type querySetWrapper[T any] struct {
	qs QuerySet[T]
}

func (w *querySetWrapper[T]) Update(ctx context.Context, updates UpdateMap) (int64, error) {
	return w.qs.Update(ctx, updates)
}
