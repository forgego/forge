package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractOptionFromMethod_AllOptions(t *testing.T) {
	const modelSrc = `package models

import "github.com/forgego/forge/schema"

type TestModel struct {
	schema.BaseSchema
}

func (TestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("f_primary").Primary().Build(),
		schema.Int64("f_autoincrement").AutoIncrement().Build(),
		schema.String("f_required").Required().Build(),
		schema.String("f_optional").Optional().Build(),
		schema.String("f_unique").Unique().Build(),
		schema.String("f_blank").Blank().Build(),
		schema.String("f_dbindex").DBIndex().Build(),
		schema.String("f_dbcolumn").DBColumn("custom_col").Build(),
		schema.String("f_maxlength").MaxLength(255).Build(),
		schema.String("f_minlength").MinLength(5).Build(),
		schema.Float64("f_maxvalue").MaxValue(99.5).Build(),
		schema.Float64("f_minvalue").MinValue(1.5).Build(),
		schema.String("f_default").Default("default_value").Build(),
		schema.String("f_helptext").HelpText("help text").Build(),
		schema.String("f_verbosename").VerboseName("verbose name").Build(),
		schema.Time("f_autonow").AutoNow().Build(),
		schema.Time("f_autonowadd").AutoNowAdd().Build(),
		schema.String("f_writeonly").WriteOnly().Build(),
		schema.Bool("f_editable").Editable(false).Build(),
		schema.String("f_choices_variadic").Choices("active", "inactive").Build(),
		schema.String("f_choices_composite").Choices([]string{"draft", "published"}).Build(),
		schema.Decimal("f_maxdigits").MaxDigits(10).Build(),
		schema.Decimal("f_decimalplaces").DecimalPlaces(2).Build(),
		schema.String("f_dbdefault").DBDefault("CURRENT_TIMESTAMP").Build(),
		schema.String("f_generatedcolumn").GeneratedColumn("col_a + col_b", true).Build(),
		schema.String("f_generatedcolumn_nostored").GeneratedColumn("col_a").Build(),
		schema.String("f_dbcomment").DBComment("column comment").Build(),
		schema.String("f_dbcollation").DBCollation("utf8mb4_bin").Build(),
		schema.String("f_dbtablespace").DBTablespace("fast_space").Build(),
		schema.String("f_dbtype").DBType("varchar(255)").Build(),
	}
}
`

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, "models.go", modelSrc, 0)
	require.NoError(t, err)

	callsByMethod := make(map[string][]*ast.CallExpr)
	ast.Inspect(fileNode, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				callsByMethod[sel.Sel.Name] = append(callsByMethod[sel.Sel.Name], call)
			}
		}
		return true
	})

	tests := []struct {
		name       string
		methodName string
		index      int
		expected   map[string]interface{}
	}{
		{
			name:       "Primary",
			methodName: "Primary",
			expected: map[string]interface{}{
				"primary":     true,
				"primary_key": true,
			},
		},
		{
			name:       "AutoIncrement",
			methodName: "AutoIncrement",
			expected: map[string]interface{}{
				"auto_increment": true,
			},
		},
		{
			name:       "Required",
			methodName: "Required",
			expected: map[string]interface{}{
				"required": true,
			},
		},
		{
			name:       "Optional",
			methodName: "Optional",
			expected: map[string]interface{}{
				"required": false,
			},
		},
		{
			name:       "Unique",
			methodName: "Unique",
			expected: map[string]interface{}{
				"unique": true,
			},
		},
		{
			name:       "Blank",
			methodName: "Blank",
			expected: map[string]interface{}{
				"blank": true,
			},
		},
		{
			name:       "DBIndex",
			methodName: "DBIndex",
			expected: map[string]interface{}{
				"db_index": true,
			},
		},
		{
			name:       "DBColumn",
			methodName: "DBColumn",
			expected: map[string]interface{}{
				"db_column": "custom_col",
			},
		},
		{
			name:       "MaxLength",
			methodName: "MaxLength",
			expected: map[string]interface{}{
				"max_length": 255,
			},
		},
		{
			name:       "MinLength",
			methodName: "MinLength",
			expected: map[string]interface{}{
				"min_length": 5,
			},
		},
		{
			name:       "MaxValue",
			methodName: "MaxValue",
			expected: map[string]interface{}{
				"max_value": 99.5,
			},
		},
		{
			name:       "MinValue",
			methodName: "MinValue",
			expected: map[string]interface{}{
				"min_value": 1.5,
			},
		},
		{
			name:       "Default",
			methodName: "Default",
			expected: map[string]interface{}{
				"default": "default_value",
			},
		},
		{
			name:       "HelpText",
			methodName: "HelpText",
			expected: map[string]interface{}{
				"help_text": "help text",
			},
		},
		{
			name:       "VerboseName",
			methodName: "VerboseName",
			expected: map[string]interface{}{
				"verbose_name": "verbose name",
			},
		},
		{
			name:       "AutoNow",
			methodName: "AutoNow",
			expected: map[string]interface{}{
				"auto_now": true,
			},
		},
		{
			name:       "AutoNowAdd",
			methodName: "AutoNowAdd",
			expected: map[string]interface{}{
				"auto_now_add": true,
			},
		},
		{
			name:       "WriteOnly",
			methodName: "WriteOnly",
			expected: map[string]interface{}{
				"write_only": true,
			},
		},
		{
			name:       "Editable",
			methodName: "Editable",
			expected: map[string]interface{}{
				"editable": false,
			},
		},
		{
			name:       "Choices_Variadic",
			methodName: "Choices",
			index:      0,
			expected: map[string]interface{}{
				"has_choices": true,
				"choices":     []string{"active", "inactive"},
			},
		},
		{
			name:       "Choices_CompositeLiteral",
			methodName: "Choices",
			index:      1,
			expected: map[string]interface{}{
				"has_choices": true,
				"choices":     []string{"draft", "published"},
			},
		},
		{
			name:       "MaxDigits",
			methodName: "MaxDigits",
			expected: map[string]interface{}{
				"max_digits": 10,
			},
		},
		{
			name:       "DecimalPlaces",
			methodName: "DecimalPlaces",
			expected: map[string]interface{}{
				"decimal_places": 2,
			},
		},
		{
			name:       "DBDefault",
			methodName: "DBDefault",
			expected: map[string]interface{}{
				"db_default": "CURRENT_TIMESTAMP",
			},
		},
		{
			name:       "GeneratedColumn_WithStored",
			methodName: "GeneratedColumn",
			index:      0,
			expected: map[string]interface{}{
				"generated":        true,
				"generated_expr":   "col_a + col_b",
				"generated_stored": true,
			},
		},
		{
			name:       "GeneratedColumn_WithoutStored",
			methodName: "GeneratedColumn",
			index:      1,
			expected: map[string]interface{}{
				"generated":      true,
				"generated_expr": "col_a",
			},
		},
		{
			name:       "DBComment",
			methodName: "DBComment",
			expected: map[string]interface{}{
				"db_comment": "column comment",
			},
		},
		{
			name:       "DBCollation",
			methodName: "DBCollation",
			expected: map[string]interface{}{
				"db_collation": "utf8mb4_bin",
			},
		},
		{
			name:       "DBTablespace",
			methodName: "DBTablespace",
			expected: map[string]interface{}{
				"db_tablespace": "fast_space",
			},
		},
		{
			name:       "DBType",
			methodName: "DBType",
			expected: map[string]interface{}{
				"db_type": "varchar(255)",
			},
		},
	}

	p := NewASTParser()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			calls, ok := callsByMethod[tc.methodName]
			require.True(t, ok, "calls for method %s should exist", tc.methodName)
			require.Greater(t, len(calls), tc.index, "call index %d for method %s out of range", tc.index, tc.methodName)

			options := make(map[string]interface{})
			p.extractOptionFromMethod(tc.methodName, calls[tc.index], options)
			assert.Equal(t, tc.expected, options)
		})
	}
}
