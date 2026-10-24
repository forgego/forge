package admin

import (
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ExtensionAPIVersion is the currently supported client API version.
const ExtensionAPIVersion = 1

var extensionNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

// Extension represents an admin UI extension module.
type Extension struct {
	Name    string
	Assets  fs.FS
	Version int
}

// Validate checks whether the extension satisfies registry invariants:
// name matches ^[a-z0-9-]+$, Assets is non-nil and contains index.js (a regular file),
// and Version == ExtensionAPIVersion.
func (e Extension) Validate() error {
	if !extensionNameRegex.MatchString(e.Name) {
		return fmt.Errorf("admin extension: invalid name %q: must match ^[a-z0-9-]+$", e.Name)
	}
	if e.Version != ExtensionAPIVersion {
		return fmt.Errorf("admin extension %q: unsupported version %d (supported: %d)", e.Name, e.Version, ExtensionAPIVersion)
	}
	if e.Assets == nil {
		return fmt.Errorf("admin extension %q: assets filesystem cannot be nil", e.Name)
	}

	f, err := e.Assets.Open("index.js")
	if err != nil {
		return fmt.Errorf("admin extension %q: missing index.js: %w", e.Name, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("admin extension %q: cannot stat index.js: %w", e.Name, err)
	}
	if stat.IsDir() {
		return fmt.Errorf("admin extension %q: index.js must be a file, not a directory", e.Name)
	}

	return nil
}

// RegisterExtension registers an extension on the site.
// Duplicate names or invalid extensions return a descriptive error.
func (s *Site) RegisterExtension(ext Extension) error {
	if err := ext.Validate(); err != nil {
		return err
	}

	s.extensionsMu.Lock()
	defer s.extensionsMu.Unlock()

	for _, existing := range s.extensions {
		if existing.Name == ext.Name {
			return fmt.Errorf("admin extension %q already registered", ext.Name)
		}
	}

	s.extensions = append(s.extensions, ext)
	return nil
}

// RegisterExtension registers an extension with the default admin site.
func RegisterExtension(ext Extension) error {
	return DefaultSite.RegisterExtension(ext)
}

// ValidateExtensions re-checks that all registered extensions are valid and have unique names.
func (s *Site) ValidateExtensions() error {
	s.extensionsMu.RLock()
	defer s.extensionsMu.RUnlock()

	seen := make(map[string]bool, len(s.extensions))
	for _, ext := range s.extensions {
		if err := ext.Validate(); err != nil {
			return err
		}
		if seen[ext.Name] {
			return fmt.Errorf("admin extension %q already registered", ext.Name)
		}
		seen[ext.Name] = true
	}
	return nil
}

// GetExtensions returns a copy of registered extensions.
func (s *Site) GetExtensions() []Extension {
	s.extensionsMu.RLock()
	defer s.extensionsMu.RUnlock()

	result := make([]Extension, len(s.extensions))
	copy(result, s.extensions)
	return result
}

// mountExtensions mounts extension asset handlers onto the router.
func (s *Site) mountExtensions(r chi.Router, prefix string) {
	s.extensionsMu.RLock()
	exts := make([]Extension, len(s.extensions))
	copy(exts, s.extensions)
	s.extensionsMu.RUnlock()

	if len(exts) == 0 {
		return
	}

	extBase := prefix + "/ext"
	r.Route(extBase, func(extR chi.Router) {
		for _, ext := range exts {
			mountPrefix := extBase + "/" + ext.Name
			handler := extensionAssetHandler(mountPrefix, ext)
			extR.Handle("/"+ext.Name, handler)
			extR.Handle("/"+ext.Name+"/*", handler)
		}
		extR.NotFound(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.NotFound(w, req)
		})
	})
}

// extensionAssetHandler serves extension files.
func extensionAssetHandler(mountPrefix string, ext Extension) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		p := r.URL.Path
		if p == "" || strings.HasSuffix(p, "/") {
			http.NotFound(w, r)
			return
		}

		cleanPath := strings.TrimPrefix(p, "/")
		if !fs.ValidPath(cleanPath) {
			http.NotFound(w, r)
			return
		}

		stat, err := fs.Stat(ext.Assets, cleanPath)
		if err != nil || stat.IsDir() {
			http.NotFound(w, r)
			return
		}

		http.ServeFileFS(w, r, ext.Assets, cleanPath)
	})

	return http.StripPrefix(mountPrefix, handler)
}
