package permissions

import (
	"net/http"
	"reflect"

	"github.com/forgego/forge/api/authentication"
)

// IsOwnerOrReadOnly allows read access to all users
// but only allows write access to the owner of the object
type IsOwnerOrReadOnly struct {
	// OwnerField is the field name on the object that contains the owner/user ID
	// Default: "user_id" or "owner_id"
	OwnerField string
	// UserIDField is the field name on the user that contains the user ID
	// Default: "id"
	UserIDField string
}

// NewIsOwnerOrReadOnly creates a new IsOwnerOrReadOnly permission
func NewIsOwnerOrReadOnly(ownerField string) *IsOwnerOrReadOnly {
	if ownerField == "" {
		ownerField = "user_id"
	}
	return &IsOwnerOrReadOnly{
		OwnerField:  ownerField,
		UserIDField: "id",
	}
}

// HasPermission checks if the request is a safe method or user is authenticated
func (p *IsOwnerOrReadOnly) HasPermission(r *http.Request, view ViewSet) bool {
	// Allow safe methods without authentication
	if IsSafeMethod(r.Method) {
		return true
	}
	// Require authentication for write operations
	_, ok := authentication.GetUserFromRequest(r)
	return ok
}

// HasObjectPermission checks if the user owns the object or request is a safe method
func (p *IsOwnerOrReadOnly) HasObjectPermission(r *http.Request, view ViewSet, obj interface{}) bool {
	// Allow safe methods
	if IsSafeMethod(r.Method) {
		return true
	}

	// Get authenticated user
	user, ok := authentication.GetUserFromRequest(r)
	if !ok {
		return false
	}

	// Get user ID
	userID := getUserID(user, p.UserIDField)
	if userID == nil {
		return false
	}

	// Get object owner ID
	ownerID := getOwnerID(obj, p.OwnerField)
	if ownerID == nil {
		return false
	}

	// Check if user ID matches owner ID
	return reflect.DeepEqual(userID, ownerID)
}

// GetMessage returns the error message
func (p *IsOwnerOrReadOnly) GetMessage() string {
	return "You do not have permission to perform this action"
}

// GetCode returns the error code
func (p *IsOwnerOrReadOnly) GetCode() string {
	return "permission_denied"
}

// getUserID gets the user ID from a user object
func getUserID(user interface{}, fieldName string) interface{} {
	if fieldName == "" {
		fieldName = "id"
	}

	// Try method first
	if idMethod := getMethod(user, "GetID"); idMethod.IsValid() {
		results := idMethod.Call(nil)
		if len(results) > 0 {
			return results[0].Interface()
		}
	}

	// Try field
	if idField := getField(user, fieldName); idField.IsValid() {
		return idField.Interface()
	}

	return nil
}

// getOwnerID gets the owner ID from an object
func getOwnerID(obj interface{}, fieldName string) interface{} {
	if fieldName == "" {
		// Try common field names
		for _, name := range []string{"user_id", "owner_id", "UserID", "OwnerID"} {
			if id := getOwnerID(obj, name); id != nil {
				return id
			}
		}
		return nil
	}

	// Try field
	if idField := getField(obj, fieldName); idField.IsValid() {
		return idField.Interface()
	}

	return nil
}
