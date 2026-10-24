package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/forgego/forge/admin"
)

//go:embed web/dist/*
var embeddedAssets embed.FS

// Extension returns a configured admin.Extension targeting ExtensionAPIVersion.
func Extension() (admin.Extension, error) {
	distFS, err := fs.Sub(embeddedAssets, "web/dist")
	if err != nil {
		return admin.Extension{}, fmt.Errorf("failed to create sub filesystem: %w", err)
	}

	return admin.Extension{
		Name:    "analytics",
		Assets:  distFS,
		Version: admin.ExtensionAPIVersion,
	}, nil
}

func main() {
	ext, err := Extension()
	if err != nil {
		log.Fatalf("extension setup failed: %v", err)
	}

	site := admin.DefaultSite
	if err := site.RegisterExtension(ext); err != nil {
		log.Fatalf("failed to register extension: %v", err)
	}

	fmt.Printf("Registered admin extension %q (API v%d)\n", ext.Name, ext.Version)
	fmt.Printf("Assets available under /admin/ext/%s/\n", ext.Name)

	mux := http.NewServeMux()
	mux.Handle("/admin/", site.Handler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Starting admin server on http://localhost:8080/admin/")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server terminated: %v", err)
	}
}
