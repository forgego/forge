package api

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

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

// BaseViewSet provides common viewset functionality
type BaseViewSet struct {
	Serializer func() Serializer
	Queryset   interface{} // This would be a QuerySet in real implementation
	Model      interface{}
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
		if qsValue.MethodByName("Create").IsValid() {
			return qsValue
		}
	}
	
	// Fallback to finding manager from model instance
	// This creates a new instance of the model type to search for manager
	modelType := reflect.TypeOf(vs.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	instance := reflect.New(modelType).Interface()
	return getManagerFromModel(instance)
}

// List handles GET /resource/
func (vs *BaseViewSet) List(w http.ResponseWriter, r *http.Request) {
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
			// Check for error first
			if errVal := allResults[1]; !errVal.IsNil() {
				if err, ok := errVal.Interface().(error); ok {
					_ = forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
					return
				}
			}
			if allResults[0].CanInterface() {
				if objects, ok := allResults[0].Interface().([]interface{}); ok {
					results = objects
				} else if allResults[0].Kind() == reflect.Slice {
					// Try to convert slice of pointers
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
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = SendPaginatedResponse(w, r, serialized, int(totalCount), page, pageSize)
}

// Create handles POST /resource/
func (vs *BaseViewSet) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	serializer := vs.Serializer()
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

	createMethod := manager.MethodByName("Create")
	if !createMethod.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Create method not found")
		return
	}

	results := createMethod.Call([]reflect.Value{
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
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = forgehttp.SendJSON(w, http.StatusCreated, serialized)
}

// Retrieve handles GET /resource/{id}/
func (vs *BaseViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
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

	getMethod := manager.MethodByName("Get")
	if !getMethod.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Get method not found")
		return
	}

	results := getMethod.Call([]reflect.Value{
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
	serialized := SerializeModel(instance)
	// nolint:errcheck // HTTP response errors can't be handled meaningfully
	_ = forgehttp.SendJSON(w, http.StatusOK, serialized)
}

// Update handles PUT /resource/{id}/
func (vs *BaseViewSet) Update(w http.ResponseWriter, r *http.Request) {
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

	getMethod := manager.MethodByName("Get")
	if !getMethod.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Get method not found")
		return
	}

	getResults := getMethod.Call([]reflect.Value{
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

	// Populate from data
	populateFromMap(instance, data)

	// Update
	updateMethod := manager.MethodByName("Update")
	if !updateMethod.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Update method not found")
		return
	}

	updateResults := updateMethod.Call([]reflect.Value{
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
	forgehttp.SendJSON(w, http.StatusOK, serialized)
}

// PartialUpdate handles PATCH /resource/{id}/
func (vs *BaseViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	// Similar to Update but only updates provided fields
	vs.Update(w, r)
}

// Destroy handles DELETE /resource/{id}/
func (vs *BaseViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
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

	getMethod := manager.MethodByName("Get")
	if !getMethod.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Get method not found")
		return
	}

	getResults := getMethod.Call([]reflect.Value{
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

	// Delete
	deleteMethod := manager.MethodByName("Delete")
	if !deleteMethod.IsValid() {
		// nolint:errcheck // HTTP response errors can't be handled meaningfully
		_ = forgehttp.SendError(w, http.StatusInternalServerError, "Delete method not found")
		return
	}

	deleteResults := deleteMethod.Call([]reflect.Value{
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

// Router is a router for viewsets (DRF-like)
type Router struct {
	prefix string
	routes map[string]ViewSet
}

// NewRouter creates a new API router
func NewRouter(prefix string) *Router {
	return &Router{
		prefix: prefix,
		routes: make(map[string]ViewSet),
	}
}

// Register registers a viewset with a resource name
func (r *Router) Register(resource string, vs ViewSet) {
	r.routes[resource] = vs
}

// RegisterRoutes registers all routes on a chi router
func (r *Router) RegisterRoutes(router *forgehttp.Router) {
	for resource, vs := range r.routes {
		path := r.prefix + "/" + resource

		// List and Create
		router.Route(path, func(sub *forgehttp.Router) {
			sub.Get("/", vs.List)
			sub.Post("/", vs.Create)

			// Retrieve, Update, PartialUpdate, Destroy
			sub.Get("/{id}", vs.Retrieve)
			sub.Put("/{id}", vs.Update)
			sub.Patch("/{id}", vs.PartialUpdate)
			sub.Delete("/{id}", vs.Destroy)
		})
	}
}

// Helper functions for viewset operations

// applyFilters applies query parameter filters to queryset
func applyFilters(qs reflect.Value, r *http.Request) reflect.Value {
	// Get filter parameters from query string
	query := r.URL.Query()

	for key, values := range query {
		if len(values) == 0 || key == "page" || key == "page_size" || key == "ordering" || key == "search" {
			continue
		}

		filterMethod := qs.MethodByName("Filter")
		if filterMethod.IsValid() {
			value := values[0]

			expr := buildFilterExpr(key, value)
			if expr != nil {
				results := filterMethod.Call([]reflect.Value{reflect.ValueOf(expr)})
				if len(results) > 0 {
					if newQS, ok := results[0].Interface().(interface{}); ok {
						qs = reflect.ValueOf(newQS)
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

	orderByMethod := qs.MethodByName("OrderBy")
	if orderByMethod.IsValid() {
		rawFields := strings.Split(ordering, ",")
		args := make([]reflect.Value, 0, len(rawFields))
		for _, field := range rawFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			args = append(args, reflect.ValueOf(field))
		}
		if len(args) == 0 {
			return qs
		}
		results := orderByMethod.Call(args)
		if len(results) > 0 {
			if newQS, ok := results[0].Interface().(interface{}); ok {
				return reflect.ValueOf(newQS)
			}
		}
	}

	return qs
}

// getManagerFromModel gets the manager for a model using reflection
func getManagerFromModel(instance interface{}) reflect.Value {
	// This is a simplified approach - assumes model has package-level manager
	// Full implementation would use model registry
	modelType := reflect.TypeOf(instance)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Try to find package-level manager variable
	// This requires the model to be in a package with a manager variable
	// For MVP, return invalid value - full implementation needed
	return reflect.Value{}
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
