package parse

import "github.com/forgego/forge/db/migrate/core"

// DestructiveOperations returns data-deleting operations in migration SQL.
func DestructiveOperations(sql string) ([]string, error) {
	statements, err := NewLexer().Scan(sql)
	if err != nil {
		return nil, err
	}
	var operations []string
	parser := NewSQLParser()
	for _, statement := range statements {
		changes, err := parser.ParseUpSQL(statement.Text)
		if err != nil {
			return nil, err
		}
		if len(changes) == 0 {
			changes = []core.Change{&core.UnknownChange{SQL: statement.Text}}
		}
		for _, change := range changes {
			if name := core.DestructiveOperation(change); name != "" {
				operations = append(operations, name)
			}
		}
	}
	return operations, nil
}
