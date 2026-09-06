package server

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestETag(t *testing.T) {
	testBody := []byte("hello world")
	expectedHash := sha256.Sum256(testBody)
	expectedETag := fmt.Sprintf(`"%x"`, expectedHash)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(testBody)
	})

	tests := []struct {
		name           string
		opts           *ETagOptions
		ifNoneMatch    string
		expectedStatus int
		expectedETag   string
		expectedBody   []byte
	}{
		{
			name:           "Strong ETag",
			opts:           nil,
			ifNoneMatch:    "",
			expectedStatus: http.StatusOK,
			expectedETag:   expectedETag,
			expectedBody:   testBody,
		},
		{
			name:           "Weak ETag",
			opts:           &ETagOptions{Weak: true},
			ifNoneMatch:    "",
			expectedStatus: http.StatusOK,
			expectedETag:   "W/" + expectedETag,
			expectedBody:   testBody,
		},
		{
			name:           "If-None-Match matches strong ETag",
			opts:           nil,
			ifNoneMatch:    expectedETag,
			expectedStatus: http.StatusNotModified,
			expectedETag:   expectedETag,
			expectedBody:   []byte{},
		},
		{
			name:           "If-None-Match matches weak ETag",
			opts:           &ETagOptions{Weak: true},
			ifNoneMatch:    "W/" + expectedETag,
			expectedStatus: http.StatusNotModified,
			expectedETag:   "W/" + expectedETag,
			expectedBody:   []byte{},
		},
		{
			name:           "If-None-Match matches strong ETag with weak header",
			opts:           nil,
			ifNoneMatch:    "W/" + expectedETag,
			expectedStatus: http.StatusNotModified,
			expectedETag:   expectedETag,
			expectedBody:   []byte{},
		},
		{
			name:           "If-None-Match matches weak ETag with strong header",
			opts:           &ETagOptions{Weak: true},
			ifNoneMatch:    expectedETag,
			expectedStatus: http.StatusNotModified,
			expectedETag:   "W/" + expectedETag,
			expectedBody:   []byte{},
		},
		{
			name:           "If-None-Match does not match",
			opts:           nil,
			ifNoneMatch:    `"different"`,
			expectedStatus: http.StatusOK,
			expectedETag:   expectedETag,
			expectedBody:   testBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := ETag(tt.opts)
			req := httptest.NewRequest("GET", "/", nil)
			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}
			rec := httptest.NewRecorder()

			middleware(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Header().Get("ETag") != tt.expectedETag {
				t.Errorf("Expected ETag %s, got %s", tt.expectedETag, rec.Header().Get("ETag"))
			}

			if !bytes.Equal(rec.Body.Bytes(), tt.expectedBody) {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rec.Body.Bytes())
			}
		})
	}
}
