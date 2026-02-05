package serializers

import (
	"fmt"

	"github.com/forgego/forge/api"
)

// RegisterSerializer serializes user registration requests
type RegisterSerializer struct {
	*api.BaseSerializer
}

// NewRegisterSerializer creates a new registration serializer
func NewRegisterSerializer(data map[string]interface{}) *RegisterSerializer {
	return &RegisterSerializer{
		BaseSerializer: api.NewBaseSerializer(data),
	}
}

// Validate validates registration data
func (s *RegisterSerializer) Validate() error {
	username := s.GetString("username")
	if username == "" {
		s.AddError("username", "username is required")
	} else if len(username) < 3 {
		s.AddError("username", "username must be at least 3 characters")
	} else if len(username) > 150 {
		s.AddError("username", "username must be 150 characters or less")
	}

	email := s.GetString("email")
	if email == "" {
		s.AddError("email", "email is required")
	} else if !isValidEmailAuth(email) {
		s.AddError("email", "invalid email format")
	}

	password := s.GetString("password")
	if password == "" {
		s.AddError("password", "password is required")
	} else if len(password) < 8 {
		s.AddError("password", "password must be at least 8 characters")
	}

	if !s.IsValid() {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// GetUsername returns the username
func (s *RegisterSerializer) GetUsername() string {
	return s.GetString("username")
}

// GetEmail returns the email
func (s *RegisterSerializer) GetEmail() string {
	return s.GetString("email")
}

// GetPassword returns the password
func (s *RegisterSerializer) GetPassword() string {
	return s.GetString("password")
}

// LoginSerializer serializes login requests
type LoginSerializer struct {
	*api.BaseSerializer
}

// NewLoginSerializer creates a new login serializer
func NewLoginSerializer(data map[string]interface{}) *LoginSerializer {
	return &LoginSerializer{
		BaseSerializer: api.NewBaseSerializer(data),
	}
}

// Validate validates login data
func (s *LoginSerializer) Validate() error {
	usernameOrEmail := s.GetString("username")
	if usernameOrEmail == "" {
		usernameOrEmail = s.GetString("email")
	}
	if usernameOrEmail == "" {
		s.AddError("username", "username or email is required")
	}

	password := s.GetString("password")
	if password == "" {
		s.AddError("password", "password is required")
	}

	if !s.IsValid() {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// GetUsernameOrEmail returns username or email
func (s *LoginSerializer) GetUsernameOrEmail() string {
	username := s.GetString("username")
	if username != "" {
		return username
	}
	return s.GetString("email")
}

// GetPassword returns the password
func (s *LoginSerializer) GetPassword() string {
	return s.GetString("password")
}

// GetRememberMe returns remember me flag
func (s *LoginSerializer) GetRememberMe() bool {
	return s.GetBool("remember_me")
}

// PasswordResetSerializer serializes password reset requests
type PasswordResetSerializer struct {
	*api.BaseSerializer
}

// NewPasswordResetSerializer creates a new password reset serializer
func NewPasswordResetSerializer(data map[string]interface{}) *PasswordResetSerializer {
	return &PasswordResetSerializer{
		BaseSerializer: api.NewBaseSerializer(data),
	}
}

// Validate validates password reset data
func (s *PasswordResetSerializer) Validate() error {
	email := s.GetString("email")
	if email == "" {
		s.AddError("email", "email is required")
	} else if !isValidEmailAuth(email) {
		s.AddError("email", "invalid email format")
	}

	if !s.IsValid() {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// GetEmail returns the email
func (s *PasswordResetSerializer) GetEmail() string {
	return s.GetString("email")
}

// PasswordChangeSerializer serializes password change requests
type PasswordChangeSerializer struct {
	*api.BaseSerializer
}

// NewPasswordChangeSerializer creates a new password change serializer
func NewPasswordChangeSerializer(data map[string]interface{}) *PasswordChangeSerializer {
	return &PasswordChangeSerializer{
		BaseSerializer: api.NewBaseSerializer(data),
	}
}

// Validate validates password change data
func (s *PasswordChangeSerializer) Validate() error {
	currentPassword := s.GetString("current_password")
	if currentPassword == "" {
		s.AddError("current_password", "current password is required")
	}

	newPassword := s.GetString("new_password")
	if newPassword == "" {
		s.AddError("new_password", "new password is required")
	} else if len(newPassword) < 8 {
		s.AddError("new_password", "new password must be at least 8 characters")
	}

	if !s.IsValid() {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// GetCurrentPassword returns the current password
func (s *PasswordChangeSerializer) GetCurrentPassword() string {
	return s.GetString("current_password")
}

// GetNewPassword returns the new password
func (s *PasswordChangeSerializer) GetNewPassword() string {
	return s.GetString("new_password")
}

// isValidEmailAuth performs basic email validation (for auth serializers)
func isValidEmailAuth(email string) bool {
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
