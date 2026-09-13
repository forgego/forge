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
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/validate"
	"github.com/go-viper/mapstructure/v2"
)

// Admin represents a type-safe admin configuration for a model
// It fully integrates with schema, ORM, and filter systems
type Admin[T any] struct {
	// Core dependencies
	schema      schema.Schema
	manager     *orm.Manager[T]
	modelSchema *orm.ModelSchema
	config      *Config[T]

	// Discovered information (cached)
	metadata *Metadata

	// Internal
	name string
}

// NewAdmin creates a new Admin instance
func NewAdmin[T any](
	schemaInstance schema.Schema,
	manager *orm.Manager[T],
	config *Config[T],
) (*Admin[T], error) {
	// Get model schema from ORM
	modelSchema, err := orm.GetModelSchema[T]()
	if err != nil {
		return nil, fmt.Errorf("failed to get model schema: %w", err)
	}

	// Get model name
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	name := typ.Name()

	// Use table name from meta if available
	meta := schemaInstance.Meta()
	if meta.TableName != "" {
		name = meta.TableName
	}

	// Apply defaults to config
	if config == nil {
		config = &Config[T]{}
	}
	applyConfigDefaults(config, schemaInstance)

	admin := &Admin[T]{
		schema:      schemaInstance,
		manager:     manager,
		modelSchema: modelSchema,
		config:      config,
		name:        name,
	}

	return admin, nil
}

// SetDB sets the database connection for the admin manager
func (a *Admin[T]) SetDB(database *db.DB) {
	if a.manager != nil {
		a.manager.SetDB(database)
	}
}

// ModelName returns the name of the model
func (a *Admin[T]) ModelName() string {
	return a.name
}

// ModelType returns the model type
func (a *Admin[T]) ModelType() reflect.Type {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

// Schema returns the schema instance
func (a *Admin[T]) Schema() schema.Schema {
	return a.schema
}

// Manager returns the ORM manager
func (a *Admin[T]) Manager() *orm.Manager[T] {
	return a.manager
}

// ModelSchema returns the ORM model schema
func (a *Admin[T]) ModelSchema() *orm.ModelSchema {
	return a.modelSchema
}

// Config returns the configuration
func (a *Admin[T]) Config() *Config[T] {
	return a.config
}

// GetMetadata returns the metadata for this admin
// This is called by API handlers to send to frontend.
// Schema-derived parts are cached; per-request Permissions are computed
// fresh on a copy so concurrent users never see each other's permissions.
func (a *Admin[T]) GetMetadata(ctx context.Context, user interface{}) (*Metadata, error) {
	// Return cached metadata if available
	if a.metadata != nil {
		// Copy: never mutate the shared cache per-user
		meta := *a.metadata
		meta.Permissions = a.getPermissionsMetadata(ctx, user)
		return &meta, nil
	}

	// Build metadata from schema
	meta, err := buildMetadata(a.schema, a.config, a.name)
	if err != nil {
		return nil, err
	}

	// Add permissions
	meta.Permissions = a.getPermissionsMetadata(ctx, user)

	// Add page type
	meta.PageType = a.PageType()

	// Cache it
	a.metadata = meta

	return meta, nil
}

// GetQueryset returns the base queryset for this admin
func (a *Admin[T]) GetQueryset(ctx context.Context) (orm.QuerySet[T], error) {
	// Get base queryset from manager
	// Use pointer to Q for empty filter (all records)
	qs, err := a.manager.Filter(&orm.Q{})
	if err != nil {
		return nil, err
	}

	// Apply custom queryset hook if provided
	if a.config.GetQueryset != nil {
		return a.config.GetQueryset(ctx, a, qs)
	}

	return qs, nil
}

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

// PageType returns the preferred layout for this admin
func (a *Admin[T]) PageType() string {
	if a.config.PageType != "" {
		return a.config.PageType
	}
	return "list" // Default
}

// Data operations

func (a *Admin[T]) ListObjects(ctx context.Context, params ListParams) (*PaginatedResponse, error) {
	qs, err := a.GetQueryset(ctx)
	if err != nil {
		return nil, err
	}

	// Apply search
	if params.Search != "" && len(a.config.SearchFields) > 0 {
		searchQueries := make([]orm.Expression, 0, len(a.config.SearchFields))
		for _, field := range a.config.SearchFields {
			// Use F() helper for dynamic field names
			searchQueries = append(searchQueries, orm.F(orm.ExtractPathFromAny(field)).IContains(params.Search))
		}
		qs = qs.Filter(orm.Or(searchQueries...))
	}

	// Apply filters
	for key, value := range params.Filters {
		// Check for manual filter overrides
		applied := false
		for _, filter := range a.config.Filters {
			if filter.Name == key && filter.Handler != nil {
				qs = filter.Handler(ctx, qs, value)
				applied = true
				break
			}
		}
		if !applied {
			// Parse lookup (e.g. price__gt, name__contains)
			field, lookup := parseLookup(key)

			// Create expression based on lookup
			var expr orm.Expression
			f := orm.F(field)

			switch lookup {
			case "exact":
				expr = f.Eq(value)
			case "ne":
				expr = f.Ne(value)
			case "gt":
				expr = f.Gt(value)
			case "gte":
				expr = f.Gte(value)
			case "lt":
				expr = f.Lt(value)
			case "lte":
				expr = f.Lte(value)
			case "contains":
				if s, ok := value.(string); ok {
					expr = f.Contains(s)
				} else {
					expr = f.Eq(value) // Fallback
				}
			case "icontains":
				if s, ok := value.(string); ok {
					expr = f.IContains(s)
				} else {
					expr = f.Eq(value) // Fallback
				}
			case "startswith":
				if s, ok := value.(string); ok {
					expr = f.StartsWith(s)
				} else {
					expr = f.Eq(value)
				}
			case "endswith":
				if s, ok := value.(string); ok {
					expr = f.EndsWith(s)
				} else {
					expr = f.Eq(value)
				}
			case "in":
				// value should be slice or comma-separated string
				// For now assume value is single string from query param
				if s, ok := value.(string); ok {
					parts := strings.Split(s, ",")
					args := make([]interface{}, 0, len(parts))
					for _, v := range parts {
						if trimmed := strings.TrimSpace(v); trimmed != "" {
							args = append(args, trimmed)
						}
					}
					if len(args) > 0 {
						expr = f.In(args...)
					} else {
						expr = f.Eq(value)
					}
				} else {
					expr = f.Eq(value)
				}
			case "isnull":
				if s, ok := value.(string); ok && (s == "true" || s == "1") {
					expr = f.IsNull()
				} else {
					expr = f.IsNotNull()
				}
			default:
				expr = f.Eq(value)
			}

			qs = qs.Filter(expr)
		}
	}

	// Apply ordering (request ordering wins; unknown fields are dropped
	// so a bad `ordering` param can never produce a SQL error)
	if validOrdering := sanitizeOrdering(params.Ordering, a.orderableFields()); len(validOrdering) > 0 {
		ordering := make([]any, len(validOrdering))
		for i, v := range validOrdering {
			ordering[i] = v
		}
		qs = qs.OrderBy(ordering...)
	} else if len(a.config.Ordering) > 0 {
		ordering := make([]any, len(a.config.Ordering))
		for i, v := range a.config.Ordering {
			ordering[i] = v
		}
		qs = qs.OrderBy(ordering...)
	}

	// Apply pagination (guard against zero/negative/huge values so
	// direct callers can never produce a negative offset or div-by-zero)
	page := params.Page
	if page < 1 {
		page = 1
	}
	limit := params.PageSize
	if limit <= 0 {
		limit = a.config.ListPerPage
	}
	if limit <= 0 {
		limit = 25
	}
	const maxListLimit = 1000
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := (page - 1) * limit

	// Apply pagination
	count, err := qs.Count(ctx)
	if err != nil {
		return nil, err
	}

	results, err := qs.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := (int(count) + limit - 1) / limit

	return &PaginatedResponse{
		Count:      count,
		PageSize:   limit,
		Page:       page,
		TotalPages: totalPages,
		Results:    results,
	}, nil
}

// orderableFields returns the set of schema field names that may be used
// for ordering (plus the conventional "id"/"pk" aliases).
func (a *Admin[T]) orderableFields() map[string]bool {
	fields := make(map[string]bool)
	if a.schema != nil {
		for _, f := range a.schema.Fields() {
			if f.Name != "" {
				fields[f.Name] = true
			}
		}
	}
	fields["id"] = true
	fields["pk"] = true
	return fields
}

// sanitizeOrdering drops ordering keys that reference unknown fields or
// contain unsafe characters, so user input can never break the query.
func sanitizeOrdering(ordering []string, valid map[string]bool) []string {
	sanitized := make([]string, 0, len(ordering))
	for _, key := range ordering {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		desc := strings.HasPrefix(key, "-")
		name := strings.TrimPrefix(key, "-")
		if !isSafeOrderField(name) {
			continue
		}
		if !valid[name] {
			continue
		}
		if desc {
			sanitized = append(sanitized, "-"+name)
		} else {
			sanitized = append(sanitized, name)
		}
	}
	return sanitized
}

// isSafeOrderField reports whether name is a plain field identifier.
func isSafeOrderField(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' {
			continue
		}
		return false
	}
	return true
}

func (a *Admin[T]) Autocomplete(ctx context.Context, query string, limit int) ([]AutocompleteItem, error) {
	qs, err := a.GetQueryset(ctx)
	if err != nil {
		return nil, err
	}

	// Apply search
	if query != "" && len(a.config.SearchFields) > 0 {
		searchQueries := make([]orm.Expression, 0, len(a.config.SearchFields))
		for _, field := range a.config.SearchFields {
			searchQueries = append(searchQueries, orm.F(orm.ExtractPathFromAny(field)).IContains(query))
		}
		qs = qs.Filter(orm.Or(searchQueries...))
	}

	// Fetch results
	results, err := qs.Limit(limit).All(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to AutocompleteItem
	items := make([]AutocompleteItem, 0, len(results))
	for _, res := range results {
		items = append(items, AutocompleteItem{
			Value: a.getObjectID(res),
			Label: a.getObjectLabel(res),
		})
	}

	return items, nil
}

// LabelResolver is implemented by admins that can label objects by id in bulk.
type LabelResolver interface {
	ObjectLabels(ctx context.Context, ids []interface{}) (map[string]string, error)
}

// ObjectLabels resolves human-readable labels for objects by their primary keys in bulk.
func (a *Admin[T]) ObjectLabels(ctx context.Context, ids []interface{}) (map[string]string, error) {
	if len(ids) == 0 {
		return make(map[string]string), nil
	}

	if len(ids) > 1000 {
		ids = ids[:1000]
	}

	qs, err := a.GetQueryset(ctx)
	if err != nil {
		return nil, err
	}

	pkField := "id"
	if a.modelSchema != nil && a.modelSchema.PrimaryKey != "" {
		pkField = a.modelSchema.PrimaryKey
	}

	normalizedIDs := make([]interface{}, len(ids))
	for i, id := range ids {
		if f, ok := id.(float64); ok && f == math.Floor(f) && !math.IsNaN(f) && !math.IsInf(f, 0) {
			normalizedIDs[i] = int64(f)
		} else {
			normalizedIDs[i] = id
		}
	}

	qs = qs.Filter(orm.F(pkField).In(normalizedIDs...))

	results, err := qs.All(ctx)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string, len(results))
	for _, obj := range results {
		objID := a.getObjectID(obj)
		if objID != nil {
			labels[fmt.Sprint(objID)] = a.getObjectLabel(obj)
		}
	}

	return labels, nil
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

func (a *Admin[T]) CreateObject(ctx context.Context, data map[string]interface{}) (interface{}, error) { // Validate incoming data against the schema before touching the DB.
	// Full validation: missing required fields are rejected.
	if err := a.validateData(data, false); err != nil {
		return nil, err
	}

	// Create new instance
	var instance T

	// Map data to instance fields
	if err := a.decodeData(data, &instance); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	if err := a.SaveModel(ctx, &instance, true); err != nil {
		return nil, err
	}

	user, _ := apicore.UserFromContext(ctx)
	objID := a.getObjectID(&instance)
	repr := a.getObjectLabel(&instance)
	changesJSON, _ := json.Marshal(data)
	_ = a.LogAction(ctx, user, fmt.Sprintf("%v", objID), repr, ActionAdd, string(changesJSON))

	return &instance, nil
}

func (a *Admin[T]) UpdateObject(ctx context.Context, id interface{}, data map[string]interface{}) (interface{}, error) {
	intID, err := toInt64(id)
	if err != nil {
		return nil, err
	}

	// Partial validation: provided fields must be valid, but omitted
	// required fields are fine (PATCH semantics).
	if err := a.validateData(data, true); err != nil {
		return nil, err
	}

	// Ensure object exists (and permission hooks receive a concrete object path).
	instance, err := a.safeGetObjectByID(ctx, intID)
	if err != nil {
		return nil, err
	}

	// PATCH semantics: only update provided fields and avoid writing zero-values
	// for fields omitted from request payload.
	updates := orm.UpdateMap{}
	for key, value := range data {
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
	changesJSON, _ := json.Marshal(data)
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

func (a *Admin[T]) ExecuteAction(ctx context.Context, actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
	_ = params

	// Find action
	var selectedAction *Action[T]
	for _, action := range a.config.Actions {
		if action.Name == actionName {
			selectedAction = &action
			break
		}
	}

	if selectedAction == nil {
		return nil, fmt.Errorf("action %s not found", actionName)
	}

	user, _ := apicore.UserFromContext(ctx)

	// Fetch instances
	instances := make([]*T, 0, len(ids))
	actionErrors := make([]BulkActionError, 0)
	for _, id := range ids {
		intID, err := toInt64(id)
		if err != nil {
			actionErrors = append(actionErrors, BulkActionError{
				ID:      0,
				Code:    "invalid_id",
				Message: fmt.Sprintf("invalid id %v", id),
			})
			continue
		}

		instance, err := a.safeGetObjectByID(ctx, intID)
		if err != nil {
			actionErrors = append(actionErrors, BulkActionError{
				ID:      intID,
				Code:    "not_found",
				Message: "object not found",
			})
			continue
		}

		if !a.HasChangePermission(ctx, user, instance) {
			actionErrors = append(actionErrors, BulkActionError{
				ID:      intID,
				Code:    "permission_denied",
				Message: "permission denied",
			})
			continue
		}

		instances = append(instances, instance)
	}

	if len(instances) == 0 {
		response := &BulkActionResponse{
			Success:  false,
			Affected: 0,
			Message:  "No permitted objects found for action",
		}
		if len(actionErrors) > 0 {
			response.Errors = actionErrors
		}
		return response, nil
	}

	// Execute handler
	err := selectedAction.Handler(ctx, instances)
	if err != nil {
		return nil, err
	}

	response := &BulkActionResponse{
		Success:  len(actionErrors) == 0,
		Affected: len(instances),
		Message:  fmt.Sprintf("Successfully executed %s on %d objects", selectedAction.Label, len(instances)),
	}
	if len(actionErrors) > 0 {
		response.Errors = actionErrors
		response.Message = fmt.Sprintf("Executed %s on %d objects; %d skipped", selectedAction.Label, len(instances), len(actionErrors))
	}
	return response, nil
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

// History methods
func (a *Admin[T]) GetHistory(ctx context.Context, objectID string) ([]LogEntry, error) {
	if a.config.HistoryManager != nil {
		return a.config.HistoryManager.GetHistory(ctx, a.name, objectID)
	}
	return []LogEntry{}, nil
}

func (a *Admin[T]) LogAction(ctx context.Context, user interface{}, objectID string, repr string, action ActionType, changes string) error {
	if a.config.HistoryManager != nil {
		entry := LogEntry{
			Timestamp:   time.Now(),
			ModelName:   a.name,
			ObjectID:    objectID,
			ObjectRepr:  repr,
			Action:      action,
			ChangeStats: changes,
			UserID:      fmt.Sprintf("%v", a.resolveUserID(user)),
		}

		return a.config.HistoryManager.LogAction(ctx, entry)
	}
	return nil
}

// resolveUserID tries to extract a string identifier from the user object
func (a *Admin[T]) resolveUserID(user interface{}) interface{} {
	if user == nil {
		return "anonymous"
	}

	val := reflect.ValueOf(user)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 1. Try common ID fields
	if val.Kind() == reflect.Struct {
		for _, name := range []string{"ID", "Id", "id", "UserID", "Username", "Email"} {
			f := val.FieldByName(name)
			if f.IsValid() {
				return f.Interface()
			}
		}
	}

	// 2. Fallback to string representation
	return fmt.Sprintf("%v", user)
}

func isSuperuser(user interface{}) bool {
	if user == nil {
		return false
	}
	switch u := user.(type) {
	case map[string]interface{}:
		if role, ok := u["role"].(string); ok && (role == "superuser" || role == "admin") {
			return true
		}
		if isSuper, ok := u["is_superuser"].(bool); ok && isSuper {
			return true
		}
	case map[string]string:
		if u["role"] == "superuser" || u["role"] == "admin" {
			return true
		}
	}

	val := reflect.ValueOf(user)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false
		}
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		for _, fieldName := range []string{"IsSuperuser", "IsAdmin", "Role"} {
			f := val.FieldByName(fieldName)
			if f.IsValid() {
				if f.Kind() == reflect.Bool && f.Bool() {
					return true
				}
				if f.Kind() == reflect.String && (f.String() == "superuser" || f.String() == "admin") {
					return true
				}
			}
		}
	}
	return false
}

// Permission methods

func (a *Admin[T]) HasAddPermission(ctx context.Context, user interface{}) bool {
	if a.config.HasAddPermission != nil {
		return a.config.HasAddPermission(ctx, a, user)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermAdd))
	}
	if isSuperuser(user) {
		return true
	}
	return false // Default deny: configure Has*Permission or PermissionChecker to grant access
}

func (a *Admin[T]) HasChangePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	var typedObj *T
	if obj != nil {
		var ok bool
		typedObj, ok = obj.(*T)
		if !ok {
			return false
		}
	}

	if a.config.HasChangePermission != nil {
		return a.config.HasChangePermission(ctx, a, user, typedObj)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermChange))
	}
	if isSuperuser(user) {
		return true
	}
	return false // Default deny: configure Has*Permission or PermissionChecker to grant access
}

func (a *Admin[T]) HasDeletePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	var typedObj *T
	if obj != nil {
		var ok bool
		typedObj, ok = obj.(*T)
		if !ok {
			return false
		}
	}

	if a.config.HasDeletePermission != nil {
		return a.config.HasDeletePermission(ctx, a, user, typedObj)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermDelete))
	}
	if isSuperuser(user) {
		return true
	}
	return false // Default deny: configure Has*Permission or PermissionChecker to grant access
}

func (a *Admin[T]) HasViewPermission(ctx context.Context, user interface{}, obj interface{}) bool {
	var typedObj *T
	if obj != nil {
		var ok bool
		typedObj, ok = obj.(*T)
		if !ok {
			return false
		}
	}

	if a.config.HasViewPermission != nil {
		return a.config.HasViewPermission(ctx, a, user, typedObj)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermView))
	}
	if isSuperuser(user) {
		return true
	}
	return false // Default deny: configure Has*Permission or PermissionChecker to grant access
}

func (a *Admin[T]) HasModulePermission(ctx context.Context, user interface{}) bool {
	if a.config.HasModulePermission != nil {
		return a.config.HasModulePermission(ctx, a, user)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermView))
	}
	if isSuperuser(user) {
		return true
	}
	return a.HasViewPermission(ctx, user, nil) ||
		a.HasAddPermission(ctx, user) ||
		a.HasChangePermission(ctx, user, nil) ||
		a.HasDeletePermission(ctx, user, nil)
}

// Interface implementation for type-agnostic access

func (a *Admin[T]) ManagerInterface() interface{} {
	return a.manager
}

func (a *Admin[T]) ConfigInterface() interface{} {
	return a.config
}

// Helper methods

func (a *Admin[T]) getPermissionsMetadata(ctx context.Context, user interface{}) PermissionMetadata {
	return PermissionMetadata{
		Add:    a.HasAddPermission(ctx, user),
		Change: a.HasChangePermission(ctx, user, nil),
		Delete: a.HasDeletePermission(ctx, user, nil),
		View:   a.HasViewPermission(ctx, user, nil),
	}
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

// applyConfigDefaults applies default configuration from schema
func applyConfigDefaults[T any](config *Config[T], s schema.Schema) {
	meta := s.Meta()

	// Set verbose names if not provided
	if config.VerboseName == "" {
		config.VerboseName = meta.VerboseName
	}
	if config.VerboseNamePlural == "" {
		config.VerboseNamePlural = meta.VerboseNamePlural
	}

	// Set pagination defaults
	if config.ListPerPage == 0 {
		config.ListPerPage = 25
	}
	if config.ListMaxShowAll == 0 {
		config.ListMaxShowAll = 100
	}

	// Set default history manager if not provided
	if config.HistoryManager == nil {
		config.HistoryManager = NewMemoryHistoryManager()
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

// parseLookup splits key into field path and lookup (e.g. "price__gt" -> "price", "gt")
func parseLookup(key string) (string, string) {
	parts := strings.Split(key, "__")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		// Check if last part is a known lookup
		switch last {
		case "exact", "ne", "gt", "gte", "lt", "lte", "contains", "icontains", "startswith", "endswith", "in", "isnull":
			return strings.Join(parts[:len(parts)-1], "__"), last
		}
	}
	return key, "exact"
}
