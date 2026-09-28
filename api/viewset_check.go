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
		return errors.New("serializer factory is nil")
	}
	defer func() {
		if recover() != nil {
			err = errors.New("serializer factory panicked")
		}
	}()
	serializer := factory()
	if serializer == nil || isNilReflectValue(reflect.ValueOf(serializer)) {
		return errors.New("serializer factory returned nil")
	}
	return nil
}

func checkModel(model interface{}) (reflect.Type, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}
	value := reflect.ValueOf(model)
	modelType := value.Type()
	if modelType.Kind() != reflect.Ptr || modelType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a pointer to a struct, got %s", modelType)
	}
	if value.IsNil() {
		return nil, fmt.Errorf("model is a nil %s", modelType)
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

// List-chain method requirements. Count and All are required on the types
// List calls them on; the chain methods are optional, but when present they
// must be callable the way List calls them.
const (
	countWant   = "Count(context.Context) (int64, error)"
	allWant     = "All(context.Context) ([]T, error)"
	filterWant  = "Filter(orm.Expression) T"
	orderByWant = "OrderBy(...string) T"
)

// listMethodError reports whether t's method name can be called the way List
// calls it. Count and All are required; Offset, Limit, Filter and OrderBy are
// optional and only checked when present.
func listMethodError(t reflect.Type, name string) error {
	in, out, variadic, ok := methodSignature(t, name)
	switch name {
	case "Count":
		if !ok {
			return fmt.Errorf("missing %s, needed by list", countWant)
		}
		if len(in) != 1 || variadic || !acceptsContext(in[0]) || len(out) != 2 ||
			(out[0] != int64Type && out[0].Kind() != reflect.Interface) || !isErrorResult(out[1]) {
			return methodMismatch("Count", countWant, "list", in, out, variadic)
		}
	case "All":
		if !ok {
			return fmt.Errorf("missing %s, needed by list", allWant)
		}
		if len(in) != 1 || variadic || !acceptsContext(in[0]) || len(out) != 2 ||
			(out[0].Kind() != reflect.Slice && out[0].Kind() != reflect.Interface) || !isErrorResult(out[1]) {
			return methodMismatch("All", allWant, "list", in, out, variadic)
		}
	case "Offset", "Limit":
		if ok && (len(in) != 1 || variadic || !intType.AssignableTo(in[0]) || len(out) == 0) {
			return methodMismatch(name, name+"(int) T", "list pagination", in, out, variadic)
		}
	case "Filter":
		if ok && (len(in) != 1 || variadic || in[0].Kind() != reflect.Interface ||
			!expressionInterfaceType.Implements(in[0]) || len(out) == 0) {
			return methodMismatch("Filter", filterWant, "list filtering", in, out, variadic)
		}
	case "OrderBy":
		if ok && (len(in) != 1 || !variadic || !stringType.AssignableTo(in[0].Elem()) || len(out) == 0) {
			return methodMismatch("OrderBy", orderByWant, "list ordering", in, out, variadic)
		}
	}
	return nil
}

// isDynamicType reports whether values of t only reveal their methods at run
// time: an interface type whose method set does not include the list methods.
func isDynamicType(t reflect.Type) bool {
	return t.Kind() == reflect.Interface
}

// listChainNode is a type List may call methods on, and how it was reached.
type listChainNode struct {
	t   reflect.Type
	via string // e.g. "QuerySet().Filter()"; empty for the queryset itself
}

func (n listChainNode) label() string {
	if n.via == "" {
		return ""
	}
	return fmt.Sprintf("%s result %s: ", n.via, n.t)
}

// chainStep returns the node a present, callable chain method name leads to.
// It reports false when the method is absent or invalid (reported elsewhere),
// or when its result type is only known per request.
func chainStep(n listChainNode, name string) (listChainNode, bool) {
	_, out, _, ok := methodSignature(n.t, name)
	if !ok || listMethodError(n.t, name) != nil || isDynamicType(out[0]) {
		return listChainNode{}, false
	}
	via := name + "()"
	if n.via != "" {
		via = n.via + "." + via
	}
	return listChainNode{t: out[0], via: via}, true
}

// checkListSource follows the calls List makes: Filter (repeatedly), then
// OrderBy, then Count, then Offset and Limit, then All. Each call may return a
// different type, so every type reachable with a static result type is
// checked for the methods List calls on it. A result declared as an interface
// is resolved per request, where a mismatch answers 500 (see
// BaseViewSet.listConfigurationError) instead of panicking.
func checkListSource(qsType reflect.Type) error {
	source := listSourceType(qsType)
	if isDynamicType(source) && source.NumMethod() == 0 {
		// An untyped QuerySet() result can only be checked per request.
		return nil
	}
	root := listChainNode{t: source}
	if source != qsType {
		root.via = "QuerySet()"
	}

	var problems []string
	reported := map[string]bool{}
	add := func(n listChainNode, err error) {
		if err == nil {
			return
		}
		message := n.label() + err.Error()
		if !reported[message] {
			reported[message] = true
			problems = append(problems, message)
		}
	}
	checkedChain := map[reflect.Type]bool{}
	checkChainMethods := func(n listChainNode) {
		if checkedChain[n.t] {
			return
		}
		checkedChain[n.t] = true
		for _, name := range []string{"Offset", "Limit", "Filter", "OrderBy"} {
			add(n, listMethodError(n.t, name))
		}
	}

	// Filter is applied once per query parameter, each time to the previous result.
	filtered := []listChainNode{root}
	seen := map[reflect.Type]bool{root.t: true}
	for i := 0; i < len(filtered); i++ {
		checkChainMethods(filtered[i])
		if next, ok := chainStep(filtered[i], "Filter"); ok && !seen[next.t] {
			seen[next.t] = true
			filtered = append(filtered, next)
		}
	}
	// OrderBy is applied at most once, after filtering.
	counted := append([]listChainNode(nil), filtered...)
	for _, n := range filtered {
		if next, ok := chainStep(n, "OrderBy"); ok && !seen[next.t] {
			seen[next.t] = true
			counted = append(counted, next)
		}
	}
	allChecked := map[reflect.Type]bool{}
	for _, n := range counted {
		checkChainMethods(n)
		add(n, listMethodError(n.t, "Count"))
		paged := n
		if _, _, _, ok := methodSignature(paged.t, "Offset"); ok {
			next, static := chainStep(paged, "Offset")
			if !static {
				continue
			}
			paged = next
			checkChainMethods(paged)
		}
		if _, _, _, ok := methodSignature(paged.t, "Limit"); ok {
			next, static := chainStep(paged, "Limit")
			if !static {
				continue
			}
			paged = next
			checkChainMethods(paged)
		}
		if !allChecked[paged.t] {
			allChecked[paged.t] = true
			add(paged, listMethodError(paged.t, "All"))
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
