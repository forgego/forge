package identity

import (
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/principal"
)

// PrincipalFromUser adapts an active user model to a principal.
func PrincipalFromUser(user *models.User) (principal.Principal, error) {
	return principal.FromUser(user)
}
