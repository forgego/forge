package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/docs"
	apierrors "github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	forgeerrors "github.com/forgego/forge/errors"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/forgego/forge/validate"
	"go.uber.org/zap"
)

// ViewSet is the base interface for all viewsets
type ViewSet interface {
	// List handles GET /resource/
	List(w http.ResponseWriter, r *http.Request)
	// Create handles POST /resource/
	Create(w http.ResponseWriter, r *http.Request)
	// Retrieve handles GET /resource/{id}/
	Retrieve(w http.ResponseWriter, r *http.Request)
	// Update handles PUT /resource/{id}/
	Update(w http.ResponseWriter, r *http.Request)
	// PartialUpdate handles PATCH /resource/{id}/
	PartialUpdate(w http.ResponseWriter, r *http.Request)
	// Destroy handles DELETE /resource/{id}/
	Destroy(w http.ResponseWriter, r *http.Request)
}

// QuerySetInterface defines the common methods needed from QuerySet for viewset operations.
// This interface allows type-safe operations without reflection when the QuerySet implements it.
type QuerySetInterface interface {
	Count(ctx context.Context) (int64, error)
	All(ctx context.Context) (interface{}, error)
	Filter(expr interface{}) interface{}
	OrderBy(fields ...interface{}) interface{}
	Limit(limit int) interface{}
	Offset(offset int) interface{}
}

// ManagerInterface defines the common methods needed from Manager for viewset operations.
type ManagerInterface interface {
	Get(ctx context.Context, id int64) (interface{}, error)
	Create(ctx context.Context, model interface{}) error
	Update(ctx context.Context, model interface{}) error
	Delete(ctx context.Context, model interface{}) error
}

// BaseViewSet provides common viewset functionality
type BaseViewSet struct {
	Serializer func() Serializer
	Queryset   interface{} // This would be a QuerySet in real implementation
	Model      interface{}
	// ExcludeResponseFields holds response keys removed from serialized output; nil keeps every field.
	ExcludeResponseFields []string
	// ReadOnlyRequestFields holds request keys ignored on create and update; nil accepts every field.
	ReadOnlyRequestFields []string
	// Authentication uses the current defaults when nil; a non-nil empty slice disables authentication.
	Authentication []authentication.Authentication
	// Permissions uses the current defaults when nil; a non-nil empty slice disables permission checks.
	Permissions []permissions.Permission
	// Throttles uses the current defaults when nil; a non-nil empty slice disables throttling.
	Throttles   []throttling.Throttle
	ErrorWriter func(http.ResponseWriter, *http.Request, error)

	actionMu sync.RWMutex
	action   string
}

// ActionFromContext returns the action name stored in ctx, or an empty string if none is set.
func ActionFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if action, ok := core.ActionFromContext(ctx); ok {
		return action
	}
	return ""
}

// GetActionFromRequest returns the action name stored in r's context, or an empty string if none is set.
func GetActionFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return ActionFromContext(r.Context())
}

func withAction(r *http.Request, action string) *http.Request {
	if r == nil {
		return nil
	}
	return r.WithContext(core.WithAction(r.Context(), action))
}

// NewBaseViewSet creates a new base viewset
func NewBaseViewSet(serializer func() Serializer, queryset, model interface{}) *BaseViewSet {
	return &BaseViewSet{
		Serializer: serializer,
		Queryset:   queryset,
		Model:      model,
	}
}

// getManager gets the manager for operations
func (vs *BaseViewSet) getManager() reflect.Value {
	// If Queryset is set and looks like a manager (has Create method), use it
	if vs.Queryset != nil {
		qsValue := reflect.ValueOf(vs.Queryset)
		qsType := qsValue.Type()

		// Use cached method lookup instead of MethodByName
		if _, ok := globalCache.GetMethod(qsType, "Create"); ok {
			return qsValue
		}
	}

	return reflect.Value{}
}

// GetAction returns the dispatched action. Inside permission and throttle checks
// the view is request-scoped, so it returns that request's action; on the shared
// viewset it returns the most recently dispatched action.
func (vs *BaseViewSet) GetAction() string {
	vs.actionMu.RLock()
	defer vs.actionMu.RUnlock()
	return vs.action
}

// SetAction sets the fallback action name on BaseViewSet for backward compatibility.
// During normal HTTP request processing, actions are carried in the request context.
func (vs *BaseViewSet) SetAction(action string) {
	vs.actionMu.Lock()
	defer vs.actionMu.Unlock()
	vs.action = action
}

func (vs *BaseViewSet) authenticateRequest(r *http.Request) error {
	authClasses := vs.Authentication
	if authClasses == nil {
		authClasses = GetDefaultAuthentication()
	}
	result, err := authentication.AuthenticateRequest(r, authClasses)
	if err != nil {
		return exceptions.NewAuthenticationFailed(err.Error())
	}
	if result == nil {
		return nil
	}
	authentication.SetUserOnRequest(r, result.User)
	authentication.SetAuthOnRequest(r, result.Auth)
	return nil
}

// viewForRequest returns a request-scoped view carrying the request's action.
// When the request carries no action it returns vs unchanged.
func (vs *BaseViewSet) viewForRequest(r *http.Request) *BaseViewSet {
	action := GetActionFromRequest(r)
	if action == "" {
		return vs
	}
	return &BaseViewSet{
		Serializer:            vs.Serializer,
		Queryset:              vs.Queryset,
		Model:                 vs.Model,
		ExcludeResponseFields: vs.ExcludeResponseFields,
		ReadOnlyRequestFields: vs.ReadOnlyRequestFields,
		Authentication:        vs.Authentication,
		Permissions:           vs.Permissions,
		Throttles:             vs.Throttles,
		ErrorWriter:           vs.ErrorWriter,
		action:                action,
	}
}

func (vs *BaseViewSet) checkPermissions(r *http.Request) error {
	perms := vs.Permissions
	if perms == nil {
		perms = GetDefaultPermissions()
	}
	reqView := vs.viewForRequest(r)
	if permissions.CheckPermissions(r, reqView, perms) {
		return nil
	}
	for _, permission := range perms {
		if !permission.HasPermission(r, reqView) {
			return exceptions.NewPermissionDenied(permission.GetMessage())
		}
	}
	return exceptions.NewPermissionDenied("Permission denied")
}

func (vs *BaseViewSet) checkObjectPermissions(r *http.Request, object interface{}) error {
	perms := vs.Permissions
	if perms == nil {
		perms = GetDefaultPermissions()
	}
	reqView := vs.viewForRequest(r)
	if permissions.CheckObjectPermissions(r, reqView, object, perms) {
		return nil
	}
	for _, permission := range perms {
		if !permission.HasObjectPermission(r, reqView, object) {
			return exceptions.NewPermissionDenied(permission.GetMessage())
		}
	}
	return exceptions.NewPermissionDenied("Permission denied")
}

func (vs *BaseViewSet) checkThrottles(r *http.Request) error {
	throttles := vs.Throttles
	if throttles == nil {
		throttles = GetDefaultThrottles()
	}
	err := throttling.CheckThrottles(r, vs.viewForRequest(r), throttles)
	if err == nil {
		return nil
	}
	throttled, ok := err.(*throttling.ThrottledError)
	if !ok {
		return err
	}
	return exceptions.NewThrottled("Request was throttled", throttled.WaitDuration)
}

func (vs *BaseViewSet) checkRequest(w http.ResponseWriter, r *http.Request, action string) bool {
	vs.SetAction(action)
	if GetActionFromRequest(r) == "" {
		r = withAction(r, action)
	}
	checks := []func(*http.Request) error{
		vs.authenticateRequest,
		vs.checkPermissions,
		vs.checkThrottles,
	}
	for _, check := range checks {
		if err := check(r); err != nil {
			vs.handleException(w, r, err)
			return false
		}
	}
	return true
}

func (vs *BaseViewSet) handleException(w http.ResponseWriter, r *http.Request, err error) {
	if vs.ErrorWriter != nil {
		vs.ErrorWriter(w, r, err)
		return
	}
	apierrors.WriteError(w, r, err)
}

func (vs *BaseViewSet) allowObject(w http.ResponseWriter, r *http.Request, object interface{}) bool {
	if err := vs.checkObjectPermissions(r, object); err != nil {
		vs.handleException(w, r, err)
		return false
	}
	return true
}

// List handles GET /resource/
func (vs *BaseViewSet) List(w http.ResponseWriter, r *http.Request) {
	r = withAction(r, "list")
	if !vs.checkRequest(w, r, "list") {
		return
	}
	ctx := r.Context()

	// Get pagination parameters, defaulting to the configured page size.
	defaultPageSize := 20
	if s := GetSettings(); s != nil && s.PageSize > 0 {
		defaultPageSize = s.PageSize
	}
	page, pageSize, _ := ParsePaginationParams(r, defaultPageSize)

	// Get queryset using reflection. A bare manager that cannot paginate is
	// converted to a queryset via its QuerySet constructor when available, so
	// Offset/Limit are applied through the queryset instead of loading all rows.
	querysetValue := reflect.ValueOf(vs.Queryset)
	if !querysetValue.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Queryset not set")
		return
	}
	querysetValue = paginatableQueryset(querysetValue)

	// Apply filtering from query params
	qs := applyFilters(querysetValue, r, vs.Model)

	// Apply explicit ordering, model Meta ordering, or a stable primary-key fallback.
	qs = applyOrdering(qs, r, vs.Model, defaultOrdering(vs.Model)...)

	// Get total count using cached method lookup
	qsType := qs.Type()
	countMethod, ok := globalCache.GetMethod(qsType, "Count")
	if !ok {
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Count method not found")
		return
	}

	var totalCount int64
	results := countMethod.Func.Call([]reflect.Value{qs, reflect.ValueOf(ctx)})
	if len(results) >= 2 {
		if !results[1].IsNil() {
			if err, ok := results[1].Interface().(error); ok {
				vs.handleException(w, r, err)
				return
			}
		}
		if results[0].CanInterface() {
			if count, ok := results[0].Interface().(int64); ok {
				totalCount = count
			}
		}
	}

	// Apply pagination using cached method lookup
	offset := (page - 1) * pageSize

	offsetMethod, ok := globalCache.GetMethod(qsType, "Offset")
	if ok {
		offsetResults := offsetMethod.Func.Call([]reflect.Value{qs, reflect.ValueOf(offset)})
		if len(offsetResults) > 0 {
			if newQS := offsetResults[0].Interface(); newQS != nil {
				qs = reflect.ValueOf(newQS)
				qsType = qs.Type()
			}
		}
	}

	limitMethod, ok := globalCache.GetMethod(qsType, "Limit")
	if ok {
		limitResults := limitMethod.Func.Call([]reflect.Value{qs, reflect.ValueOf(pageSize)})
		if len(limitResults) > 0 {
			if newQS := limitResults[0].Interface(); newQS != nil {
				qs = reflect.ValueOf(newQS)
				qsType = qs.Type()
			}
		}
	}

	// Execute query using cached method lookup
	allMethod, ok := globalCache.GetMethod(qsType, "All")
	if !ok {
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "All method not found")
		return
	}

	var resultList []interface{}
	allResults := allMethod.Func.Call([]reflect.Value{qs, reflect.ValueOf(ctx)})
	if len(allResults) >= 2 {
		// Check for error first
		if errVal := allResults[1]; !errVal.IsNil() {
			if err, ok := errVal.Interface().(error); ok {
				vs.handleException(w, r, err)
				return
			}
		}
		if allResults[0].CanInterface() {
			if objects, ok := allResults[0].Interface().([]interface{}); ok {
				resultList = objects
			} else if allResults[0].Kind() == reflect.Slice {
				// Try to convert slice of pointers
				sliceValue := allResults[0]
				for i := 0; i < sliceValue.Len(); i++ {
					resultList = append(resultList, sliceValue.Index(i).Interface())
				}
			}
		}
	}

	// Serialize results
	serialized := vs.stripExcludedFieldsMany(SerializeMany(resultList))
	serialized = filterOutputMany(vs.Serializer(), serialized)

	// Send paginated response
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = SendPaginatedResponse(w, r, serialized, int(totalCount), page, pageSize)
}

// Create handles POST /resource/
func (vs *BaseViewSet) Create(w http.ResponseWriter, r *http.Request) {
	r = withAction(r, "create")
	if !vs.checkRequest(w, r, "create") {
		return
	}
	ctx := r.Context()

	var data map[string]interface{}
	if err := forgehttp.GetJSONUseNumber(r, &data); err != nil || data == nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	applySchemaDefaults(vs.Model, data)

	serializer := vs.Serializer()
	stripReadOnlyInput(serializer, data)
	serializer.SetData(data)

	if err := serializer.Validate(); err != nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errors": serializer.Errors(),
		})
		return
	}

	// Create model instance from serializer data
	modelValue := reflect.New(reflect.TypeOf(vs.Model).Elem())
	instance := modelValue.Interface()

	// Drop read-only request fields (serializer read-only fields and
	// viewset-level ReadOnlyRequestFields) before populating the model.
	var ignoredKeys []string
	ignoredKeys = append(ignoredKeys, vs.ReadOnlyRequestFields...)
	if ro, ok := serializer.(interface{ ReadOnlyFields() []string }); ok {
		ignoredKeys = append(ignoredKeys, ro.ReadOnlyFields()...)
	}

	// Populate instance from data
	if err := populateFromMap(instance, data, ignoredKeys...); err != nil {
		vs.handleException(w, r, validationErrorForPopulate(err))
		return
	}

	// Validate the populated model before the manager runs business hooks
	// (BeforeCreate/BeforeSave), so an invalid request fails with 400 without
	// triggering hook side effects.
	if err := validateModelInstance(instance); err != nil {
		vs.handleException(w, r, modelValidationException(err))
		return
	}

	// Get manager and call Create
	manager := vs.getManager()
	if !manager.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Manager not found")
		return
	}

	managerType := manager.Type()
	createMethod, ok := globalCache.GetMethod(managerType, "Create")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Create method not found")
		return
	}

	results := createMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(orm.WithModelValidationCompleted(ctx)),
		reflect.ValueOf(instance),
	})

	if len(results) > 0 && !results[0].IsNil() {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			vs.handleException(w, r, persistenceException(err))
			return
		}
	}

	pkGoName, pkDBName, pkJSONName := vs.getSchemaPrimaryKeyField()
	if id, ok := primaryKeyInt64(instance, pkGoName, pkDBName, pkJSONName); ok && id != 0 {
		refetched, err := getManagerInstance(manager, managerType, ctx, id)
		if err != nil {
			logRefetchFailure(r, "create", id, err)
		} else {
			instance = refetched
		}
	}

	// Serialize and return created instance
	serialized := vs.stripExcludedFields(SerializeModel(instance))
	serialized = filterOutputMap(vs.Serializer(), serialized)
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = forgehttp.SendJSON(w, http.StatusCreated, serialized)
}

// Retrieve handles GET /resource/{id}/
func (vs *BaseViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
	r = withAction(r, "retrieve")
	if !vs.checkRequest(w, r, "retrieve") {
		return
	}
	ctx := r.Context()

	// Get ID from URL
	idStr := forgehttp.GetParam(r, "id")
	if idStr == "" {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Get manager and call Get
	manager := vs.getManager()
	if !manager.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Manager not found")
		return
	}

	managerType := manager.Type()
	getMethod, ok := globalCache.GetMethod(managerType, "Get")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Get method not found")
		return
	}

	results := getMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(ctx),
		reflect.ValueOf(id),
	})

	if len(results) < 2 {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Invalid Get method")
		return
	}

	if !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok && err != nil {
			vs.handleException(w, r, lookupException(err))
			return
		}
	}

	instance := results[0].Interface()
	if !vs.allowObject(w, r, instance) {
		return
	}
	serialized := vs.stripExcludedFields(SerializeModel(instance))
	serialized = filterOutputMap(vs.Serializer(), serialized)
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = forgehttp.SendJSON(w, http.StatusOK, serialized)
}

// Update handles PUT /resource/{id}/
func (vs *BaseViewSet) Update(w http.ResponseWriter, r *http.Request) {
	vs.update(w, r, "update")
}

func (vs *BaseViewSet) update(w http.ResponseWriter, r *http.Request, action string) {
	r = withAction(r, action)
	if !vs.checkRequest(w, r, action) {
		return
	}
	ctx := r.Context()

	idStr := forgehttp.GetParam(r, "id")
	if idStr == "" {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var data map[string]interface{}
	if err := forgehttp.GetJSONUseNumber(r, &data); err != nil || data == nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	serializer := vs.Serializer()
	stripReadOnlyInput(serializer, data)
	serializer.SetData(data)

	if err := serializer.Validate(); err != nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errors": serializer.Errors(),
		})
		return
	}

	// Get existing instance
	manager := vs.getManager()
	if !manager.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Manager not found")
		return
	}

	managerType := manager.Type()
	getMethod, ok := globalCache.GetMethod(managerType, "Get")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Get method not found")
		return
	}

	getResults := getMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(ctx),
		reflect.ValueOf(id),
	})

	if len(getResults) < 2 || !getResults[1].IsNil() {
		if err, ok := getResults[1].Interface().(error); ok && err != nil {
			vs.handleException(w, r, lookupException(err))
			return
		}
	}

	instance := getResults[0].Interface()
	if !vs.allowObject(w, r, instance) {
		return
	}

	// Get schema primary key field info
	pkGoName, pkDBName, pkJSONName := vs.getSchemaPrimaryKeyField()

	// Capture primary key from existing instance before populating
	origPK := getPrimaryKeyValue(instance, pkGoName, pkDBName, pkJSONName)

	// Collect primary-key fields and read-only fields to ignore from body
	ignoredKeys := []string{"id", "ID", "Id"}
	if pkJSONName != "" {
		ignoredKeys = append(ignoredKeys, pkJSONName)
	}
	if pkGoName != "" {
		ignoredKeys = append(ignoredKeys, pkGoName)
	}
	if pkDBName != "" && pkDBName != pkGoName && pkDBName != pkJSONName {
		ignoredKeys = append(ignoredKeys, pkDBName)
	}
	if ro, ok := serializer.(interface{ ReadOnlyFields() []string }); ok {
		ignoredKeys = append(ignoredKeys, ro.ReadOnlyFields()...)
	}
	ignoredKeys = append(ignoredKeys, vs.ReadOnlyRequestFields...)

	// Populate from data, ignoring primary-key fields
	if err := populateFromMap(instance, data, ignoredKeys...); err != nil {
		vs.handleException(w, r, validationErrorForPopulate(err))
		return
	}

	// Restore primary key from the object loaded via the URL after populating
	restorePrimaryKey(instance, origPK, id, pkGoName, pkDBName, pkJSONName)

	// Validate the populated model before the manager runs business hooks
	// (BeforeUpdate/BeforeSave), so an invalid request fails with 400 without
	// triggering hook side effects.
	if err := validateModelInstance(instance); err != nil {
		vs.handleException(w, r, modelValidationException(err))
		return
	}

	// Update
	updateMethod, ok := globalCache.GetMethod(managerType, "Update")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Update method not found")
		return
	}

	updateResults := updateMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(orm.WithModelValidationCompleted(ctx)),
		reflect.ValueOf(instance),
	})

	if len(updateResults) > 0 && !updateResults[0].IsNil() {
		if err, ok := updateResults[0].Interface().(error); ok && err != nil {
			vs.handleException(w, r, persistenceException(err))
			return
		}
	}

	refetched, err := getManagerInstance(manager, managerType, ctx, id)
	if err != nil {
		logRefetchFailure(r, action, id, err)
	} else {
		instance = refetched
	}

	serialized := vs.stripExcludedFields(SerializeModel(instance))
	serialized = filterOutputMap(vs.Serializer(), serialized)
	forgehttp.SendJSON(w, http.StatusOK, serialized)
}

func logRefetchFailure(r *http.Request, action string, id int64, err error) {
	logger, ok := forgehttp.GetLogger(r).(interface {
		Warn(string, ...zap.Field)
	})
	if !ok {
		logger = zap.L()
	}
	logger.Warn("failed to refetch model after committed write",
		zap.String("action", action),
		zap.Int64("id", id),
		zap.Error(err),
	)
}

// PartialUpdate handles PATCH /resource/{id}/
func (vs *BaseViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	vs.update(w, r, "partial_update")
}

// Destroy handles DELETE /resource/{id}/
func (vs *BaseViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
	r = withAction(r, "destroy")
	if !vs.checkRequest(w, r, "destroy") {
		return
	}
	ctx := r.Context()

	idStr := forgehttp.GetParam(r, "id")
	if idStr == "" {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "ID is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Get instance first
	manager := vs.getManager()
	if !manager.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Manager not found")
		return
	}

	managerType := manager.Type()
	getMethod, ok := globalCache.GetMethod(managerType, "Get")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Get method not found")
		return
	}

	getResults := getMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(ctx),
		reflect.ValueOf(id),
	})

	if len(getResults) < 2 || !getResults[1].IsNil() {
		if err, ok := getResults[1].Interface().(error); ok && err != nil {
			vs.handleException(w, r, lookupException(err))
			return
		}
	}

	instance := getResults[0].Interface()
	if !vs.allowObject(w, r, instance) {
		return
	}

	// Delete
	deleteMethod, ok := globalCache.GetMethod(managerType, "Delete")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Delete method not found")
		return
	}

	deleteResults := deleteMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(ctx),
		reflect.ValueOf(instance),
	})

	if len(deleteResults) > 0 && !deleteResults[0].IsNil() {
		if err, ok := deleteResults[0].Interface().(error); ok && err != nil {
			vs.handleException(w, r, persistenceException(err))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func lookupException(err error) error {
	if forgeerrors.IsNotFound(err) || errors.Is(err, sql.ErrNoRows) {
		return exceptions.NewNotFound("Not found")
	}
	var notFound *exceptions.NotFound
	if errors.As(err, &notFound) {
		return exceptions.NewNotFound("Not found")
	}
	return persistenceException(err)
}

// ViewSetHandler creates an HTTP handler from a viewset
func ViewSetHandler(vs ViewSet, actions []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Check if it's a list or retrieve
			if forgehttp.GetParam(r, "id") != "" {
				vs.Retrieve(w, r)
			} else {
				vs.List(w, r)
			}
		case http.MethodPost:
			vs.Create(w, r)
		case http.MethodPut:
			vs.Update(w, r)
		case http.MethodPatch:
			vs.PartialUpdate(w, r)
		case http.MethodDelete:
			vs.Destroy(w, r)
		default:
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

type customRoute struct {
	method  string
	path    string
	handler http.HandlerFunc
}

// ActionConfig describes an extra endpoint on a registered resource.
type ActionConfig struct {
	Methods []string // e.g. http.MethodPost; defaults to GET
	Detail  bool     // true: /{resource}/{id}/{path}; false: /{resource}/{path}
	URLPath string   // defaults to the action name
}

type actionRoute struct {
	resource string
	name     string
	cfg      ActionConfig
	handler  http.HandlerFunc
}

func isValidHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// Router is a router for viewsets (DRF-like)
type Router struct {
	prefix       string
	routes       map[string]ViewSet
	customRoutes []customRoute
	actions      []actionRoute
}

// NewRouter creates a new API router
func NewRouter(prefix string) *Router {
	return &Router{
		prefix:       prefix,
		routes:       make(map[string]ViewSet),
		customRoutes: make([]customRoute, 0),
		actions:      make([]actionRoute, 0),
	}
}

// Register registers a viewset with a resource name
func (r *Router) Register(resource string, vs ViewSet) {
	r.routes[resource] = vs
}

// Action registers an extra endpoint on a resource.
func (r *Router) Action(resource, name string, cfg ActionConfig, handler http.HandlerFunc) {
	if len(cfg.Methods) == 0 {
		cfg.Methods = []string{http.MethodGet}
	} else {
		for _, method := range cfg.Methods {
			if !isValidHTTPMethod(method) {
				panic(fmt.Sprintf("unknown HTTP method: %s", method))
			}
		}
	}
	if cfg.URLPath == "" {
		cfg.URLPath = name
	}
	r.actions = append(r.actions, actionRoute{
		resource: resource,
		name:     name,
		cfg:      cfg,
		handler:  handler,
	})
}

// Get registers a custom GET route on the API router
func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.customRoutes = append(r.customRoutes, customRoute{
		method:  http.MethodGet,
		path:    path,
		handler: handler,
	})
}

// Post registers a custom POST route on the API router
func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.customRoutes = append(r.customRoutes, customRoute{
		method:  http.MethodPost,
		path:    path,
		handler: handler,
	})
}

// Handle registers a custom route with a specific method on the API router
func (r *Router) Handle(method, path string, handler http.HandlerFunc) {
	r.customRoutes = append(r.customRoutes, customRoute{
		method:  method,
		path:    path,
		handler: handler,
	})
}

func (r *Router) registerResourceRoutes(router *forgehttp.Router, resource string, vs ViewSet) {
	path := r.prefix + "/" + resource
	router.Route(path, func(sub *forgehttp.Router) {
		sub.Get("/", vs.List)
		sub.Post("/", vs.Create)
		sub.Get("/{id}", vs.Retrieve)
		sub.Put("/{id}", vs.Update)
		sub.Patch("/{id}", vs.PartialUpdate)
		sub.Delete("/{id}", vs.Destroy)
		sub.Options("/", func(w http.ResponseWriter, r *http.Request) {
			docs.OptionsHandler(w, r, vs, nil)
		})
	})
}

func (r *Router) registerActions(router *forgehttp.Router) {
	cleanPrefix := strings.TrimSuffix(r.prefix, "/")
	for _, act := range r.actions {
		cleanResource := strings.Trim(act.resource, "/")
		cleanActionPath := strings.Trim(act.cfg.URLPath, "/")
		var fullPath string
		if act.cfg.Detail {
			fullPath = cleanPrefix + "/" + cleanResource + "/{id}/" + cleanActionPath
		} else {
			fullPath = cleanPrefix + "/" + cleanResource + "/" + cleanActionPath
		}
		for _, method := range act.cfg.Methods {
			router.Method(method, fullPath, act.handler)
		}
	}
}

// RegisterRoutes registers all routes on a chi router
func (r *Router) RegisterRoutes(router *forgehttp.Router) {
	for resource, vs := range r.routes {
		r.registerResourceRoutes(router, resource, vs)
	}
	r.registerActions(router)
	for _, cr := range r.customRoutes {
		cleanPath := "/" + strings.TrimPrefix(cr.path, "/")
		fullPath := r.prefix + cleanPath
		router.Method(cr.method, fullPath, cr.handler)
	}
}

// Helper functions for viewset operations

// paginatableQueryset returns a queryset capable of Offset/Limit. If qs already
// supports Offset and Limit it is returned unchanged. Otherwise, when it exposes
// a zero-argument QuerySet constructor (like orm managers and the generated
// API managers), that constructor is invoked and its result is used, so
// pagination is applied through the queryset instead of loading all rows.
func paginatableQueryset(qs reflect.Value) reflect.Value {
	if !qs.IsValid() {
		return qs
	}
	qsType := qs.Type()
	if _, ok := globalCache.GetMethod(qsType, "Offset"); !ok {
		return maybeConvertToQueryset(qs)
	}
	if _, ok := globalCache.GetMethod(qsType, "Limit"); !ok {
		return maybeConvertToQueryset(qs)
	}
	return qs
}

func maybeConvertToQueryset(qs reflect.Value) reflect.Value {
	qsType := qs.Type()
	qsMethod, ok := globalCache.GetMethod(qsType, "QuerySet")
	if !ok {
		return qs
	}
	if !qsMethod.Func.IsValid() || qsMethod.Func.Type().NumIn() != 1 {
		return qs
	}
	results := qsMethod.Func.Call([]reflect.Value{qs})
	if len(results) == 0 {
		return qs
	}
	if len(results) >= 2 {
		if errVal := results[1]; errVal.IsValid() && errVal.CanInterface() {
			if err, ok := errVal.Interface().(error); ok && err != nil {
				return qs
			}
		}
	}
	first := results[0]
	if !first.IsValid() || (first.Kind() == reflect.Ptr && first.IsNil()) {
		return qs
	}
	if !first.CanInterface() || first.Interface() == nil {
		return qs
	}
	if first.Kind() == reflect.Interface && !first.IsNil() {
		first = first.Elem()
	}
	return first
}

// applyFilters applies query parameter filters to queryset
func applyFilters(qs reflect.Value, r *http.Request, model interface{}) reflect.Value {
	// Get filter parameters from query string
	query := r.URL.Query()
	qsType := qs.Type()

	// Get cached Filter method
	filterMethod, hasFilter := globalCache.GetMethod(qsType, "Filter")

	for key, values := range query {
		if len(values) == 0 || key == "page" || key == "page_size" || key == "ordering" || key == "search" {
			continue
		}

		if hasFilter {
			value := values[0]
			fieldName, lookup := parseLookup(key)
			if resolvedName, field, ok := resolveQueryField(model, fieldName); ok {
				if !field.Serialize {
					continue
				}
				fieldName = resolvedName
			}

			expr := buildFilterExpr(strings.Join([]string{fieldName, lookup}, "__"), value)
			if expr != nil {
				results := filterMethod.Func.Call([]reflect.Value{qs, reflect.ValueOf(expr)})
				if len(results) > 0 {
					if newQS := results[0].Interface(); newQS != nil {
						qs = reflect.ValueOf(newQS)
						qsType = qs.Type()
						// Update cached method for new queryset type
						filterMethod, hasFilter = globalCache.GetMethod(qsType, "Filter")
					}
				}
			}
		}
	}

	return qs
}

func buildFilterExpr(rawKey string, rawValue string) orm.Expression {
	field, lookup := parseLookup(rawKey)
	parsed := parseFilterValue(rawValue)

	f := orm.F(field)
	switch lookup {
	case "exact":
		return f.Eq(parsed)
	case "ne":
		return f.Ne(parsed)
	case "gt":
		return f.Gt(parsed)
	case "gte":
		return f.Gte(parsed)
	case "lt":
		return f.Lt(parsed)
	case "lte":
		return f.Lte(parsed)
	case "contains":
		return f.Contains(rawValue)
	case "icontains":
		return f.IContains(rawValue)
	case "startswith":
		return f.StartsWith(rawValue)
	case "endswith":
		return f.EndsWith(rawValue)
	case "in":
		parts := strings.Split(rawValue, ",")
		args := make([]interface{}, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			args = append(args, parseFilterValue(part))
		}
		if len(args) == 0 {
			return f.Eq(parsed)
		}
		return f.In(args...)
	case "isnull":
		lower := strings.ToLower(rawValue)
		if lower == "true" || lower == "1" {
			return f.IsNull()
		}
		return f.IsNotNull()
	default:
		return f.Eq(parsed)
	}
}

func parseLookup(key string) (string, string) {
	if strings.Contains(key, "__") {
		parts := strings.SplitN(key, "__", 2)
		return parts[0], parts[1]
	}
	return key, "exact"
}

func parseFilterValue(raw string) interface{} {
	if raw == "" {
		return raw
	}
	if boolVal, err := strconv.ParseBool(raw); err == nil {
		return boolVal
	}
	if intVal, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return intVal
	}
	if floatVal, err := strconv.ParseFloat(raw, 64); err == nil {
		return floatVal
	}
	return raw
}

// applyOrdering applies request ordering or the supplied model defaults.
func applyOrdering(qs reflect.Value, r *http.Request, model interface{}, defaults ...string) reflect.Value {
	ordering := r.URL.Query().Get("ordering")
	if ordering == "" {
		ordering = strings.Join(defaults, ",")
		if ordering == "" {
			return qs
		}
	}

	qsType := qs.Type()
	orderByMethod, ok := globalCache.GetMethod(qsType, "OrderBy")
	if !ok {
		return qs
	}

	rawFields := strings.Split(ordering, ",")
	args := make([]reflect.Value, 0, len(rawFields)+1)
	args = append(args, qs) // Add receiver
	for _, field := range rawFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		descending := strings.HasPrefix(field, "-")
		fieldName := strings.TrimLeft(field, "-")
		if resolvedName, schemaField, ok := resolveQueryField(model, fieldName); ok {
			if !schemaField.Serialize {
				continue
			}
			field = resolvedName
			if descending {
				field = "-" + field
			}
		}
		args = append(args, reflect.ValueOf(field))
	}
	if len(args) <= 1 { // Only receiver, no fields
		return qs
	}
	results := orderByMethod.Func.Call(args)
	if len(results) > 0 {
		if newQS := results[0].Interface(); newQS != nil {
			return reflect.ValueOf(newQS)
		}
	}

	return qs
}

func defaultOrdering(model interface{}) []string {
	s, ok := model.(schema.Schema)
	if !ok {
		return []string{"id"}
	}
	if ordering := s.Meta().OrderBy; len(ordering) > 0 {
		return append([]string(nil), ordering...)
	}
	for _, field := range s.Fields() {
		if field.PrimaryKey {
			if field.DBColumn != "" {
				return []string{field.DBColumn}
			}
			if resolved, ok := schema.ResolveField(model, field); ok && resolved.DBTag != "" && resolved.DBTag != "-" {
				return []string{resolved.DBTag}
			}
			if field.Name != "" {
				return []string{field.Name}
			}
		}
	}
	return []string{"id"}
}

func resolveQueryField(model interface{}, name string) (string, schema.Field, bool) {
	s, ok := model.(schema.Schema)
	if !ok {
		return "", schema.Field{}, false
	}
	for _, field := range s.Fields() {
		for _, alias := range resolvedFieldNames(model, field) {
			if !strings.EqualFold(alias, name) {
				continue
			}
			column := field.DBColumn
			if column == "" {
				if resolved, found := schema.ResolveField(model, field); found && resolved.DBTag != "" && resolved.DBTag != "-" {
					column = resolved.DBTag
				}
			}
			if column == "" {
				column = field.Name
			}
			return column, field, true
		}
	}
	return "", schema.Field{}, false
}

// fieldError describes a single request-field conversion failure.
type fieldError struct {
	Field   string
	Message string
}

func (e *fieldError) Error() string {
	return fmt.Sprintf("invalid value for field %q: %s", e.Field, e.Message)
}

// validateModelInstance runs the model's own validation (Clean method,
// schema Clean hook, Validate method) in the same order as orm.Manager, so
// the API request path can reject invalid payloads before business hooks run.
func validateModelInstance(instance interface{}) error {
	if validatable, ok := instance.(interface{ Clean() error }); ok {
		if err := validatable.Clean(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}
	if s, ok := instance.(schema.Schema); ok {
		if hooks := s.Hooks(); hooks != nil && hooks.Clean != nil {
			if err := hooks.Clean(instance); err != nil {
				return fmt.Errorf("schema validation failed: %w", err)
			}
		}
	}
	if validatable, ok := instance.(interface{ Validate() error }); ok {
		if err := validatable.Validate(); err != nil {
			return fmt.Errorf("model validation failed: %w", err)
		}
	}
	if s, ok := instance.(schema.Schema); ok {
		if err := validation.ValidateModelWithSchema(validation.NewValidator(), instance, s.Fields()); err != nil {
			return schemaModelValidationError(instance, s.Fields(), err)
		}
	}
	return nil
}

func schemaModelValidationError(instance interface{}, fields []schema.Field, err error) error {
	message := err.Error()
	schemaName := ""
	if idx := strings.Index(message, ":"); idx >= 0 {
		schemaName = strings.TrimSpace(message[:idx])
		message = strings.TrimSpace(message[idx+1:])
	}
	fieldName := schemaName
	for _, field := range fields {
		if field.Name != schemaName && field.DBColumn != schemaName {
			continue
		}
		if resolved, ok := schema.ResolveField(instance, field); ok && resolved.JSONName != "" && resolved.JSONName != "-" {
			fieldName = resolved.JSONName
		}
		break
	}
	if fieldName == "" {
		fieldName = "non_field_errors"
	}
	return forgeerrors.NewInvalidInputError(fieldName, message)
}

// modelValidationException converts a validateModelInstance failure into a
// 400 validation exception. Structured validation errors keep their field
// detail via persistenceException; anything else becomes a non-field error.
func modelValidationException(err error) error {
	var validationErrs *validation.ValidationErrors
	if errors.As(err, &validationErrs) {
		return persistenceException(err)
	}
	var validationErr *validation.ValidationError
	if errors.As(err, &validationErr) {
		return persistenceException(err)
	}
	var invalidInput *forgeerrors.InvalidInputError
	if errors.As(err, &invalidInput) {
		return persistenceException(err)
	}
	return exceptions.NewValidationError(map[string][]string{"non_field_errors": {err.Error()}})
}

// validationErrorForPopulate converts a populateFromMap failure into a 400
// validation exception so the ErrorWriter (or the default RFC 7807 writer)
// renders it instead of leaking anything verbatim.
func validationErrorForPopulate(err error) error {
	if fe, ok := err.(*fieldError); ok {
		return exceptions.NewValidationError(map[string][]string{fe.Field: {fe.Message}})
	}
	return exceptions.NewValidationError(map[string][]string{"non_field_errors": {err.Error()}})
}

// populateFromMap populates a model instance from a map, optionally ignoring specified keys.
// Keys are matched case-insensitively against the Go field name, the json tag
// name and the db tag name. A value that cannot be converted to the field type
// returns a *fieldError; nothing is silently dropped.
func populateFromMap(instance interface{}, data map[string]interface{}, ignoredKeys ...string) error {
	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	if instanceValue.Kind() != reflect.Struct {
		return nil
	}

	ignored := make(map[string]bool, len(ignoredKeys))
	for _, k := range ignoredKeys {
		ignored[strings.ToLower(k)] = true
	}
	schemaFields := schemaFieldsByRequestName(instance)

	var applyToStruct func(target reflect.Value) error
	applyToStruct = func(target reflect.Value) error {
		targetType := target.Type()
		for i := 0; i < targetType.NumField(); i++ {
			field := targetType.Field(i)
			if !field.IsExported() {
				continue
			}
			fieldValue := target.Field(i)

			if field.Anonymous {
				switch fieldValue.Kind() {
				case reflect.Struct:
					if err := applyToStruct(fieldValue); err != nil {
						return err
					}
				case reflect.Ptr:
					if fieldValue.IsNil() {
						continue
					}
					if fieldValue.Elem().Kind() == reflect.Struct {
						if err := applyToStruct(fieldValue.Elem()); err != nil {
							return err
						}
					}
				}
				continue
			}

			jsonTag := field.Tag.Get("json")
			tagParts := strings.Split(jsonTag, ",")
			key := tagParts[0]
			dbTag := strings.Split(field.Tag.Get("db"), ",")[0]
			fieldSchema := schemaFieldForStructField(schemaFields, field, key, dbTag)
			if fieldSchema == nil && (key == "" || key == "-") {
				continue
			}
			fieldNames := []string{key, field.Name, dbTag}
			if fieldSchema != nil {
				fieldNames = append(fieldNames, resolvedFieldNames(instance, *fieldSchema)...)
			}
			isIgnored := false
			for _, name := range fieldNames {
				if ignored[strings.ToLower(name)] {
					isIgnored = true
					break
				}
			}
			if isIgnored {
				continue
			}
			requestKey := key
			value, valueExists := data[requestKey]
			if !valueExists && fieldSchema != nil {
				for _, alias := range resolvedFieldNames(instance, *fieldSchema) {
					if candidate, exists := data[alias]; exists {
						requestKey = alias
						value = candidate
						valueExists = true
						break
					}
				}
			}
			if valueExists {
				if fieldValue.CanSet() {
					if err := setFieldValue(fieldValue, value, fieldSchema); err != nil {
						return &fieldError{Field: requestKey, Message: err.Error()}
					}
				}
			}
		}
		return nil
	}

	return applyToStruct(instanceValue)
}

func schemaFieldsByRequestName(instance interface{}) map[string]*schema.Field {
	modelSchema, ok := instance.(schema.Schema)
	if !ok {
		return nil
	}
	fields := modelSchema.Fields()
	result := make(map[string]*schema.Field, len(fields)*3)
	for i := range fields {
		field := &fields[i]
		for _, name := range resolvedFieldNames(instance, *field) {
			result[strings.ToLower(name)] = field
		}
	}
	return result
}

func schemaFieldForStructField(fields map[string]*schema.Field, field reflect.StructField, jsonName, dbName string) *schema.Field {
	for _, name := range []string{jsonName, dbName, field.Name} {
		if schemaField := fields[strings.ToLower(name)]; schemaField != nil {
			return schemaField
		}
	}
	return nil
}

func applySchemaDefaults(model interface{}, data map[string]interface{}) {
	if data == nil {
		return
	}
	modelSchema, ok := model.(schema.Schema)
	if !ok {
		return
	}
	for _, field := range modelSchema.Fields() {
		if field.Default == nil {
			continue
		}
		resolved, _ := schema.ResolveField(model, field)
		requestName := resolved.JSONName
		if requestName == "" || requestName == "-" {
			requestName = field.Name
		}
		present := false
		for _, name := range resolvedFieldNames(model, field) {
			if name == "" {
				continue
			}
			if _, exists := data[name]; exists {
				present = true
				break
			}
		}
		if !present {
			data[requestName] = field.Default
		}
	}
}

// getSchemaPrimaryKeyField returns the primary key field from the model's schema if available.
// Returns the Go struct field name, db column name, and json tag name.
// When no schema is available it returns empty strings and callers fall back
// to the "id"/"ID"/"Id" variants.
func (vs *BaseViewSet) getSchemaPrimaryKeyField() (goName, dbName, jsonName string) {
	if vs.Model == nil {
		return "", "", ""
	}
	// Check if model implements schema.Schema
	modelSchema, ok := vs.Model.(schema.Schema)
	if !ok {
		return "", "", ""
	}
	var primaryKey *schema.Field
	for _, f := range modelSchema.Fields() {
		if f.PrimaryKey {
			field := f
			primaryKey = &field
			break
		}
	}
	if primaryKey == nil {
		return "", "", ""
	}
	resolved, found := schema.ResolveField(vs.Model, *primaryKey)
	if !found {
		dbName := primaryKey.DBColumn
		if dbName == "" {
			dbName = primaryKey.Name
		}
		return "", dbName, primaryKey.Name
	}
	dbName = resolved.DBColumn
	if dbName == "" || dbName == "-" {
		dbName = resolved.DBTag
	}
	if dbName == "" || dbName == "-" {
		dbName = resolved.SchemaName
	}
	jsonName = resolved.JSONName
	if jsonName == "" || jsonName == "-" {
		jsonName = resolved.SchemaName
	}
	return resolved.GoName, dbName, jsonName
}

// getPrimaryKeyValue reads the current primary key value from instance, preferring
// the schema primary key field names and falling back to "ID"/"Id"/"id".
func getPrimaryKeyValue(instance interface{}, pkGoName, pkDBName, pkJSONName string) interface{} {
	if modelSchema, ok := instance.(schema.Schema); ok {
		for _, field := range modelSchema.Fields() {
			if !field.PrimaryKey {
				continue
			}
			if resolved, found := schema.ResolveField(instance, field); found {
				if value, exists := concreteFieldValue(reflect.ValueOf(instance), resolved.StructField.Index); exists && value.CanInterface() {
					return value.Interface()
				}
			}
			break
		}
	}
	if modelWithID, ok := instance.(interface{ GetID() int64 }); ok {
		return modelWithID.GetID()
	}
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	if fv, ok := findPrimaryKeyField(v, []string{pkGoName, pkDBName, pkJSONName}); ok {
		return fv.Interface()
	}
	if pkGoName == "" && pkDBName == "" && pkJSONName == "" {
		if fv, ok := findPrimaryKeyField(v, []string{"ID", "Id", "id"}); ok {
			return fv.Interface()
		}
	}
	return nil
}

func primaryKeyInt64(instance interface{}, pkGoName, pkDBName, pkJSONName string) (int64, bool) {
	value := getPrimaryKeyValue(instance, pkGoName, pkDBName, pkJSONName)
	if value == nil {
		return 0, false
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if rv.Uint() > math.MaxInt64 {
			return 0, false
		}
		return int64(rv.Uint()), true
	default:
		return 0, false
	}
}

func getManagerInstance(manager reflect.Value, managerType reflect.Type, ctx context.Context, id int64) (interface{}, error) {
	getMethod, ok := globalCache.GetMethod(managerType, "Get")
	if !ok {
		return nil, errors.New("Get method not found")
	}
	results := getMethod.Func.Call([]reflect.Value{manager, reflect.ValueOf(ctx), reflect.ValueOf(id)})
	if len(results) < 2 {
		return nil, errors.New("invalid Get method")
	}
	if !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok {
			return nil, err
		}
		return nil, errors.New("Get returned a non-error failure value")
	}
	if isNilReflectValue(results[0]) {
		return nil, errors.New("Get returned a nil instance")
	}
	return results[0].Interface(), nil
}

func isNilReflectValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// restorePrimaryKey restores the primary key value on instance from the original value or urlID.
// Schema primary key field names take precedence; "ID"/"Id"/"id" are only used
// as a fallback when no schema primary key is known. A body value for the
// primary key never changes which row is written.
func restorePrimaryKey(instance interface{}, origPK interface{}, urlID int64, pkGoName, pkDBName, pkJSONName string) {
	if instance == nil {
		return
	}
	hasSchemaPK := pkGoName != "" || pkDBName != "" || pkJSONName != ""
	if !hasSchemaPK {
		if m, ok := instance.(interface{ SetID(int64) }); ok {
			if origInt, ok := origPK.(int64); ok && origInt != 0 {
				m.SetID(origInt)
			} else {
				m.SetID(urlID)
			}
			return
		}
	}
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	setField := func(f reflect.Value) bool {
		if !f.CanSet() {
			return false
		}
		if origPK != nil {
			origVal := reflect.ValueOf(origPK)
			if origVal.IsValid() && origVal.Type().AssignableTo(f.Type()) {
				f.Set(origVal)
				return true
			}
		}
		if isIntKind(f.Kind()) {
			if f.Kind() >= reflect.Uint && f.Kind() <= reflect.Uint64 {
				f.SetUint(uint64(urlID))
			} else {
				f.SetInt(urlID)
			}
			return true
		}
		return false
	}
	if modelSchema, ok := instance.(schema.Schema); ok {
		for _, field := range modelSchema.Fields() {
			if !field.PrimaryKey {
				continue
			}
			if resolved, found := schema.ResolveField(instance, field); found {
				if value, exists := concreteFieldValue(reflect.ValueOf(instance), resolved.StructField.Index); exists && setField(value) {
					return
				}
			}
			break
		}
	}
	if modelWithID, ok := instance.(interface{ SetID(int64) }); ok {
		if original, ok := origPK.(int64); ok && original != 0 {
			modelWithID.SetID(original)
		} else {
			modelWithID.SetID(urlID)
		}
		return
	}

	// Prefer the schema primary key when known.
	if pkGoName != "" || pkDBName != "" || pkJSONName != "" {
		if f, ok := findPrimaryKeyField(v, []string{pkGoName, pkDBName, pkJSONName}); ok {
			if setField(f) {
				return
			}
		}
	}

	for _, name := range []string{"ID", "Id", "id"} {
		f := v.FieldByName(name)
		if f.IsValid() {
			if setField(f) {
				return
			}
		}
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		parts := strings.Split(tag, ",")
		if len(parts) > 0 && strings.EqualFold(parts[0], "id") {
			f := v.Field(i)
			if f.IsValid() {
				if setField(f) {
					return
				}
			}
		}
	}
}

func concreteFieldValue(value reflect.Value, index []int) (reflect.Value, bool) {
	for value.IsValid() && value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}
	for _, fieldIndex := range index {
		for value.IsValid() && value.Kind() == reflect.Ptr {
			if value.IsNil() {
				return reflect.Value{}, false
			}
			value = value.Elem()
		}
		if !value.IsValid() || value.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		value = value.Field(fieldIndex)
	}
	return value, value.IsValid()
}

// findPrimaryKeyField locates the settable struct field matching any of the given
// names (Go field name case-insensitively, or exact json/db tag match),
// searching into embedded structs.
func findPrimaryKeyField(v reflect.Value, names []string) (reflect.Value, bool) {
	clean := make([]string, 0, len(names))
	lowered := make(map[string]bool)
	for _, n := range names {
		if n == "" {
			continue
		}
		clean = append(clean, n)
		lowered[strings.ToLower(n)] = true
	}
	if len(clean) == 0 {
		return reflect.Value{}, false
	}
	// Fast path: direct Go field name lookup.
	for _, n := range clean {
		if f := v.FieldByName(n); f.IsValid() {
			return f, true
		}
	}
	t := v.Type()
	var search func(rv reflect.Value, rt reflect.Type) (reflect.Value, bool)
	search = func(rv reflect.Value, rt reflect.Type) (reflect.Value, bool) {
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			if !sf.IsExported() {
				continue
			}
			fv := rv.Field(i)
			if sf.Anonymous {
				ft := sf.Type
				fvv := fv
				if ft.Kind() == reflect.Ptr {
					if fvv.IsNil() {
						continue
					}
					fvv = fvv.Elem()
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct {
					if res, ok := search(fvv, ft); ok {
						return res, true
					}
				}
				continue
			}
			jsonTag := strings.Split(sf.Tag.Get("json"), ",")[0]
			dbTag := strings.Split(sf.Tag.Get("db"), ",")[0]
			if lowered[strings.ToLower(sf.Name)] {
				return fv, true
			}
			for _, n := range clean {
				if jsonTag != "" && jsonTag != "-" && n == jsonTag {
					return fv, true
				}
				if dbTag != "" && n == dbTag {
					return fv, true
				}
			}
		}
		return reflect.Value{}, false
	}
	return search(v, t)
}

func isIntKind(k reflect.Kind) bool {
	return (k >= reflect.Int && k <= reflect.Int64) || (k >= reflect.Uint && k <= reflect.Uint64)
}

// setFieldValue sets a field value from interface{}. It supports the Go types
// codegen emits (float32/float64 for Float/Decimal, []byte for JSON/Bytes)
// plus re-marshaling for map, slice, struct and interface fields. JSON numbers
// arrive as float64 (or json.Number) and numeric strings are accepted for
// numeric fields. An incompatible value returns an error; it is never dropped
// silently.
func setFieldValue(field reflect.Value, value interface{}, schemaField ...*schema.Field) error {
	if !field.CanSet() {
		return nil
	}
	var fieldSchema *schema.Field
	if len(schemaField) > 0 {
		fieldSchema = schemaField[0]
	}

	if fieldSchema != nil && fieldSchema.Type == schema.TypeJSON {
		encoded, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("invalid JSON value: %w", err)
		}
		if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Uint8 {
			field.SetBytes(encoded)
			return nil
		}
		decoded := reflect.New(field.Type())
		if err := json.Unmarshal(encoded, decoded.Interface()); err != nil {
			return fmt.Errorf("invalid JSON value for %s: %w", field.Type(), err)
		}
		field.Set(decoded.Elem())
		return nil
	}

	if value == nil {
		switch field.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}

	fieldType := field.Type()

	// []byte fields (codegen JSON/Bytes): accept a base64 string like
	// encoding/json does, or an array of byte values.
	if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8 {
		strictByteArray := fieldSchema != nil && fieldSchema.Type == schema.TypeBytes
		return setBytesField(field, value, strictByteArray)
	}

	switch field.Kind() {
	case reflect.String:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		field.SetString(s)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setIntField(field, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUintField(field, value)
	case reflect.Float32, reflect.Float64:
		return setFloatField(field, value)
	case reflect.Bool:
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
		field.SetBool(b)
		return nil
	case reflect.Ptr:
		elem := reflect.New(fieldType.Elem())
		if err := setFieldValue(elem.Elem(), value, fieldSchema); err != nil {
			return err
		}
		field.Set(elem)
		return nil
	case reflect.Slice, reflect.Map:
		return setCompositeField(field, value)
	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			return setTimeField(field, value, fieldSchema)
		}
		return setCompositeField(field, value)
	case reflect.Interface:
		valueValue := reflect.ValueOf(value)
		if valueValue.Type().AssignableTo(fieldType) {
			field.Set(valueValue)
			return nil
		}
		if valueValue.Type().Implements(fieldType) {
			field.Set(valueValue)
			return nil
		}
		return setCompositeField(field, value)
	default:
		valueValue := reflect.ValueOf(value)
		if valueValue.Type().AssignableTo(fieldType) {
			field.Set(valueValue)
			return nil
		}
		return fmt.Errorf("cannot assign %T to %s", value, fieldType)
	}
}

// setFloatField assigns JSON numbers (float64 or json.Number), numeric strings
// and integers to float32/float64 fields.
func setFloatField(field reflect.Value, value interface{}) error {
	bitSize := 64
	if field.Kind() == reflect.Float32 {
		bitSize = 32
	}
	f, err := parseFloatValue(value, bitSize)
	if err != nil {
		return err
	}
	field.SetFloat(f)
	return nil
}

func parseFloatValue(value interface{}, bitSize int) (float64, error) {
	switch v := value.(type) {
	case float64:
		if bitSize == 32 {
			if _, err := strconv.ParseFloat(strconv.FormatFloat(v, 'g', -1, 64), 32); err != nil {
				return 0, fmt.Errorf("value %v out of range for float32", v)
			}
		}
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, fmt.Errorf("invalid numeric value %v", v)
		}
		return v, nil
	case float32:
		return float64(v), nil
	case json.Number:
		f, err := strconv.ParseFloat(strings.TrimSpace(v.String()), bitSize)
		if err != nil {
			return 0, fmt.Errorf("invalid numeric value %q", v.String())
		}
		return f, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, fmt.Errorf("empty string is not a number")
		}
		f, err := strconv.ParseFloat(s, bitSize)
		if err != nil {
			return 0, fmt.Errorf("invalid numeric value %q", v)
		}
		return f, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return parseFloatValue(rv.Float(), bitSize)
		}
		return 0, fmt.Errorf("expected number, got %T", value)
	}
}

// setIntField assigns JSON numbers, integers, json.Number values and numeric
// strings to signed integer fields.
func setIntField(field reflect.Value, value interface{}) error {
	var n int64
	switch v := value.(type) {
	case float64:
		var err error
		n, err = integralFloatToInt64(v)
		if err != nil {
			return err
		}
	case float32:
		var err error
		n, err = integralFloatToInt64(float64(v))
		if err != nil {
			return err
		}
	case int:
		n = int64(v)
	case int8:
		n = int64(v)
	case int16:
		n = int64(v)
	case int32:
		n = int64(v)
	case int64:
		n = v
	case uint:
		n = int64(v)
	case uint8:
		n = int64(v)
	case uint16:
		n = int64(v)
	case uint32:
		n = int64(v)
	case uint64:
		n = int64(v)
	case json.Number:
		parsed, err := parseIntegralInt64(v.String())
		if err != nil {
			return fmt.Errorf("invalid integer value %q", v.String())
		}
		n = parsed
	case string:
		parsed, err := parseIntegralInt64(v)
		if err != nil {
			return fmt.Errorf("invalid integer value %q", v)
		}
		n = parsed
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n = rv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n = int64(rv.Uint())
		case reflect.Float32, reflect.Float64:
			var err error
			n, err = integralFloatToInt64(rv.Float())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	}
	if field.OverflowInt(n) {
		return fmt.Errorf("value %d out of range for %s", n, field.Type())
	}
	field.SetInt(n)
	return nil
}

func parseIntegralInt64(value string) (int64, error) {
	rational, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
		return 0, fmt.Errorf("not an integer")
	}
	return rational.Num().Int64(), nil
}

func integralFloatToInt64(value float64) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value {
		return 0, fmt.Errorf("non-integral value %v invalid for integer field", value)
	}
	formatted := strconv.FormatFloat(value, 'f', -1, 64)
	return parseIntegralInt64(formatted)
}

// setUintField assigns JSON numbers, integers, json.Number values and numeric
// strings to unsigned integer fields.
func setUintField(field reflect.Value, value interface{}) error {
	var n uint64
	switch v := value.(type) {
	case float64:
		var err error
		n, err = integralFloatToUint64(v)
		if err != nil {
			return err
		}
	case float32:
		var err error
		n, err = integralFloatToUint64(float64(v))
		if err != nil {
			return err
		}
	case int:
		if v < 0 {
			return fmt.Errorf("negative value %d invalid for %s", v, field.Type())
		}
		n = uint64(v)
	case int8:
		if v < 0 {
			return fmt.Errorf("negative value %d invalid for %s", v, field.Type())
		}
		n = uint64(v)
	case int16:
		if v < 0 {
			return fmt.Errorf("negative value %d invalid for %s", v, field.Type())
		}
		n = uint64(v)
	case int32:
		if v < 0 {
			return fmt.Errorf("negative value %d invalid for %s", v, field.Type())
		}
		n = uint64(v)
	case int64:
		if v < 0 {
			return fmt.Errorf("negative value %d invalid for %s", v, field.Type())
		}
		n = uint64(v)
	case uint:
		n = uint64(v)
	case uint8:
		n = uint64(v)
	case uint16:
		n = uint64(v)
	case uint32:
		n = uint64(v)
	case uint64:
		n = v
	case json.Number:
		parsed, err := parseIntegralUint64(v.String())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value %q", v.String())
		}
		n = parsed
	case string:
		parsed, err := parseIntegralUint64(v)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value %q", v)
		}
		n = parsed
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n = rv.Uint()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if rv.Int() < 0 {
				return fmt.Errorf("negative value %d invalid for %s", rv.Int(), field.Type())
			}
			n = uint64(rv.Int())
		case reflect.Float32, reflect.Float64:
			var err error
			n, err = integralFloatToUint64(rv.Float())
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("expected unsigned integer, got %T", value)
		}
	}
	if field.OverflowUint(n) {
		return fmt.Errorf("value %d out of range for %s", n, field.Type())
	}
	field.SetUint(n)
	return nil
}

func parseIntegralUint64(value string) (uint64, error) {
	rational, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	if !ok || !rational.IsInt() || rational.Sign() < 0 || !rational.Num().IsUint64() {
		return 0, fmt.Errorf("not an unsigned integer")
	}
	return rational.Num().Uint64(), nil
}

func integralFloatToUint64(value float64) (uint64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || math.Trunc(value) != value {
		return 0, fmt.Errorf("non-integral value %v invalid for unsigned integer field", value)
	}
	formatted := strconv.FormatFloat(value, 'f', -1, 64)
	return parseIntegralUint64(formatted)
}

// setBytesField assigns a base64 string (like encoding/json does for []byte),
// a byte array, or another []uint8 value to a []byte field.
func setBytesField(field reflect.Value, value interface{}, strictByteArray bool) error {
	switch v := value.(type) {
	case string:
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return fmt.Errorf("invalid base64 string: %v", err)
		}
		field.SetBytes(decoded)
		return nil
	case []byte:
		field.SetBytes(v)
		return nil
	case []interface{}:
		b := make([]byte, len(v))
		for i, e := range v {
			n, ok := byteValue(e)
			if !ok {
				if !strictByteArray {
					encoded, err := json.Marshal(v)
					if err != nil {
						return fmt.Errorf("invalid JSON array: %w", err)
					}
					field.SetBytes(encoded)
					return nil
				}
				return fmt.Errorf("element %d must be an integer between 0 and 255", i)
			}
			b[i] = n
		}
		field.SetBytes(b)
		return nil
	case map[string]interface{}:
		if strictByteArray {
			return fmt.Errorf("expected base64 string or byte array, got %T", value)
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("invalid JSON object: %w", err)
		}
		field.SetBytes(encoded)
		return nil
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
			field.SetBytes(rv.Bytes())
			return nil
		}
		return fmt.Errorf("expected base64 string, got %T", value)
	}
}

func byteValue(value interface{}) (byte, bool) {
	switch v := value.(type) {
	case float64:
		if v < 0 || v > 255 || v != float64(int(v)) {
			return 0, false
		}
		return byte(int(v)), true
	case int:
		if v < 0 || v > 255 {
			return 0, false
		}
		return byte(v), true
	case json.Number:
		n, err := parseIntegralUint64(v.String())
		if err != nil || n > 255 {
			return 0, false
		}
		return byte(n), true
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if n := rv.Int(); n >= 0 && n <= 255 {
				return byte(n), true
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if n := rv.Uint(); n <= 255 {
				return byte(n), true
			}
		case reflect.Float32, reflect.Float64:
			if f := rv.Float(); f >= 0 && f <= 255 && f == float64(int(f)) {
				return byte(int(f)), true
			}
		}
		return 0, false
	}
}

// setCompositeField assigns decoded JSON values (maps, slices, structs) to map,
// slice, struct or compatible fields by re-marshaling through encoding/json.
func setCompositeField(field reflect.Value, value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cannot encode value for %s: %v", field.Type(), err)
	}
	target := reflect.New(field.Type())
	if err := json.Unmarshal(raw, target.Interface()); err != nil {
		return fmt.Errorf("cannot decode value into %s: %v", field.Type(), err)
	}
	field.Set(target.Elem())
	return nil
}

// setTimeField assigns date/time strings to time.Time fields.
func setTimeField(field reflect.Value, value interface{}, schemaField *schema.Field) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected date/time string, got %T", value)
	}
	if str == "" {
		field.Set(reflect.ValueOf(time.Time{}))
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02"}
	if schemaField != nil {
		switch schemaField.Type {
		case schema.TypeTime:
			layouts = []string{"15:04:05", "15:04", time.RFC3339, "2006-01-02"}
		case schema.TypeDate:
			layouts = []string{"2006-01-02"}
		case schema.TypeDateTime:
			layouts = []string{time.RFC3339}
		}
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, str); err == nil {
			field.Set(reflect.ValueOf(parsed))
			return nil
		}
	}
	return fmt.Errorf("invalid date/time value %q", str)
}
