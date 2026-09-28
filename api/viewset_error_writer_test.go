package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseViewSet_UsesCustomErrorWriter(t *testing.T) {
	var customWriterCalled bool
	var capturedError error

	customWriter := func(w http.ResponseWriter, r *http.Request, err error) {
		customWriterCalled = true
		capturedError = err
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"title":"Custom Conflict","status":409}`))
	}

	viewset := NewBaseViewSet(
		newAccessCheckSerializer,
		&accessCheckQueryset{},
		&accessCheckModel{},
	)
	// Add a denying permission to trigger handleException
	viewset.Permissions = []permissions.Permission{denyingPermission{}}
	viewset.ErrorWriter = customWriter

	router := NewRouter("/api")
	router.Register("items", viewset)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/items/", nil)
	handler.ServeHTTP(rec, req)

	assert.True(t, customWriterCalled)
	require.NotNil(t, capturedError)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.JSONEq(t, `{"title":"Custom Conflict","status":409}`, rec.Body.String())
}

func TestViewSetConfig_PassesErrorWriterToBaseViewSet(t *testing.T) {
	customWriter := func(w http.ResponseWriter, r *http.Request, err error) {}

	config := &ViewSetConfig{
		Model:       &accessCheckModel{},
		Queryset:    &accessCheckQueryset{},
		Serializer:  newAccessCheckSerializer(),
		ErrorWriter: customWriter,
	}

	viewSet := NewConfigurableViewSet(config)
	assert.NotNil(t, viewSet.ErrorWriter)
}
