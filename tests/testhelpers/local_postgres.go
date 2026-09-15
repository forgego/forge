package testhelpers

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func testDatabaseURL(defaultURL string) string {
	if u := os.Getenv("FORGE_TEST_DATABASE_URL"); u != "" {
		return u
	}
	return defaultURL
}

func requireDB() bool {
	return os.Getenv("FORGE_REQUIRE_DB") == "1"
}

func dbUnavailableAction() string {
	if requireDB() {
		return "fatal"
	}
	return "skip"
}

func skipOrFailNoDB(t testing.TB, format string, args ...any) {
	t.Helper()
	msg := fmt.Sprintf(format, args...)
	if requireDB() {
		t.Fatalf("%s (FORGE_REQUIRE_DB=1 is set)", msg)
		return
	}
	t.Skipf("%s (set FORGE_REQUIRE_DB=1 to turn into failure)", msg)
}

// SkipOrFailNoDB calls t.Fatalf if FORGE_REQUIRE_DB=1 is set, or t.Skipf otherwise.
func SkipOrFailNoDB(t testing.TB, format string, args ...any) {
	t.Helper()
	skipOrFailNoDB(t, format, args...)
}

// RequirePostgresURL returns the PostgreSQL connection URL to use:
// when FORGE_TEST_DATABASE_URL is set, it returns it;
// when FORGE_TEST_DATABASE_URL is not set and DATABASE_URL is set, it returns DATABASE_URL;
// otherwise it calls the testhelpers skip-or-fail helper (which fails if FORGE_REQUIRE_DB=1, skips otherwise).
func RequirePostgresURL(t testing.TB) string {
	t.Helper()
	if u := os.Getenv("FORGE_TEST_DATABASE_URL"); u != "" {
		return u
	}
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	skipOrFailNoDB(t, "Postgres not available")
	return ""
}

// applyPostgresPrecedence applies connection options precedence to opts:
// When FORGE_TEST_DATABASE_URL is set, its user, password, host, port and query parameters
// win over POSTGRES_* env vars and built-in defaults; only the database name may differ
// (per-test databases). Without the variable, today's behaviour is preserved using
// POSTGRES_* env vars and built-in defaults.
func applyPostgresPrecedence(opts *PostgresOpts) {
	if raw := os.Getenv("FORGE_TEST_DATABASE_URL"); raw != "" {
		u, err := url.Parse(raw)
		if err == nil {
			if h := u.Hostname(); h != "" {
				opts.Host = h
			} else {
				opts.Host = "127.0.0.1"
			}
			if p := u.Port(); p != "" {
				opts.Port = p
			} else {
				opts.Port = "5432"
			}
			if u.User != nil {
				opts.User = u.User.Username()
				if pass, ok := u.User.Password(); ok {
					opts.Password = pass
				} else {
					opts.Password = ""
				}
			} else {
				opts.User = ""
				opts.Password = ""
			}
			opts.RawQuery = u.RawQuery
			opts.UseDirect = true

			if opts.DBName == "" {
				if db := strings.TrimPrefix(u.Path, "/"); db != "" {
					opts.DBName = db
				} else {
					opts.DBName = "testdb"
				}
			}
			opts.DBName = strings.ToLower(opts.DBName)
			opts.DBName = truncateDBName(opts.DBName)
			return
		}
	}

	// Without FORGE_TEST_DATABASE_URL, keep today's behaviour:
	if opts.Host == "" {
		if envHost := os.Getenv("POSTGRES_HOST"); envHost != "" {
			opts.Host = envHost
		} else {
			opts.Host = "127.0.0.1"
		}
	}
	if opts.Port == "" {
		if envPort := os.Getenv("POSTGRES_PORT"); envPort != "" {
			opts.Port = envPort
		} else {
			opts.Port = "5432"
		}
	}
	if opts.User == "" {
		if envUser := os.Getenv("POSTGRES_USER"); envUser != "" {
			opts.User = envUser
		} else {
			opts.User = "postgres"
		}
	}
	if opts.Password == "" {
		if envPass := os.Getenv("POSTGRES_PASSWORD"); envPass != "" {
			opts.Password = envPass
		} else {
			opts.Password = "123"
		}
	}
	if opts.DBName == "" {
		if envDB := os.Getenv("POSTGRES_DB"); envDB != "" {
			opts.DBName = envDB
		} else {
			opts.DBName = "testdb"
		}
	}
	opts.DBName = strings.ToLower(opts.DBName)
	opts.DBName = truncateDBName(opts.DBName)
}

func applyDatabaseURLToOpts(opts *PostgresOpts) {
	applyPostgresPrecedence(opts)
}

// defaultAdminDSN returns a DSN for connecting to the administrative/default Postgres DB
// to create or drop per-test databases, respecting FORGE_TEST_DATABASE_URL credentials.
func defaultAdminDSN(opts PostgresOpts) string {
	applyPostgresPrecedence(&opts)
	query := "sslmode=disable"
	if opts.RawQuery != "" {
		query = opts.RawQuery
	}
	var authority string
	if opts.User != "" {
		if opts.Password != "" {
			authority = fmt.Sprintf("%s:%s@%s:%s", opts.User, opts.Password, opts.Host, opts.Port)
		} else {
			authority = fmt.Sprintf("%s@%s:%s", opts.User, opts.Host, opts.Port)
		}
	} else {
		authority = fmt.Sprintf("%s:%s", opts.Host, opts.Port)
	}
	defaultURL := fmt.Sprintf("postgres://%s/postgres?%s", authority, query)

	if raw := os.Getenv("FORGE_TEST_DATABASE_URL"); raw != "" {
		if u, err := url.Parse(raw); err == nil {
			db := strings.TrimPrefix(u.Path, "/")
			if db == "" || db == opts.DBName {
				u.Path = "/postgres"
				return u.String()
			}
		}
		return raw
	}
	return defaultURL
}

// LocalPostgresOpts returns PostgreSQL options for local database
// Uses localhost with user "postgres" and password "123"
// Creates a unique database name for each test
func LocalPostgresOpts(testName string) PostgresOpts {
	opts := PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", sanitizeTestName(testName), time.Now().UnixNano()),
	}
	applyPostgresPrecedence(&opts)
	return opts
}
