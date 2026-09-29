package state

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
)

// applyRenameTable renames a table, as ALTER TABLE .. RENAME TO does:
// foreign keys of other tables that target it follow it. Its own foreign
// keys are recorded under the names the builder gives them on the new table
// (fk_<table>_<column>), which the migrations guide asks a hand-written
// rename to use as well.
func (s *InMemoryState) applyRenameTable(c *core.RenameTable) {
	table, exists := s.state.Tables[c.OldName]
	if !exists || c.OldName == c.NewName {
		return
	}
	// A rename onto an existing name replaces that table.
	delete(s.state.Tables, c.NewName)
	delete(s.state.Tables, c.OldName)
	table.Name = c.NewName
	s.state.Tables[c.NewName] = table

	foreignKeys := make(map[string]*ForeignKeyState, len(table.ForeignKeys))
	for _, fk := range table.ForeignKeys {
		fk.Name = foreignKeyName(c.NewName, fk.Column)
		foreignKeys[fk.Name] = fk
	}
	table.ForeignKeys = foreignKeys

	for _, other := range s.state.Tables {
		for _, fk := range other.ForeignKeys {
			if fk.TargetTable == c.OldName {
				fk.TargetTable = c.NewName
			}
		}
	}
}

// applyRenameColumn renames a column, as ALTER TABLE .. RENAME COLUMN does:
// the indexes, constraints and foreign key of the table that refer to it
// follow it.
func (s *InMemoryState) applyRenameColumn(c *core.RenameColumn) {
	table, exists := s.state.Tables[c.Table]
	if !exists || c.OldName == c.NewName {
		return
	}
	col, exists := table.Columns[c.OldName]
	if !exists {
		return
	}
	col.Name = c.NewName
	table.Columns[c.NewName] = col
	delete(table.Columns, c.OldName)

	for _, idx := range table.Indexes {
		idx.Fields = renameField(idx.Fields, c.OldName, c.NewName)
	}
	for _, constraint := range table.Constraints {
		constraint.Fields = renameField(constraint.Fields, c.OldName, c.NewName)
		constraint.Condition = renameIdentifier(constraint.Condition, c.OldName, c.NewName)
	}
	// Generated columns that compute from it follow it too.
	for _, other := range table.Columns {
		if expr, ok := other.Options["generated_expr"].(string); ok && expr != "" {
			other.Options["generated_expr"] = renameIdentifier(expr, c.OldName, c.NewName)
		}
	}
	for name, fk := range table.ForeignKeys {
		if fk.Column != c.OldName {
			continue
		}
		delete(table.ForeignKeys, name)
		fk.Column = c.NewName
		fk.Name = foreignKeyName(c.Table, c.NewName)
		table.ForeignKeys[fk.Name] = fk
	}
	for _, other := range s.state.Tables {
		for _, fk := range other.ForeignKeys {
			if fk.TargetTable == c.Table && fk.TargetColumn == c.OldName {
				fk.TargetColumn = c.NewName
			}
		}
	}
}

// foreignKeyName is the name the SQL builder gives a foreign key.
func foreignKeyName(table, column string) string {
	return fmt.Sprintf("fk_%s_%s", table, column)
}

// renameField returns fields with old replaced by new, in a new slice.
func renameField(fields []string, old, new string) []string {
	if fields == nil {
		return nil
	}
	renamed := make([]string, len(fields))
	for i, field := range fields {
		if field == old {
			field = new
		}
		renamed[i] = field
	}
	return renamed
}

// renameIdentifier replaces the identifier old, bare or double-quoted, with
// new in a SQL expression such as a CHECK condition. Single-quoted literals
// and longer identifiers that contain old are left alone.
func renameIdentifier(expr, old, new string) string {
	if !strings.Contains(expr, old) {
		return expr
	}
	var b strings.Builder
	for i := 0; i < len(expr); {
		c := expr[i]
		switch {
		case c == '\'':
			end := i + 1
			for end < len(expr) {
				if expr[end] == '\'' {
					if end+1 < len(expr) && expr[end+1] == '\'' {
						end += 2
						continue
					}
					end++
					break
				}
				end++
			}
			b.WriteString(expr[i:end])
			i = end
		case c == '"' && strings.HasPrefix(expr[i+1:], old+`"`):
			b.WriteString(`"` + new + `"`)
			i += len(old) + 2
		case isIdentifierByte(c):
			end := i
			for end < len(expr) && isIdentifierByte(expr[end]) {
				end++
			}
			if expr[i:end] == old && !notColumnReference(expr, i, end) {
				b.WriteString(new)
			} else {
				b.WriteString(expr[i:end])
			}
			i = end
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

// notColumnReference reports whether the word expr[start:end] is not a
// column name: a type after ::, a function name before (, or the field of
// EXTRACT(field FROM ...).
func notColumnReference(expr string, start, end int) bool {
	before := strings.TrimRight(expr[:start], " \t\r\n")
	if strings.HasSuffix(before, "::") {
		return true
	}
	after := strings.TrimLeft(expr[end:], " \t\r\n")
	if strings.HasPrefix(after, "(") {
		return true
	}
	if strings.HasSuffix(before, "(") && len(after) >= 4 && strings.EqualFold(after[:4], "FROM") &&
		(len(after) == 4 || !isIdentifierByte(after[4])) {
		fn := strings.TrimRight(before[:len(before)-1], " \t\r\n")
		return len(fn) >= 7 && strings.EqualFold(fn[len(fn)-7:], "EXTRACT")
	}
	return false
}

func isIdentifierByte(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
