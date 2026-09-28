package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/stretchr/testify/require"
)

func TestViewForRequest_CopiesAllExportedFields(t *testing.T) {
	src := &BaseViewSet{}
	sv := reflect.ValueOf(src).Elem()
	st := sv.Type()

	setFunc := func(fv reflect.Value) {
		fv.Set(reflect.MakeFunc(fv.Type(), func(args []reflect.Value) []reflect.Value {
			out := make([]reflect.Value, fv.Type().NumOut())
			for i := range out {
				out[i] = reflect.Zero(fv.Type().Out(i))
			}
			return out
		}))
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if !field.IsExported() {
			continue
		}
		fv := sv.Field(i)
		switch field.Name {
		case "Serializer":
			fv.Set(reflect.ValueOf(func() Serializer { return NewBaseSerializer(nil) }))
		case "Queryset":
			fv.Set(reflect.ValueOf(&dummyQueryset{}))
		case "Model":
			fv.Set(reflect.ValueOf(&map[string]interface{}{}))
		case "Authentication":
			fv.Set(reflect.ValueOf([]authentication.Authentication{nil}))
		case "Permissions":
			fv.Set(reflect.ValueOf([]permissions.Permission{listOnlyPermission{}}))
		case "Throttles":
			fv.Set(reflect.ValueOf([]throttling.Throttle{nil}))
		case "ErrorWriter":
			setFunc(fv)
		case "ExcludeResponseFields":
			fv.Set(reflect.ValueOf([]string{"password"}))
		case "ReadOnlyRequestFields":
			fv.Set(reflect.ValueOf([]string{"created_at"}))
		default:
			t.Fatalf("unhandled exported field %s — update viewForRequest to copy it", field.Name)
		}
	}

	req := withAction(httptest.NewRequest(http.MethodGet, "/api/items/", nil), "list")
	got := src.viewForRequest(req)
	require.NotSame(t, src, got, "viewForRequest must return a new viewset for requests with an action")
	require.Equal(t, "list", got.GetAction())

	gv := reflect.ValueOf(got).Elem()
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if !field.IsExported() {
			continue
		}
		srcField := sv.Field(i)
		gotField := gv.Field(i)
		if srcField.Kind() == reflect.Func {
			require.Equal(t, srcField.Pointer(), gotField.Pointer(), "field %s must be copied", field.Name)
			continue
		}
		require.Equal(t, srcField.Interface(), gotField.Interface(), "field %s must be copied", field.Name)
	}

	// A request without an action returns the shared viewset unchanged.
	plain := httptest.NewRequest(http.MethodGet, "/api/items/", nil)
	require.Same(t, src, src.viewForRequest(plain))
	require.Same(t, src, src.viewForRequest(nil))
}
