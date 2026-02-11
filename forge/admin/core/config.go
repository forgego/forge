package core

import (
	"context"

	"github.com/forgego/forge/orm"
)

// Field represents any item that can be displayed or filtered
// implementation normally provided by orm.Field[T] or admin.Method
type Field interface {
	Path() string
}

// Method represents a method on type T or a computed field
type Method string

func (m Method) Path() string { return string(m) }

// Computed creates a safe reference to a method or computed field
func Computed(name string) Method { return Method(name) }

// Config represents the configuration for an admin instance
type Config[T any] struct {
	// Basic information
	VerboseName       string
	VerboseNamePlural string
	Description       string
	Icon              string
	PageType          string // "list", "form", "list-form", "detail"
	HistoryManager    HistoryManager
	UIOverrides       map[string]string // Target (e.g. "list.actions") -> Component Name in React registry

	// List view configuration
	ListDisplay         []Field
	ListDisplayLinks    []Field
	ListEditable        []Field
	ListFilter          []Field
	SearchFields        []Field
	SearchHelpText      string
	DateHierarchy       string
	Ordering            []Field
	ListPerPage         int
	ListMaxShowAll      int
	ListSelectRelated   []string
	ListPrefetchRelated []string
	ShowFullResultCount bool
	PreserveFilters     bool
	EmptyValueDisplay   string
	SortableBy          []Field

	// Form configuration
	Fields                []string
	Exclude               []string
	ReadOnlyFields        []string
	PrepopulatedFields    map[string][]string
	RawIDFields           []string
	AutocompleteFields    []string
	RadioFields           map[string]RadioLayout
	Fieldsets             []Fieldset[T]
	InlineRelations       map[string]InlineRelationConfig
	GetFields             func(ctx context.Context, instance *T, isNew bool) []string
	GetFieldsets          func(ctx context.Context, instance *T, isNew bool) []Fieldset[T]
	GetReadOnlyFields     func(ctx context.Context, instance *T, isNew bool) []string
	GetPrepopulatedFields func(ctx context.Context, instance *T, isNew bool) map[string][]string

	// Actions
	Actions []Action[T]
	Filters []Filter[T]

	// Permissions
	PermissionChecker   PermissionChecker
	HasAddPermission    func(ctx context.Context, admin *Admin[T], user interface{}) bool
	HasChangePermission func(ctx context.Context, admin *Admin[T], user interface{}, obj *T) bool
	HasDeletePermission func(ctx context.Context, admin *Admin[T], user interface{}, obj *T) bool
	HasViewPermission   func(ctx context.Context, admin *Admin[T], user interface{}, obj *T) bool
	HasModulePermission func(ctx context.Context, admin *Admin[T], user interface{}) bool

	// View customization hooks
	GetQueryset func(ctx context.Context, admin *Admin[T], qs orm.QuerySet[T]) (orm.QuerySet[T], error)
	SaveModel   func(ctx context.Context, admin *Admin[T], instance *T, isNew bool) error
	DeleteModel func(ctx context.Context, admin *Admin[T], instance *T) error

	// Advanced features
	SaveAs         bool
	SaveAsContinue bool
	SaveOnTop      bool
	ViewOnSite     func(ctx context.Context, instance *T) string
	ShowChangeLink bool
}

// RadioLayout specifies the layout for radio buttons
type RadioLayout string

const (
	RadioHorizontal RadioLayout = "horizontal"
	RadioVertical   RadioLayout = "vertical"
)

// Fieldset represents a fieldset for grouping form fields
type Fieldset[T any] struct {
	Name        string
	Fields      []string
	Collapsed   bool
	Description string
}

// InlineConfig configures UI behavior for inline relations.
type InlineConfig struct {
	ListDisplay []string
	Fieldsets   any
}

// InlineRelationConfig configures inline relation behavior in forms.
type InlineRelationConfig struct {
	Type         string   `json:"type,omitempty"` // one_to_many, one_to_one, many_to_many
	Label        string   `json:"label,omitempty"`
	Fields       []string `json:"fields,omitempty"`
	RelatedModel string   `json:"related_model,omitempty"`
	RelatedField string   `json:"related_field,omitempty"`
	InlineConfig InlineConfig `json:"inline_config,omitempty"`
}

// NewFieldset creates a new fieldset
func NewFieldset[T any](name string, fields ...string) Fieldset[T] {
	return Fieldset[T]{
		Name:   name,
		Fields: fields,
	}
}

// WithCollapsed sets the fieldset as collapsed
func (f Fieldset[T]) WithCollapsed(collapsed bool) Fieldset[T] {
	f.Collapsed = collapsed
	return f
}

// WithDescription sets the fieldset description
func (f Fieldset[T]) WithDescription(desc string) Fieldset[T] {
	f.Description = desc
	return f
}

// Action represents a bulk action
type Action[T any] struct {
	Name         string
	Label        string
	Description  string
	Handler      func(ctx context.Context, instances []*T) error
	Permissions  []string
	Confirmation string
	Dangerous    bool
	Icon         string
	UIComponent  string
}

// NewAction creates a new action
func NewAction[T any](
	name string,
	label string,
	handler func(ctx context.Context, instances []*T) error,
) Action[T] {
	return Action[T]{
		Name:    name,
		Label:   label,
		Handler: handler,
	}
}

// WithDescription sets the action description
func (a Action[T]) WithDescription(desc string) Action[T] {
	a.Description = desc
	return a
}

// WithPermissions sets required permissions
func (a Action[T]) WithPermissions(perms ...string) Action[T] {
	a.Permissions = perms
	return a
}

// WithConfirmation sets the confirmation message
func (a Action[T]) WithConfirmation(msg string) Action[T] {
	a.Confirmation = msg
	return a
}

// WithDangerous marks the action as dangerous
func (a Action[T]) WithDangerous(dangerous bool) Action[T] {
	a.Dangerous = dangerous
	return a
}

// WithIcon sets the action icon
func (a Action[T]) WithIcon(icon string) Action[T] {
	a.Icon = icon
	return a
}

// WithUIComponent sets the action UI component
func (a Action[T]) WithUIComponent(component string) Action[T] {
	a.UIComponent = component
	return a
}

// Filter represents a manual filter definition
type Filter[T any] struct {
	Name        string
	Label       string
	Type        string // text, number, boolean, choice, date_range, relation
	Choices     []Choice
	Handler     func(ctx context.Context, qs orm.QuerySet[T], value interface{}) orm.QuerySet[T]
	UIComponent string
}
