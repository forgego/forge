package api

import (
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/schema"
)

// unknownRequestFieldMessage is the per-key error for a rejected request key.
const unknownRequestFieldMessage = "Unknown field."

// allowRequestFields enforces RejectUnknownRequestFields. It runs before
// serializer validation, model validation, hooks and persistence, so a
// rejected request changes nothing.
func (vs *BaseViewSet) allowRequestFields(w http.ResponseWriter, r *http.Request, serializer Serializer, data map[string]interface{}) bool {
	if !vs.RejectUnknownRequestFields {
		return true
	}
	ignored := append([]string{"id", "ID", "Id"}, vs.ReadOnlyRequestFields...)
	pkGoName, pkDBName, pkJSONName := vs.getSchemaPrimaryKeyField()
	ignored = append(ignored, pkGoName, pkDBName, pkJSONName)
	unknown := unknownRequestFields(vs.Model, serializer, data, ignored)
	if len(unknown) == 0 {
		return true
	}
	errs := make(map[string][]string, len(unknown))
	for _, key := range unknown {
		errs[key] = []string{unknownRequestFieldMessage}
	}
	vs.handleException(w, r, exceptions.NewValidationError(errs))
	return false
}

// unknownRequestFields returns the sorted keys of data that the viewset
// neither writes nor deliberately ignores. A key is known when it is:
//   - the request key of a model field without a schema entry (its json
//     name, or its Go name under json:",omitempty", as responses name it);
//   - any name of a schema field (schema, column, db tag, json or Go name)
//     that clients can see or write: Serialize or Editable, or write-only
//     (Editable and not Serialize) when the struct hides it with json:"-";
//   - declared by the serializer in Fields, ReadOnlyFields or
//     WriteOnlyFields;
//   - an ignored key (primary-key aliases and ReadOnlyRequestFields),
//     matched case-insensitively like populateFromMap does, unless it names
//     a hidden field.
//
// Read-only keys that clients echo back, such as id or created_at, are
// therefore known and stay silently ignored. A hidden field (neither
// serialized nor writable) is reported like a key that does not exist, so
// the response does not reveal it.
func unknownRequestFields(model interface{}, serializer Serializer, data map[string]interface{}, ignoredKeys []string) []string {
	if len(data) == 0 {
		return nil
	}
	known, hidden := knownRequestFields(model)
	for _, names := range [][]string{serializerFields(serializer), serializerReadOnly(serializer), serializerWriteOnly(serializer)} {
		for _, name := range names {
			known[name] = struct{}{}
		}
	}
	ignored := make(map[string]struct{}, len(ignoredKeys))
	for _, key := range ignoredKeys {
		if key != "" {
			ignored[strings.ToLower(key)] = struct{}{}
		}
	}

	var unknown []string
	for key := range data {
		if _, ok := known[key]; ok {
			continue
		}
		if _, isHidden := hidden[strings.ToLower(key)]; !isHidden {
			if _, ok := ignored[strings.ToLower(key)]; ok {
				continue
			}
		}
		unknown = append(unknown, key)
	}
	sort.Strings(unknown)
	return unknown
}

// knownRequestFields returns the request keys that name a model field
// clients can see or write, and the lowercased names of the hidden schema
// fields.
func knownRequestFields(model interface{}) (known, hidden map[string]struct{}) {
	known = make(map[string]struct{})
	hidden = make(map[string]struct{})
	walkRequestFields(model, func(field requestField) {
		if field.schema == nil {
			if field.key != "" {
				known[field.key] = struct{}{}
			}
			return
		}
		names := resolvedFieldNames(model, *field.schema)
		visible := field.key != "" && (field.schema.Serialize || field.schema.Editable)
		writeOnly := field.key == "" && field.writable()
		if !visible && !writeOnly {
			for _, name := range names {
				hidden[strings.ToLower(name)] = struct{}{}
			}
			return
		}
		if field.key != "" {
			known[field.key] = struct{}{}
		}
		for _, name := range names {
			known[name] = struct{}{}
		}
	})
	for name := range known {
		delete(hidden, strings.ToLower(name))
	}
	return known, hidden
}

// requestField is a struct field as request input sees it. populateFromMap,
// knownRequestFields and missingRequiredFields share it, so they agree on
// which fields a request writes and under which keys.
type requestField struct {
	goName string
	dbTag  string
	// key is the field's own request key, named as modelToMap names it in
	// responses: the json tag name, or the Go name when the tag leaves the
	// name empty (json:",omitempty"). It is "" when the field has no json
	// tag or is tagged json:"-".
	key string
	// schema is the schema field backing the struct field, or nil.
	schema *schema.Field
}

func newRequestField(schemaFields map[string]*schema.Field, field reflect.StructField) requestField {
	f := requestField{goName: field.Name, dbTag: strings.Split(field.Tag.Get("db"), ",")[0]}
	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		f.key = strings.Split(tag, ",")[0]
		if f.key == "" {
			f.key = field.Name
		}
		if f.key == "-" {
			f.key = ""
		}
	}
	f.schema = schemaFieldForStructField(schemaFields, field, f.key, f.dbTag)
	return f
}

// writable reports whether request input may write the field. A field
// hidden from JSON is writable only when the schema declares it write-only
// (editable, never serialized), such as a password; anything else hidden
// stays out of reach of request input.
func (f requestField) writable() bool {
	if f.key != "" {
		return true
	}
	return f.schema != nil && f.schema.Editable && !f.schema.Serialize
}

// ignoredBy reports whether ignored, a set of lowercased keys, names the
// field under any of its names.
func (f requestField) ignoredBy(model interface{}, ignored map[string]bool) bool {
	names := []string{f.key, f.goName, f.dbTag}
	if f.schema != nil {
		names = append(names, resolvedFieldNames(model, *f.schema)...)
	}
	for _, name := range names {
		if name != "" && ignored[strings.ToLower(name)] {
			return true
		}
	}
	return false
}

// lookup returns the value data holds for the field and the key it was
// found under. Keys match exactly: the field's own key first, then its
// schema aliases.
func (f requestField) lookup(model interface{}, data map[string]interface{}) (string, interface{}, bool) {
	if f.key != "" {
		if value, ok := data[f.key]; ok {
			return f.key, value, true
		}
	}
	if f.schema != nil {
		for _, alias := range resolvedFieldNames(model, *f.schema) {
			if value, ok := data[alias]; ok {
				return alias, value, true
			}
		}
	}
	return "", nil, false
}

// walkRequestFields calls fn for each exported, non-embedded field of
// model's struct type, descending into embedded structs.
func walkRequestFields(model interface{}, fn func(requestField)) {
	if model == nil {
		return
	}
	modelType := reflect.TypeOf(model)
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return
	}
	schemaFields := schemaFieldsByRequestName(model)

	var walk func(t reflect.Type, depth int)
	walk = func(t reflect.Type, depth int) {
		if depth > 8 {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			if field.Anonymous {
				embedded := field.Type
				if embedded.Kind() == reflect.Ptr {
					embedded = embedded.Elem()
				}
				if embedded.Kind() == reflect.Struct {
					walk(embedded, depth+1)
				}
				continue
			}
			fn(newRequestField(schemaFields, field))
		}
	}
	walk(modelType, 0)
}
