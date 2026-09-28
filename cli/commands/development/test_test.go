package development

import (
	"reflect"
	"testing"
)

func TestBuildGoTestArgs(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		coverage bool
		want     []string
	}{
		{"default", false, false, []string{"test", "./..."}},
		{"verbose", true, false, []string{"test", "-v", "./..."}},
		{"coverage", false, true, []string{"test", "-cover", "./..."}},
		{"verbose coverage", true, true, []string{"test", "-v", "-cover", "./..."}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildGoTestArgs(tt.verbose, tt.coverage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildGoTestArgs(%v, %v) = %v, want %v", tt.verbose, tt.coverage, got, tt.want)
			}
		})
	}
}
