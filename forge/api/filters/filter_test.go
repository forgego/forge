package filters

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockQueryset for testing
type MockQueryset struct{}

func (m *MockQueryset) Search(query string, fields []string) interface{} {
	return m
}

func (m *MockQueryset) OrderBy(fields []string) interface{} {
	return m
}

func TestSearchFilter_FilterQueryset(t *testing.T) {
	filter := NewSearchFilter([]string{"name", "email"})
	queryset := &MockQueryset{}

	req := httptest.NewRequest("GET", "/test/?search=john", nil)
	result := filter.FilterQueryset(req, queryset)

	// Should return filtered queryset
	assert.NotNil(t, result)
}

func TestSearchFilter_NoSearchQuery(t *testing.T) {
	filter := NewSearchFilter([]string{"name", "email"})
	queryset := &MockQueryset{}

	req := httptest.NewRequest("GET", "/test/", nil)
	result := filter.FilterQueryset(req, queryset)

	// Should return original queryset when no search query
	assert.Equal(t, queryset, result)
}

func TestSearchFilter_GetSchema(t *testing.T) {
	filter := NewSearchFilter([]string{"name", "email"})
	req := httptest.NewRequest("GET", "/test/", nil)

	schema := filter.GetSchema(req, nil)

	assert.NotNil(t, schema)
	assert.Contains(t, schema, "search")
}

func TestOrderingFilter_FilterQueryset(t *testing.T) {
	filter := NewOrderingFilter([]string{"name", "created_at"})
	queryset := &MockQueryset{}

	req := httptest.NewRequest("GET", "/test/?ordering=name", nil)
	result := filter.FilterQueryset(req, queryset)

	assert.NotNil(t, result)
}

func TestOrderingFilter_Descending(t *testing.T) {
	filter := NewOrderingFilter([]string{"name", "created_at"})
	queryset := &MockQueryset{}

	req := httptest.NewRequest("GET", "/test/?ordering=-name", nil)
	result := filter.FilterQueryset(req, queryset)

	assert.NotNil(t, result)
}

func TestOrderingFilter_InvalidField(t *testing.T) {
	filter := NewOrderingFilter([]string{"name", "created_at"})
	queryset := &MockQueryset{}

	// Try to order by invalid field
	req := httptest.NewRequest("GET", "/test/?ordering=invalid_field", nil)
	result := filter.FilterQueryset(req, queryset)

	// Should return original queryset
	assert.Equal(t, queryset, result)
}

func TestOrderingFilter_MultipleFields(t *testing.T) {
	filter := NewOrderingFilter([]string{"name", "created_at"})
	queryset := &MockQueryset{}

	req := httptest.NewRequest("GET", "/test/?ordering=name,-created_at", nil)
	result := filter.FilterQueryset(req, queryset)

	assert.NotNil(t, result)
}

func TestOrderingFilter_GetSchema(t *testing.T) {
	filter := NewOrderingFilter([]string{"name", "created_at"})
	req := httptest.NewRequest("GET", "/test/", nil)

	schema := filter.GetSchema(req, nil)

	assert.NotNil(t, schema)
	assert.Contains(t, schema, "ordering")
}

func TestFilterBackendList_ApplyFilters(t *testing.T) {
	queryset := &MockQueryset{}

	backends := FilterBackendList{
		NewSearchFilter([]string{"name"}),
		NewOrderingFilter([]string{"name"}),
	}

	req := httptest.NewRequest("GET", "/test/?search=john&ordering=name", nil)
	result := backends.ApplyFilters(req, queryset)

	assert.NotNil(t, result)
}
