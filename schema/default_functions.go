package schema

import "strings"

// databaseFunctionDefaults are the Default strings that name a database
// function instead of a literal value: the migration builder writes them
// unquoted, so the database evaluates them for each row. Any other SQL
// expression belongs in DBDefault.
var databaseFunctionDefaults = []string{
	"now()", "CURRENT_TIMESTAMP", "CURRENT_DATE", "CURRENT_TIME",
	"uuid_generate_v4()", "gen_random_uuid()", "random()",
}

// IsDatabaseFunctionDefault reports whether a field's Default is one of the
// database functions the migration builder leaves for the database to
// evaluate, such as now() or gen_random_uuid(). Such a default is never
// assigned in Go (by the ORM's ApplyDefaults, the REST API or the admin): the
// column is left out of the INSERT so the database fills it.
func IsDatabaseFunctionDefault(def interface{}) bool {
	s, ok := def.(string)
	if !ok {
		return false
	}
	for _, fn := range databaseFunctionDefaults {
		if strings.EqualFold(strings.TrimSpace(s), fn) {
			return true
		}
	}
	return false
}
