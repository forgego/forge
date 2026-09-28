package core

// SQLTypeOption is the field option under which migration-file parsing records
// a column's SQL type exactly as it was declared. The SQL builder renders a field
// carrying it with that type, so schema state reconstructed from migration files
// renders back to the DDL it was parsed from and regeneration stays stable.
const SQLTypeOption = "sql_type"

// Driver represents a database driver name
type Driver string

const (
	DriverPostgreSQL Driver = "postgres"
	DriverSQLite     Driver = "sqlite"
	DriverSQLite3    Driver = "sqlite3"
)

// IsPostgreSQL returns true if the driver is PostgreSQL
func (d Driver) IsPostgreSQL() bool {
	return d == DriverPostgreSQL || d == "postgresql"
}

// IsSQLite returns true if the driver is SQLite
func (d Driver) IsSQLite() bool {
	return d == DriverSQLite || d == DriverSQLite3
}

// String returns the driver name as a string
func (d Driver) String() string {
	return string(d)
}
