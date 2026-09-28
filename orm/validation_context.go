package orm

import "context"

type modelValidationContextKey struct{}

// WithModelValidationCompleted marks a single manager call as already having
// completed model validation. It is intended for request adapters that must
// validate before lifecycle hooks run. Direct manager calls remain unchanged.
func WithModelValidationCompleted(ctx context.Context) context.Context {
	return context.WithValue(ctx, modelValidationContextKey{}, true)
}

func modelValidationCompleted(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	validated, _ := ctx.Value(modelValidationContextKey{}).(bool)
	return validated
}

func withoutModelValidationCompleted(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, modelValidationContextKey{}, false)
}
