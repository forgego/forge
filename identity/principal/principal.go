// Package principal defines the identity and authorization attributes carried
// through request contexts.
package principal

import (
	"context"
	"fmt"
	"reflect"

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
	id := backendID(user)
	if id == "" || backendAccountDisabled(user) {
		return Principal{}, false
	}
	p := Principal{ID: id, Kind: KindUser}
	for _, role := range []struct{ name, member string }{{"staff", "IsStaff"}, {"superuser", "IsSuperuser"}, {"admin", "IsAdmin"}} {
		if backendBool(user, role.member) {
			p.RoleNames = append(p.RoleNames, role.name)
		}
	}
	if named, ok := user.(interface{ GetDisplayName() string }); ok {
		p.DisplayName = named.GetDisplayName()
	} else if field := backendField(user, "Username"); field.IsValid() && field.Kind() == reflect.String {
		p.DisplayName = field.String()
	}
	return p, true
}

func backendValue(user interface{}) reflect.Value {
	v := reflect.ValueOf(user)
	if v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func backendField(user interface{}, name string) reflect.Value {
	v := backendValue(user)
	if v.IsValid() && v.Kind() == reflect.Struct {
		return v.FieldByName(name)
	}
	return reflect.Value{}
}

func backendID(user interface{}) string {
	if idProvider, ok := user.(interface{ GetID() string }); ok {
		return idProvider.GetID()
	}
	if idProvider, ok := user.(interface{ GetID() int64 }); ok {
		return fmt.Sprint(idProvider.GetID())
	}
	field := backendField(user, "ID")
	if field.IsValid() && field.CanInterface() {
		return fmt.Sprint(field.Interface())
	}
	return ""
}

func backendAccountDisabled(user interface{}) bool {
	active := backendField(user, "IsActive")
	if active.IsValid() && active.Kind() == reflect.Bool && !active.Bool() {
		return true
	}
	locked := backendField(user, "IsLocked")
	return locked.IsValid() && locked.Kind() == reflect.Bool && locked.Bool()
}

func backendBool(user interface{}, name string) bool {
	v := reflect.ValueOf(user)
	if !v.IsValid() {
		return false
	}
	if m := v.MethodByName(name); m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 1 && m.Type().Out(0).Kind() == reflect.Bool {
		return m.Call(nil)[0].Bool()
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	field := v.FieldByName(name)
	return field.IsValid() && field.Kind() == reflect.Bool && field.Bool()
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
