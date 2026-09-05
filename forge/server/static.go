package server

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// StaticOptions configures static file serving
type StaticOptions struct {
	// IndexFiles are filenames to try when serving directories (e.g., "index.html")
	IndexFiles []string
	// ShowIndexes enables directory listing (dev mode only)
	ShowIndexes bool
	// Prefix is the URL prefix to strip from the request path
	Prefix string
	// MaxAge sets the cache max age in seconds
	MaxAge int
	// DisableCache disables caching entirely
	DisableCache bool
	// Fallback is the file to serve if the requested file is not found (SPA support)
	Fallback string
}

// DefaultStaticOptions returns default static file options
func DefaultStaticOptions() *StaticOptions {
	return &StaticOptions{
		IndexFiles:   []string{"index.html"},
		ShowIndexes:  false,
		Prefix:       "",
		MaxAge:       3600, // 1 hour
		DisableCache: false,
		Fallback:     "",
	}
}

// StaticOption is a function that configures StaticOptions
type StaticOption func(*StaticOptions)

// WithFallback sets the fallback file (SPA support)
func WithFallback(file string) StaticOption {
	return func(o *StaticOptions) {
		o.Fallback = file
	}
}

// WithIndexFiles sets the index files
func WithIndexFiles(files ...string) StaticOption {
	return func(o *StaticOptions) {
		o.IndexFiles = files
	}
}

// WithShowIndexes enables directory listing
func WithShowIndexes(enabled bool) StaticOption {
	return func(o *StaticOptions) {
		o.ShowIndexes = enabled
	}
}

// WithPrefix sets the URL prefix to strip
func WithPrefix(prefix string) StaticOption {
	return func(o *StaticOptions) {
		o.Prefix = prefix
	}
}

// WithMaxAge sets the cache max age
func WithMaxAge(seconds int) StaticOption {
	return func(o *StaticOptions) {
		o.MaxAge = seconds
	}
}

// WithDisableCache disables caching
func WithDisableCache(disabled bool) StaticOption {
	return func(o *StaticOptions) {
		o.DisableCache = disabled
	}
}

// StaticFiles serves static files from a directory
func StaticFiles(pattern, root string, options ...StaticOption) http.Handler {
	opts := DefaultStaticOptions()
	for _, opt := range options {
		opt(opts)
	}

	fs := os.DirFS(root)
	return StaticFS(pattern, fs, options...)
}

// StaticFS serves static files from an fs.FS (including embed.FS)
func StaticFS(pattern string, filesystem fs.FS, options ...StaticOption) http.Handler {
	opts := DefaultStaticOptions()
	for _, opt := range options {
		opt(opts)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the file path from the URL
		urlPath := r.URL.Path

		// Remove the pattern prefix if specified
		if opts.Prefix != "" {
			if !strings.HasPrefix(urlPath, opts.Prefix) {
				http.NotFound(w, r)
				return
			}
			urlPath = strings.TrimPrefix(urlPath, opts.Prefix)
		} else if pattern != "" && pattern != "/" {
			// Remove pattern from path
			if !strings.HasPrefix(urlPath, pattern) {
				http.NotFound(w, r)
				return
			}
			urlPath = strings.TrimPrefix(urlPath, pattern)
		}

		// Clean the path to prevent directory traversal
		urlPath = path.Clean("/" + urlPath)
		if strings.Contains(urlPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Remove leading slash for filesystem
		filePath := strings.TrimPrefix(urlPath, "/")
		if filePath == "" {
			filePath = "."
		}

		// Try to open the file
		file, err := filesystem.Open(filePath)
		if err != nil {
			// If file not found, try index files if it's a directory request
			if os.IsNotExist(err) && len(opts.IndexFiles) > 0 {
				for _, indexFile := range opts.IndexFiles {
					indexPath := path.Join(filePath, indexFile)
					if f, err := filesystem.Open(indexPath); err == nil {
						file = f
						filePath = indexPath
						break
					}
				}
			}

			if file == nil {
				// Try fallback if configured (SPA support)
				if opts.Fallback != "" {
					if f, err := filesystem.Open(opts.Fallback); err == nil {
						file = f
						filePath = opts.Fallback
					}
				}

				if file == nil {
					http.NotFound(w, r)
					return
				}
			}
		}
		defer file.Close()

		// Get file info
		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If it's a directory, try index files or show listing
		if stat.IsDir() {
			if opts.ShowIndexes {
				// Directory listing (simplified - in production, use a proper template)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(w, "<html><head><title>Index of %s</title></head><body><h1>Index of %s</h1><ul>", urlPath, urlPath)
				// Note: This is a simplified listing. For production, use a proper directory listing implementation.
				fmt.Fprintf(w, "</ul></body></html>")
				return
			}

			// Try index files
			for _, indexFile := range opts.IndexFiles {
				indexPath := path.Join(filePath, indexFile)
				if f, err := filesystem.Open(indexPath); err == nil {
					// Serve index file directly
					defer f.Close()
					if stat, err := f.Stat(); err == nil {
						// Set content type
						contentType := detectContentType(indexPath)
						w.Header().Set("Content-Type", contentType)
						serveContent(w, r, f, stat)
						return
					}
				}
			}
		}

		// Set content type
		contentType := detectContentType(filePath)
		w.Header().Set("Content-Type", contentType)

		// Set cache headers
		if !opts.DisableCache {
			if opts.MaxAge > 0 {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", opts.MaxAge))
			}
			// Set ETag
			etag := generateETag(stat)
			w.Header().Set("ETag", etag)

			// Check If-None-Match
			if match := r.Header.Get("If-None-Match"); match != "" {
				if match == etag || match == "*" {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			// Set Last-Modified
			modTime := stat.ModTime()
			w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))

			// Check If-Modified-Since
			if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil {
				if modTime.UTC().Before(t.Add(1 * time.Second)) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		} else {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		// Handle range requests
		if stat.Size() > 0 {
			w.Header().Set("Accept-Ranges", "bytes")
			serveContent(w, r, file, stat)
		} else {
			// Empty file
			w.WriteHeader(http.StatusOK)
		}
	})
}

// serveContent serves file content with range request support
func serveContent(w http.ResponseWriter, r *http.Request, file fs.File, stat fs.FileInfo) {
	size := stat.Size()

	// Check for range request
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// No range request, serve entire file
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.WriteHeader(http.StatusOK)
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
		return
	}

	// Parse range header
	ranges, err := parseRange(rangeHeader, size)
	if err != nil || len(ranges) == 0 {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// For simplicity, handle single range only
	ra := ranges[0]
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.end, size))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", ra.length))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to start position
	if seeker, ok := file.(io.ReadSeeker); ok {
		seeker.Seek(ra.start, 0)
		// Copy the range
		io.CopyN(w, seeker, ra.length)
	} else {
		// Fallback: read and discard until start
		buf := make([]byte, 32*1024)
		remaining := ra.start
		for remaining > 0 {
			n := int64(len(buf))
			if n > remaining {
				n = remaining
			}
			file.Read(buf[:n])
			remaining -= n
		}
		// Now copy the range
		io.CopyN(w, file, ra.length)
	}
}

// byteRange represents a byte range
type byteRange struct {
	start, end, length int64
}

// parseRange parses a Range header
func parseRange(s string, size int64) ([]byteRange, error) {
	if !strings.HasPrefix(s, "bytes=") {
		return nil, fmt.Errorf("invalid range")
	}

	s = s[6:] // Remove "bytes="
	ranges := []byteRange{}

	for _, ra := range strings.Split(s, ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}

		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, fmt.Errorf("invalid range")
		}

		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])

		var r byteRange
		if start == "" {
			// Suffix range: -500 means last 500 bytes
			if end == "" {
				return nil, fmt.Errorf("invalid range")
			}
			suffix, err := parseInt64(end)
			if err != nil {
				return nil, err
			}
			if suffix > size {
				suffix = size
			}
			r.start = size - suffix
			r.end = size - 1
		} else {
			// Start specified
			var err error
			r.start, err = parseInt64(start)
			if err != nil {
				return nil, err
			}
			if r.start < 0 {
				r.start = 0
			}
			if r.start >= size {
				continue // Skip unsatisfiable range
			}

			if end == "" {
				// No end specified, serve to end
				r.end = size - 1
			} else {
				r.end, err = parseInt64(end)
				if err != nil {
					return nil, err
				}
				if r.end >= size {
					r.end = size - 1
				}
			}
		}

		if r.start > r.end {
			continue // Skip invalid range
		}

		r.length = r.end - r.start + 1
		ranges = append(ranges, r)
	}

	return ranges, nil
}

// parseInt64 parses an integer string
func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// generateETag generates an ETag from file info
func generateETag(stat fs.FileInfo) string {
	// Simple ETag based on modification time and size
	// In production, you might want to use a hash of the file content
	return fmt.Sprintf(`"%x-%x"`, stat.Size(), stat.ModTime().Unix())
}

// detectContentType detects the content type from file extension
func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".xml":
		return "application/xml; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

