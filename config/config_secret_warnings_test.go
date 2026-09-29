package config

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewConfig must not print the ephemeral-secret warnings: every CLI command
// loads the config, and only server startup reports them (#294).
func TestNewConfigDoesNotLogSecretWarnings(t *testing.T) {
	t.Setenv("FORGE_SECURITY_SECRET_KEY", "")
	t.Setenv("FORGE_SECURITY_CSRF_SECRET_KEY", "change-me")
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", "")
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(prev) })

	cfg := NewConfig()

	assert.Empty(t, buf.String())
	warnings := cfg.SecretWarnings()
	require.Len(t, warnings, 3)
	assert.Contains(t, warnings, "security.csrf_secret_key is set to an insecure placeholder; overriding with a generated ephemeral value (set it explicitly for production)")
	assert.Contains(t, warnings, "security.secret_key is not configured; using a generated ephemeral value (set it explicitly for production)")

	cfg.Set("security.secret_key", "an-explicit-secret-that-is-long-enough-000000")
	assert.Len(t, cfg.SecretWarnings(), 2, "an explicitly set secret is no longer reported")
}
