// Package migrate provides a comprehensive migration system for database schema management.
//
// The migrate package is organized into workflow-oriented sub-packages:
//
//   - generate: Migration generation from model definitions
//   - apply: Migration execution and application
//   - schema: Schema state management
//   - sqlparse: SQL parsing for state reconstruction
//   - sqlgen: SQL generation from changes
//   - verify: Validation and verification
//
// The root package provides the public API facade and core types.
package migrate
