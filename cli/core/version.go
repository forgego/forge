package core

import (
	"runtime/debug"
	"strings"
)

// modulePath is the Go module that contains the CLI and the framework.
const modulePath = "github.com/forgego/forge"

// fallbackVersion is used when the binary carries no module version, as in a
// local `go build` from a checkout. Update it when tagging a release.
const fallbackVersion = "v0.2.0"

// Version reports the version of the Forge module this binary was built from.
// `go install github.com/forgego/forge/cmd/forge@vX.Y.Z` records vX.Y.Z (or a
// pseudo-version for a commit) in the binary's build information.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fallbackVersion
	}
	return versionFromBuildInfo(info)
}

func versionFromBuildInfo(info *debug.BuildInfo) string {
	if info.Main.Path == modulePath && isReleaseVersion(info.Main.Version) {
		return info.Main.Version
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath && isReleaseVersion(dep.Version) {
			return dep.Version
		}
	}
	return fallbackVersion
}

// isReleaseVersion reports whether v can be required from the module proxy.
// Local builds report "(devel)" or a VCS pseudo-version with "+dirty" build
// metadata, neither of which another module can download.
func isReleaseVersion(v string) bool {
	return v != "" && v != "(devel)" && !strings.Contains(v, "+")
}
