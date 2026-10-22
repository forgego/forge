package errors

import (
	"strings"
	"testing"
)

func TestSanitizedErrorsHideStackTracePaths(t *testing.T) {
	for _, hide := range []bool{false, true} {
		sanitizer := NewSanitizer(SanitizerConfig{HideStackTraces: hide})
		got := sanitizer.SanitizeString("failure at /srv/app/handler.go:42")
		if strings.Contains(got, "handler.go:42") {
			t.Fatalf("hide=%v leaked stack path: %q", hide, got)
		}
	}
}
