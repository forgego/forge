package rest

import (
	"encoding/json"
	"net/http"

	"github.com/forgego/forge/admin/core"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code string, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(core.ErrorResponse{
		Error: core.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
