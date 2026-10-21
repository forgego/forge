package admin

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/utils"
)

// createsuperuser persists an active staff superuser with a hashed password.
func TestCreateSuperuserCreatesAdminAccount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.sqlite")
	database, err := db.NewDBWithDriver("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE NOT NULL, email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL, first_name TEXT, last_name TEXT, is_active BOOLEAN, is_staff BOOLEAN,
		is_superuser BOOLEAN, is_locked BOOLEAN, email_verified BOOLEAN, phone_number TEXT,
		phone_verified BOOLEAN, timezone TEXT, locale TEXT, language TEXT, bio TEXT, website TEXT,
		location TEXT, avatar TEXT, password_changed_at DATETIME, password_expires_at DATETIME,
		must_change_password BOOLEAN DEFAULT FALSE, locked_at DATETIME, locked_reason TEXT DEFAULT '', failed_login_attempts INTEGER DEFAULT 0,
		last_failed_login_at DATETIME, email_verified_at DATETIME, date_joined DATETIME, last_login DATETIME,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.name", path)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(w, "owner\nowner@example.invalid\nStrongPass123!\nStrongPass123!\n"); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old; _ = r.Close() })
	if err := NewCreateSuperUserCommand().Execute(&core.Context{Config: cfg}, nil); err != nil {
		t.Fatal(err)
	}
	database, err = db.NewDBWithDriver("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	var username, hash string
	var active, staff, superuser bool
	if err := database.QueryRow("SELECT username,password,is_active,is_staff,is_superuser FROM users WHERE email=?", "owner@example.invalid").Scan(&username, &hash, &active, &staff, &superuser); err != nil {
		t.Fatal(err)
	}
	if username != "owner" || !active || !staff || !superuser || !utils.CheckPassword("StrongPass123!", hash) {
		t.Fatalf("account: username=%q active=%t staff=%t superuser=%t hash_valid=%t", username, active, staff, superuser, utils.CheckPassword("StrongPass123!", hash))
	}
}
