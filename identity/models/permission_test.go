package models

import (
	"testing"
)

func TestPermission_GetFullCodename(t *testing.T) {
	tests := []struct {
		name       string
		permission Permission
		expected   string
	}{
		{
			name:       "with app label",
			permission: Permission{AppLabel: "blog", Codename: "add_post"},
			expected:   "blog.add_post",
		},
		{
			name:       "without app label",
			permission: Permission{AppLabel: "", Codename: "view_user"},
			expected:   "view_user",
		},
		{
			name:       "empty codename",
			permission: Permission{AppLabel: "admin", Codename: ""},
			expected:   "admin.",
		},
		{
			name:       "both empty",
			permission: Permission{AppLabel: "", Codename: ""},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.permission.GetFullCodename()
			if result != tt.expected {
				t.Errorf("GetFullCodename() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPermission_Fields(t *testing.T) {
	perm := Permission{
		ID:          1,
		Name:        "Can add post",
		Codename:    "add_post",
		ContentType: "post",
		AppLabel:    "blog",
	}

	if perm.ID != 1 {
		t.Errorf("ID = %d, want 1", perm.ID)
	}
	if perm.Name != "Can add post" {
		t.Errorf("Name = %q, want 'Can add post'", perm.Name)
	}
	if perm.Codename != "add_post" {
		t.Errorf("Codename = %q, want 'add_post'", perm.Codename)
	}
	if perm.ContentType != "post" {
		t.Errorf("ContentType = %q, want 'post'", perm.ContentType)
	}
	if perm.AppLabel != "blog" {
		t.Errorf("AppLabel = %q, want 'blog'", perm.AppLabel)
	}
}

func TestPermission_CommonPermissions(t *testing.T) {
	// Test common permission patterns
	permissions := []struct {
		name        string
		codename    string
		contentType string
		appLabel    string
	}{
		{"Add", "add_user", "user", "auth"},
		{"Change", "change_user", "user", "auth"},
		{"Delete", "delete_user", "user", "auth"},
		{"View", "view_user", "user", "auth"},
	}

	for _, tt := range permissions {
		t.Run(tt.name, func(t *testing.T) {
			perm := Permission{
				Name:        tt.name,
				Codename:    tt.codename,
				ContentType: tt.contentType,
				AppLabel:    tt.appLabel,
			}
			fullName := perm.GetFullCodename()
			expected := tt.appLabel + "." + tt.codename
			if fullName != expected {
				t.Errorf("GetFullCodename() = %q, want %q", fullName, expected)
			}
		})
	}
}

func TestPermission_JSONTags(t *testing.T) {
	// Verify JSON tags are correctly set
	perm := Permission{
		ID:          1,
		Name:        "Test Permission",
		Codename:    "test_perm",
		ContentType: "test",
		AppLabel:    "test_app",
	}

	// These fields should be accessible
	if perm.ID == 0 {
		t.Error("ID should be set")
	}
	if perm.Name == "" {
		t.Error("Name should be set")
	}
	if perm.Codename == "" {
		t.Error("Codename should be set")
	}
	if perm.ContentType == "" {
		t.Error("ContentType should be set")
	}
	if perm.AppLabel == "" {
		t.Error("AppLabel should be set")
	}
}
