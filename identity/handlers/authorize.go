package handlers

import (
	"net/http"

	"github.com/forgego/forge/identity/models"
	forgehttp "github.com/forgego/forge/server"
)

type userAccess int

const (
	accessSelf userAccess = iota
	accessList
	accessRead
	accessUpdate
	accessAdmin
)

// authorizeUser applies the actor/target policy for identity handlers.
func authorizeUser(w http.ResponseWriter, r *http.Request, targetID int64, access userAccess) bool {
	actor, ok := GetUserFromContext(r.Context())
	if !ok || actor == nil || actor.ID == 0 {
		forgehttp.SendError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}
	if canAccessUser(actor, targetID, access) {
		return true
	}
	forgehttp.SendError(w, http.StatusForbidden, "Permission denied")
	return false
}

func canAccessUser(actor *models.User, targetID int64, access userAccess) bool {
	switch access {
	case accessSelf:
		return actor.ID == targetID
	case accessList:
		return actor.IsStaff || actor.IsSuperuser
	case accessRead:
		return actor.ID == targetID || actor.IsStaff || actor.IsSuperuser
	case accessUpdate:
		return actor.ID == targetID || actor.IsSuperuser
	case accessAdmin:
		return actor.IsSuperuser
	default:
		return false
	}
}
