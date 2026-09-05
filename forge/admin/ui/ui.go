//go:build embed
// +build embed

package ui

import (
	"embed"
	"io/fs"
)

// dist contains the built React Admin UI files
//go:embed dist/*
var dist embed.FS

// GetFS returns a sub-filesystem for the dist folder
func GetFS() fs.FS {
	f, _ := fs.Sub(dist, "dist")
	return f
}

// Handler returns a default handler for the admin UI (can be overridden)
// For use in simpler setups
func Handler() fs.FS {
	return GetFS()
}

// Force rebuild 4
