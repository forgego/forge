package core

import (
	"time"

	"github.com/forgego/forge/admin/components"
)

// Metadata represents the complete metadata for an admin model
// This is what gets serialized to JSON for the frontend
type Metadata struct {
	Name              string                `json:"name"`
	VerboseName       string                `json:"verbose_name"`
	VerboseNamePlural string                `json:"verbose_name_plural"`
	Description       string                `json:"description,omitempty"`
	Icon              string                `json:"icon,omitempty"`
	Fields            []FieldMetadata       `json:"fields"`
	Relations         []RelationMetadata    `json:"relations,omitempty"`
	Permissions       PermissionMetadata    `json:"permissions"`
	Actions           []ActionMetadata      `json:"actions"`
	Filters           []FilterMetadata      `json:"filters"`
	ListDisplay       []string              `json:"list_display"`
	ListFilter        []string              `json:"list_filter,omitempty"`
	SearchFields      []string              `json:"search_fields,omitempty"`
	Ordering          []string              `json:"ordering,omitempty"`
	Pagination        PaginationConfig      `json:"pagination"`
	PageType          string                `json:"page_type,omitempty"` // "list", "form", "list-form", "detail"
	UIOverrides       map[string]string     `json:"ui_overrides,omitempty"`
	Layout            *components.Component `json:"layout,omitempty"` // Server-Driven UI Layout
}

// FieldMetadata represents metadata for a single field
type FieldMetadata struct {
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	Label        string              `json:"label"`
	HelpText     string              `json:"help_text,omitempty"`
	Required     bool                `json:"required"`
	ReadOnly     bool                `json:"read_only"`
	Choices      []Choice            `json:"choices,omitempty"`
	Widget       string              `json:"widget"`
	Validators   []ValidatorMetadata `json:"validators,omitempty"`
	DefaultValue interface{}         `json:"default_value,omitempty"`
	MaxLength    int                 `json:"max_length,omitempty"`
	MinLength    int                 `json:"min_length,omitempty"`
	MaxValue     interface{}         `json:"max_value,omitempty"`
	MinValue     interface{}         `json:"min_value,omitempty"`
	Accept       string              `json:"accept,omitempty"`       // For file uploads
	MaxSize      int64               `json:"max_size,omitempty"`     // For file uploads
	Multiple     bool                `json:"multiple,omitempty"`     // For file uploads
	AllowCreate  bool                `json:"allow_create,omitempty"` // For relations
}

// Choice represents a choice for a field
type Choice struct {
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

// ValidatorMetadata represents a validator
type ValidatorMetadata struct {
	Name    string                 `json:"name"`
	Message string                 `json:"message,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// RelationMetadata represents metadata for a relation
type RelationMetadata struct {
	Name         string                `json:"name"`
	Type         string                `json:"type"` // foreign_key, one_to_one, many_to_many
	RelatedModel string                `json:"related_model"`
	RelatedField string                `json:"related_field,omitempty"`
	Label        string                `json:"label"`
	Inline       *InlineRelationConfig `json:"inline,omitempty"`
}

// PermissionMetadata represents permissions for this model
type PermissionMetadata struct {
	Add    bool `json:"add"`
	Change bool `json:"change"`
	Delete bool `json:"delete"`
	View   bool `json:"view"`
}

// ActionMetadata represents metadata for a bulk action
type ActionMetadata struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Description  string   `json:"description,omitempty"`
	Confirmation string   `json:"confirmation,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	Dangerous    bool     `json:"dangerous,omitempty"`
	Icon         string   `json:"icon,omitempty"`
	UIComponent  string   `json:"ui_component,omitempty"`
}

// FilterMetadata represents metadata for a filter
type FilterMetadata struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // text, number, boolean, choice, date_range, relation
	Label        string                 `json:"label"`
	Choices      []Choice               `json:"choices,omitempty"`
	RelatedModel string                 `json:"related_model,omitempty"`
	Multiple     bool                   `json:"multiple,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
	UIComponent  string                 `json:"ui_component,omitempty"`
}

// PaginationConfig represents pagination configuration
type PaginationConfig struct {
	PageSize    int `json:"page_size"`
	MaxPageSize int `json:"max_page_size"`
}

// ModelListMetadata represents metadata for the model list endpoint
type ModelListMetadata struct {
	Name              string             `json:"name"`
	VerboseName       string             `json:"verbose_name"`
	VerboseNamePlural string             `json:"verbose_name_plural"`
	Icon              string             `json:"icon,omitempty"`
	Count             int64              `json:"count"`
	Permissions       PermissionMetadata `json:"permissions"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Count      int64       `json:"count"`
	Next       string      `json:"next,omitempty"`
	Previous   string      `json:"previous,omitempty"`
	PageSize   int         `json:"page_size"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
	Results    interface{} `json:"results"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// BulkActionRequest represents a bulk action request
type BulkActionRequest struct {
	IDs    []interface{}          `json:"ids"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// BulkActionResponse represents a bulk action response
type BulkActionResponse struct {
	Success  bool              `json:"success"`
	Affected int               `json:"affected"`
	Message  string            `json:"message"`
	Errors   []BulkActionError `json:"errors,omitempty"`
}

// BulkActionError represents an error for a specific item in bulk action
type BulkActionError struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// SearchRequest represents a global search request
type SearchRequest struct {
	Query  string   `json:"query"`
	Models []string `json:"models,omitempty"`
}

// SearchResponse represents a global search response
type SearchResponse struct {
	Results []SearchResultGroup `json:"results"`
}

// SearchResultGroup represents search results for a model
type SearchResultGroup struct {
	Model string             `json:"model"`
	Count int                `json:"count"`
	Items []SearchResultItem `json:"items"`
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	ID        interface{} `json:"id"`
	Title     string      `json:"title"`
	Highlight string      `json:"highlight,omitempty"`
	URL       string      `json:"url"`
}

// AutocompleteRequest represents an autocomplete request
type AutocompleteRequest struct {
	Field string `json:"field"`
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// AutocompleteResponse represents an autocomplete response
type AutocompleteResponse struct {
	Results []AutocompleteItem `json:"results"`
}

// AutocompleteItem represents an autocomplete item
type AutocompleteItem struct {
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

// UploadResponse represents a file upload response
type UploadResponse struct {
	URL        string    `json:"url"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	Width      int       `json:"width,omitempty"`
	Height     int       `json:"height,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
