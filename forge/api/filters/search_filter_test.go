package filters

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeSearchFilterQueryset struct {
	calls int
	arg   any
}

func (f *fakeSearchFilterQueryset) Filter(expr any) interface{} {
	f.calls++
	f.arg = expr
	return f
}

func TestSearchFilter_CallsFilterWithExpression(t *testing.T) {
	filter := NewSearchFilter([]string{"name", "email"})

	t.Run("search query calls Filter once with non-nil expression", func(t *testing.T) {
		qs := &fakeSearchFilterQueryset{}
		req := httptest.NewRequest("GET", "/test/?search=abc", nil)
		result := filter.FilterQueryset(req, qs)

		assert.NotNil(t, result)
		assert.Equal(t, 1, qs.calls)
		assert.NotNil(t, qs.arg)
	})

	t.Run("empty search does not call Filter", func(t *testing.T) {
		qs := &fakeSearchFilterQueryset{}
		req := httptest.NewRequest("GET", "/test/?search=", nil)
		result := filter.FilterQueryset(req, qs)

		assert.NotNil(t, result)
		assert.Equal(t, 0, qs.calls)
		assert.Nil(t, qs.arg)
	})
}

func TestSearchFilter_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		searchFields  []string
		queryURL      string
		expectedCalls int
		expectArgNil  bool
	}{
		{
			name:          "single search field with query",
			searchFields:  []string{"title"},
			queryURL:      "/test/?search=golang",
			expectedCalls: 1,
			expectArgNil:  false,
		},
		{
			name:          "multiple search fields with query",
			searchFields:  []string{"title", "body", "author"},
			queryURL:      "/test/?search=framework",
			expectedCalls: 1,
			expectArgNil:  false,
		},
		{
			name:          "empty search fields list",
			searchFields:  []string{},
			queryURL:      "/test/?search=test",
			expectedCalls: 0,
			expectArgNil:  true,
		},
		{
			name:          "missing search parameter",
			searchFields:  []string{"title"},
			queryURL:      "/test/",
			expectedCalls: 0,
			expectArgNil:  true,
		},
		{
			name:          "search fields with whitespace only",
			searchFields:  []string{"   ", ""},
			queryURL:      "/test/?search=query",
			expectedCalls: 0,
			expectArgNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewSearchFilter(tt.searchFields)
			qs := &fakeSearchFilterQueryset{}
			req := httptest.NewRequest("GET", tt.queryURL, nil)

			result := filter.FilterQueryset(req, qs)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCalls, qs.calls)
			if tt.expectArgNil {
				assert.Nil(t, qs.arg)
			} else {
				assert.NotNil(t, qs.arg)
			}
		})
	}
}
