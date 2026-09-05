package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/forgego/forge/api/errors"
)

// GetJSON parses JSON from request body
func GetJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

// GetQueryInt gets an integer query parameter
func GetQueryInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetQueryString gets a string query parameter
func GetQueryString(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryBool gets a boolean query parameter
func GetQueryBool(r *http.Request, key string, defaultValue bool) bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// GetParam gets a URL parameter (from chi router)
func GetParam(r *http.Request, key string) string {
	return URLParam(r, key)
}

// SendJSON sends a JSON response
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// SendError sends a JSON error response using RFC 7807 Problem Details
func SendError(w http.ResponseWriter, statusCode int, message string) error {
	errType := errors.ErrorTypeInternal
	if statusCode >= 400 && statusCode < 500 {
		errType = errors.ErrorTypeBadRequest
		if statusCode == http.StatusNotFound {
			errType = errors.ErrorTypeNotFound
		} else if statusCode == http.StatusUnauthorized {
			errType = errors.ErrorTypeAuthentication
		} else if statusCode == http.StatusForbidden {
			errType = errors.ErrorTypeAuthorization
		} else if statusCode == http.StatusMethodNotAllowed {
			errType = errors.ErrorTypeMethodNotAllowed
		} else if statusCode == http.StatusNotAcceptable {
			errType = errors.ErrorTypeNotAcceptable
		} else if statusCode == http.StatusConflict {
			errType = errors.ErrorTypeConflict
		} else if statusCode == http.StatusUnsupportedMediaType {
			errType = errors.ErrorTypeUnsupportedMediaType
		} else if statusCode == http.StatusTooManyRequests {
			errType = errors.ErrorTypeRateLimit
		}
	}

	problem := errors.NewProblem(
		errType,
		strconv.Itoa(statusCode),
		statusCode,
		http.StatusText(statusCode),
		message,
	)

	return problem.WriteJSON(w)
}

// SendSuccess sends a JSON success response
// Note: Consider using a standardized response envelope or just returning data directly
func SendSuccess(w http.ResponseWriter, statusCode int, data interface{}) error {
	return SendJSON(w, statusCode, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

