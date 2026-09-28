package config

import (
	"errors"
	"testing"
)

func TestEnsureSecretsFailClosedWhenRandomnessFails(t *testing.T) {
	t.Setenv("FORGE_SECURITY_SECRET_KEY", "")
	t.Setenv("FORGE_SECURITY_CSRF_SECRET_KEY", "")
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", "")

	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errors.New("simulated entropy failure") }
	t.Cleanup(func() { randRead = orig })

	cfg := NewConfig()
	got := cfg.GeneratedSecrets()
	want := map[string]bool{
		"security.secret_key":      false,
		"security.csrf_secret_key": false,
		"security.session_secret":  false,
	}
	for _, key := range got {
		want[key] = true
	}
	for key, seen := range want {
		if !seen {
			t.Errorf("GeneratedSecrets() = %v, missing %q after randomness failure (fail-open)", got, key)
		}
	}
}
