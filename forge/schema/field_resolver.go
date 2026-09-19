package schema

import (
	"reflect"
	"strings"
	"unicode"
)

// ResolvedField describes the concrete Go field backing a schema field.
// StructField.Index contains the complete path through anonymous embeddings.
type ResolvedField struct {
	SchemaName  string
	DBColumn    string
	DBTag       string
	JSONName    string
	GoName      string
	StructField reflect.StructField
}

// Names returns every non-empty name by which the field can be addressed.
func (f ResolvedField) Names() []string {
	names := []string{f.SchemaName, f.DBColumn, f.DBTag, f.JSONName, f.GoName}
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name == "" || name == "-" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}

// ResolveField finds the concrete struct field backing field. It searches
// anonymous embedded structs and matches schema, column, tag, and Go names.
func ResolveField(model interface{}, field Field) (ResolvedField, bool) {
	t := reflect.TypeOf(model)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return ResolvedField{}, false
	}

	wanted := compactFieldName(field.Name)
	wantedColumn := compactFieldName(field.DBColumn)
	var normalized *ResolvedField
	var search func(reflect.Type, []int) (ResolvedField, bool)
	search = func(current reflect.Type, prefix []int) (ResolvedField, bool) {
		for i := 0; i < current.NumField(); i++ {
			structField := current.Field(i)
			if !structField.IsExported() {
				continue
			}
			index := append(append([]int(nil), prefix...), i)
			fieldType := structField.Type
			if structField.Anonymous {
				for fieldType.Kind() == reflect.Ptr {
					fieldType = fieldType.Elem()
				}
				if fieldType.Kind() == reflect.Struct {
					if found, ok := search(fieldType, index); ok {
						return found, true
					}
				}
				continue
			}

			jsonName := tagName(structField.Tag.Get("json"))
			dbTag := tagName(structField.Tag.Get("db"))
			candidate := ResolvedField{
				SchemaName:  field.Name,
				DBColumn:    field.DBColumn,
				DBTag:       dbTag,
				JSONName:    jsonName,
				GoName:      structField.Name,
				StructField: structField,
			}
			candidate.StructField.Index = index

			for _, name := range []string{structField.Name, jsonName, dbTag} {
				if name != "" && name != "-" &&
					(strings.EqualFold(name, field.Name) || strings.EqualFold(name, field.DBColumn)) {
					return candidate, true
				}
			}
			if normalized == nil {
				for _, name := range []string{structField.Name, jsonName, dbTag} {
					compact := compactFieldName(name)
					if compact != "" && (compact == wanted || compact == wantedColumn) {
						copy := candidate
						normalized = &copy
						break
					}
				}
			}
		}
		return ResolvedField{}, false
	}

	if found, ok := search(t, nil); ok {
		return found, true
	}
	if normalized != nil {
		return *normalized, true
	}
	return ResolvedField{}, false
}

func tagName(tag string) string {
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func compactFieldName(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, name)
}
