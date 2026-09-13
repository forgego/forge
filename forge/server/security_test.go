package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXSSEscapeHTML(t *testing.T) {
	xss := NewXSS()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic HTML escaping",
			input:    `<script>alert("xss")</script>`,
			expected: `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`,
		},
		{
			name:     "Ampersand escaping",
			input:    `Tom & Jerry`,
			expected: `Tom &amp; Jerry`,
		},
		{
			name:     "Quote escaping",
			input:    `"Hello World"`,
			expected: `&#34;Hello World&#34;`,
		},
		{
			name:     "Single quote escaping",
			input:    `'Hello World'`,
			expected: `&#39;Hello World&#39;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xss.EscapeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLInjectionSanitizeIdentifier(t *testing.T) {
	sqli := NewSQLInjection()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Safe identifier",
			input:    "user_name",
			expected: "user_name",
		},
		{
			name:     "Identifier with special chars",
			input:    "user@name!",
			expected: "username",
		},
		{
			name:     "Identifier with spaces",
			input:    "user name",
			expected: "username",
		},
		{
			name:     "Identifier with SQL chars",
			input:    "user;DROP TABLE",
			expected: "userDROPTABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqli.SanitizeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContentSecurityPolicy(t *testing.T) {
	xss := NewXSS()
	csp := xss.ContentSecurityPolicy()

	t.Run("Required headers", func(t *testing.T) {
		assert.Contains(t, csp, "Content-Security-Policy")
		assert.Contains(t, csp, "X-Content-Type-Options")
		assert.Contains(t, csp, "X-Frame-Options")
		assert.Contains(t, csp, "X-XSS-Protection")
	})

	t.Run("Header values", func(t *testing.T) {
		assert.Equal(t, "nosniff", csp["X-Content-Type-Options"])
		assert.Equal(t, "DENY", csp["X-Frame-Options"])
	})
}

func BenchmarkXSSSanitization(b *testing.B) {
	xss := NewXSS()
	input := `<div onclick="alert('xss')">Hello <script>alert('xss')</script> World</div>`

	b.Run("EscapeHTML", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			xss.EscapeHTML(input)
		}
	})
}

func BenchmarkSQLInjection(b *testing.B) {
	sqli := NewSQLInjection()

	b.Run("SanitizeIdentifier", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sqli.SanitizeIdentifier("user_table_name")
		}
	})
}
