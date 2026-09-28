package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	apicore "github.com/forgego/forge/api/core"
	"github.com/forgego/forge/orm"
	validation "github.com/forgego/forge/validate"
	"github.com/go-viper/mapstructure/v2"
)

// SaveModel saves a model instance
func (a *Admin[T]) SaveModel(ctx context.Context, instance *T, isNew bool) error {
	// Apply save hooks if provided
	if a.config.SaveModel != nil {
		return a.config.SaveModel(ctx, a, instance, isNew)
	}

	// Default save behavior
	if isNew {
		return a.manager.Create(ctx, instance)
	}
	return a.manager.Update(ctx, instance)
}

// DeleteModel deletes a model instance
func (a *Admin[T]) DeleteModel(ctx context.Context, instance *T) error {
	// Apply delete hooks if provided
	if a.config.DeleteModel != nil {
		return a.config.DeleteModel(ctx, a, instance)
	}

	// Default delete behavior
	return a.manager.Delete(ctx, instance)
}

func (a *Admin[T]) GetObject(ctx context.Context, id interface{}) (interface{}, error) {
	intID, err := toInt64(id)
	if err != nil {
		return nil, err
	}

	instance, err := a.safeGetObjectByID(ctx, intID)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// validateData checks incoming mutation data against the schema field
// definitions. With partial=false (create) missing required fields are
// rejected; with partial=true (PATCH) only provided fields are checked.
func (a *Admin[T]) validateData(data map[string]interface{}, partial bool) error {
	if a.schema == nil {
		return nil
	}
	fv := validation.NewFieldValidator(validation.NewValidator())
	errs := &validation.ValidationErrors{}
	for _, field := range a.schema.Fields() {
		value, present := data[field.Name]
		if !present {
			if !partial && field.Required {
				errs.Add(field.Name, "is required")
			}
			continue
		}
		if err := fv.ValidateField(field, value); err != nil {
			errs.Add(field.Name, err.Error())
		}
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (a *Admin[T]) CreateObject(ctx context.Context, data map[string]interface{}) (interface{}, error) {
	filtered, err := a.filterWritable(ctx, nil, true, data)
	if err != nil {
		return nil, err
	}

	// Full validation: missing required fields are rejected.
	if err := a.validateData(filtered, false); err != nil {
		return nil, err
	}

	// Create new instance
	var instance T

	// Map data to instance fields
	if err := a.decodeData(filtered, &instance); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	if err := a.SaveModel(ctx, &instance, true); err != nil {
		return nil, err
	}

	user, _ := apicore.UserFromContext(ctx)
	objID := a.getObjectID(&instance)
	repr := a.getObjectLabel(&instance)
	changesJSON, _ := json.Marshal(filtered)
	_ = a.LogAction(ctx, user, fmt.Sprintf("%v", objID), repr, ActionAdd, string(changesJSON))

	return &instance, nil
}

func (a *Admin[T]) UpdateObject(ctx context.Context, id interface{}, data map[string]interface{}) (interface{}, error) {
	intID, err := toInt64(id)
	if err != nil {
		return nil, err
	}

	// Ensure object exists (and permission hooks receive a concrete object path).
	instance, err := a.safeGetObjectByID(ctx, intID)
	if err != nil {
		return nil, err
	}

	filtered, err := a.filterWritable(ctx, instance, false, data)
	if err != nil {
		return nil, err
	}

	// Partial validation: provided fields must be valid, but omitted
	// required fields are fine (PATCH semantics).
	if err := a.validateData(filtered, true); err != nil {
		return nil, err
	}

	// PATCH semantics: only update provided fields and avoid writing zero-values
	// for fields omitted from request payload.
	updates := orm.UpdateMap{}
	for key, value := range filtered {
		if strings.EqualFold(key, "id") {
			continue
		}
		updates[key] = value
	}
	if len(updates) == 0 {
		return instance, nil
	}
	if err := a.manager.UpdateFields(ctx, intID, updates); err != nil {
		return nil, err
	}

	updated, err := a.safeGetObjectByID(ctx, intID)
	if err != nil {
		return nil, err
	}

	user, _ := apicore.UserFromContext(ctx)
	repr := a.getObjectLabel(updated)
	changesJSON, _ := json.Marshal(filtered)
	_ = a.LogAction(ctx, user, fmt.Sprintf("%v", intID), repr, ActionChange, string(changesJSON))

	return updated, nil
}

func (a *Admin[T]) DeleteObject(ctx context.Context, id interface{}) error {
	intID, err := toInt64(id)
	if err != nil {
		return err
	}

	instance, err := a.safeGetObjectByID(ctx, intID)
	if err != nil {
		return err
	}
	repr := a.getObjectLabel(instance)
	if err := a.DeleteModel(ctx, instance); err != nil {
		return err
	}

	user, _ := apicore.UserFromContext(ctx)
	_ = a.LogAction(ctx, user, fmt.Sprintf("%v", intID), repr, ActionDelete, "")

	return nil
}

func (a *Admin[T]) safeGetObjectByID(ctx context.Context, id int64) (*T, error) {
	// Primary path: direct lookup by manager.
	// Some model/config combinations can panic inside typed filter resolution.
	var recovered any
	var getErr error
	var obj *T
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = r
			}
		}()
		obj, getErr = a.manager.Get(ctx, id)
	}()
	if recovered == nil && getErr == nil && obj != nil {
		return obj, nil
	}

	// Fallback path: scan all records and match by extracted ID.
	// This is slower but keeps admin operations working while ORM lookup
	// expression typing is being hardened.
	all, listErr := a.manager.All(ctx)
	if listErr != nil {
		if getErr != nil {
			return nil, getErr
		}
		if recovered != nil {
			return nil, fmt.Errorf("failed to get object with id %d: recovered panic %v", id, recovered)
		}
		return nil, listErr
	}
	for _, candidate := range all {
		candidateID := a.getObjectID(candidate)
		parsedID, parseErr := toInt64(candidateID)
		if parseErr != nil {
			continue
		}
		if parsedID == id {
			return candidate, nil
		}
	}

	if getErr != nil {
		return nil, getErr
	}
	if recovered != nil {
		return nil, fmt.Errorf("object with id %d not found after recovered panic: %v", id, recovered)
	}
	return nil, fmt.Errorf("object with id %d not found", id)
}

// decodeData uses mapstructure to decode a map into the model struct
func (a *Admin[T]) decodeData(data map[string]interface{}, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           result,
		TagName:          "json",
		WeaklyTypedInput: true,
		Squash:           true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			stringToDateTimeHook(),
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(data)
}

// stringToDateTimeHook handles conversion from string (ISO8601 or similar) to time.Time
func stringToDateTimeHook() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		str := data.(string)
		if str == "" {
			return time.Time{}, nil
		}

		// Try multiple formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04",
			"2006-01-02",
		}

		for _, format := range formats {
			if b, err := time.Parse(format, str); err == nil {
				return b, nil
			}
		}

		return data, nil
	}
}

// getObjectID returns the primary key value of an object
func (a *Admin[T]) getObjectID(obj *T) interface{} {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Try common ID fields
	for _, name := range []string{"ID", "Id", "id"} {
		f := val.FieldByName(name)
		if f.IsValid() {
			return f.Interface()
		}
	}

	return nil
}

// getObjectLabel returns a descriptive label for an object
func (a *Admin[T]) getObjectLabel(obj *T) string {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 1. Try common label fields
	for _, name := range []string{"Name", "Title", "Label", "Email", "Username", "DisplayName", "Subject", "Description", "Code", "Slug"} {
		f := val.FieldByName(name)
		if f.IsValid() {
			str := fmt.Sprintf("%v", f.Interface())
			if len(str) > 100 {
				return str[:97] + "..."
			}
			return str
		}
	}

	// 2. Try ID
	modelLabel := a.name
	if a.metadata != nil && a.metadata.VerboseName != "" {
		modelLabel = a.metadata.VerboseName
	} else if a.config != nil && a.config.VerboseName != "" {
		modelLabel = a.config.VerboseName
	}

	id := a.getObjectID(obj)
	if id != nil {
		return fmt.Sprintf("%s #%v", modelLabel, id)
	}

	return modelLabel
}

// toInt64 converts interface{} to int64
func toInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("id is nil")
	}

	// Handle standard types
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		if val > math.MaxInt64 {
			return 0, fmt.Errorf("uint value %d exceeds max int64", val)
		}
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		if val > math.MaxInt64 {
			return 0, fmt.Errorf("uint64 value %d exceeds max int64", val)
		}
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil // Care needed for precision loss
	case string:
		// Attempt parsing
		var i int64
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i, nil
		}
		return 0, fmt.Errorf("cannot parse string %q as int64", val)
	}

	// Fallback to absolute strict reflection (unlikely needed with above switch)
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := rv.Uint()
		if u > math.MaxInt64 {
			return 0, fmt.Errorf("uint value %d exceeds max int64", u)
		}
		return int64(u), nil
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), nil
	}

	return 0, fmt.Errorf("cannot convert %T to int64", v)
}
