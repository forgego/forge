package api

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	forgehttp "github.com/forgego/forge/server"
)

// EnhancedBaseViewSet provides enhanced viewset functionality with authentication and permissions
// This is the new recommended implementation
type EnhancedBaseViewSet struct {
	Serializer func() Serializer
	Queryset   interface{}
	Model      interface{}

	// Configuration
	AuthenticationClasses []authentication.Authentication
	PermissionClasses     []permissions.Permission
	ThrottleClasses       []throttling.Throttle
	Action                string // Current action name
}

// NewEnhancedBaseViewSet creates a new enhanced base viewset
func NewEnhancedBaseViewSet(serializer func() Serializer, queryset, model interface{}) *EnhancedBaseViewSet {
	return &EnhancedBaseViewSet{
		Serializer: serializer,
		Queryset:   queryset,
		Model:      model,
	}
}

// GetAction returns the current action name (for permissions compatibility)
func (vs *EnhancedBaseViewSet) GetAction() string {
	return vs.Action
}

// SetAction sets the current action name
func (vs *EnhancedBaseViewSet) SetAction(action string) {
	vs.Action = action
}

// GetAuthentication returns the authentication classes for this viewset
func (vs *EnhancedBaseViewSet) GetAuthentication() []authentication.Authentication {
	if len(vs.AuthenticationClasses) > 0 {
		return vs.AuthenticationClasses
	}
	return []authentication.Authentication{}
}

// GetPermissions returns the permission classes for this viewset
func (vs *EnhancedBaseViewSet) GetPermissions() []permissions.Permission {
	if len(vs.PermissionClasses) > 0 {
		return vs.PermissionClasses
	}
	return []permissions.Permission{}
}

// GetThrottles returns the throttle classes for this viewset
func (vs *EnhancedBaseViewSet) GetThrottles() []throttling.Throttle {
	if len(vs.ThrottleClasses) > 0 {
		return vs.ThrottleClasses
	}
	return []throttling.Throttle{}
}

// authenticateRequest authenticates the request
func (vs *EnhancedBaseViewSet) authenticateRequest(r *http.Request) error {
	authClasses := vs.GetAuthentication()
	if len(authClasses) == 0 {
		return nil
	}

	result, err := authentication.AuthenticateRequest(r, authClasses)
	if err != nil {
		return exceptions.NewAuthenticationFailed(err.Error())
	}

	if result != nil {
		authentication.SetUserOnRequest(r, result.User)
		authentication.SetAuthOnRequest(r, result.Auth)
	}

	return nil
}

// checkPermissions checks view-level permissions
func (vs *EnhancedBaseViewSet) checkPermissions(r *http.Request) error {
	permClasses := vs.GetPermissions()
	if len(permClasses) == 0 {
		return nil
	}

	if !permissions.CheckPermissions(r, vs, permClasses) {
		for _, perm := range permClasses {
			if !perm.HasPermission(r, vs) {
				return exceptions.NewPermissionDenied(perm.GetMessage())
			}
		}
		return exceptions.NewPermissionDenied("Permission denied")
	}

	return nil
}

// checkObjectPermissions checks object-level permissions
func (vs *EnhancedBaseViewSet) checkObjectPermissions(r *http.Request, obj interface{}) error {
	permClasses := vs.GetPermissions()
	if len(permClasses) == 0 {
		return nil
	}

	if !permissions.CheckObjectPermissions(r, vs, obj, permClasses) {
		for _, perm := range permClasses {
			if !perm.HasObjectPermission(r, vs, obj) {
				return exceptions.NewPermissionDenied(perm.GetMessage())
			}
		}
		return exceptions.NewPermissionDenied("Permission denied")
	}

	return nil
}

// checkThrottles checks throttling
func (vs *EnhancedBaseViewSet) checkThrottles(r *http.Request) error {
	throttleClasses := vs.GetThrottles()
	if len(throttleClasses) == 0 {
		return nil
	}

	err := throttling.CheckThrottles(r, vs, throttleClasses)
	if err != nil {
		if throttledErr, ok := err.(*throttling.ThrottledError); ok {
			return exceptions.NewThrottled("Request was throttled", throttledErr.WaitDuration)
		}
		return err
	}

	return nil
}

// handleException handles exceptions and writes HTTP response
func (vs *EnhancedBaseViewSet) handleException(w http.ResponseWriter, r *http.Request, err error) {
	exceptions.HandleExceptionHTTP(w, r, err, nil)
}

// List handles GET /resource/
// Returns a paginated list of resources.
// Supports filtering, ordering, and search via query parameters.
// Validates authentication and permissions before processing.
func (vs *EnhancedBaseViewSet) List(w http.ResponseWriter, r *http.Request) {
	vs.SetAction("list")
	ctx := r.Context()

	// Authenticate
	if err := vs.authenticateRequest(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check permissions
	if err := vs.checkPermissions(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check throttles
	if err := vs.checkThrottles(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Get pagination parameters
	page, pageSize, _ := ParsePaginationParams(r, 20)

	// Get queryset using reflection
	querysetValue := reflect.ValueOf(vs.Queryset)
	if !querysetValue.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Queryset not set",
			nil,
		))
		return
	}

	// Apply filtering from query params
	qs := applyFilters(querysetValue, r)

	// Apply ordering
	qs = applyOrdering(qs, r)

	// Get total count
	countMethod := qs.MethodByName("Count")
	var totalCount int64
	if countMethod.IsValid() {
		results := countMethod.Call([]reflect.Value{reflect.ValueOf(ctx)})
		if len(results) > 0 && results[0].CanInterface() {
			if count, ok := results[0].Interface().(int64); ok {
				totalCount = count
			}
		}
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	offsetMethod := qs.MethodByName("Offset")
	if offsetMethod.IsValid() {
		results := offsetMethod.Call([]reflect.Value{reflect.ValueOf(offset)})
		if len(results) > 0 {
			if newQS, ok := results[0].Interface().(interface{}); ok {
				qs = reflect.ValueOf(newQS)
			}
		}
	}

	limitMethod := qs.MethodByName("Limit")
	if limitMethod.IsValid() {
		results := limitMethod.Call([]reflect.Value{reflect.ValueOf(pageSize)})
		if len(results) > 0 {
			if newQS, ok := results[0].Interface().(interface{}); ok {
				qs = reflect.ValueOf(newQS)
			}
		}
	}

	// Execute query
	allMethod := qs.MethodByName("All")
	var results []interface{}
	if allMethod.IsValid() {
		allResults := allMethod.Call([]reflect.Value{reflect.ValueOf(ctx)})
		if len(allResults) >= 2 {
			if errVal := allResults[1]; !errVal.IsNil() {
				if err, ok := errVal.Interface().(error); ok {
					vs.handleException(w, r, exceptions.NewAPIException(
						http.StatusInternalServerError,
						"internal_error",
						err.Error(),
						nil,
					))
					return
				}
			}
			if allResults[0].CanInterface() {
				if objects, ok := allResults[0].Interface().([]interface{}); ok {
					results = objects
				} else if allResults[0].Kind() == reflect.Slice {
					sliceValue := allResults[0]
					for i := 0; i < sliceValue.Len(); i++ {
						results = append(results, sliceValue.Index(i).Interface())
					}
				}
			}
		}
	}

	// Serialize results
	serialized := SerializeMany(results)

	// Send paginated response
	if err := SendPaginatedResponse(w, r, serialized, int(totalCount), page, pageSize); err != nil {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Failed to send response",
			nil,
		))
	}
}

// Create handles POST /resource/
// Creates a new resource instance from request body.
// Validates authentication, permissions, and throttling before processing.
// Validates serializer data and returns the created resource on success.
func (vs *EnhancedBaseViewSet) Create(w http.ResponseWriter, r *http.Request) {
	vs.SetAction("create")
	ctx := r.Context()

	// Authenticate
	if err := vs.authenticateRequest(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check permissions
	if err := vs.checkPermissions(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check throttles
	if err := vs.checkThrottles(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"parse_error",
			"Invalid JSON",
			nil,
		))
		return
	}

	serializer := vs.Serializer()
	baseSerializer, ok := serializer.(*BaseSerializer)
	if !ok {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Invalid serializer type",
			nil,
		))
		return
	}
	baseSerializer.SetData(data)

	if err := serializer.Validate(); err != nil {
		errors := serializer.Errors()
		vs.handleException(w, r, exceptions.NewValidationError(errors))
		return
	}

	// Create model instance from serializer data
	modelValue := reflect.New(reflect.TypeOf(vs.Model).Elem())
	instance := modelValue.Interface()

	// Populate instance from data
	populateFromMap(instance, data)

	// Get manager and call Create
	manager := getManagerFromModel(instance)
	if !manager.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Manager not found",
			nil,
		))
		return
	}

	createMethod := manager.MethodByName("Create")
	if !createMethod.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Create method not found",
			nil,
		))
		return
	}

	results := createMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(instance),
	})

	if len(results) > 0 && !results[0].IsNil() {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			vs.handleException(w, r, exceptions.NewValidationError(map[string][]string{
				"non_field_errors": {err.Error()},
			}))
			return
		}
	}

	// Serialize and return created instance
	serialized := SerializeModel(instance)
	response := core.NewResponse(w)
	response.Status(http.StatusCreated).JSON(serialized)
}

// Retrieve handles GET /resource/{id}/
// Returns a single resource instance by ID.
// Validates authentication and permissions before processing.
// Returns 404 if resource not found.
func (vs *EnhancedBaseViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
	vs.SetAction("retrieve")
	ctx := r.Context()

	// Authenticate
	if err := vs.authenticateRequest(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check permissions
	if err := vs.checkPermissions(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check throttles
	if err := vs.checkThrottles(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Get ID from URL
	idStr := forgehttp.GetParam(r, "id")
	if idStr == "" {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"invalid_request",
			"ID is required",
			nil,
		))
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"invalid_request",
			"Invalid ID",
			nil,
		))
		return
	}

	// Get manager and call Get
	modelValue := reflect.New(reflect.TypeOf(vs.Model).Elem())
	instance := modelValue.Interface()

	manager := getManagerFromModel(instance)
	if !manager.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Manager not found",
			nil,
		))
		return
	}

	getMethod := manager.MethodByName("Get")
	if !getMethod.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Get method not found",
			nil,
		))
		return
	}

	results := getMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(id),
	})

	if len(results) < 2 {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Invalid Get method",
			nil,
		))
		return
	}

	if !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok && err != nil {
			vs.handleException(w, r, exceptions.NewNotFound("Resource not found"))
			return
		}
	}

	instance = results[0].Interface()

	// Check object permissions
	if err := vs.checkObjectPermissions(r, instance); err != nil {
		vs.handleException(w, r, err)
		return
	}

	serialized := SerializeModel(instance)
	response := core.NewResponse(w)
	response.JSON(serialized)
}

// Update handles PUT /resource/{id}/
// Performs full update of a resource instance.
// Validates authentication, permissions, and throttling before processing.
// Returns the updated resource on success, or appropriate error response.
func (vs *EnhancedBaseViewSet) Update(w http.ResponseWriter, r *http.Request) {
	vs.SetAction("update")
	ctx := r.Context()

	// Authenticate
	if err := vs.authenticateRequest(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check permissions
	if err := vs.checkPermissions(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check throttles
	if err := vs.checkThrottles(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	idStr := forgehttp.GetParam(r, "id")
	if idStr == "" {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"invalid_request",
			"ID is required",
			nil,
		))
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"invalid_request",
			"Invalid ID",
			nil,
		))
		return
	}

	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"parse_error",
			"Invalid JSON",
			nil,
		))
		return
	}

	serializer := vs.Serializer()
	baseSerializer, ok := serializer.(*BaseSerializer)
	if !ok {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Invalid serializer type",
			nil,
		))
		return
	}
	baseSerializer.SetData(data)

	if err := serializer.Validate(); err != nil {
		errors := serializer.Errors()
		vs.handleException(w, r, exceptions.NewValidationError(errors))
		return
	}

	// Get existing instance
	modelValue := reflect.New(reflect.TypeOf(vs.Model).Elem())
	instance := modelValue.Interface()

	manager := getManagerFromModel(instance)
	if !manager.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Manager not found",
			nil,
		))
		return
	}

	getMethod := manager.MethodByName("Get")
	if !getMethod.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Get method not found",
			nil,
		))
		return
	}

	getResults := getMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(id),
	})

	if len(getResults) < 2 || !getResults[1].IsNil() {
		if err, ok := getResults[1].Interface().(error); ok && err != nil {
			vs.handleException(w, r, exceptions.NewNotFound("Resource not found"))
			return
		}
	}

	instance = getResults[0].Interface()

	// Check object permissions
	if err := vs.checkObjectPermissions(r, instance); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Populate from data
	populateFromMap(instance, data)

	// Update
	updateMethod := manager.MethodByName("Update")
	if !updateMethod.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Update method not found",
			nil,
		))
		return
	}

	updateResults := updateMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(instance),
	})

	if len(updateResults) > 0 && !updateResults[0].IsNil() {
		if err, ok := updateResults[0].Interface().(error); ok && err != nil {
			vs.handleException(w, r, exceptions.NewValidationError(map[string][]string{
				"non_field_errors": {err.Error()},
			}))
			return
		}
	}

	serialized := SerializeModel(instance)
	response := core.NewResponse(w)
	response.JSON(serialized)
}

// PartialUpdate handles PATCH /resource/{id}/
// PartialUpdate handles PATCH /resource/{id}/
// Performs partial update of a resource instance (only provided fields).
// Currently delegates to Update - will be enhanced to support partial updates.
// Validates authentication, permissions, and throttling before processing.
func (vs *EnhancedBaseViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	// For now, same as Update - will be enhanced later
	vs.Update(w, r)
}

// Destroy handles DELETE /resource/{id}/
// Deletes a resource instance by ID.
// Validates authentication, permissions, and throttling before processing.
// Returns 204 No Content on success, or appropriate error response.
func (vs *EnhancedBaseViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
	vs.SetAction("destroy")
	ctx := r.Context()

	// Authenticate
	if err := vs.authenticateRequest(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check permissions
	if err := vs.checkPermissions(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Check throttles
	if err := vs.checkThrottles(r); err != nil {
		vs.handleException(w, r, err)
		return
	}

	idStr := forgehttp.GetParam(r, "id")
	if idStr == "" {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"invalid_request",
			"ID is required",
			nil,
		))
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusBadRequest,
			"invalid_request",
			"Invalid ID",
			nil,
		))
		return
	}

	// Get instance first
	modelValue := reflect.New(reflect.TypeOf(vs.Model).Elem())
	instance := modelValue.Interface()

	manager := getManagerFromModel(instance)
	if !manager.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Manager not found",
			nil,
		))
		return
	}

	getMethod := manager.MethodByName("Get")
	if !getMethod.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Get method not found",
			nil,
		))
		return
	}

	getResults := getMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(id),
	})

	if len(getResults) < 2 || !getResults[1].IsNil() {
		if err, ok := getResults[1].Interface().(error); ok && err != nil {
			vs.handleException(w, r, exceptions.NewNotFound("Resource not found"))
			return
		}
	}

	instance = getResults[0].Interface()

	// Check object permissions
	if err := vs.checkObjectPermissions(r, instance); err != nil {
		vs.handleException(w, r, err)
		return
	}

	// Delete
	deleteMethod := manager.MethodByName("Delete")
	if !deleteMethod.IsValid() {
		vs.handleException(w, r, exceptions.NewAPIException(
			http.StatusInternalServerError,
			"internal_error",
			"Delete method not found",
			nil,
		))
		return
	}

	deleteResults := deleteMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(instance),
	})

	if len(deleteResults) > 0 && !deleteResults[0].IsNil() {
		if err, ok := deleteResults[0].Interface().(error); ok && err != nil {
			vs.handleException(w, r, exceptions.NewValidationError(map[string][]string{
				"non_field_errors": {err.Error()},
			}))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Ensure EnhancedBaseViewSet implements ViewSet interface
var _ ViewSet = (*EnhancedBaseViewSet)(nil)
