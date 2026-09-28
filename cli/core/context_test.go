package core

import (
	"testing"

	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/require"
)

func TestNewContextPreservesCallerConfig(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite")

	ctx := NewContextWithConfig(cfg)

	require.Same(t, cfg, ctx.Config)
	require.Equal(t, "sqlite", ctx.Config.GetDriver())
}

func TestNewContext_DefaultConfig(t *testing.T) {
	ctx := NewContext()
	defaultCtx := NewContextWithConfig(config.NewConfig())

	require.NotNil(t, ctx)
	require.NotNil(t, ctx.Config)
	require.Equal(t, defaultCtx.Config.GetDriver(), ctx.Config.GetDriver())
}
