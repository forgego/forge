package db

import (
	"strconv"
	"strings"
)

// dsnValueEscaper escapes the two characters that are special inside a
// single-quoted libpq keyword/value connection string.
var dsnValueEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)

// quoteDSNValue quotes v for a libpq keyword/value connection string, so a
// value containing spaces, quotes or backslashes stays one value.
func quoteDSNValue(v string) string {
	return "'" + dsnValueEscaper.Replace(v) + "'"
}

// postgresKeywordDSN builds a libpq keyword/value connection string from
// the database.* settings. Every value is quoted and escaped. An empty
// password is omitted so PGPASSWORD or a pgpass file can supply it.
func postgresKeywordDSN(host string, port int, user, password, name, sslmode string) string {
	parts := []string{
		"host=" + quoteDSNValue(host),
		"port=" + quoteDSNValue(strconv.Itoa(port)),
		"user=" + quoteDSNValue(user),
	}
	if password != "" {
		parts = append(parts, "password="+quoteDSNValue(password))
	}
	parts = append(parts,
		"dbname="+quoteDSNValue(name),
		"sslmode="+quoteDSNValue(sslmode),
	)
	return strings.Join(parts, " ")
}

// DetectDriverFromDSN returns "postgres" or "sqlite3".
// Rules, applied in order:
//
//	a. trimmed dsn starts with "postgres://" or "postgresql://" -> "postgres"
//	b. starts with "file:" OR equals ":memory:" OR starts with ":memory:" -> "sqlite3"
//	c. strip any "?query" part; if the remaining path ends with ".db", ".sqlite" or ".sqlite3" (case-insensitive) -> "sqlite3"
//	d. split on whitespace; if ANY whole field has prefix "host=", "dbname=", "user=", "sslmode=" or "port=" -> "postgres"
//	e. otherwise -> "postgres"
func DetectDriverFromDSN(dsn string) string {
	trimmed := strings.TrimSpace(dsn)

	// a. trimmed dsn starts with "postgres://" or "postgresql://" -> "postgres"
	if strings.HasPrefix(trimmed, "postgres://") || strings.HasPrefix(trimmed, "postgresql://") {
		return "postgres"
	}

	// b. starts with "file:" OR equals ":memory:" OR starts with ":memory:" -> "sqlite3"
	if strings.HasPrefix(trimmed, "file:") || strings.HasPrefix(trimmed, ":memory:") {
		return "sqlite3"
	}

	// c. strip any "?query" part; if the remaining path ends with ".db", ".sqlite" or ".sqlite3" (case-insensitive) -> "sqlite3"
	pathPart := trimmed
	if idx := strings.Index(pathPart, "?"); idx != -1 {
		pathPart = pathPart[:idx]
	}
	lowerPath := strings.ToLower(pathPart)
	if strings.HasSuffix(lowerPath, ".db") || strings.HasSuffix(lowerPath, ".sqlite") || strings.HasSuffix(lowerPath, ".sqlite3") {
		return "sqlite3"
	}

	// d. split on whitespace; if ANY whole field has prefix "host=", "dbname=", "user=", "sslmode=" or "port=" -> "postgres"
	for _, field := range strings.Fields(trimmed) {
		if strings.HasPrefix(field, "host=") ||
			strings.HasPrefix(field, "dbname=") ||
			strings.HasPrefix(field, "user=") ||
			strings.HasPrefix(field, "sslmode=") ||
			strings.HasPrefix(field, "port=") {
			return "postgres"
		}
	}

	// e. otherwise -> "postgres"
	return "postgres"
}
