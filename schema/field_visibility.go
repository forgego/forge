package schema

import (
	"reflect"
	"strings"
)

// NonSerializableFieldNames returns schema, JSON, database and Go aliases for
// fields whose Serialize flag is false. It describes shared field facts; each
// API applies its response policy at its own boundary.
func NonSerializableFieldNames(model any) []string {
	if model == nil {
		return nil
	}
	value := reflect.ValueOf(model)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil
	}
	definition, ok := model.(Schema)
	if !ok && value.Kind() != reflect.Ptr {
		pointer := reflect.New(value.Type())
		pointer.Elem().Set(value)
		definition, ok = pointer.Interface().(Schema)
	}
	if !ok {
		return nil
	}
	var names []string
	seen := make(map[string]bool)
	for _, field := range definition.Fields() {
		if field.Serialize {
			continue
		}
		resolved, ok := ResolveField(model, field)
		if !ok {
			resolved = ResolvedField{SchemaName: field.Name, DBColumn: field.DBColumn}
		}
		for _, name := range resolved.Names() {
			if !seen[name] {
				names = append(names, name)
				seen[name] = true
			}
		}
	}
	return names
}

// OperationOnlyFieldNames returns schema, JSON, database and Go aliases for
// fields marked as operation-only on the model.
func OperationOnlyFieldNames(model any) []string {
	if model == nil {
		return nil
	}
	value := reflect.ValueOf(model)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil
	}
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	var explicit []string
	if provider, ok := model.(interface{ OperationOnlyFields() []string }); ok {
		explicit = append(explicit, provider.OperationOnlyFields()...)
	}
	if value.IsValid() && value.CanAddr() {
		if provider, ok := value.Addr().Interface().(interface{ OperationOnlyFields() []string }); ok {
			explicit = append(explicit, provider.OperationOnlyFields()...)
		}
	}

	var definition Schema
	if def, ok := model.(Schema); ok {
		definition = def
	} else if value.IsValid() && value.CanAddr() {
		if def, ok := value.Addr().Interface().(Schema); ok {
			definition = def
		}
	}
	if definition == nil && value.IsValid() && value.Kind() == reflect.Struct {
		pointer := reflect.New(value.Type())
		pointer.Elem().Set(value)
		if def, ok := pointer.Interface().(Schema); ok {
			definition = def
		}
	}

	if definition != nil {
		meta := definition.Meta()
		if len(meta.OperationOnlyFields) > 0 {
			explicit = append(explicit, meta.OperationOnlyFields...)
		}
	}

	var names []string
	seen := make(map[string]bool)

	addName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || name == "-" {
			return
		}
		lower := strings.ToLower(name)
		if !seen[lower] {
			seen[lower] = true
			names = append(names, name)
		}
	}

	for _, name := range explicit {
		addName(name)
	}

	if definition != nil {
		for _, field := range definition.Fields() {
			isOpOnly := field.OperationOnly
			if !isOpOnly {
				for _, exp := range explicit {
					if strings.EqualFold(exp, field.Name) || (field.DBColumn != "" && strings.EqualFold(exp, field.DBColumn)) {
						isOpOnly = true
						break
					}
				}
			}
			if isOpOnly {
				addName(field.Name)
				if field.DBColumn != "" {
					addName(field.DBColumn)
				}
				resolved, ok := ResolveField(model, field)
				if ok {
					for _, name := range resolved.Names() {
						addName(name)
					}
				}
			}
		}
	}

	return names
}

// FindOperationOnlyField checks whether data contains any operation-only field.
// It returns the offending key from data, and true if found.
func FindOperationOnlyField(data map[string]interface{}, operationOnlyFields []string) (string, bool) {
	if len(data) == 0 || len(operationOnlyFields) == 0 {
		return "", false
	}
	for key := range data {
		for _, opField := range operationOnlyFields {
			if strings.EqualFold(key, opField) {
				return key, true
			}
		}
	}
	return "", false
}
