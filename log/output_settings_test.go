package log

import (
	"github.com/forgego/forge/config"
	forgeerrors "github.com/forgego/forge/errors"
	"testing"
)

func TestRemoteOutputFailsWithNotImplemented(t *testing.T) {
	logger, err := NewBuilder().AddRemoteOutput("http://example.invalid/logs", LevelInfo, 5).Build()
	if logger != nil || !forgeerrors.IsNotImplemented(err) {
		t.Fatalf("logger = %v, err = %v", logger, err)
	}
}

func TestFileOutputUsesDocumentedSizeAndBackupDefaults(t *testing.T) {
	cfg := ConfigFromSettings(config.LoggingSettings{Outputs: []config.LoggingOutputConfig{{Type: "file", Enabled: true, Path: "app.log"}}})
	if len(cfg.Outputs) != 1 {
		t.Fatalf("outputs = %+v", cfg.Outputs)
	}
	rotation := cfg.Outputs[0].File.Rotation
	if rotation.MaxSize != 100 || rotation.MaxBackups != 10 {
		t.Fatalf("rotation = %+v", rotation)
	}
}
