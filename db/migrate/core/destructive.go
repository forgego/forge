package core

import (
	"fmt"
	"regexp"
	"strings"
)

var destructiveSQL = regexp.MustCompile(`(?is)^\s*(?:--[^\n]*\n\s*)*(DROP\s+(?:DATABASE|SCHEMA|TABLE)|TRUNCATE(?:\s+TABLE)?|DELETE\s+FROM)\b`)
var dropColumnSQL = regexp.MustCompile(`(?is)^\s*(?:--[^\n]*\n\s*)*ALTER\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:ONLY\s+)?\S+\s+(DROP\s+COLUMN)\b`)

// DestructiveOperation names an operation that can delete stored data.
// An empty result means the change does not directly delete data.
func DestructiveOperation(change Change) string {
	switch c := change.(type) {
	case *DropTable:
		return fmt.Sprintf("DropTable %s", c.Table)
	case *DropColumn:
		return fmt.Sprintf("DropColumn %s.%s", c.Table, c.ColumnName)
	case *RunSQL:
		return destructiveSQLName(c.SQL)
	case *UnknownChange:
		return destructiveSQLName(c.SQL)
	default:
		return ""
	}
}

func destructiveSQLName(sql string) string {
	match := destructiveSQL.FindStringSubmatch(sql)
	if len(match) == 0 {
		match = dropColumnSQL.FindStringSubmatch(sql)
		if len(match) == 0 {
			return ""
		}
	}
	return strings.Join(strings.Fields(strings.ToUpper(match[1])), " ")
}
