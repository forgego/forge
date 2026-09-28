package api

import (
	"reflect"

	"github.com/forgego/forge/schema"
)

// NonSerializableFields returns the names of fields on model whose Serialize
// flag is false (write-only values such as passwords). Both the field Name
// and every concrete Go, JSON, and database alias are returned. It returns
// nil when model does not implement schema.Schema.
func NonSerializableFields(model interface{}) []string {
	if model == nil {
		return nil
	}
	if s, ok := model.(schema.Schema); ok {
		return nonSerializableFromFields(model, s.Fields())
	}
	// Handle a non-pointer model whose pointer implements schema.Schema.
	v := reflect.ValueOf(model)
	if v.IsValid() && v.Kind() != reflect.Ptr {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		if s, ok := ptr.Interface().(schema.Schema); ok {
			return nonSerializableFromFields(ptr.Interface(), s.Fields())
		}
	}
	return nil
}

func nonSerializableFromFields(model interface{}, fields []schema.Field) []string {
	var out []string
	seen := make(map[string]bool)
	appendUnique := func(name string) {
		if name == "" || name == "-" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	for _, f := range fields {
		if f.Serialize {
			continue
		}
		for _, name := range resolvedFieldNames(model, f) {
			appendUnique(name)
		}
	}
	return out
}

// NonEditableFields returns the names of fields on model that must not be
// written through create or update: fields whose Editable flag is false as
// well as database-owned auto-managed fields (AutoNow, AutoNowAdd and
// generated columns), which normally keep Editable: true. Every schema,
// concrete Go, JSON, and database alias is returned. It returns nil when model
// does not implement schema.Schema.
func NonEditableFields(model interface{}) []string {
	if model == nil {
		return nil
	}
	if s, ok := model.(schema.Schema); ok {
		return nonEditableFromFields(model, s.Fields())
	}
	// Handle a non-pointer model whose pointer implements schema.Schema.
	v := reflect.ValueOf(model)
	if v.IsValid() && v.Kind() != reflect.Ptr {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		if s, ok := ptr.Interface().(schema.Schema); ok {
			return nonEditableFromFields(ptr.Interface(), s.Fields())
		}
	}
	return nil
}

func nonEditableFromFields(model interface{}, fields []schema.Field) []string {
	var out []string
	seen := make(map[string]bool)
	appendUnique := func(name string) {
		if name == "" || name == "-" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	for _, f := range fields {
		if f.Editable && !isRequestReadOnly(f) {
			continue
		}
		for _, name := range resolvedFieldNames(model, f) {
			appendUnique(name)
		}
	}
	return out
}

func resolvedFieldNames(model interface{}, field schema.Field) []string {
	if resolved, ok := schema.ResolveField(model, field); ok {
		return resolved.Names()
	}
	return schema.ResolvedField{SchemaName: field.Name, DBColumn: field.DBColumn}.Names()
}

// isRequestReadOnly reports whether a field is owned by the database and must
// therefore be ignored on requests even when it keeps Editable: true.
// Mirrors admin/core isAutoManaged for the request path and includes an
// auto-increment primary key, whose value is always owned by the database.
func isRequestReadOnly(f schema.Field) bool {
	return (f.PrimaryKey && f.AutoIncrement) || f.AutoNow || f.AutoNowAdd || f.Generated
}

// stripExcludedFields removes ExcludeResponseFields keys from a serialized map.
func (vs *BaseViewSet) stripExcludedFields(m map[string]interface{}) map[string]interface{} {
	if m == nil || len(vs.ExcludeResponseFields) == 0 {
		return m
	}
	for _, k := range vs.ExcludeResponseFields {
		delete(m, k)
	}
	return m
}

// stripExcludedFieldsMany removes ExcludeResponseFields keys from each map.
func (vs *BaseViewSet) stripExcludedFieldsMany(items []map[string]interface{}) []map[string]interface{} {
	if len(vs.ExcludeResponseFields) == 0 {
		return items
	}
	for _, m := range items {
		vs.stripExcludedFields(m)
	}
	return items
}
