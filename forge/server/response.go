package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Response provides helper methods for HTTP responses
type Response struct {
	http.ResponseWriter
	statusCode int
	request    *http.Request
}

// NewResponse wraps an http.ResponseWriter with helper methods
func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		request:        nil,
	}
}

// NewResponseWithRequest wraps an http.ResponseWriter with helper methods and stores the request
func NewResponseWithRequest(w http.ResponseWriter, r *http.Request) *Response {
	return &Response{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		request:        r,
	}
}

// Status sets the HTTP status code
func (r *Response) Status(code int) *Response {
	r.statusCode = code
	return r
}

// JSON sends a JSON response
func (r *Response) JSON(data interface{}) error {
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(r.statusCode)
	return json.NewEncoder(r).Encode(data)
}

// JSONError sends a JSON error response
func (r *Response) JSONError(message string, code int) error {
	return r.Status(code).JSON(map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    code,
	})
}

// Text sends a plain text response
func (r *Response) Text(text string) error {
	r.Header().Set("Content-Type", "text/plain")
	r.WriteHeader(r.statusCode)
	_, err := r.Write([]byte(text))
	return err
}

// HTML sends an HTML response
func (r *Response) HTML(html string) error {
	r.Header().Set("Content-Type", "text/html")
	r.WriteHeader(r.statusCode)
	_, err := r.Write([]byte(html))
	return err
}

// Redirect sends a redirect response
func (r *Response) Redirect(url string, code int) {
	http.Redirect(r, r.Request(), url, code)
}

// Cookie sets a cookie
func (r *Response) Cookie(cookie *http.Cookie) {
	http.SetCookie(r, cookie)
}

// Header sets a response header
func (r *Response) SetHeader(key, value string) {
	r.Header().Set(key, value)
}

// WriteHeader writes the status code
func (r *Response) WriteHeader(code int) {
	if r.statusCode == 0 {
		r.statusCode = code
	}
	r.ResponseWriter.WriteHeader(r.statusCode)
}

// Request returns the associated request (if available)
func (r *Response) Request() *http.Request {
	return r.request
}

// SetRequest sets the associated request
func (r *Response) SetRequest(req *http.Request) {
	r.request = req
}

// ETag sets the ETag header
func (r *Response) ETag(etag string, weak bool) *Response {
	if weak {
		etag = "W/" + etag
	}
	r.Header().Set("ETag", etag)
	return r
}

// LastModified sets the Last-Modified header
func (r *Response) LastModified(t time.Time) *Response {
	r.Header().Set("Last-Modified", t.UTC().Format(http.TimeFormat))
	return r
}

// Cache sets cache control headers
func (r *Response) Cache(maxAge int, public bool) *Response {
	directive := "private"
	if public {
		directive = "public"
	}
	r.Header().Set("Cache-Control", fmt.Sprintf("%s, max-age=%d", directive, maxAge))
	return r
}

// NoCache disables caching
func (r *Response) NoCache() *Response {
	r.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
	r.Header().Set("Pragma", "no-cache")
	r.Header().Set("Expires", "0")
	return r
}

// Stream streams data from a reader
func (r *Response) Stream(reader io.Reader, contentType string) error {
	if contentType != "" {
		r.Header().Set("Content-Type", contentType)
	}
	r.WriteHeader(r.statusCode)
	_, err := io.Copy(r, reader)
	return err
}

// File sends a file response
func (r *Response) File(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Set content type
	contentType := detectResponseContentType(filepath)
	r.Header().Set("Content-Type", contentType)

	// Set content length
	r.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	// Set last modified
	r.LastModified(stat.ModTime())

	// Handle range requests if supported
	if r.request != nil {
		rangeHeader := r.request.Header.Get("Range")
		if rangeHeader != "" {
			return r.serveFileRange(file, stat, rangeHeader)
		}
	}

	// Serve entire file
	r.WriteHeader(r.statusCode)
	_, err = io.Copy(r, file)
	return err
}

// serveFileRange serves a file with range request support
func (r *Response) serveFileRange(file *os.File, stat os.FileInfo, rangeHeader string) error {
	size := stat.Size()
	ranges, err := parseRangeHeader(rangeHeader, size)
	if err != nil || len(ranges) == 0 {
		r.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		r.statusCode = http.StatusRequestedRangeNotSatisfiable
		r.WriteHeader(r.statusCode)
		return nil
	}

	// Handle single range (simplified)
	ra := ranges[0]
	r.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.end, size))
	r.Header().Set("Content-Length", strconv.FormatInt(ra.length, 10))
	r.statusCode = http.StatusPartialContent

	// Seek to start
	if _, err := file.Seek(ra.start, 0); err != nil {
		return err
	}

	r.WriteHeader(r.statusCode)
	_, err = io.CopyN(r, file, ra.length)
	return err
}

// parseRangeHeader parses a Range header (simplified version)
func parseRangeHeader(s string, size int64) ([]responseByteRange, error) {
	if !strings.HasPrefix(s, "bytes=") {
		return nil, fmt.Errorf("invalid range")
	}

	s = s[6:]
		ranges := []responseByteRange{}

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

			var br responseByteRange
		if start == "" {
			// Suffix range
			if end == "" {
				return nil, fmt.Errorf("invalid range")
			}
			suffix, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, err
			}
			if suffix > size {
				suffix = size
			}
			br.start = size - suffix
			br.end = size - 1
		} else {
			var err error
			br.start, err = strconv.ParseInt(start, 10, 64)
			if err != nil {
				return nil, err
			}
			if br.start < 0 {
				br.start = 0
			}
			if br.start >= size {
				continue
			}

			if end == "" {
				br.end = size - 1
			} else {
				br.end, err = strconv.ParseInt(end, 10, 64)
				if err != nil {
					return nil, err
				}
				if br.end >= size {
					br.end = size - 1
				}
			}
		}

		if br.start > br.end {
			continue
		}

		br.length = br.end - br.start + 1
		ranges = append(ranges, br)
	}

	return ranges, nil
}

// responseByteRange represents a byte range for response
type responseByteRange struct {
	start, end, length int64
}

// Download forces a file download with a specific filename
func (r *Response) Download(filePath string, filename string) error {
	if filename == "" {
		filename = filepath.Base(filePath)
	}

	// Set content disposition
	r.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return r.File(filePath)
}

// detectResponseContentType detects content type from file extension
func detectResponseContentType(filename string) string {
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
