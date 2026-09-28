package server

import (
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
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
	// IndexTransform rewrites the bytes of index/fallback HTML files before they are served
	IndexTransform func([]byte) []byte
}

// DefaultStaticOptions returns default static file options
func DefaultStaticOptions() *StaticOptions {
	return &StaticOptions{
		IndexFiles:     []string{"index.html"},
		ShowIndexes:    false,
		Prefix:         "",
		MaxAge:         3600, // 1 hour
		DisableCache:   false,
		Fallback:       "",
		IndexTransform: nil,
	}
}

// StaticOption is a function that configures StaticOptions
type StaticOption func(*StaticOptions)

// WithIndexTransform rewrites the bytes of index/fallback HTML files before they are served.
func WithIndexTransform(fn func([]byte) []byte) StaticOption {
	return func(o *StaticOptions) {
		o.IndexTransform = fn
	}
}

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
	for _, option := range options {
		option(opts)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveStaticRequest(w, r, pattern, filesystem, opts)
	})
}

func serveStaticRequest(w http.ResponseWriter, r *http.Request, pattern string, filesystem fs.FS, opts *StaticOptions) {
	filePath, urlPath, ok := staticFilePath(w, r, pattern, opts)
	if !ok {
		return
	}
	file, filePath, fallback, index, ok := openStaticFile(filesystem, filePath, opts)
	if !ok {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		file, stat, filePath, fallback, index, ok = resolveStaticDirectory(w, r, filesystem, filePath, urlPath, opts)
		if !ok {
			return
		}
		defer file.Close()
	}
	serveStaticResponse(w, r, file, stat, filePath, opts, index || fallback)
}

func staticFilePath(w http.ResponseWriter, r *http.Request, pattern string, opts *StaticOptions) (string, string, bool) {
	urlPath := r.URL.Path
	if opts.Prefix != "" {
		if !strings.HasPrefix(urlPath, opts.Prefix) {
			http.NotFound(w, r)
			return "", "", false
		}
		urlPath = strings.TrimPrefix(urlPath, opts.Prefix)
	} else if pattern != "" && pattern != "/" {
		if !strings.HasPrefix(urlPath, pattern) {
			http.NotFound(w, r)
			return "", "", false
		}
		urlPath = strings.TrimPrefix(urlPath, pattern)
	}
	urlPath = path.Clean("/" + urlPath)
	if strings.Contains(urlPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return "", "", false
	}
	filePath := strings.TrimPrefix(urlPath, "/")
	if filePath == "" {
		filePath = "."
	}
	return filePath, urlPath, true
}

func openStaticFile(filesystem fs.FS, filePath string, opts *StaticOptions) (fs.File, string, bool, bool, bool) {
	file, err := filesystem.Open(filePath)
	if err == nil {
		return file, filePath, false, false, true
	}
	if os.IsNotExist(err) {
		if file, indexPath, ok := openIndexFile(filesystem, filePath, opts.IndexFiles); ok {
			return file, indexPath, false, true, true
		}
	}
	if file, ok := openFallbackFile(filesystem, opts.Fallback); ok {
		return file, opts.Fallback, true, false, true
	}
	return nil, "", false, false, false
}

func openIndexFile(filesystem fs.FS, filePath string, indexFiles []string) (fs.File, string, bool) {
	for _, indexFile := range indexFiles {
		indexPath := path.Join(filePath, indexFile)
		if file, err := filesystem.Open(indexPath); err == nil {
			return file, indexPath, true
		}
	}
	return nil, "", false
}

func openFallbackFile(filesystem fs.FS, fallback string) (fs.File, bool) {
	if fallback == "" {
		return nil, false
	}
	file, err := filesystem.Open(fallback)
	return file, err == nil
}

func resolveStaticDirectory(w http.ResponseWriter, r *http.Request, filesystem fs.FS, filePath, urlPath string, opts *StaticOptions) (fs.File, fs.FileInfo, string, bool, bool, bool) {
	if opts.ShowIndexes {
		serveDirectoryIndex(w, urlPath)
		return nil, nil, "", false, false, false
	}
	if file, stat, indexPath, ok := openDirectoryIndex(filesystem, filePath, opts.IndexFiles); ok {
		return file, stat, indexPath, false, true, true
	}
	if file, stat, ok := openRegularFallback(filesystem, opts.Fallback); ok {
		return file, stat, opts.Fallback, true, false, true
	}
	http.NotFound(w, r)
	return nil, nil, "", false, false, false
}

func serveDirectoryIndex(w http.ResponseWriter, urlPath string) {
	safePath := html.EscapeString(urlPath)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<html><head><title>Index of %s</title></head><body><h1>Index of %s</h1><ul>", safePath, safePath)
	fmt.Fprint(w, "</ul></body></html>")
}

func openDirectoryIndex(filesystem fs.FS, filePath string, indexFiles []string) (fs.File, fs.FileInfo, string, bool) {
	for _, indexName := range indexFiles {
		indexPath := path.Join(filePath, indexName)
		file, err := filesystem.Open(indexPath)
		if err != nil {
			continue
		}
		stat, statErr := file.Stat()
		if statErr == nil && !stat.IsDir() {
			return file, stat, indexPath, true
		}
		file.Close()
	}
	return nil, nil, "", false
}

func openRegularFallback(filesystem fs.FS, fallback string) (fs.File, fs.FileInfo, bool) {
	file, ok := openFallbackFile(filesystem, fallback)
	if !ok {
		return nil, nil, false
	}
	stat, err := file.Stat()
	if err == nil && !stat.IsDir() {
		return file, stat, true
	}
	file.Close()
	return nil, nil, false
}

func serveStaticResponse(w http.ResponseWriter, r *http.Request, file fs.File, stat fs.FileInfo, filePath string, opts *StaticOptions, matched bool) {
	w.Header().Set("Content-Type", detectContentType(filePath))
	if writeStaticCacheHeaders(w, r, stat, opts) {
		return
	}
	if opts.IndexTransform != nil && isIndexOrFallback(filePath, opts, matched) {
		serveTransformedStaticFile(w, r, file, opts.IndexTransform)
		return
	}
	if stat.Size() > 0 {
		w.Header().Set("Accept-Ranges", "bytes")
		serveContent(w, r, file, stat)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func writeStaticCacheHeaders(w http.ResponseWriter, r *http.Request, stat fs.FileInfo, opts *StaticOptions) bool {
	if opts.DisableCache {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		return false
	}
	if opts.MaxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", opts.MaxAge))
	}
	etag := generateETag(stat)
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag || r.Header.Get("If-None-Match") == "*" {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	modTime := stat.ModTime()
	w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	if modifiedSince(r, modTime) {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}

func modifiedSince(r *http.Request, modTime time.Time) bool {
	modified, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
	return err == nil && modTime.UTC().Before(modified.Add(time.Second))
}

func serveTransformedStaticFile(w http.ResponseWriter, r *http.Request, file fs.File, transform func([]byte) []byte) {
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data = transform(data)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		w.Write(data)
	}
}

func isIndexOrFallback(filePath string, opts *StaticOptions, matched bool) bool {
	if matched {
		return true
	}
	cleanPath := path.Clean("/" + filePath)
	if opts.Fallback != "" && cleanPath == path.Clean("/"+opts.Fallback) {
		return true
	}
	base := path.Base(filePath)
	for _, idx := range opts.IndexFiles {
		if filePath == idx || base == idx || cleanPath == path.Clean("/"+idx) {
			return true
		}
	}
	return false
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
	var ranges []byteRange
	for _, value := range strings.Split(s[6:], ",") {
		byteRange, valid, err := parseByteRange(strings.TrimSpace(value), size)
		if err != nil {
			return nil, err
		}
		if valid {
			ranges = append(ranges, byteRange)
		}
	}
	return ranges, nil
}

func parseByteRange(value string, size int64) (byteRange, bool, error) {
	if value == "" {
		return byteRange{}, false, nil
	}
	separator := strings.Index(value, "-")
	if separator < 0 {
		return byteRange{}, false, fmt.Errorf("invalid range")
	}
	start, end := strings.TrimSpace(value[:separator]), strings.TrimSpace(value[separator+1:])
	if start == "" {
		return parseSuffixRange(end, size)
	}
	return parseStartRange(start, end, size)
}

func parseSuffixRange(end string, size int64) (byteRange, bool, error) {
	if end == "" {
		return byteRange{}, false, fmt.Errorf("invalid range")
	}
	suffix, err := parseInt64(end)
	if err != nil {
		return byteRange{}, false, err
	}
	if suffix > size {
		suffix = size
	}
	return completeRange(byteRange{start: size - suffix, end: size - 1})
}

func parseStartRange(start, end string, size int64) (byteRange, bool, error) {
	rangeStart, err := parseInt64(start)
	if err != nil {
		return byteRange{}, false, err
	}
	if rangeStart < 0 {
		rangeStart = 0
	}
	if rangeStart >= size {
		return byteRange{}, false, nil
	}
	rangeEnd, err := rangeEnd(end, size)
	if err != nil {
		return byteRange{}, false, err
	}
	return completeRange(byteRange{start: rangeStart, end: rangeEnd})
}

func rangeEnd(end string, size int64) (int64, error) {
	if end == "" {
		return size - 1, nil
	}
	rangeEnd, err := parseInt64(end)
	if err != nil {
		return 0, err
	}
	if rangeEnd >= size {
		return size - 1, nil
	}
	return rangeEnd, nil
}

func completeRange(value byteRange) (byteRange, bool, error) {
	if value.start > value.end {
		return byteRange{}, false, nil
	}
	value.length = value.end - value.start + 1
	return value, true, nil
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
