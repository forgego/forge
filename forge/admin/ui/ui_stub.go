//go:build !embed
// +build !embed

package ui

import (
	"embed"
	"io/fs"
)

// stub contains a minimal placeholder UI for non-embedded builds.
//go:embed stub/*
var stub embed.FS

// GetFS returns a sub-filesystem for the stub folder.
func GetFS() fs.FS {
	f, _ := fs.Sub(stub, "stub")
	return f
}

// Handler returns a default handler for the admin UI (can be overridden).
// For use in simpler setups.
func Handler() fs.FS {
	return GetFS()
}
