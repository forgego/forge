package server

import (
	"context"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A malformed logging.outputs list cannot fail config.LoadSettings, so the
// server reports it when it starts instead of silently logging to the
// console.
func TestMalformedLoggingOutputsWarnsOnServerStart(t *testing.T) {
	withCleanSecretEnv(t)
	logged := captureStdLog(t)

	cfg := config.NewConfig()
	cfg.Set("logging.outputs", "file")
	settings := config.LoadSettings(cfg)
	settings.App.Env = "development"
	settings.Server.Host = "127.0.0.1"
	settings.Server.Port = "0"
	assert.NotContains(t, logged.String(), "logging.outputs", "loading the config must not warn")

	srv, err := NewServer(cfg, settings, nil)
	require.NoError(t, err)
	assert.NotContains(t, logged.String(), "logging.outputs", "building the server must not warn")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, srv.serveUntil(ctx, nil))
	assert.Contains(t, logged.String(), "forge/server: WARNING: logging.outputs is malformed and was ignored")
}

func TestValidLoggingOutputsDoNotWarnOnServerStart(t *testing.T) {
	withCleanSecretEnv(t)
	logged := captureStdLog(t)

	cfg := config.NewConfig()
	cfg.Set("logging.outputs", []map[string]interface{}{{"type": "console", "enabled": true}})
	srv, err := NewServer(cfg, &config.Settings{
		App:    config.AppSettings{Env: "development"},
		Server: config.ServerSettings{Host: "127.0.0.1", Port: "0"},
	}, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, srv.serveUntil(ctx, nil))
	assert.NotContains(t, logged.String(), "logging.outputs")
}
