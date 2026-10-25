package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	forgeerrors "github.com/forgego/forge/errors"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/require"
)

type configConcurrencyModel struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type configConcurrencySerializer struct {
	*BaseSerializer
}

func newConfigConcurrencySerializer() Serializer {
	return &configConcurrencySerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *configConcurrencySerializer) Fields() []string {
	return []string{"id", "name"}
}

func (s *configConcurrencySerializer) ReadOnlyFields() []string {
	return []string{"id"}
}

func (s *configConcurrencySerializer) New() Serializer {
	return newConfigConcurrencySerializer()
}

type configConcurrencyQueryset struct {
	unusedUpdateOperation
	unusedDeleteOperation
}

func (q *configConcurrencyQueryset) All(context.Context) ([]*configConcurrencyModel, error) {
	return []*configConcurrencyModel{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "second"},
	}, nil
}

func (q *configConcurrencyQueryset) Count(context.Context) (int64, error) {
	return 2, nil
}

func (q *configConcurrencyQueryset) Offset(int) *configConcurrencyQueryset {
	return q
}

func (q *configConcurrencyQueryset) Limit(int) *configConcurrencyQueryset {
	return q
}

func (q *configConcurrencyQueryset) Filter(interface{}) *configConcurrencyQueryset {
	return q
}

func (q *configConcurrencyQueryset) OrderBy(...string) *configConcurrencyQueryset {
	return q
}

func (q *configConcurrencyQueryset) Get(_ context.Context, id int64) (*configConcurrencyModel, error) {
	if id == 1 {
		return &configConcurrencyModel{ID: 1, Name: "first"}, nil
	}
	if id == 2 {
		return &configConcurrencyModel{ID: 2, Name: "second"}, nil
	}
	return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
}

func (q *configConcurrencyQueryset) Create(context.Context, *configConcurrencyModel) error {
	return nil
}

func (q *configConcurrencyQueryset) Update(context.Context, *configConcurrencyModel) error {
	return nil
}

func (q *configConcurrencyQueryset) Delete(context.Context, *configConcurrencyModel) error {
	return nil
}

func TestViewSetConfigConcurrency(t *testing.T) {
	config := &ViewSetConfig{
		Model:          &configConcurrencyModel{},
		Queryset:       &configConcurrencyQueryset{},
		Serializer:     newConfigConcurrencySerializer(),
		Filterable:     []string{"name"},
		Ordering:       []string{"id"},
		Authentication: []authentication.Authentication{},
		Permissions:    []permissions.Permission{},
		Throttles:      []throttling.Throttle{},
		ErrorWriter:    func(w http.ResponseWriter, r *http.Request, err error) {},
	}

	viewSet := NewConfigurableViewSet(config)
	require.NoError(t, viewSet.CheckConfiguration())

	router := NewRouter("/api")
	router.Register("items", viewSet)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	const (
		numRequestsPerEndpoint = 50
		numReaders             = 10
		numSetters             = 5
		iterations             = 50
	)

	var startWg sync.WaitGroup
	startWg.Add(1)

	var httpWg sync.WaitGroup
	var bgWg sync.WaitGroup

	// Slice to record list request status codes
	listStatuses := make([]int, numRequestsPerEndpoint)
	for i := 0; i < numRequestsPerEndpoint; i++ {
		httpWg.Add(1)
		go func(idx int) {
			defer httpWg.Done()
			startWg.Wait()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/items/", nil)
			handler.ServeHTTP(rec, req)
			listStatuses[idx] = rec.Code
		}(i)
	}

	// Slice to record retrieve request status codes
	retrieveStatuses := make([]int, numRequestsPerEndpoint)
	for i := 0; i < numRequestsPerEndpoint; i++ {
		httpWg.Add(1)
		go func(idx int) {
			defer httpWg.Done()
			startWg.Wait()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/items/1", nil)
			handler.ServeHTTP(rec, req)
			retrieveStatuses[idx] = rec.Code
		}(i)
	}

	// Goroutines reading configuration through getters and config accessors
	var stopReaders atomic.Bool
	for i := 0; i < numReaders; i++ {
		bgWg.Add(1)
		go func() {
			defer bgWg.Done()
			startWg.Wait()
			for !stopReaders.Load() {
				// Public getter
				_ = viewSet.GetAction()

				// Public configuration method
				_ = viewSet.CheckConfiguration()

				// Configuration field accessors
				_ = viewSet.Filterable
				_ = viewSet.ExcludeResponseFields
				_ = viewSet.ReadOnlyRequestFields
				_ = viewSet.OperationOnlyFields
				_ = viewSet.RejectUnknownRequestFields
				_ = viewSet.ReadOnly
				_ = viewSet.Authentication
				_ = viewSet.Permissions
				_ = viewSet.Throttles
				_ = viewSet.ErrorWriter
				_ = viewSet.Model
				_ = viewSet.Queryset
				if viewSet.Serializer != nil {
					_ = viewSet.Serializer()
				}

				// Config accessors on the underlying config
				_ = config.CheckConfiguration()
				_ = config.Model
				_ = config.Queryset
				_ = config.Serializer
				_ = config.ListFields
				_ = config.DetailFields
				_ = config.Filterable
				_ = config.Searchable
				_ = config.Ordering
				_ = config.PerPage
				_ = config.Authentication
				_ = config.Permissions
				_ = config.Throttles
				_ = config.ErrorWriter
				_ = config.ReadOnly
				_ = config.RejectUnknownRequestFields
				_ = config.OperationOnlyFields

				runtime.Gosched()
			}
		}()
	}

	// Goroutines calling public setters or configuration methods meant for use after construction
	var stopSetters atomic.Bool
	for i := 0; i < numSetters; i++ {
		bgWg.Add(1)
		go func() {
			defer bgWg.Done()
			startWg.Wait()
			for iter := 0; iter < iterations && !stopSetters.Load(); iter++ {
				switch iter % 3 {
				case 0:
					viewSet.SetAction("list")
				case 1:
					viewSet.SetAction("retrieve")
				case 2:
					viewSet.SetAction("custom")
				}
				_ = viewSet.CheckConfiguration()
				runtime.Gosched()
			}
		}()
	}

	// Release all goroutines simultaneously
	startWg.Done()

	// Wait for all HTTP requests to finish
	httpWg.Wait()

	// Signal background readers and setters to terminate and wait for them
	stopReaders.Store(true)
	stopSetters.Store(true)
	bgWg.Wait()

	// Assert every request returned the expected status
	for i, code := range listStatuses {
		require.Equal(t, http.StatusOK, code, "list request %d status code", i)
	}
	for i, code := range retrieveStatuses {
		require.Equal(t, http.StatusOK, code, "retrieve request %d status code", i)
	}
}
