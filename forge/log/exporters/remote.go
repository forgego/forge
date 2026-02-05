package exporters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap/zapcore"
)

// RemoteExporter sends logs to a remote service
type RemoteExporter struct {
	url     string
	format  string
	timeout time.Duration
	level   zapcore.Level
	client  *http.Client
}

// RemoteConfig configures remote output
type RemoteConfig struct {
	URL     string
	Format  string
	Timeout time.Duration
}

// NewRemoteExporter creates a new remote exporter
func NewRemoteExporter(config RemoteConfig, level zapcore.Level) *RemoteExporter {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &RemoteExporter{
		url:     config.URL,
		format:  config.Format,
		timeout: timeout,
		level:   level,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetWriter returns a write syncer that sends to remote service
func (e *RemoteExporter) GetWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(&remoteWriter{
		exporter: e,
		buffer:   make([]byte, 0, 1024),
	})
}

// GetLevel returns the log level
func (e *RemoteExporter) GetLevel() zapcore.Level {
	return e.level
}

// remoteWriter implements io.Writer for remote logging
type remoteWriter struct {
	exporter *RemoteExporter
	buffer   []byte
}

func (w *remoteWriter) Write(p []byte) (int, error) {
	// Append to buffer
	w.buffer = append(w.buffer, p...)

	// If buffer ends with newline, send it
	if len(w.buffer) > 0 && w.buffer[len(w.buffer)-1] == '\n' {
		if err := w.send(); err != nil {
			// Log error but don't fail - remote logging should be non-blocking
			return len(p), nil
		}
	}

	return len(p), nil
}

func (w *remoteWriter) send() error {
	if len(w.buffer) == 0 {
		return nil
	}

	// Prepare payload
	var payload interface{}
	if w.exporter.format == "json" {
		// Try to parse as JSON
		var logEntry map[string]interface{}
		if err := json.Unmarshal(w.buffer, &logEntry); err == nil {
			payload = logEntry
		} else {
			payload = map[string]interface{}{
				"message": string(w.buffer),
			}
		}
	} else {
		payload = map[string]interface{}{
			"message": string(w.buffer),
		}
	}

	// Marshal to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send HTTP POST request
	req, err := http.NewRequest("POST", w.exporter.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.exporter.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("remote logging failed with status %d", resp.StatusCode)
	}

	// Clear buffer
	w.buffer = w.buffer[:0]
	return nil
}

// Sync flushes the buffer
func (w *remoteWriter) Sync() error {
	if len(w.buffer) > 0 {
		return w.send()
	}
	return nil
}
