package errors

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

// Problem implements RFC 7807 Problem Details for HTTP APIs
type Problem struct {
	// Type is a URI reference that identifies the problem type
	Type string `json:"type" xml:"type"`
	// Title is a short, human-readable summary of the problem type
	Title string `json:"title" xml:"title"`
	// Status is the HTTP status code
	Status int `json:"status" xml:"status"`
	// Detail is a human-readable explanation specific to this occurrence
	Detail string `json:"detail,omitempty" xml:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence
	Instance string `json:"instance,omitempty" xml:"instance,omitempty"`
	// Code is an application-level error code (extension to RFC 7807)
	Code string `json:"code,omitempty" xml:"code,omitempty"`
	// Meta contains additional metadata (extension to RFC 7807)
	Meta map[string]interface{} `json:"meta,omitempty" xml:"meta,omitempty"`
	// Errors contains field-level validation errors (extension for validation errors)
	Errors map[string][]FieldError `json:"errors,omitempty" xml:"errors,omitempty"`
}

// FieldError represents a field-level validation error
type FieldError struct {
	Message string `json:"message" xml:"message"`
	Code    string `json:"code,omitempty" xml:"code,omitempty"`
}

// NewProblem creates a new Problem Details structure
func NewProblem(errType ErrorType, code string, status int, title, detail string) *Problem {
	return &Problem{
		Type:   getTypeURI(errType),
		Title:  title,
		Status: status,
		Detail: detail,
		Code:   code,
		Meta:   make(map[string]interface{}),
	}
}

// NewProblemFromError creates a Problem from a ProblemError
func NewProblemFromError(err ProblemError, instance string) *Problem {
	problem := err.ToProblem()
	if instance != "" {
		problem.Instance = instance
	}
	return problem
}

// WithInstance sets the instance URI
func (p *Problem) WithInstance(instance string) *Problem {
	p.Instance = instance
	return p
}

// WithMeta sets metadata
func (p *Problem) WithMeta(key string, value interface{}) *Problem {
	if p.Meta == nil {
		p.Meta = make(map[string]interface{})
	}
	p.Meta[key] = value
	return p
}

// WithMetaMap sets multiple metadata entries
func (p *Problem) WithMetaMap(meta map[string]interface{}) *Problem {
	if p.Meta == nil {
		p.Meta = make(map[string]interface{})
	}
	for k, v := range meta {
		p.Meta[k] = v
	}
	return p
}

// WithFieldError adds a field-level validation error
func (p *Problem) WithFieldError(field, message, code string) *Problem {
	if p.Errors == nil {
		p.Errors = make(map[string][]FieldError)
	}
	p.Errors[field] = append(p.Errors[field], FieldError{
		Message: message,
		Code:    code,
	})
	return p
}

// WithFieldErrors sets multiple field-level errors
func (p *Problem) WithFieldErrors(errors map[string][]FieldError) *Problem {
	if p.Errors == nil {
		p.Errors = make(map[string][]FieldError)
	}
	for field, fieldErrors := range errors {
		p.Errors[field] = append(p.Errors[field], fieldErrors...)
	}
	return p
}

// getTypeURI generates the type URI for the problem
func getTypeURI(errType ErrorType) string {
	baseURL := GetTypeBaseURL()
	return fmt.Sprintf("%s/%s", baseURL, string(errType))
}

// WriteJSON writes the problem as JSON to the HTTP response
func (p *Problem) WriteJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	return json.NewEncoder(w).Encode(p)
}

// WriteXML writes the problem as XML to the HTTP response
func (p *Problem) WriteXML(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/problem+xml")
	w.WriteHeader(p.Status)
	w.Write([]byte(xml.Header))
	return xml.NewEncoder(w).Encode(p)
}

// Write writes the problem to the HTTP response based on Accept header
func (p *Problem) Write(w http.ResponseWriter, r *http.Request) error {
	accept := r.Header.Get("Accept")

	// Default to JSON, but support XML if requested
	if accept == "application/xml" || accept == "text/xml" || accept == "application/problem+xml" {
		return p.WriteXML(w)
	}

	// Default to JSON (including application/problem+json)
	return p.WriteJSON(w)
}

// MarshalJSON implements json.Marshaler
func (p *Problem) MarshalJSON() ([]byte, error) {
	type Alias Problem
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	})
}

// String returns a string representation of the problem
func (p *Problem) String() string {
	if p.Detail != "" {
		return fmt.Sprintf("%s: %s", p.Title, p.Detail)
	}
	return p.Title
}

// Error implements the error interface
func (p *Problem) Error() string {
	return p.String()
}
