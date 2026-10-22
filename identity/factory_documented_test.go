package identity

import (
	"testing"

	"github.com/forgego/forge/db"
)

func TestSetupIdentitySystemWiresDocumentedComponents(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	identitySystem, err := SetupIdentitySystem(database, nil)
	if err != nil {
		t.Fatal(err)
	}
	if identitySystem.Config == nil || identitySystem.UserRepo == nil || identitySystem.SessionRepo == nil || identitySystem.TokenRepo == nil || identitySystem.PermissionRepo == nil || identitySystem.GroupRepo == nil || identitySystem.UserService == nil || identitySystem.AuthService == nil || identitySystem.PasswordService == nil || identitySystem.PermissionService == nil || identitySystem.BackendRegistry == nil {
		t.Fatalf("documented identity components were not wired: %+v", identitySystem)
	}
}
