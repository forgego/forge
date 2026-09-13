package filters

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeVariadicOrderByQueryset struct {
	recorded []string
}

func (f *fakeVariadicOrderByQueryset) OrderBy(fields ...string) interface{} {
	f.recorded = append(f.recorded, fields...)
	return f
}

type fakeAnyVariadicOrderByQueryset struct {
	recorded []any
}

func (f *fakeAnyVariadicOrderByQueryset) OrderBy(fields ...any) interface{} {
	f.recorded = append(f.recorded, fields...)
	return f
}

func TestOrderingFilter_VariadicOrderByRecording(t *testing.T) {
	filter := NewOrderingFilter([]string{"price", "name"})
	qs := &fakeVariadicOrderByQueryset{}

	req := httptest.NewRequest("GET", "/test/?ordering=-price,name", nil)
	result := filter.FilterQueryset(req, qs)

	assert.NotNil(t, result)
	assert.Equal(t, []string{"-price", "name"}, qs.recorded)
}

func TestOrderingFilter_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		allowedFields  []string
		queryURL       string
		expectedString []string
		expectedAny    []any
	}{
		{
			name:           "single ascending field",
			allowedFields:  []string{"name"},
			queryURL:       "/test/?ordering=name",
			expectedString: []string{"name"},
			expectedAny:    []any{"name"},
		},
		{
			name:           "single descending field",
			allowedFields:  []string{"price"},
			queryURL:       "/test/?ordering=-price",
			expectedString: []string{"-price"},
			expectedAny:    []any{"-price"},
		},
		{
			name:           "multiple fields with unallowed field filtered out",
			allowedFields:  []string{"price", "name"},
			queryURL:       "/test/?ordering=-price,secret,name",
			expectedString: []string{"-price", "name"},
			expectedAny:    []any{"-price", "name"},
		},
		{
			name:           "empty ordering param returns unchanged",
			allowedFields:  []string{"price", "name"},
			queryURL:       "/test/?ordering=",
			expectedString: nil,
			expectedAny:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" (string variadic)", func(t *testing.T) {
			filter := NewOrderingFilter(tt.allowedFields)
			qs := &fakeVariadicOrderByQueryset{}
			req := httptest.NewRequest("GET", tt.queryURL, nil)

			result := filter.FilterQueryset(req, qs)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedString, qs.recorded)
		})

		t.Run(tt.name+" (any variadic)", func(t *testing.T) {
			filter := NewOrderingFilter(tt.allowedFields)
			qs := &fakeAnyVariadicOrderByQueryset{}
			req := httptest.NewRequest("GET", tt.queryURL, nil)

			result := filter.FilterQueryset(req, qs)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedAny, qs.recorded)
		})
	}
}
