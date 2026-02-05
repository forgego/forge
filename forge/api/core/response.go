package core

import (
	"encoding/json"
	"net/http"
)

// Response wraps http.ResponseWriter with API-specific functionality
type Response struct {
	http.ResponseWriter
	statusCode int
	headers    map[string]string
}

// NewResponse creates a new API response
func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		headers:        make(map[string]string),
	}
}

// Status sets the HTTP status code
func (r *Response) Status(code int) *Response {
	r.statusCode = code
	return r
}

// Header sets a response header
func (r *Response) Header(key, value string) *Response {
	r.headers[key] = value
	return r
}

// JSON sends a JSON response
func (r *Response) JSON(data interface{}) error {
	// Set headers
	r.ResponseWriter.Header().Set("Content-Type", "application/json")
	for k, v := range r.headers {
		r.ResponseWriter.Header().Set(k, v)
	}

	// Write status
	r.ResponseWriter.WriteHeader(r.statusCode)

	// Encode JSON
	return json.NewEncoder(r.ResponseWriter).Encode(data)
}

// Error sends an error response
func (r *Response) Error(code int, message string, details interface{}) error {
	return r.Status(code).JSON(map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    code,
		"details": details,
	})
}

// WriteHeader writes the status code
func (r *Response) WriteHeader(code int) {
	if r.statusCode == 0 {
		r.statusCode = code
	}
	r.ResponseWriter.WriteHeader(r.statusCode)
}
