// Package deleteerrors maps delete restrictions to a client-safe HTTP failure.
package deleteerrors

import (
	"net/http"

	"github.com/forgego/forge/internal/dberrors"
)

// Protection returns the shared response for a database-enforced protected
// reference. Driver details may contain table names and values; never expose them.
func Protection(err error) (status int, code, message string, ok bool) {
	violation, recognized := dberrors.Classify(err)
	if !recognized || violation.Kind != dberrors.ForeignKey {
		return 0, "", "", false
	}
	return http.StatusBadRequest, "in_use", "Other records still refer to this record, so it cannot be deleted.", true
}
