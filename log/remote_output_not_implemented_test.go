package log

import (
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
)

func TestRemoteLogOutput_NotImplemented(t *testing.T) {
	cfg := &LoggingConfig{
		Level:  LevelInfo,
		Format: FormatJSON,
		Outputs: []OutputConfig{
			{
				Type:    OutputRemote,
				Enabled: true,
				Level:   LevelInfo,
				Remote: RemoteOutputConfig{
					URL: "http://localhost:8080",
				},
			},
		},
	}

	logger, err := NewLoggerFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error when building logger with OutputRemote, got nil")
	}
	if !forgeerrors.IsNotImplemented(err) {
		t.Fatalf("expected NotImplementedError, got: %v", err)
	}
	if logger != nil {
		t.Errorf("expected nil logger, got: %v", logger)
	}
}
