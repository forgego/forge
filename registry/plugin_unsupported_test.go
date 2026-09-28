package registry

import (
	"fmt"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
)

func TestRegisterPluginUnsupportedPluginTypes(t *testing.T) {
	tests := []struct {
		name   string
		plugin Plugin
		wantOK bool
	}{
		{"admin", NewAdminPluginBase("admin_"+t.Name(), "1.0.0"), false},
		{"api", NewAPIPluginBase("api_"+t.Name(), "1.0.0"), false},
		{"model", NewModelPluginBase("model_"+t.Name(), "1.0.0"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterPlugin(tt.plugin)
			if tt.wantOK {
				if err != nil {
					t.Fatalf("RegisterPlugin() error = %v", err)
				}
				return
			}
			if !forgeerrors.IsNotImplemented(err) {
				t.Fatalf("RegisterPlugin() error = %v, want not implemented", err)
			}
		})
	}

	if _, err := GetPlugin(fmt.Sprintf("model_%s", t.Name())); err != nil {
		t.Fatalf("model-only plugin was not registered: %v", err)
	}
}
