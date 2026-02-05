package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// APIClient is a test client for API testing
type APIClient struct {
	Handler http.Handler
	BaseURL string
	Headers map[string]string
}

// NewAPIClient creates a new API test client
func NewAPIClient(handler http.Handler) *APIClient {
	return &APIClient{
		Handler: handler,
		BaseURL: "",
		Headers: make(map[string]string),
	}
}

// SetHeader sets a header for all requests
func (c *APIClient) SetHeader(key, value string) {
	c.Headers[key] = value
}

// SetAuth sets authentication header
func (c *APIClient) SetAuth(token string) {
	c.SetHeader("Authorization", "Token "+token)
}

// Get performs a GET request
func (c *APIClient) Get(path string) *Response {
	return c.Request("GET", path, nil)
}

// Post performs a POST request
func (c *APIClient) Post(path string, data interface{}) *Response {
	return c.Request("POST", path, data)
}

// Put performs a PUT request
func (c *APIClient) Put(path string, data interface{}) *Response {
	return c.Request("PUT", path, data)
}

// Patch performs a PATCH request
func (c *APIClient) Patch(path string, data interface{}) *Response {
	return c.Request("PATCH", path, data)
}

// Delete performs a DELETE request
func (c *APIClient) Delete(path string) *Response {
	return c.Request("DELETE", path, nil)
}

// Request performs an HTTP request
func (c *APIClient) Request(method, path string, data interface{}) *Response {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return &Response{Error: err}
		}
		body = bytes.NewBuffer(jsonData)
	}

	req := httptest.NewRequest(method, path, body)

	// Set headers
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// Set Content-Type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	c.Handler.ServeHTTP(recorder, req)

	return &Response{
		StatusCode: recorder.Code,
		Body:       recorder.Body.Bytes(),
		Headers:    recorder.Header(),
	}
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Error      error
}

// JSON parses the response body as JSON
func (r *Response) JSON() map[string]interface{} {
	var data map[string]interface{}
	if r.Error != nil {
		return nil
	}
	json.Unmarshal(r.Body, &data)
	return data
}

// Status returns the status code
func (r *Response) Status() int {
	return r.StatusCode
}
