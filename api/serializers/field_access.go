package serializers

import (
	"fmt"
	"reflect"

	"github.com/forgego/forge/schema"
)

// resolveTypedField is the shared read/write path for both schema-derived and
// manually configured fields. Schema fields keep their resolved embedded index.
func resolveTypedField[T any](instance *T, field TypedField[T]) (reflect.Value, error) {
	v := reflect.ValueOf(instance)
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, fmt.Errorf("instance is nil")
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("instance must be a struct")
	}
	index := field.fieldIndex
	if len(index) == 0 {
		resolved, ok := schema.ResolveField(instance, schema.Field{Name: field.fieldExpr.Path()})
		if !ok {
			return reflect.Value{}, fmt.Errorf("model field %q not found", field.fieldExpr.Path())
		}
		index = resolved.StructField.Index
	}
	value, err := v.FieldByIndexErr(index)
	if err != nil {
		return reflect.Value{}, err
	}
	if !value.IsValid() || !value.CanInterface() {
		return reflect.Value{}, fmt.Errorf("model field %q is inaccessible", field.fieldExpr.Path())
	}
	return value, nil
}

func valueForField(value interface{}, target reflect.Type) (reflect.Value, error) {
	if value == nil {
		switch target.Kind() {
		case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Func, reflect.Chan:
			return reflect.Zero(target), nil
		default:
			return reflect.Value{}, fmt.Errorf("null cannot be assigned to %s", target)
		}
	}
	v := reflect.ValueOf(value)
	if v.Type().AssignableTo(target) {
		return v, nil
	}
	if v.Type().ConvertibleTo(target) && numericConversionFits(v, target) {
		return v.Convert(target), nil
	}
	return reflect.Value{}, fmt.Errorf("expected %s, got %s", target, v.Type())
}

func numericConversionFits(v reflect.Value, target reflect.Type) bool {
	sourceKind, targetKind := v.Kind(), target.Kind()
	signed := func(k reflect.Kind) bool { return k >= reflect.Int && k <= reflect.Int64 }
	unsigned := func(k reflect.Kind) bool { return k >= reflect.Uint && k <= reflect.Uintptr }
	float := func(k reflect.Kind) bool { return k == reflect.Float32 || k == reflect.Float64 }
	switch {
	case signed(sourceKind) && signed(targetKind):
		return !reflect.Zero(target).OverflowInt(v.Int())
	case unsigned(sourceKind) && unsigned(targetKind):
		return !reflect.Zero(target).OverflowUint(v.Uint())
	case signed(sourceKind) && unsigned(targetKind):
		return v.Int() >= 0 && !reflect.Zero(target).OverflowUint(uint64(v.Int()))
	case unsigned(sourceKind) && signed(targetKind):
		return v.Uint() <= uint64(^uint64(0)>>1) && !reflect.Zero(target).OverflowInt(int64(v.Uint()))
	case float(sourceKind) && float(targetKind):
		return !reflect.Zero(target).OverflowFloat(v.Float()) && v.Convert(target).Float() == v.Float()
	default:
		return false
	}
}
