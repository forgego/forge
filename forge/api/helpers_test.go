package api

import (
	"testing"
)

func TestSetupCompleteAPI(t *testing.T) {
	orig := GetSettings()
	t.Cleanup(func() {
		SetSettings(orig)
	})

	SetupCompleteAPI()

	renderers := GetDefaultRenderers()
	if len(renderers) != 3 {
		t.Fatalf("expected 3 default renderers, got %d", len(renderers))
	}

	parsers := GetDefaultParsers()
	if len(parsers) != 3 {
		t.Fatalf("expected 3 default parsers, got %d", len(parsers))
	}

	// Verify getters for auth, permissions, and throttles do not panic and return slices
	_ = GetDefaultAuthentication()
	_ = GetDefaultPermissions()
	_ = GetDefaultThrottles()
}
