package handlers

import (
	"errors"
	"net/http"

	forgehttp "github.com/forgego/forge/server"
)

type publicServiceError struct {
	err     error
	status  int
	message string
}

// sendServiceError only exposes explicitly classified errors to clients.
func sendServiceError(w http.ResponseWriter, err error, known ...publicServiceError) {
	for _, item := range known {
		if errors.Is(err, item.err) {
			forgehttp.SendError(w, item.status, item.message)
			return
		}
	}
	forgehttp.SendError(w, http.StatusInternalServerError, "Internal server error")
}
