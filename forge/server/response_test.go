package server

import (
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseConstructors(t *testing.T) {
	w := httptest.NewRecorder()

	t.Run("NewResponse", func(t *testing.T) {
		r := NewResponse(w)
		assert.Equal(t, w, r.ResponseWriter)
		assert.Equal(t, http.StatusOK, r.statusCode)
		assert.Nil(t, r.request)
	})

	t.Run("NewResponseWithRequest", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		r := NewResponseWithRequest(w, req)
		assert.Equal(t, w, r.ResponseWriter)
		assert.Equal(t, http.StatusOK, r.statusCode)
		assert.Equal(t, req, r.request)

		// Test Request() and SetRequest()
		r.SetRequest(nil)
		assert.Nil(t, r.Request())
		r.SetRequest(req)
		assert.Equal(t, req, r.Request())
	})
}

func TestResponseContentMethods(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.Status(http.StatusCreated).JSON(map[string]string{"foo": "bar"})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var data map[string]string
		json.Unmarshal(w.Body.Bytes(), &data)
		assert.Equal(t, "bar", data["foo"])
	})

	t.Run("JSONError", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.JSONError("not found", http.StatusNotFound)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)

		var data map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &data)
		assert.True(t, data["error"].(bool))
		assert.Equal(t, "not found", data["message"])
	})

	t.Run("Text", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.Status(http.StatusAccepted).Text("hello world")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("HTML", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.HTML("<h1>hello</h1>")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Equal(t, "<h1>hello</h1>", w.Body.String())
	})

	t.Run("Stream", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		content := "stream content"
		reader := strings.NewReader(content)

		err := r.Status(http.StatusPartialContent).Stream(reader, "text/csv")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusPartialContent, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Equal(t, content, w.Body.String())
	})
}

func TestResponseHeaders(t *testing.T) {
	t.Run("SetHeader and WriteHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		r.SetHeader("X-Custom", "value")
		r.WriteHeader(http.StatusAccepted)

		assert.Equal(t, "value", w.Header().Get("X-Custom"))
		assert.Equal(t, http.StatusAccepted, w.Code)

		// WriteHeader should not change code if already set
		r.WriteHeader(http.StatusOK)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("Redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		r := NewResponseWithRequest(w, req)

		r.Redirect("/login", http.StatusFound)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})

	t.Run("Cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		cookie := &http.Cookie{Name: "session", Value: "123"}
		r.Cookie(cookie)

		assert.Contains(t, w.Header().Get("Set-Cookie"), "session=123")
	})

	t.Run("ETag", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		r.ETag("12345", false)
		assert.Equal(t, "12345", w.Header().Get("ETag"))

		r.ETag("12345", true)
		assert.Equal(t, "W/12345", w.Header().Get("ETag"))
	})

	t.Run("LastModified", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		modTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		r.LastModified(modTime)
		assert.Equal(t, "Sun, 01 Jan 2023 12:00:00 GMT", w.Header().Get("Last-Modified"))
	})

	t.Run("Cache Controls", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		r1 := NewResponse(w1)
		r1.Cache(3600, true)
		assert.Equal(t, "public, max-age=3600", w1.Header().Get("Cache-Control"))

		w2 := httptest.NewRecorder()
		r2 := NewResponse(w2)
		r2.Cache(60, false)
		assert.Equal(t, "private, max-age=60", w2.Header().Get("Cache-Control"))

		w3 := httptest.NewRecorder()
		r3 := NewResponse(w3)
		r3.NoCache()
		assert.Equal(t, "no-store, no-cache, must-revalidate, private", w3.Header().Get("Cache-Control"))
		assert.Equal(t, "no-cache", w3.Header().Get("Pragma"))
		assert.Equal(t, "0", w3.Header().Get("Expires"))
	})
}

func TestResponseFileHandling(t *testing.T) {
	// Create a temporary file
	content := []byte("0123456789")
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	t.Run("File", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.File(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, "10", w.Header().Get("Content-Length"))
		assert.Equal(t, "0123456789", w.Body.String())
	})

	t.Run("File - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.File("non_existent_file.txt")
		assert.Error(t, err)
	})

	t.Run("File - Range Request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Range", "bytes=2-5")

		w := httptest.NewRecorder()
		r := NewResponseWithRequest(w, req)

		err := r.File(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusPartialContent, w.Code)
		assert.Equal(t, "4", w.Header().Get("Content-Length"))
		assert.Equal(t, "bytes 2-5/10", w.Header().Get("Content-Range"))
		assert.Equal(t, "2345", w.Body.String())
	})

	t.Run("File - Invalid Range Request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Range", "bytes=20-30")

		w := httptest.NewRecorder()
		r := NewResponseWithRequest(w, req)

		err := r.File(tmpFile.Name())
		assert.NoError(t, err) // No error returned from serving file, but status is 416
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
		assert.Equal(t, "bytes */10", w.Header().Get("Content-Range"))
	})

	t.Run("Download", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.Download(tmpFile.Name(), "custom.txt")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `attachment; filename="custom.txt"`, w.Header().Get("Content-Disposition"))
		assert.Equal(t, "0123456789", w.Body.String())
	})

	t.Run("Download Default Name", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := NewResponse(w)

		err := r.Download(tmpFile.Name(), "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		expectedDisp := `attachment; filename="` + filepath.Base(tmpFile.Name()) + `"`
		assert.Equal(t, expectedDisp, w.Header().Get("Content-Disposition"))
	})
}

func TestRangeParsing(t *testing.T) {
	size := int64(100)

	tests := []struct {
		header string
		size   int64
		ranges []responseByteRange
		err    bool
	}{
		{"bytes=0-49", size, []responseByteRange{{0, 49, 50}}, false},
		{"bytes=50-99", size, []responseByteRange{{50, 99, 50}}, false},
		{"bytes=-50", size, []responseByteRange{{50, 99, 50}}, false}, // suffix length
		{"bytes=50-", size, []responseByteRange{{50, 99, 50}}, false}, // prefix
		{"bytes=0-0,-1", size, []responseByteRange{{0, 0, 1}, {99, 99, 1}}, false}, // multiple ranges
		{"invalid", size, nil, true}, // invalid format
		{"bytes=a-b", size, nil, true}, // invalid numbers
		{"bytes=200-300", size, nil, false}, // out of bounds (ignored, empty array returned if all invalid)
	}

	for _, tc := range tests {
		ranges, err := parseRangeHeader(tc.header, tc.size)
		if tc.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, len(tc.ranges), len(ranges))
			for i := range ranges {
				if i < len(tc.ranges) {
					assert.Equal(t, tc.ranges[i].start, ranges[i].start)
					assert.Equal(t, tc.ranges[i].end, ranges[i].end)
					assert.Equal(t, tc.ranges[i].length, ranges[i].length)
				}
			}
		}
	}
}

func TestContentTypeDetection(t *testing.T) {
	tests := map[string]string{
		"index.html": "text/html; charset=utf-8",
		"style.css":  "text/css; charset=utf-8",
		"app.js":     "application/javascript; charset=utf-8",
		"data.json":  "application/json; charset=utf-8",
		"data.xml":   "application/xml; charset=utf-8",
		"image.png":  "image/png",
		"image.jpg":  "image/jpeg",
		"image.jpeg": "image/jpeg",
		"image.gif":  "image/gif",
		"image.svg":  "image/svg+xml",
		"icon.ico":   "image/x-icon",
		"doc.pdf":    "application/pdf",
		"file.zip":   "application/zip",
		"note.txt":   "text/plain; charset=utf-8",
		"unknown.xyz":"application/octet-stream",
	}

	for filename, expectedType := range tests {
		assert.Equal(t, expectedType, detectResponseContentType(filename))
	}
}
