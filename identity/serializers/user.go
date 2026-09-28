package serializers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/identity/models"
)

// UserSerializer serializes User model for API
type UserSerializer struct {
	*api.BaseSerializer
	user *models.User
}

// NewUserSerializer creates a new user serializer
func NewUserSerializer() *UserSerializer {
	return &UserSerializer{
		BaseSerializer: api.NewBaseSerializer(make(map[string]interface{})),
	}
}

// FromUser creates a serializer from a user model
func FromUser(user *models.User) *UserSerializer {
	data := map[string]interface{}{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"is_active":      user.IsActive,
		"is_staff":       user.IsStaff,
		"is_superuser":   user.IsSuperuser,
		"is_locked":      user.IsLocked,
		"email_verified": user.EmailVerified,
		"phone_number":   user.PhoneNumber,
		"phone_verified": user.PhoneVerified,
		"timezone":       user.Timezone,
		"locale":         user.Locale,
		"language":       user.Language,
		"bio":            user.Bio,
		"website":        user.Website,
		"location":       user.Location,
		"avatar":         user.Avatar,
		"date_joined":    user.DateJoined.Format(time.RFC3339),
		"created_at":     user.CreatedAt.Format(time.RFC3339),
		"updated_at":     user.UpdatedAt.Format(time.RFC3339),
	}

	if user.LastLogin != nil {
		data["last_login"] = user.LastLogin.Format(time.RFC3339)
	}
	if user.EmailVerifiedAt != nil {
		data["email_verified_at"] = user.EmailVerifiedAt.Format(time.RFC3339)
	}

	return &UserSerializer{
		BaseSerializer: api.NewBaseSerializer(data),
		user:           user,
	}
}

// Validate validates the serializer data
func (s *UserSerializer) Validate() error {
	// Username validation
	username := s.GetString("username")
	if username == "" {
		s.AddError("username", "username is required")
	} else if len(username) > 150 {
		s.AddError("username", "username must be 150 characters or less")
	}

	// Email validation
	email := s.GetString("email")
	if email == "" {
		s.AddError("email", "email is required")
	} else if len(email) > 254 {
		s.AddError("email", "email must be 254 characters or less")
	} else if !isValidEmail(email) {
		s.AddError("email", "invalid email format")
	}

	// First name validation
	firstName := s.GetString("first_name")
	if firstName != "" && len(firstName) > 150 {
		s.AddError("first_name", "first name must be 150 characters or less")
	}

	// Last name validation
	lastName := s.GetString("last_name")
	if lastName != "" && len(lastName) > 150 {
		s.AddError("last_name", "last name must be 150 characters or less")
	}

	if !s.IsValid() {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// GetUsername returns the username
func (s *UserSerializer) GetUsername() string {
	return s.GetString("username")
}

// GetEmail returns the email
func (s *UserSerializer) GetEmail() string {
	return s.GetString("email")
}

// GetPassword returns the password (for create/update)
func (s *UserSerializer) GetPassword() string {
	return s.GetString("password")
}

// GetFirstName returns the first name
func (s *UserSerializer) GetFirstName() string {
	return s.GetString("first_name")
}

// GetLastName returns the last name
func (s *UserSerializer) GetLastName() string {
	return s.GetString("last_name")
}

// ToUser converts serializer to User model (for create/update)
func (s *UserSerializer) ToUser() *models.User {
	user := &models.User{
		Username:  s.GetUsername(),
		Email:     s.GetEmail(),
		FirstName: s.GetFirstName(),
		LastName:  s.GetLastName(),
		IsActive:  s.GetBool("is_active"),
		IsStaff:   s.GetBool("is_staff"),
	}

	return user
}

// ToJSONMap converts serializer to map for JSON response
func (s *UserSerializer) ToJSONMap() map[string]interface{} {
	// Serialize to JSON and unmarshal to get map
	jsonBytes, err := s.ToJSON()
	if err != nil {
		return make(map[string]interface{})
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return make(map[string]interface{})
	}

	return result
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	atIndex := -1
	dotIndex := -1

	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Multiple @
			}
			atIndex = i
		} else if char == '.' {
			dotIndex = i
		}
	}

	return atIndex > 0 && dotIndex > atIndex && dotIndex < len(email)-1
}
