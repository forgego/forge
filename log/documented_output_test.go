package log

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Console and JSON encoders produce their documented output formats.
func TestLoggerBuilderSupportsConsoleAndJSONFormats(t *testing.T) {
	for _, format := range []Format{FormatConsole, FormatJSON} {
		path := filepath.Join(t.TempDir(), "format.log")
		builder := NewBuilder().Format(format)
		builder.config.Outputs = []OutputConfig{{Type: OutputFile, Enabled: true, Level: LevelInfo, Format: format, File: FileOutputConfig{Path: path}}}
		builder.config.Production.Sampling.Enabled = false
		logger, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		logger.Info("format-probe")
		if err := logger.Close(); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "format-probe") {
			t.Fatalf("missing message: %s", data)
		}
		var record map[string]any
		err = json.Unmarshal(data, &record)
		if format == FormatJSON && (err != nil || record["msg"] != "format-probe") {
			t.Fatalf("JSON output=%s err=%v", data, err)
		}
		if format == FormatConsole && err == nil {
			t.Fatalf("console output was JSON: %s", data)
		}
	}
}

// Production sampling suppresses repeated log entries after the first.
func TestLoggerBuilderSamplesRepeatedMessages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.log")
	logger, err := NewBuilder().Production().Sampling(1, 2).AddFileOutput(path, LevelInfo, 1, 1, 1, false).Build()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		logger.Info("repeated-sample")
	}
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "repeated-sample"); got != 3 {
		t.Fatalf("sampled entries=%d, want 3", got)
	}
}
