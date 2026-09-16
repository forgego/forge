package api

// Serializer field-limit interfaces. These are optional: a Serializer may
// implement any subset. When absent, the viewset preserves its historical
// behavior (all model fields serialized, all input keys populated).
//
//   - Fields() trims response maps to a whitelist.
//   - Exclude() removes response keys.
//   - WriteOnlyFields() removes response keys (accepted on input, never output).
//   - ReadOnlyFields() removes input keys (present in output, ignored on write).
type fieldsProvider interface {
	Fields() []string
}

type excludeProvider interface {
	Exclude() []string
}

type readOnlyFieldsProvider interface {
	ReadOnlyFields() []string
}

type readonlyFieldsProvider interface {
	ReadonlyFields() []string
}

type writeOnlyFieldsProvider interface {
	WriteOnlyFields() []string
}

func serializerFields(s Serializer) []string {
	if s == nil {
		return nil
	}
	if p, ok := s.(fieldsProvider); ok {
		return p.Fields()
	}
	return nil
}

func serializerExclude(s Serializer) []string {
	if s == nil {
		return nil
	}
	if p, ok := s.(excludeProvider); ok {
		return p.Exclude()
	}
	return nil
}

func serializerWriteOnly(s Serializer) []string {
	if s == nil {
		return nil
	}
	if p, ok := s.(writeOnlyFieldsProvider); ok {
		return p.WriteOnlyFields()
	}
	return nil
}

func serializerReadOnly(s Serializer) []string {
	if s == nil {
		return nil
	}
	if p, ok := s.(readOnlyFieldsProvider); ok {
		return p.ReadOnlyFields()
	}
	if p, ok := s.(readonlyFieldsProvider); ok {
		return p.ReadonlyFields()
	}
	return nil
}

// filterOutputMap trims a serialized model map according to the serializer's
// Fields/Exclude/WriteOnlyFields declarations. A nil or empty Fields list
// means "all fields" (historical behavior).
func filterOutputMap(s Serializer, m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return m
	}
	if fields := serializerFields(s); len(fields) > 0 {
		keep := make(map[string]struct{}, len(fields))
		for _, f := range fields {
			keep[f] = struct{}{}
		}
		for k := range m {
			if _, ok := keep[k]; !ok {
				delete(m, k)
			}
		}
	}
	for _, f := range serializerExclude(s) {
		delete(m, f)
	}
	for _, f := range serializerWriteOnly(s) {
		delete(m, f)
	}
	return m
}

// filterOutputMany applies filterOutputMap to each serialized model.
func filterOutputMany(s Serializer, items []map[string]interface{}) []map[string]interface{} {
	for i := range items {
		items[i] = filterOutputMap(s, items[i])
	}
	return items
}

// stripReadOnlyInput removes serializer-declared read-only keys from decoded
// request data before validation and model population.
func stripReadOnlyInput(s Serializer, data map[string]interface{}) {
	if data == nil {
		return
	}
	for _, f := range serializerReadOnly(s) {
		delete(data, f)
	}
}
