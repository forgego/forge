package schema

import "reflect"

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
