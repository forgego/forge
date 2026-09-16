package server

import (
	"os"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/require"
)

const (
	testSecureSecretKey   = "test-production-secret-key-that-is-long-and-explicit-001"
	testSecureCSRFKey     = "test-production-csrf-secret-that-is-long-and-explicit-002"
	testSecureSessionKey  = "test-production-session-secret-long-and-explicit-003"
	testExplicitSecretKey = "explicit-config-secret-key-long-and-secure-value-001"
	testExplicitCSRFKey   = "explicit-config-csrf-secret-long-and-secure-value-002"
	testExplicitSessKey   = "explicit-config-session-secret-long-secure-value-003"
)

// withCleanSecretEnv guarantees NewConfig generates ephemeral secrets by
// removing any explicit secret configuration from the environment.
func withCleanSecretEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"FORGE_SECURITY_SECRET_KEY",
		"FORGE_SECURITY_CSRF_SECRET_KEY",
		"FORGE_SECURITY_SESSION_SECRET",
	}
	old := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			old[k] = v
		}
		require.NoError(t, os.Unsetenv(k))
	}
	t.Cleanup(func() {
		for _, k := range keys {
			_ = os.Unsetenv(k)
		}
		for k, v := range old {
			_ = os.Setenv(k, v)
		}
	})
}

func productionSettingsWith(security config.SecuritySettings) *config.Settings {
	return &config.Settings{
		App:      config.AppSettings{Env: "production"},
		Security: security,
	}
}

func secureTestSecurity() config.SecuritySettings {
	return config.SecuritySettings{
		SecretKey:     testSecureSecretKey,
		CSRFSecretKey: testSecureCSRFKey,
		SessionSecret: testSecureSessionKey,
	}
}

// Production with explicit secure settings must start even when the Config
// holds generated ephemeral secrets.
func TestProductionWithExplicitSecureSettingsStarts(t *testing.T) {
	withCleanSecretEnv(t)
	cfg := config.NewConfig()
	require.NotEmpty(t, cfg.GeneratedSecrets(), "test requires a config with generated secrets")
	srv, err := NewServer(cfg, productionSettingsWith(secureTestSecurity()), nil)
	require.NoError(t, err)
	require.NoError(t, srv.validateProductionSecrets())
}

// Production with empty effective security must fail, even when the Config
// itself holds explicit secrets (the server uses Settings, not Config).
func TestProductionWithEmptySettingsSecurityFails(t *testing.T) {
	t.Setenv("FORGE_SECURITY_SECRET_KEY", testExplicitSecretKey)
	t.Setenv("FORGE_SECURITY_CSRF_SECRET_KEY", testExplicitCSRFKey)
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", testExplicitSessKey)
	cfg := config.NewConfig()
	require.Empty(t, cfg.GeneratedSecrets())
	srv, err := NewServer(cfg, productionSettingsWith(config.SecuritySettings{}), nil)
	require.NoError(t, err)
	err = srv.validateProductionSecrets()
	require.Error(t, err)
	require.Contains(t, err.Error(), "security.secret_key")
	require.Contains(t, err.Error(), "security.csrf_secret_key")
	require.Contains(t, err.Error(), "security.session_secret")
}

// Production with placeholder effective security must fail.
func TestProductionWithPlaceholderSettingsSecurityFails(t *testing.T) {
	t.Setenv("FORGE_SECURITY_SECRET_KEY", testExplicitSecretKey)
	t.Setenv("FORGE_SECURITY_CSRF_SECRET_KEY", testExplicitCSRFKey)
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", testExplicitSessKey)
	cfg := config.NewConfig()
	require.Empty(t, cfg.GeneratedSecrets())
	srv, err := NewServer(cfg, productionSettingsWith(config.SecuritySettings{
		SecretKey:     "change-me-in-production-please",
		CSRFSecretKey: "change-me-csrf",
		SessionSecret: "secret",
	}), nil)
	require.NoError(t, err)
	require.Error(t, srv.validateProductionSecrets())
}

// Non-production environments are unaffected by secret validation.
func TestNonProductionIgnoresInsecureSettingsSecurity(t *testing.T) {
	withCleanSecretEnv(t)
	cfg := config.NewConfig()
	for _, env := range []string{"development", "test", "staging", ""} {
		srv, err := NewServer(cfg, &config.Settings{
			App:      config.AppSettings{Env: env},
			Security: config.SecuritySettings{},
		}, nil)
		require.NoError(t, err)
		require.NoError(t, srv.validateProductionSecrets(), "env %q should not be validated", env)
	}
}
