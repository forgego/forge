package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// awkwardPassword contains every character that broke the unquoted
// keyword/value DSN: a space, a single quote and a backslash (#290).
const awkwardPassword = `p a'ss\word`

func TestPostgresKeywordDSN_QuotesValues(t *testing.T) {
	dsn := PostgresKeywordDSN("db.internal", 6543, "app user", awkwardPassword, "my db", "verify-full")

	cfg, err := pq.NewConfig(dsn)
	require.NoError(t, err, "dsn %q must parse", dsn)
	require.Equal(t, "db.internal", cfg.Host)
	require.Equal(t, uint16(6543), cfg.Port)
	require.Equal(t, "app user", cfg.User)
	require.Equal(t, awkwardPassword, cfg.Password)
	require.Equal(t, "my db", cfg.Database)
	require.Equal(t, pq.SSLModeVerifyFull, cfg.SSLMode)
	require.Equal(t, "postgres", DetectDriverFromDSN(dsn))
}

func TestPostgresKeywordDSN_EmptyPasswordOmitted(t *testing.T) {
	dsn := PostgresKeywordDSN("localhost", 5432, "postgres", "", "forge", "disable")
	require.NotContains(t, dsn, "password=")

	cfg, err := pq.NewConfig(dsn)
	require.NoError(t, err)
	require.Equal(t, "forge", cfg.Database, "an empty password must not swallow the next key")
	require.Equal(t, "postgres", cfg.User)
}

// TestNewDBFromConfig_PasswordWithSpecialCharacters logs in to the local test
// PostgreSQL as a scratch role whose password needs quoting.
func TestNewDBFromConfig_PasswordWithSpecialCharacters(t *testing.T) {
	rawURL := os.Getenv("FORGE_TEST_DATABASE_URL")
	if rawURL == "" {
		t.Skip("FORGE_TEST_DATABASE_URL not set")
	}
	u, err := url.Parse(rawURL)
	require.NoError(t, err)

	admin, err := sql.Open("postgres", rawURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	if err := admin.Ping(); err != nil {
		t.Skipf("test database unreachable: %v", err)
	}

	role := fmt.Sprintf("forge_dsn_test_%d", os.Getpid())
	_, _ = admin.Exec("DROP ROLE IF EXISTS " + pq.QuoteIdentifier(role))
	if _, err := admin.Exec(fmt.Sprintf("CREATE ROLE %s LOGIN PASSWORD %s",
		pq.QuoteIdentifier(role), pq.QuoteLiteral(awkwardPassword))); err != nil {
		t.Skipf("cannot create a scratch role: %v", err)
	}
	t.Cleanup(func() { _, _ = admin.Exec("DROP ROLE IF EXISTS " + pq.QuoteIdentifier(role)) })

	port := 5432
	if p := u.Port(); p != "" {
		port, err = strconv.Atoi(p)
		require.NoError(t, err)
	}
	sslmode := u.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}

	cfg := config.NewConfig()
	cfg.Set("database.driver", "postgres")
	cfg.Set("database.host", u.Hostname())
	cfg.Set("database.port", port)
	cfg.Set("database.user", role)
	cfg.Set("database.password", awkwardPassword)
	cfg.Set("database.name", u.Path[1:])
	cfg.Set("database.sslmode", sslmode)

	database, err := NewDBFromConfig(cfg)
	require.NoError(t, err)
	defer database.Close()

	var current string
	require.NoError(t, database.QueryRow("SELECT current_user").Scan(&current))
	require.Equal(t, role, current)
}
