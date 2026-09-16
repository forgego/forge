package api

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/docs"
	apierrors "github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/orm"
	forgehttp "github.com/forgego/forge/server"
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
	Serializer     func() Serializer
	Queryset       interface{} // This would be a QuerySet in real implementation
	Model          interface{}
	Authentication []authentication.Authentication
	Permissions    []permissions.Permission
	Throttles      []throttling.Throttle
	ErrorWriter    func(http.ResponseWriter, *http.Request, error)

	actionMu sync.RWMutex
	action   string
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

// GetAction returns the current action name for permission checks.
func (vs *BaseViewSet) GetAction() string {
	vs.actionMu.RLock()
	defer vs.actionMu.RUnlock()
	return vs.action
}

// SetAction sets the current action name for permission checks.
func (vs *BaseViewSet) SetAction(action string) {
	vs.actionMu.Lock()
	defer vs.actionMu.Unlock()
	vs.action = action
}

func (vs *BaseViewSet) authenticateRequest(r *http.Request) error {
	result, err := authentication.AuthenticateRequest(r, vs.Authentication)
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

func (vs *BaseViewSet) checkPermissions(r *http.Request) error {
	if permissions.CheckPermissions(r, vs, vs.Permissions) {
		return nil
	}
	for _, permission := range vs.Permissions {
		if !permission.HasPermission(r, vs) {
			return exceptions.NewPermissionDenied(permission.GetMessage())
		}
	}
	return exceptions.NewPermissionDenied("Permission denied")
}

func (vs *BaseViewSet) checkObjectPermissions(r *http.Request, object interface{}) error {
	if permissions.CheckObjectPermissions(r, vs, object, vs.Permissions) {
		return nil
	}
	for _, permission := range vs.Permissions {
		if !permission.HasObjectPermission(r, vs, object) {
			return exceptions.NewPermissionDenied(permission.GetMessage())
		}
	}
	return exceptions.NewPermissionDenied("Permission denied")
}

func (vs *BaseViewSet) checkThrottles(r *http.Request) error {
	err := throttling.CheckThrottles(r, vs, vs.Throttles)
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
	if !vs.checkRequest(w, r, "list") {
		return
	}
	ctx := r.Context()

	// Get pagination parameters
	page, pageSize, _ := ParsePaginationParams(r, 20)

	// Get queryset using reflection
	querysetValue := reflect.ValueOf(vs.Queryset)
	if !querysetValue.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Queryset not set")
		return
	}

	// Apply filtering from query params
	qs := applyFilters(querysetValue, r)

	// Apply ordering
	qs = applyOrdering(qs, r)

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
				_ = forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
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
				_ = forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
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
	serialized := SerializeMany(resultList)
	serialized = filterOutputMany(vs.Serializer(), serialized)

	// Send paginated response
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = SendPaginatedResponse(w, r, serialized, int(totalCount), page, pageSize)
}

// Create handles POST /resource/
func (vs *BaseViewSet) Create(w http.ResponseWriter, r *http.Request) {
	if !vs.checkRequest(w, r, "create") {
		return
	}
	ctx := r.Context()

	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
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

	// Create model instance from serializer data
	modelValue := reflect.New(reflect.TypeOf(vs.Model).Elem())
	instance := modelValue.Interface()

	// Populate instance from data
	populateFromMap(instance, data)

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
		reflect.ValueOf(ctx),
		reflect.ValueOf(instance),
	})

	if len(results) > 0 && !results[0].IsNil() {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Serialize and return created instance
	serialized := SerializeModel(instance)
	serialized = filterOutputMap(vs.Serializer(), serialized)
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = forgehttp.SendJSON(w, http.StatusCreated, serialized)
}

// Retrieve handles GET /resource/{id}/
func (vs *BaseViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
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
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusNotFound, err.Error())
			return
		}
	}

	instance := results[0].Interface()
	if !vs.allowObject(w, r, instance) {
		return
	}
	serialized := SerializeModel(instance)
	serialized = filterOutputMap(vs.Serializer(), serialized)
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = forgehttp.SendJSON(w, http.StatusOK, serialized)
}

// Update handles PUT /resource/{id}/
func (vs *BaseViewSet) Update(w http.ResponseWriter, r *http.Request) {
	vs.update(w, r, "update")
}

func (vs *BaseViewSet) update(w http.ResponseWriter, r *http.Request, action string) {
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
	if err := forgehttp.GetJSON(r, &data); err != nil {
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
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusNotFound, err.Error())
			return
		}
	}

	instance := getResults[0].Interface()
	if !vs.allowObject(w, r, instance) {
		return
	}

	// Populate from data
	populateFromMap(instance, data)

	// Update
	updateMethod, ok := globalCache.GetMethod(managerType, "Update")
	if !ok {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Update method not found")
		return
	}

	updateResults := updateMethod.Func.Call([]reflect.Value{
		manager,
		reflect.ValueOf(ctx),
		reflect.ValueOf(instance),
	})

	if len(updateResults) > 0 && !updateResults[0].IsNil() {
		if err, ok := updateResults[0].Interface().(error); ok && err != nil {
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	serialized := SerializeModel(instance)
	serialized = filterOutputMap(vs.Serializer(), serialized)
	forgehttp.SendJSON(w, http.StatusOK, serialized)
}

// PartialUpdate handles PATCH /resource/{id}/
func (vs *BaseViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	vs.update(w, r, "partial_update")
}

// Destroy handles DELETE /resource/{id}/
func (vs *BaseViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
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
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusNotFound, err.Error())
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
			// nolint:errcheck // HTTP response errors can't be handled meaningfully
			_ = forgehttp.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
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

// applyFilters applies query parameter filters to queryset
func applyFilters(qs reflect.Value, r *http.Request) reflect.Value {
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

			expr := buildFilterExpr(key, value)
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

// applyOrdering applies ordering from query parameters
func applyOrdering(qs reflect.Value, r *http.Request) reflect.Value {
	ordering := r.URL.Query().Get("ordering")
	if ordering == "" {
		return qs
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

// populateFromMap populates a model instance from a map
func populateFromMap(instance interface{}, data map[string]interface{}) {
	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	if instanceValue.Kind() != reflect.Struct {
		return
	}

	var applyToStruct func(target reflect.Value)
	applyToStruct = func(target reflect.Value) {
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
					applyToStruct(fieldValue)
				case reflect.Ptr:
					if fieldValue.IsNil() {
						continue
					}
					if fieldValue.Elem().Kind() == reflect.Struct {
						applyToStruct(fieldValue.Elem())
					}
				}
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			tagParts := strings.Split(jsonTag, ",")
			key := tagParts[0]
			if key == "" {
				continue
			}
			if value, ok := data[key]; ok {
				if fieldValue.CanSet() {
					setFieldValue(fieldValue, value)
				}
			}
		}
	}

	applyToStruct(instanceValue)
}

// setFieldValue sets a field value from interface{}
func setFieldValue(field reflect.Value, value interface{}) {
	if !field.CanSet() {
		return
	}

	valueValue := reflect.ValueOf(value)
	if !valueValue.IsValid() {
		switch field.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
			field.Set(reflect.Zero(field.Type()))
		}
		return
	}

	switch field.Kind() {
	case reflect.String:
		if valueValue.Kind() == reflect.String {
			field.SetString(valueValue.String())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if valueValue.Kind() == reflect.Float64 {
			field.SetInt(int64(valueValue.Float()))
		} else if valueValue.Kind() == reflect.Int || valueValue.Kind() == reflect.Int64 {
			field.SetInt(valueValue.Int())
		}
	case reflect.Bool:
		if valueValue.Kind() == reflect.Bool {
			field.SetBool(valueValue.Bool())
		}
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			if str, ok := value.(string); ok {
				if str == "" {
					field.Set(reflect.ValueOf(time.Time{}))
					return
				}
				if parsed, err := time.Parse(time.RFC3339, str); err == nil {
					field.Set(reflect.ValueOf(parsed))
					return
				}
				if parsed, err := time.Parse("2006-01-02", str); err == nil {
					field.Set(reflect.ValueOf(parsed))
					return
				}
			}
		}
	default:
		if valueValue.Type().AssignableTo(field.Type()) {
			field.Set(valueValue)
		}
	}
}
