package models

import (
	"testing"
	"time"
)

func TestUser_GetID(t *testing.T) {
	user := &User{ID: 123}
	if user.GetID() != 123 {
		t.Errorf("GetID() = %d, want 123", user.GetID())
	}
}

func TestUser_SetID(t *testing.T) {
	user := &User{}
	user.SetID(456)
	if user.ID != 456 {
		t.Errorf("SetID(456) did not set ID, got %d", user.ID)
	}
}

func TestUser_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		expected bool
	}{
		{"authenticated user", 1, true},
		{"unauthenticated user", 0, false},
		{"negative id", -1, false}, // negative IDs are not valid user IDs
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{ID: tt.id}
			if user.IsAuthenticated() != tt.expected {
				t.Errorf("IsAuthenticated() = %v, want %v", user.IsAuthenticated(), tt.expected)
			}
		})
	}
}

func TestUser_IsAnonymous(t *testing.T) {
	user := &User{ID: 1}
	if user.IsAnonymous() {
		t.Error("IsAnonymous() should always return false")
	}

	userNoID := &User{ID: 0}
	if userNoID.IsAnonymous() {
		t.Error("IsAnonymous() should always return false")
	}
}

func TestUser_GetFullName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		username  string
		expected  string
	}{
		{
			name:      "both names",
			firstName: "John",
			lastName:  "Doe",
			username:  "johndoe",
			expected:  "John Doe",
		},
		{
			name:      "first name only",
			firstName: "John",
			lastName:  "",
			username:  "johndoe",
			expected:  "John",
		},
		{
			name:      "last name only",
			firstName: "",
			lastName:  "Doe",
			username:  "johndoe",
			expected:  "Doe",
		},
		{
			name:      "no names",
			firstName: "",
			lastName:  "",
			username:  "johndoe",
			expected:  "johndoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				FirstName: tt.firstName,
				LastName:  tt.lastName,
				Username:  tt.username,
			}
			if result := user.GetFullName(); result != tt.expected {
				t.Errorf("GetFullName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestUser_GetDisplayName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		username  string
		expected  string
	}{
		{
			name:      "full name different from username",
			firstName: "John",
			lastName:  "Doe",
			username:  "johndoe",
			expected:  "John Doe",
		},
		{
			name:      "full name same as username",
			firstName: "johndoe",
			lastName:  "",
			username:  "johndoe",
			expected:  "johndoe",
		},
		{
			name:      "no names",
			firstName: "",
			lastName:  "",
			username:  "johndoe",
			expected:  "johndoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				FirstName: tt.firstName,
				LastName:  tt.lastName,
				Username:  tt.username,
			}
			if result := user.GetDisplayName(); result != tt.expected {
				t.Errorf("GetDisplayName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestUser_StatusFlags(t *testing.T) {
	user := &User{
		IsActive:      true,
		IsStaff:       true,
		IsSuperuser:   true,
		IsLocked:      false,
		EmailVerified: true,
	}

	if !user.IsActive {
		t.Error("IsActive should be true")
	}
	if !user.IsStaff {
		t.Error("IsStaff should be true")
	}
	if !user.IsSuperuser {
		t.Error("IsSuperuser should be true")
	}
	if user.IsLocked {
		t.Error("IsLocked should be false")
	}
	if !user.EmailVerified {
		t.Error("EmailVerified should be true")
	}
}

func TestUser_PasswordFields(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	user := &User{
		PasswordChangedAt:  &now,
		PasswordExpiresAt:  &future,
		MustChangePassword: true,
	}

	if user.PasswordChangedAt == nil {
		t.Error("PasswordChangedAt should not be nil")
	}
	if user.PasswordExpiresAt == nil {
		t.Error("PasswordExpiresAt should not be nil")
	}
	if !user.MustChangePassword {
		t.Error("MustChangePassword should be true")
	}
}

func TestUser_LockoutFields(t *testing.T) {
	now := time.Now()

	user := &User{
		LockedAt:            &now,
		LockedReason:        "Too many failed attempts",
		FailedLoginAttempts: 5,
		LastFailedLoginAt:   &now,
	}

	if user.LockedAt == nil {
		t.Error("LockedAt should not be nil")
	}
	if user.LockedReason != "Too many failed attempts" {
		t.Errorf("LockedReason = %q, want 'Too many failed attempts'", user.LockedReason)
	}
	if user.FailedLoginAttempts != 5 {
		t.Errorf("FailedLoginAttempts = %d, want 5", user.FailedLoginAttempts)
	}
}

func TestUser_EmailVerification(t *testing.T) {
	now := time.Now()

	user := &User{
		EmailVerified:    true,
		EmailVerifiedAt:  &now,
		PhoneVerified:    true,
	}

	if !user.EmailVerified {
		t.Error("EmailVerified should be true")
	}
	if user.EmailVerifiedAt == nil {
		t.Error("EmailVerifiedAt should not be nil")
	}
	if !user.PhoneVerified {
		t.Error("PhoneVerified should be true")
	}
}

func TestUser_Timestamps(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	user := &User{
		DateJoined: now,
		LastLogin:  &later,
		CreatedAt:  now,
		UpdatedAt:  later,
		DeletedAt:  nil,
	}

	if user.DateJoined.IsZero() {
		t.Error("DateJoined should not be zero")
	}
	if user.LastLogin == nil {
		t.Error("LastLogin should not be nil")
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
	if user.DeletedAt != nil {
		t.Error("DeletedAt should be nil for non-deleted user")
	}
}

func TestUser_SoftDelete(t *testing.T) {
	now := time.Now()
	user := &User{
		DeletedAt: &now,
	}

	if user.DeletedAt == nil {
		t.Error("DeletedAt should be set for soft-deleted user")
	}
}

func TestUser_PersonalInfo(t *testing.T) {
	user := &User{
		FirstName: "John",
		LastName:  "Doe",
		Bio:       "Software developer",
		Website:   "https://example.com",
		Location:  "New York",
		Avatar:    "https://example.com/avatar.jpg",
	}

	if user.FirstName != "John" {
		t.Errorf("FirstName = %q, want 'John'", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("LastName = %q, want 'Doe'", user.LastName)
	}
	if user.Bio != "Software developer" {
		t.Errorf("Bio = %q, want 'Software developer'", user.Bio)
	}
}

func TestUser_ContactInfo(t *testing.T) {
	user := &User{
		Email:         "john@example.com",
		PhoneNumber:   "+1234567890",
		PhoneVerified: true,
	}

	if user.Email != "john@example.com" {
		t.Errorf("Email = %q, want 'john@example.com'", user.Email)
	}
	if user.PhoneNumber != "+1234567890" {
		t.Errorf("PhoneNumber = %q, want '+1234567890'", user.PhoneNumber)
	}
}

func TestUser_Preferences(t *testing.T) {
	user := &User{
		Timezone: "America/New_York",
		Locale:   "en_US",
		Language: "en",
	}

	if user.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, want 'America/New_York'", user.Timezone)
	}
	if user.Locale != "en_US" {
		t.Errorf("Locale = %q, want 'en_US'", user.Locale)
	}
	if user.Language != "en" {
		t.Errorf("Language = %q, want 'en'", user.Language)
	}
}

func TestUser_AllFields(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:                  1,
		Username:            "johndoe",
		Email:               "john@example.com",
		Password:            "hashed_password",
		FirstName:           "John",
		LastName:            "Doe",
		Bio:                 "Developer",
		Website:             "https://example.com",
		Location:            "NYC",
		Avatar:              "avatar.jpg",
		PhoneNumber:         "+1234567890",
		PhoneVerified:       true,
		Timezone:            "UTC",
		Locale:              "en",
		Language:            "en",
		IsActive:            true,
		IsStaff:             false,
		IsSuperuser:         false,
		IsLocked:            false,
		EmailVerified:       true,
		PasswordChangedAt:   &now,
		PasswordExpiresAt:   nil,
		MustChangePassword:  false,
		LockedAt:            nil,
		LockedReason:        "",
		FailedLoginAttempts: 0,
		LastFailedLoginAt:   nil,
		EmailVerifiedAt:     &now,
		DateJoined:          now,
		LastLogin:           &now,
		CreatedAt:           now,
		UpdatedAt:           now,
		DeletedAt:           nil,
	}

	// Verify all fields are accessible
	if user.ID != 1 {
		t.Errorf("ID = %d, want 1", user.ID)
	}
	if user.Username != "johndoe" {
		t.Errorf("Username = %q, want 'johndoe'", user.Username)
	}
}
