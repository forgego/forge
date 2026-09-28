package schema

import "context"

// ModelHooks contains all Django-style lifecycle hooks
type ModelHooks struct {
	// Create hooks
	BeforeCreate func(ctx context.Context, instance interface{}) error
	AfterCreate  func(ctx context.Context, instance interface{}) error

	// Update hooks
	BeforeUpdate func(ctx context.Context, instance interface{}) error
	AfterUpdate  func(ctx context.Context, instance interface{}) error

	// Save hooks (called for both create and update)
	BeforeSave func(ctx context.Context, instance interface{}) error
	AfterSave  func(ctx context.Context, instance interface{}) error

	// Delete hooks
	BeforeDelete func(ctx context.Context, instance interface{}) error
	AfterDelete  func(ctx context.Context, instance interface{}) error

	// Validation hook
	Clean func(instance interface{}) error

	// Custom save/delete (override default behavior)
	Save   func(ctx context.Context, instance interface{}) error
	Delete func(ctx context.Context, instance interface{}) error
}

// NewModelHooks creates a new ModelHooks instance
func NewModelHooks() *ModelHooks {
	return &ModelHooks{}
}

// WithBeforeCreate sets the BeforeCreate hook
func (h *ModelHooks) WithBeforeCreate(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.BeforeCreate = fn
	return h
}

// WithAfterCreate sets the AfterCreate hook
func (h *ModelHooks) WithAfterCreate(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.AfterCreate = fn
	return h
}

// WithBeforeUpdate sets the BeforeUpdate hook
func (h *ModelHooks) WithBeforeUpdate(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.BeforeUpdate = fn
	return h
}

// WithAfterUpdate sets the AfterUpdate hook
func (h *ModelHooks) WithAfterUpdate(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.AfterUpdate = fn
	return h
}

// WithBeforeSave sets the BeforeSave hook
func (h *ModelHooks) WithBeforeSave(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.BeforeSave = fn
	return h
}

// WithAfterSave sets the AfterSave hook
func (h *ModelHooks) WithAfterSave(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.AfterSave = fn
	return h
}

// WithBeforeDelete sets the BeforeDelete hook
func (h *ModelHooks) WithBeforeDelete(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.BeforeDelete = fn
	return h
}

// WithAfterDelete sets the AfterDelete hook
func (h *ModelHooks) WithAfterDelete(fn func(ctx context.Context, instance interface{}) error) *ModelHooks {
	h.AfterDelete = fn
	return h
}

// WithClean sets the Clean validation hook
func (h *ModelHooks) WithClean(fn func(instance interface{}) error) *ModelHooks {
	h.Clean = fn
	return h
}
