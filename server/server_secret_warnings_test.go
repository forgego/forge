package server

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdLog redirects the standard logger for the duration of the test.
func captureStdLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prevOut, prevFlags := log.Writer(), log.Flags()
	log.SetOutput(&buf)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})
	return &buf
}

// Loading the config is silent about generated secrets; the server reports
// them when it starts, so CLI commands that never serve stay quiet (#294).
func TestEphemeralSecretWarningsOnlyOnServerStart(t *testing.T) {
	withCleanSecretEnv(t)
	logged := captureStdLog(t)

	cfg := config.NewConfig()
	require.Len(t, cfg.GeneratedSecrets(), 3)
	assert.NotContains(t, logged.String(), "ephemeral", "loading the config must not warn")

	srv, err := NewServer(cfg, &config.Settings{
		App:    config.AppSettings{Env: "development"},
		Server: config.ServerSettings{Host: "127.0.0.1", Port: "0"},
	}, nil)
	require.NoError(t, err)
	assert.NotContains(t, logged.String(), "ephemeral", "building the server must not warn")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, srv.serveUntil(ctx, nil))
	for _, key := range []string{"security.secret_key", "security.csrf_secret_key", "security.session_secret"} {
		assert.Contains(t, logged.String(), key+" is not configured; using a generated ephemeral value")
	}
}
