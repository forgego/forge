package server

import (
	"strings"
	"testing"
)

func TestXSSSanitizeWrappers(t *testing.T) {
	xss := NewXSS()
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "strict HTML escaping", got: xss.SanitizeHTMLStrict(`<img onerror=alert(1) src=x>`)},
		{name: "null byte removal and HTML escaping", got: xss.SanitizeInput("a\x00<b>"), want: "a&lt;b&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "strict HTML escaping" {
				if strings.Contains(tt.got, "<") {
					t.Errorf("SanitizeHTMLStrict() = %q, contains '<'", tt.got)
				}
				return
			}
			if tt.got != tt.want {
				t.Errorf("SanitizeInput() = %q, want %q", tt.got, tt.want)
			}
		})
	}
}
