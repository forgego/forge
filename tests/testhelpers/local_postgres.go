package testhelpers

import (
	"fmt"
	"time"
)

// LocalPostgresOpts returns PostgreSQL options for local database
// Uses localhost with user "postgres" and password "123"
// Creates a unique database name for each test
func LocalPostgresOpts(testName string) PostgresOpts {
	return PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", sanitizeTestName(testName), time.Now().UnixNano()),
	}
}
