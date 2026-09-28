package core

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionFromBuildInfo(t *testing.T) {
	cases := map[string]struct {
		info *debug.BuildInfo
		want string
	}{
		"installed release": {
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v0.2.0"}},
			want: "v0.2.0",
		},
		"installed commit": {
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v0.2.1-0.20260928120000-abcdef123456"}},
			want: "v0.2.1-0.20260928120000-abcdef123456",
		},
		"local build with uncommitted changes": {
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v0.0.0-20260928134643-a19c1b8a9e49+dirty"}},
			want: fallbackVersion,
		},
		"local build": {
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "(devel)"}},
			want: fallbackVersion,
		},
		"forge as a dependency": {
			info: &debug.BuildInfo{
				Main: debug.Module{Path: "example.com/app", Version: "(devel)"},
				Deps: []*debug.Module{{Path: modulePath, Version: "v0.3.0"}},
			},
			want: "v0.3.0",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, versionFromBuildInfo(tc.info))
		})
	}
}
