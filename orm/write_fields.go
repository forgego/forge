package orm

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/forgego/forge/schema"
)

// Write-column rules for Manager.Create, Manager.Update and Manager.Save.
//
// Forge follows Django: the INSERT writes the value the instance holds, and a
// schema Default is applied when the instance is constructed (by the API
// viewset for keys a request omits, or by Manager.New / ApplyDefaults in Go
// code), never at INSERT time. A Go struct cannot say "not set", so Create
// cannot tell an explicit false from an untouched one; writing the value is
// the only rule under which an explicit zero survives. Create therefore omits
// a column only when the database must fill it or when a zero value cannot
// mean a real value:
//
//   - an auto-increment or zero primary key;
//   - a generated column;
//   - a zero AutoNow or AutoNowAdd timestamp (the column default fills it);
//   - a zero value on a field with a DBDefault (the database default is the
//     field's "unset" value; use a pointer field to write an explicit zero);
//   - a nil pointer, slice, map or interface, or a zero struct such as a zero
//     time.Time (NULL, or the column default);
//   - a zero foreign key (NULL rather than a reference to row 0);
//   - a zero value on a unique, optional field (NULL, so two blank rows do not
//     collide on the unique constraint).
//
// Every other column is written, including false, 0 and "".

// insertOmitsZeroValue reports whether Create leaves column out of the
// INSERT, following the rules above. value is the field's current value.
func insertOmitsZeroValue(field schema.Field, value interface{}, fkColumns map[string]bool) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return true
	}
	if !rv.IsZero() {
		return false
	}
	if field.PrimaryKey || field.AutoNow || field.AutoNowAdd || field.DBDefault != "" {
		return true
	}
	if field.Required {
		// A required field is always written, as before: the INSERT fails
		// loudly instead of a default hiding the missing value.
		return false
	}
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Struct, reflect.Array, reflect.Chan, reflect.Func:
		return true
	}
	if fkColumns[strings.ToLower(writeColumnName(field))] || field.Type == schema.TypeForeignKey || field.Type == schema.TypeOneToOne {
		return true
	}
	return field.Unique
}

// writeColumnName returns the database column a schema field writes to.
func writeColumnName(field schema.Field) string {
	if field.DBColumn != "" {
		return field.DBColumn
	}
	return field.Name
}

// foreignKeyColumns returns the lower-cased columns of the instance's
// ForeignKey and OneToOne relations. A relation named "author" or
// "author_id" both map to the column "author_id"; a relation whose name is
// already a column name maps to that column.
func foreignKeyColumns(instance schema.Schema) map[string]bool {
	columns := map[string]bool{}
	for _, rel := range instance.Relations() {
		if rel.Type != schema.RelationForeignKey && rel.Type != schema.RelationOneToOne {
			continue
		}
		name := strings.ToLower(rel.Name)
		if name == "" {
			continue
		}
		columns[name] = true
		if !strings.HasSuffix(name, "_id") {
			columns[name+"_id"] = true
		}
	}
	return columns
}

// touchAutoNow sets every AutoNow field of instance to now, so that the
// UPDATE writes it and the caller's struct holds the stored value. It
// supports time.Time, *time.Time and string fields (RFC 3339 with
// nanoseconds, which PostgreSQL and SQLite both parse); a field of any other
// type is left unchanged.
func touchAutoNow(instance interface{}, now time.Time) {
	schemaInstance, ok := instance.(schema.Schema)
	if !ok {
		return
	}
	for _, field := range schemaInstance.Fields() {
		if !field.AutoNow {
			continue
		}
		fv, ok := settableSchemaField(instance, field)
		if !ok {
			continue
		}
		// An unsupported type (sql.NullTime, a custom type) is written as
		// the struct holds it, as before AutoNow was refreshed.
		_ = setTimeValue(fv, now)
	}
}

var timeType = reflect.TypeOf(time.Time{})

func setTimeValue(fv reflect.Value, now time.Time) error {
	switch {
	case fv.Type() == timeType:
		fv.Set(reflect.ValueOf(now))
	case fv.Kind() == reflect.Pointer && fv.Type().Elem() == timeType:
		t := now
		fv.Set(reflect.ValueOf(&t))
	case fv.Kind() == reflect.String:
		fv.SetString(now.Format(time.RFC3339Nano))
	default:
		return fmt.Errorf("unsupported type %s", fv.Type())
	}
	return nil
}

// settableSchemaField resolves field on instance to an addressable struct
// field, following embedded structs. It reports false when the field cannot
// be reached or set.
func settableSchemaField(instance interface{}, field schema.Field) (reflect.Value, bool) {
	resolved, ok := schema.ResolveField(instance, field)
	if !ok {
		return reflect.Value{}, false
	}
	value := reflect.ValueOf(instance)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return reflect.Value{}, false
	}
	value = value.Elem()
	for _, index := range resolved.StructField.Index {
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return reflect.Value{}, false
			}
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		value = value.Field(index)
	}
	if !value.IsValid() || !value.CanSet() {
		return reflect.Value{}, false
	}
	return value, true
}

// ApplyDefaults sets every field of instance that declares a schema Default
// and currently holds its Go zero value to that default. Call it on a new
// instance before assigning the caller's values, as Manager.New does: it runs
// at construction time, like a Django model's defaults, so a zero value
// assigned afterwards is written as is. instance must be a non-nil pointer to
// a struct implementing schema.Schema; fields whose default cannot be
// converted to the Go field type return an error.
func ApplyDefaults(instance interface{}) error {
	schemaInstance, ok := instance.(schema.Schema)
	if !ok {
		return fmt.Errorf("instance must implement schema.Schema")
	}
	for _, field := range schemaInstance.Fields() {
		if field.Default == nil {
			continue
		}
		fv, ok := settableSchemaField(instance, field)
		if !ok || !fv.IsZero() {
			continue
		}
		if err := assignDefault(fv, field.Default); err != nil {
			return fmt.Errorf("apply default for %s: %w", field.Name, err)
		}
	}
	return nil
}

func assignDefault(fv reflect.Value, def interface{}) error {
	dv := reflect.ValueOf(def)
	if dv.Kind() == reflect.Func {
		// A callable default, such as time.Now, is evaluated per instance.
		if dv.Type().NumIn() != 0 || dv.Type().NumOut() != 1 {
			return fmt.Errorf("callable default must take no arguments and return one value")
		}
		dv = dv.Call(nil)[0]
	}
	target := fv.Type()
	if target.Kind() == reflect.Pointer {
		converted, err := convertDefault(dv, target.Elem())
		if err != nil {
			return err
		}
		ptr := reflect.New(target.Elem())
		ptr.Elem().Set(converted)
		fv.Set(ptr)
		return nil
	}
	converted, err := convertDefault(dv, target)
	if err != nil {
		return err
	}
	fv.Set(converted)
	return nil
}

func convertDefault(dv reflect.Value, target reflect.Type) (reflect.Value, error) {
	if target.Kind() == reflect.String && dv.Kind() != reflect.String {
		// A numeric or boolean default on a string-backed field, such as
		// Default(0) on a decimal stored as string, is formatted, never
		// converted rune-wise (int(65) would become "A").
		switch dv.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return reflect.ValueOf(fmt.Sprint(dv.Interface())).Convert(target), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot convert %s to %s", dv.Type(), target)
	}
	if !dv.Type().ConvertibleTo(target) {
		return reflect.Value{}, fmt.Errorf("cannot convert %s to %s", dv.Type(), target)
	}
	return dv.Convert(target), nil
}
