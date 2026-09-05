package core

import (
	"context"
	"reflect"

	"github.com/forgego/forge/db"
)

// AdminInterface is the interface that all admin instances must implement
// This allows type-agnostic access to admin functionality
type AdminInterface interface {
	// Initialization
	SetDB(database *db.DB)

	// Metadata
	ModelName() string
	ModelType() reflect.Type
	GetMetadata(ctx context.Context, user interface{}) (*Metadata, error)

	// Permissions
	HasAddPermission(ctx context.Context, user interface{}) bool
	HasChangePermission(ctx context.Context, user interface{}, obj interface{}) bool
	HasDeletePermission(ctx context.Context, user interface{}, obj interface{}) bool
	HasViewPermission(ctx context.Context, user interface{}, obj interface{}) bool
	HasModulePermission(ctx context.Context, user interface{}) bool

	// Type conversion helpers for dynamic access
	ManagerInterface() interface{}
	ConfigInterface() interface{}

	// History
	GetHistory(ctx context.Context, objectID string) ([]LogEntry, error)
	LogAction(ctx context.Context, user interface{}, objectID string, repr string, action ActionType, changes string) error

	// Layout/UI hints
	PageType() string // "list", "form", "list-form", "detail"

	// Data operations
	ListObjects(ctx context.Context, params ListParams) (*PaginatedResponse, error)
	GetObject(ctx context.Context, id interface{}) (interface{}, error)
	CreateObject(ctx context.Context, data map[string]interface{}) (interface{}, error)
	UpdateObject(ctx context.Context, id interface{}, data map[string]interface{}) (interface{}, error)
	DeleteObject(ctx context.Context, id interface{}) error
	ExecuteAction(ctx context.Context, actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error)
	Autocomplete(ctx context.Context, query string, limit int) ([]AutocompleteItem, error)
}

// ListParams represents query parameters for list view
type ListParams struct {
	Page     int
	PageSize int
	Search   string
	Ordering []string
	Filters  map[string]interface{}
}

// PermissionChecker is an interface for checking permissions
type PermissionChecker interface {
	HasPermission(ctx context.Context, user interface{}, permission string) bool
	HasObjectPermission(ctx context.Context, user interface{}, obj interface{}, permission string) bool
}

// PermissionType represents a permission type
type PermissionType string

const (
	PermView   PermissionType = "view"
	PermAdd    PermissionType = "add"
	PermChange PermissionType = "change"
	PermDelete PermissionType = "delete"
)

// GetPermissionName returns the full permission name for a model and action
func GetPermissionName(modelName string, perm PermissionType) string {
	return string(perm) + "_" + modelName
}

