package components

// ComponentType represents the type of UI component
type ComponentType string

const (
	TypePage      ComponentType = "page"
	TypeCard      ComponentType = "card"
	TypeGrid      ComponentType = "grid"
	TypeText      ComponentType = "text"
	TypeButton    ComponentType = "button"
	TypeTable     ComponentType = "table"
	TypeForm      ComponentType = "form"
	TypeChart     ComponentType = "chart"
	TypeStats     ComponentType = "stats"
	TypeContainer ComponentType = "container"
)

// Component represents a UI component in the Server-Driven UI system
type Component struct {
	Type       ComponentType          `json:"type"`
	ID         string                 `json:"id,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
	Children   []Component            `json:"children,omitempty"`
	Layout     *LayoutConfig          `json:"layout,omitempty"`
	Visibility *VisibilityConfig      `json:"visibility,omitempty"`
}

// LayoutConfig defines how the component should be positioned
type LayoutConfig struct {
	ColSpan int    `json:"col_span,omitempty"`   // Grid column span (1-12)
	RowSpan int    `json:"row_span,omitempty"`   // Grid row span
	Align   string `json:"align,omitempty"`      // start, center, end
	Justify string `json:"justify,omitempty"`    // start, center, end, between
	Padding string `json:"padding,omitempty"`    // p-4, px-2 etc (Tailwind classes)
	Margin  string `json:"margin,omitempty"`     // m-4, my-2 etc
	Class   string `json:"class_name,omitempty"` // Custom CSS classes
}

// VisibilityConfig defines when the component should be shown
type VisibilityConfig struct {
	Permissions []string `json:"permissions,omitempty"` // User must have these permissions
	FeatureFlag string   `json:"feature_flag,omitempty"`
	Condition   string   `json:"condition,omitempty"` // Simple expression evaluation
}

// NewComponent creates a new base component
func NewComponent(t ComponentType) Component {
	return Component{
		Type:  t,
		Props: make(map[string]interface{}),
	}
}

// WithID sets the component ID
func (c Component) WithID(id string) Component {
	c.ID = id
	return c
}

// WithProp sets a property
func (c Component) WithProp(key string, value interface{}) Component {
	c.Props[key] = value
	return c
}

// WithChildren adds children components
func (c Component) WithChildren(children ...Component) Component {
	c.Children = append(c.Children, children...)
	return c
}

// WithLayout sets the layout configuration
func (c Component) WithLayout(layout LayoutConfig) Component {
	c.Layout = &layout
	return c
}

// WithColSpan is a helper to set column span
func (c Component) WithColSpan(span int) Component {
	if c.Layout == nil {
		c.Layout = &LayoutConfig{}
	}
	c.Layout.ColSpan = span
	return c
}
