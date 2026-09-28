package core

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
