package api

import (
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/forgego/forge/api/exceptions"
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
//   - the json name of a model field without a schema entry;
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
	if model == nil {
		return known, hidden
	}
	modelType := reflect.TypeOf(model)
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return known, hidden
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
			key := strings.Split(field.Tag.Get("json"), ",")[0]
			jsonHidden := key == "" || key == "-"
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			fieldSchema := schemaFieldForStructField(schemaFields, field, key, dbTag)
			if fieldSchema == nil {
				if !jsonHidden {
					known[key] = struct{}{}
				}
				continue
			}
			names := resolvedFieldNames(model, *fieldSchema)
			visible := !jsonHidden && (fieldSchema.Serialize || fieldSchema.Editable)
			writeOnly := jsonHidden && fieldSchema.Editable && !fieldSchema.Serialize
			if !visible && !writeOnly {
				for _, name := range names {
					hidden[strings.ToLower(name)] = struct{}{}
				}
				continue
			}
			if !jsonHidden {
				known[key] = struct{}{}
			}
			for _, name := range names {
				known[name] = struct{}{}
			}
		}
	}
	walk(modelType, 0)
	for name := range known {
		delete(hidden, strings.ToLower(name))
	}
	return known, hidden
}
