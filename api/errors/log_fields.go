package errors

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// errorLogFields describes err for the server log. A database driver error
// is logged by its type, code and the constraint, table and column it
// names, never by its message or detail: PostgreSQL puts row values in both
// (for example "Key (email)=(alice@example.com) already exists" or
// `invalid input syntax for type integer: "abc"`). Any other error is
// logged with its message.
func errorLogFields(err error) []zap.Field {
	if err == nil {
		return nil
	}
	errType := fmt.Sprintf("%T", err)

	var pqErr *pq.Error
	if stderrors.As(err, &pqErr) {
		fields := []zap.Field{
			zap.String("error_class", errType),
			zap.String("db_driver", "postgres"),
			zap.String("sqlstate", string(pqErr.Code)),
			zap.String("sqlstate_name", pqErr.Code.Name()),
		}
		for _, f := range []struct{ key, val string }{
			{"db_schema", pqErr.Schema},
			{"db_table", pqErr.Table},
			{"db_column", pqErr.Column},
			{"db_constraint", pqErr.Constraint},
			{"db_data_type", pqErr.DataTypeName},
		} {
			if f.val != "" {
				fields = append(fields, zap.String(f.key, sanitizeLogString(f.val)))
			}
		}
		return fields
	}

	var sqliteErr sqlite3.Error
	if stderrors.As(err, &sqliteErr) {
		fields := []zap.Field{
			zap.String("error_class", errType),
			zap.String("db_driver", "sqlite3"),
			zap.Int("sqlite_code", int(sqliteErr.Code)),
			zap.Int("sqlite_extended_code", int(sqliteErr.ExtendedCode)),
			zap.String("sqlite_code_name", sqliteErr.Code.Error()),
		}
		if target := sqliteConstraintTarget(sqliteErr); target != "" {
			fields = append(fields, zap.String("db_constraint", sanitizeLogString(target)))
		}
		return fields
	}

	return []zap.Field{
		zap.String("error_class", errType),
		zap.String("error", sanitizeLogString(err.Error())),
	}
}

// sqliteConstraintTarget returns what a SQLite constraint error names, such
// as "users.email" in "UNIQUE constraint failed: users.email". SQLite puts
// column or constraint names there, not values.
func sqliteConstraintTarget(err sqlite3.Error) string {
	if err.Code != sqlite3.ErrConstraint {
		return ""
	}
	_, target, ok := strings.Cut(err.Error(), "constraint failed: ")
	if !ok {
		return ""
	}
	return strings.TrimSpace(target)
}
