// Package principal defines the identity and authorization attributes carried
// through request contexts.
package principal

import (
	"context"
	"fmt"

	"github.com/forgego/forge/identity/models"
)

// Kind identifies the source and trust category of a principal.
type Kind string

const (
	KindUser      Kind = "user"
	KindService   Kind = "service"
	KindAnonymous Kind = "anonymous"
	KindSystem    Kind = "system"
)

type contextKey struct{}

// Principal represents the actor associated with an operation. Roles and
// permissions are snapshots supplied by the principal's trusted adapter.
type Principal struct {
	ID              string
	Kind            Kind
	DisplayName     string
	RoleNames       []string
	PermissionNames []string
	accountStateErr error
}

// Roles returns a copy of the principal's roles.
func (p Principal) Roles() []string { return append([]string(nil), p.RoleNames...) }

// Permissions returns a copy of the principal's permissions.
func (p Principal) Permissions() []string {
	return append([]string(nil), p.PermissionNames...)
}

// AccountStateError reports whether a user principal's account can authenticate.
func (p Principal) AccountStateError() error { return p.accountStateErr }

func accountStateError(user interface{}) error {
	if state, ok := user.(interface{ AccountStateError() error }); ok {
		return state.AccountStateError()
	}
	return nil
}

// FromUser adapts an active models.User into a user principal.
func FromUser(user *models.User) (Principal, error) {
	if err := accountStateError(user); err != nil {
		return Principal{}, err
	}
	roles := make([]string, 0, 2)
	if user.IsStaff {
		roles = append(roles, "staff")
	}
	if user.IsSuperuser {
		roles = append(roles, "superuser")
	}
	return Principal{ID: fmt.Sprint(user.ID), Kind: KindUser, DisplayName: user.GetDisplayName(), RoleNames: roles}, nil
}

// FromBackend adapts common custom authentication backend values. Values that
// do not expose a non-empty GetID and optional GetDisplayName are ignored.
func FromBackend(user interface{}) (Principal, bool) {
	if accountStateError(user) != nil {
		return Principal{}, false
	}
	idProvider, ok := user.(interface{ GetID() string })
	if !ok || idProvider.GetID() == "" {
		return Principal{}, false
	}
	p := Principal{ID: idProvider.GetID(), Kind: KindUser}
	if named, ok := user.(interface{ GetDisplayName() string }); ok {
		p.DisplayName = named.GetDisplayName()
	}
	return p, true
}

// Anonymous creates an anonymous principal.
func Anonymous() Principal {
	return Principal{Kind: KindAnonymous, ID: "anonymous", DisplayName: "Anonymous"}
}

// System creates a system principal with the supplied identifier.
func System(id string) Principal { return Principal{Kind: KindSystem, ID: id, DisplayName: id} }

// WithPrincipal returns a context containing p. Invalid principals are stored
// as anonymous to avoid propagating an unusable identity.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	if p.ID == "" || p.Kind == "" {
		p = Anonymous()
	}
	return context.WithValue(ctx, contextKey{}, p)
}

// PrincipalFromContext retrieves a principal from ctx.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}
