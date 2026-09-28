package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	forgehttp "github.com/forgego/forge/server"
	"go.uber.org/zap"
)

// resourceContextKey carries the resource name a viewset route was
// registered under, so request-time errors can name it.
type resourceContextKey struct{}

// withResource runs handler with the resource name in the request context.
func withResource(resource string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r.WithContext(context.WithValue(r.Context(), resourceContextKey{}, resource)))
	}
}

// resourceFromRequest returns the registered resource name, or "" when the
// viewset was invoked outside Router.RegisterRoutes.
func resourceFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	resource, _ := r.Context().Value(resourceContextKey{}).(string)
	return resource
}

type listMethodKey struct {
	t    reflect.Type
	name string
}

// listMethodErrors caches listMethodError results for the dynamic types List
// meets, so the per-request check is a map lookup.
var listMethodErrors sync.Map // listMethodKey -> error (nil stored as errNoListProblem)

var errNoListProblem = errors.New("")

// checkListCall reports whether List may call method name on a value of type
// t. Registration checks every statically known type; this covers results
// whose type is only known per request, such as a Filter declared to return
// interface{}.
func checkListCall(t reflect.Type, name string) error {
	key := listMethodKey{t: t, name: name}
	if cached, ok := listMethodErrors.Load(key); ok {
		if cached == errNoListProblem {
			return nil
		}
		return cached.(error)
	}
	err := listMethodError(t, name)
	if err == nil {
		listMethodErrors.Store(key, errNoListProblem)
		return nil
	}
	err = fmt.Errorf("list queryset value %s: %w", t, err)
	listMethodErrors.Store(key, err)
	return err
}

// errViewSetMisconfigured is what the client sees; the details are logged.
var errViewSetMisconfigured = errors.New("viewset is misconfigured")

// listConfigurationError answers a list request whose queryset chain produced
// a value List cannot use. It is a configuration error, not a client error:
// it logs the resource and the problem and responds 500.
func (vs *BaseViewSet) listConfigurationError(w http.ResponseWriter, r *http.Request, err error) {
	logger, ok := forgehttp.GetLogger(r).(interface {
		Error(string, ...zap.Field)
	})
	if !ok {
		logger = zap.L()
	}
	resource := resourceFromRequest(r)
	if resource == "" && vs.Model != nil {
		resource = reflect.TypeOf(vs.Model).String()
	}
	logger.Error("api: viewset configuration error",
		zap.String("resource", resource),
		zap.String("action", "list"),
		zap.String("problem", err.Error()),
	)
	vs.handleException(w, r, errViewSetMisconfigured)
}
