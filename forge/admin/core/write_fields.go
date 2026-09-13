package core

import (
	"context"
	"sort"
	"strings"

	validation "github.com/forgego/forge/validate"
)

// writableFields returns the set of fields allowed to be written for instance.
func (a *Admin[T]) writableFields(ctx context.Context, instance *T, isNew bool) map[string]bool {
	if a == nil || a.schema == nil {
		return nil
	}

	writable := make(map[string]bool)
	if a.config != nil && a.config.GetFields != nil {
		for _, name := range a.config.GetFields(ctx, instance, isNew) {
			writable[name] = true
		}
	} else if a.config != nil && len(a.config.Fields) > 0 {
		for _, name := range a.config.Fields {
			writable[name] = true
		}
	} else {
		for _, field := range a.schema.Fields() {
			writable[field.Name] = true
		}
	}

	if a.config != nil {
		for _, name := range a.config.Exclude {
			delete(writable, name)
		}

		readOnly := a.config.ReadOnlyFields
		if a.config.GetReadOnlyFields != nil {
			readOnly = a.config.GetReadOnlyFields(ctx, instance, isNew)
		}
		for _, name := range readOnly {
			delete(writable, name)
		}
	}

	for _, field := range a.schema.Fields() {
		if !field.Editable || isAutoManaged(field) {
			delete(writable, field.Name)
		}
	}

	return writable
}

// filterWritable drops non-writable schema fields and rejects unknown fields.
func (a *Admin[T]) filterWritable(ctx context.Context, instance *T, isNew bool, data map[string]interface{}) (map[string]interface{}, error) {
	if a == nil || a.schema == nil {
		filtered := make(map[string]interface{}, len(data))
		for k, v := range data {
			filtered[k] = v
		}
		return filtered, nil
	}

	writable := a.writableFields(ctx, instance, isNew)
	schemaFields := make(map[string]bool, len(a.schema.Fields()))
	for _, f := range a.schema.Fields() {
		schemaFields[f.Name] = true
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	errs := &validation.ValidationErrors{}
	filtered := make(map[string]interface{})
	for _, key := range keys {
		if !schemaFields[key] {
			if strings.EqualFold(key, "id") {
				continue
			}
			errs.Add(key, "unknown field")
			continue
		}
		if writable[key] {
			filtered[key] = data[key]
		}
	}

	if errs.HasErrors() {
		return nil, errs
	}
	return filtered, nil
}
