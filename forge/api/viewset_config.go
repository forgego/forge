package api

import (
	"net/http"
)

// ViewSetConfig represents configuration for a generic viewset
type ViewSetConfig struct {
	Model        interface{}
	Queryset     interface{} // Manager or QuerySet
	Serializer   Serializer
	ListFields   []string
	DetailFields []string
	Filterable   []string
	Searchable   []string
	Ordering     []string
	PerPage      int
	
	// Internal viewset created lazily
	viewSet *ConfigurableViewSet
}

// ConfigurableViewSet is a viewset created from configuration
type ConfigurableViewSet struct {
	*BaseViewSet
	config *ViewSetConfig
}

// NewConfigurableViewSet creates a new viewset from configuration
func NewConfigurableViewSet(config *ViewSetConfig) *ConfigurableViewSet {
	// Wrap the serializer instance in a factory function
	serializerFactory := func() Serializer {
		return config.Serializer.New()
	}
	
	// Create base viewset with nil queryset (will be handled dynamically)
	base := NewBaseViewSet(serializerFactory, config.Queryset, config.Model)
	
	return &ConfigurableViewSet{
		BaseViewSet: base,
		config:      config,
	}
}

// getViewSet returns the underlying viewset, creating it if necessary
func (c *ViewSetConfig) getViewSet() ViewSet {
	if c.viewSet == nil {
		c.viewSet = NewConfigurableViewSet(c)
	}
	return c.viewSet
}

// Implement ViewSet interface on ViewSetConfig

func (c *ViewSetConfig) List(w http.ResponseWriter, r *http.Request) {
	c.getViewSet().List(w, r)
}

func (c *ViewSetConfig) Create(w http.ResponseWriter, r *http.Request) {
	c.getViewSet().Create(w, r)
}

func (c *ViewSetConfig) Retrieve(w http.ResponseWriter, r *http.Request) {
	c.getViewSet().Retrieve(w, r)
}

func (c *ViewSetConfig) Update(w http.ResponseWriter, r *http.Request) {
	c.getViewSet().Update(w, r)
}

func (c *ViewSetConfig) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	c.getViewSet().PartialUpdate(w, r)
}

func (c *ViewSetConfig) Destroy(w http.ResponseWriter, r *http.Request) {
	c.getViewSet().Destroy(w, r)
}

// Ensure ViewSetConfig implements ViewSet
var _ ViewSet = (*ViewSetConfig)(nil)
