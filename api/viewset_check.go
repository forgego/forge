package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/orm"
)

// Configuration checks run when a viewset is registered and again when routes
// are mounted, so a data-access mistake stops the application at startup
// instead of surfacing as a 500 on the first request that reaches it.
//
// Messages name types and operations only. They never format configured
// values, which may hold connection strings or other secrets.

var (
	contextInterfaceType    = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterfaceType      = reflect.TypeOf((*error)(nil)).Elem()
	expressionInterfaceType = reflect.TypeOf((*orm.Expression)(nil)).Elem()
	int64Type               = reflect.TypeOf(int64(0))
	intType                 = reflect.TypeOf(0)
	stringType              = reflect.TypeOf("")
)

// readOnlyAllowedMethods lists the data methods a ReadOnly viewset answers.
var readOnlyAllowedMethods = []string{http.MethodGet}

// allowWrite rejects a write action on a ReadOnly viewset with 405.
func (vs *BaseViewSet) allowWrite(w http.ResponseWriter, r *http.Request) bool {
	if !vs.ReadOnly {
		return true
	}
	w.Header().Set("Allow", strings.Join(readOnlyAllowedMethods, ", "))
	vs.handleException(w, r, exceptions.NewMethodNotAllowed(readOnlyAllowedMethods))
	return false
}

// configurationChecker is implemented by viewsets that can verify their own
// configuration. BaseViewSet and ViewSetConfig implement it, and so does any
// viewset that embeds *BaseViewSet. A custom viewset can implement or
// override CheckConfiguration to describe its own requirements.
type configurationChecker interface {
	CheckConfiguration() error
}

// checkViewSet reports why vs cannot serve requests for resource. Viewsets
// that do not implement CheckConfiguration are only checked for nil.
func checkViewSet(resource string, vs ViewSet) error {
	if vs == nil {
		return fmt.Errorf("api: resource %q: viewset is nil", resource)
	}
	if value := reflect.ValueOf(vs); isNilReflectValue(value) {
		return fmt.Errorf("api: resource %q: viewset is a nil %s", resource, value.Type())
	}
	checker, ok := vs.(configurationChecker)
	if !ok {
		return nil
	}
	if err := checker.CheckConfiguration(); err != nil {
		return fmt.Errorf("api: resource %q: %w", resource, err)
	}
	return nil
}

// CheckConfiguration reports whether the viewset can serve every action it
// exposes: a serializer factory that returns a serializer, a non-nil pointer
// to a model struct, and a Queryset whose methods match the operations the
// viewset calls. A ReadOnly viewset needs only the read operations. The
// router calls it on Register and RegisterRoutes and panics on an error.
func (vs *BaseViewSet) CheckConfiguration() error {
	if vs == nil {
		return errors.New("BaseViewSet is nil")
	}
	var problems []string
	if err := checkSerializerFactory(vs.Serializer); err != nil {
		problems = append(problems, err.Error())
	}
	modelType, err := checkModel(vs.Model)
	if err != nil {
		problems = append(problems, err.Error())
	}
	problems = append(problems, checkQueryset(vs.Queryset, modelType, vs.ReadOnly)...)
	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "; "))
}

// CheckConfiguration reports whether the configuration can build a working
// viewset. It checks a fresh viewset, so fields changed after registration
// are still honored when the viewset is first built.
func (c *ViewSetConfig) CheckConfiguration() error {
	if c == nil {
		return errors.New("ViewSetConfig is nil")
	}
	if c.Serializer == nil || isNilReflectValue(reflect.ValueOf(c.Serializer)) {
		return errors.New("Serializer is nil")
	}
	return NewConfigurableViewSet(c).CheckConfiguration()
}

func checkSerializerFactory(factory func() Serializer) (err error) {
	if factory == nil {
		return errors.New("Serializer factory is nil")
	}
	defer func() {
		if recover() != nil {
			err = errors.New("Serializer factory panicked")
		}
	}()
	serializer := factory()
	if serializer == nil || isNilReflectValue(reflect.ValueOf(serializer)) {
		return errors.New("Serializer factory returned nil")
	}
	return nil
}

func checkModel(model interface{}) (reflect.Type, error) {
	if model == nil {
		return nil, errors.New("Model is nil")
	}
	value := reflect.ValueOf(model)
	modelType := value.Type()
	if modelType.Kind() != reflect.Ptr || modelType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("Model must be a pointer to a struct, got %s", modelType)
	}
	if value.IsNil() {
		return nil, fmt.Errorf("Model is a nil %s", modelType)
	}
	return modelType, nil
}

// checkQueryset mirrors the reflective calls BaseViewSet makes, so every
// signature accepted here can be called without a reflect panic.
func checkQueryset(queryset interface{}, modelType reflect.Type, readOnly bool) []string {
	if queryset == nil {
		return []string{"Queryset is nil"}
	}
	value := reflect.ValueOf(queryset)
	qsType := value.Type()
	if isNilReflectValue(value) {
		return []string{fmt.Sprintf("Queryset is a nil %s", qsType)}
	}

	var problems []string
	report := func(err error) {
		if err != nil {
			problems = append(problems, fmt.Sprintf("Queryset %s: %s", qsType, err))
		}
	}

	report(checkListSource(qsType))

	getOut, err := checkGet(qsType, readOnly)
	report(err)
	if readOnly {
		return problems
	}

	missingWrite := false
	for _, op := range []struct {
		name, actions string
		arg           reflect.Type
	}{
		{"Create", "create", modelType},
		{"Update", "update and partial_update", getOut},
		{"Delete", "destroy", getOut},
	} {
		if _, _, _, ok := methodSignature(qsType, op.name); !ok {
			missingWrite = true
		}
		report(checkWrite(qsType, op.name, op.actions, op.arg))
	}
	if missingWrite {
		problems = append(problems, "set ReadOnly to serve only list and retrieve")
	}
	return problems
}

// methodSignature returns the parameters and results of t's method name,
// without the receiver. Interface types list methods without a receiver.
func methodSignature(t reflect.Type, name string) (in, out []reflect.Type, variadic, ok bool) {
	method, found := t.MethodByName(name)
	if !found {
		return nil, nil, false, false
	}
	fn := method.Type
	start := 1
	if t.Kind() == reflect.Interface {
		start = 0
	}
	for i := start; i < fn.NumIn(); i++ {
		in = append(in, fn.In(i))
	}
	for i := 0; i < fn.NumOut(); i++ {
		out = append(out, fn.Out(i))
	}
	return in, out, fn.IsVariadic(), true
}

func formatSignature(name string, in, out []reflect.Type, variadic bool) string {
	params := make([]string, len(in))
	for i, p := range in {
		params[i] = p.String()
		if variadic && i == len(in)-1 {
			params[i] = "..." + p.Elem().String()
		}
	}
	results := make([]string, len(out))
	for i, r := range out {
		results[i] = r.String()
	}
	sig := name + "(" + strings.Join(params, ", ") + ")"
	switch len(results) {
	case 0:
		return sig
	case 1:
		return sig + " " + results[0]
	default:
		return sig + " (" + strings.Join(results, ", ") + ")"
	}
}

func isErrorResult(t reflect.Type) bool {
	return t.Kind() == reflect.Interface && t.Implements(errorInterfaceType)
}

func acceptsContext(t reflect.Type) bool {
	return t.Kind() == reflect.Interface && contextInterfaceType.Implements(t)
}

func methodMismatch(name, want, action string, in, out []reflect.Type, variadic bool) error {
	return fmt.Errorf("%s has signature %s; %s needs %s", name, formatSignature(name, in, out, variadic), action, want)
}

// listSourceType returns the type List calls Count and All on: the queryset
// itself, or the result of its QuerySet() constructor when it cannot paginate.
func listSourceType(qsType reflect.Type) reflect.Type {
	_, _, _, hasOffset := methodSignature(qsType, "Offset")
	_, _, _, hasLimit := methodSignature(qsType, "Limit")
	if hasOffset && hasLimit {
		return qsType
	}
	in, out, _, ok := methodSignature(qsType, "QuerySet")
	if !ok || len(in) != 0 || len(out) == 0 {
		return qsType
	}
	return out[0]
}

func checkListSource(qsType reflect.Type) error {
	source := listSourceType(qsType)
	if source.Kind() == reflect.Interface && source.NumMethod() == 0 {
		// An untyped QuerySet() result can only be checked per request.
		return nil
	}
	label := ""
	if source != qsType {
		label = fmt.Sprintf("QuerySet() result %s: ", source)
	}
	var problems []string
	add := func(err error) {
		if err != nil {
			problems = append(problems, label+err.Error())
		}
	}

	const countWant = "Count(context.Context) (int64, error)"
	if in, out, variadic, ok := methodSignature(source, "Count"); !ok {
		add(fmt.Errorf("missing %s, needed by list", countWant))
	} else if len(in) != 1 || variadic || !acceptsContext(in[0]) || len(out) != 2 ||
		(out[0] != int64Type && out[0].Kind() != reflect.Interface) || !isErrorResult(out[1]) {
		add(methodMismatch("Count", countWant, "list", in, out, variadic))
	}

	const allWant = "All(context.Context) ([]T, error)"
	if in, out, variadic, ok := methodSignature(source, "All"); !ok {
		add(fmt.Errorf("missing %s, needed by list", allWant))
	} else if len(in) != 1 || variadic || !acceptsContext(in[0]) || len(out) != 2 ||
		(out[0].Kind() != reflect.Slice && out[0].Kind() != reflect.Interface) || !isErrorResult(out[1]) {
		add(methodMismatch("All", allWant, "list", in, out, variadic))
	}

	// Optional chain methods: absent is fine, present must be callable.
	for _, name := range []string{"Offset", "Limit"} {
		if in, out, variadic, ok := methodSignature(source, name); ok {
			if len(in) != 1 || variadic || !intType.AssignableTo(in[0]) || len(out) == 0 {
				add(methodMismatch(name, name+"(int) T", "list pagination", in, out, variadic))
			}
		}
	}
	if in, out, variadic, ok := methodSignature(source, "Filter"); ok {
		if len(in) != 1 || variadic || in[0].Kind() != reflect.Interface ||
			!expressionInterfaceType.Implements(in[0]) || len(out) == 0 {
			add(methodMismatch("Filter", "Filter(orm.Expression) T", "list filtering", in, out, variadic))
		}
	}
	if in, out, variadic, ok := methodSignature(source, "OrderBy"); ok {
		if len(in) != 1 || !variadic || !stringType.AssignableTo(in[0].Elem()) || len(out) == 0 {
			add(methodMismatch("OrderBy", "OrderBy(...string) T", "list ordering", in, out, variadic))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "; "))
}

// checkGet validates Get and returns the type of the object it loads.
func checkGet(qsType reflect.Type, readOnly bool) (reflect.Type, error) {
	actions := "retrieve, update, partial_update and destroy"
	if readOnly {
		actions = "retrieve"
	}
	const want = "Get(context.Context, int64) (*Model, error)"
	in, out, variadic, ok := methodSignature(qsType, "Get")
	if !ok {
		return nil, fmt.Errorf("missing %s, needed by %s", want, actions)
	}
	if len(in) != 2 || variadic || !acceptsContext(in[0]) || !int64Type.AssignableTo(in[1]) ||
		len(out) != 2 || !isErrorResult(out[1]) {
		return nil, methodMismatch("Get", want, actions, in, out, variadic)
	}
	if !readOnly && out[0].Kind() != reflect.Ptr && out[0].Kind() != reflect.Interface {
		// A value result would be populated and saved as a copy.
		return nil, methodMismatch("Get", want, actions, in, out, variadic)
	}
	return out[0], nil
}

// checkWrite validates Create, Update or Delete, which BaseViewSet calls with
// a context and an object of type arg (nil when arg is unknown).
func checkWrite(qsType reflect.Type, name, actions string, arg reflect.Type) error {
	want := name + "(context.Context, *Model) error"
	in, out, variadic, ok := methodSignature(qsType, name)
	if !ok {
		return fmt.Errorf("missing %s, needed by %s", want, actions)
	}
	if len(in) != 2 || variadic || !acceptsContext(in[0]) || len(out) != 1 || !isErrorResult(out[0]) {
		return methodMismatch(name, want, actions, in, out, variadic)
	}
	if arg != nil && arg.Kind() != reflect.Interface && !arg.AssignableTo(in[1]) {
		return fmt.Errorf("%s takes %s but receives %s; %s needs %s", name, in[1], arg, actions, want)
	}
	return nil
}
