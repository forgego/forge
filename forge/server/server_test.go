package server

import (
	"testing"

	"github.com/forgego/forge/config"
)

func TestNewServer_NilInputs(t *testing.T) {
	// Test nil config
	settings := &config.Settings{}
	srv, err := NewServer(nil, settings, nil)
	if err == nil || err.Error() != "server config is nil" {
		t.Errorf("Expected 'server config is nil' error, got %v", err)
	}
	if srv != nil {
		t.Errorf("Expected nil server, got %v", srv)
	}

	// Test nil settings
	cfg := &config.Config{}
	srv, err = NewServer(cfg, nil, nil)
	if err == nil || err.Error() != "server settings are nil" {
		t.Errorf("Expected 'server settings are nil' error, got %v", err)
	}
	if srv != nil {
		t.Errorf("Expected nil server, got %v", srv)
	}
}
