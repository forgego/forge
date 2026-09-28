package generation

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUUIDModelSource = `package testmodels

import (
	"github.com/forgego/forge/schema"
)

type Document struct {
	schema.BaseSchema
}

func (Document) Fields() []schema.Field {
	return []schema.Field{
		schema.UUID("id").Primary().Build(),
		schema.String("title").Required().MaxLength(255).Build(),
	}
}

func (Document) Meta() schema.Meta {
	return schema.Meta{
		TableName: "test_documents",
		VerboseName: "Document",
	}
}

func (Document) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Document) Hooks() *schema.ModelHooks {
	return nil
}
`

const testStringPKModelSource = `package testmodels

import (
	"github.com/forgego/forge/schema"
)

type Category struct {
	schema.BaseSchema
}

func (Category) Fields() []schema.Field {
	return []schema.Field{
		schema.String("code").Primary().Build(),
		schema.String("title").Build(),
	}
}

func (Category) Meta() schema.Meta {
	return schema.Meta{
		TableName: "test_categories",
		VerboseName: "Category",
	}
}

func (Category) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Category) Hooks() *schema.ModelHooks {
	return nil
}
`

const testNoPKModelSource = `package testmodels

import "github.com/forgego/forge/schema"

type Log struct { schema.BaseSchema }

func (Log) Fields() []schema.Field {
	return []schema.Field{schema.String("message").Build()}
}
`

const testInt32PKModelSource = `package testmodels

import "github.com/forgego/forge/schema"

type Counter struct { schema.BaseSchema }

func (Counter) Fields() []schema.Field {
	return []schema.Field{schema.Int32("id").Primary().Build()}
}
`

func TestGenerateCommand_APIFlag_RejectsNonIntegerPrimaryKey(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "uuid primary key", source: testUUIDModelSource},
		{name: "string primary key", source: testStringPKModelSource},
		{name: "no primary key", source: testNoPKModelSource},
		{name: "int32 primary key", source: testInt32PKModelSource},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			modelFile := filepath.Join(tmpDir, "models.go")
			require.NoError(t, os.WriteFile(modelFile, []byte(tc.source), 0644))

			cmd := NewGenerateCommand()
			def := cmd.Definition()
			require.NoError(t, def.ParseFlags([]string{"--models", tmpDir, "--output", tmpDir, "--api"}))

			ctx := &core.Context{Cmd: def}
			err := cmd.Execute(ctx, []string{})
			require.Error(t, err)
			if tc.name == "no primary key" {
				assert.Contains(t, err.Error(), "has no primary key")
			} else if tc.name == "int32 primary key" {
				assert.Contains(t, err.Error(), "require an int64 primary key")
			} else {
				assert.Contains(t, err.Error(), "non-integer primary key")
			}
			assert.NoFileExists(t, filepath.Join(tmpDir, "api_gen.go"))
		})
	}
}

const testModelSource = `package testmodels

import (
	"github.com/forgego/forge/schema"
)

type Product struct {
	schema.BaseSchema
	ID int64
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.Decimal("price").Required().Build(),
	}
}

func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName: "test_products",
		VerboseName: "Product",
	}
}

func (Product) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Product) Hooks() *schema.ModelHooks {
	return nil
}
`

func TestGenerateCommand_Definition(t *testing.T) {
	cmd := NewGenerateCommand()
	def := cmd.Definition()

	assert.Equal(t, "generate", def.Use)
	assert.NotNil(t, def.Flags().Lookup("models"))
	assert.NotNil(t, def.Flags().Lookup("output"))
	apiFlag := def.Flags().Lookup("api")
	require.NotNil(t, apiFlag)
	assert.Equal(t, "false", apiFlag.DefValue)
}

func TestGenerateCommand_APIFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectAPI bool
	}{
		{
			name:      "without --api flag",
			args:      []string{},
			expectAPI: false,
		},
		{
			name:      "with --api flag",
			args:      []string{"--api"},
			expectAPI: true,
		},
	}

	var genGoOutputs [][]byte

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			modelFile := filepath.Join(tmpDir, "models.go")
			require.NoError(t, os.WriteFile(modelFile, []byte(testModelSource), 0644))

			cmd := NewGenerateCommand()
			def := cmd.Definition()

			allArgs := append([]string{"--models", tmpDir, "--output", tmpDir}, tc.args...)
			require.NoError(t, def.ParseFlags(allArgs))

			ctx := &core.Context{
				Cmd: def,
			}

			err := cmd.Execute(ctx, []string{})
			require.NoError(t, err)

			// gen.go should always be produced
			genFile := filepath.Join(tmpDir, "gen.go")
			require.FileExists(t, genFile)
			genBytes, err := os.ReadFile(genFile)
			require.NoError(t, err)
			assert.Contains(t, string(genBytes), "type ProductGenerated struct")
			assert.Contains(t, string(genBytes), "var ProductObjects = orm.MustNewManager[Product]")
			genGoOutputs = append(genGoOutputs, genBytes)

			// api_gen.go check
			apiFile := filepath.Join(tmpDir, "api_gen.go")
			if tc.expectAPI {
				assert.FileExists(t, apiFile)
				apiBytes, err := os.ReadFile(apiFile)
				require.NoError(t, err)
				assert.Contains(t, string(apiBytes), "type ProductSerializer struct")
				assert.Contains(t, string(apiBytes), "type ProductViewSet struct")
				assert.Contains(t, string(apiBytes), "RegisterAPIRoutes(router *forgehttp.Router)")
			} else {
				assert.NoFileExists(t, apiFile)
			}
		})
	}

	// Verify that other generated files (gen.go) are identical in both runs
	require.Len(t, genGoOutputs, 2)
	assert.Equal(t, genGoOutputs[0], genGoOutputs[1], "gen.go should be identical in both runs")
}

func TestGenerateCommand_AppDirectory_APIFlag(t *testing.T) {
	tests := []struct {
		name      string
		apiFlag   bool
		expectAPI bool
	}{
		{
			name:      "app mode without --api flag",
			apiFlag:   false,
			expectAPI: false,
		},
		{
			name:      "app mode with --api flag",
			apiFlag:   true,
			expectAPI: true,
		},
	}

	var genGoOutputs [][]byte

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			appDir := filepath.Join(tmpDir, "app", "catalog")
			require.NoError(t, os.MkdirAll(appDir, 0755))
			modelFile := filepath.Join(appDir, "models.go")
			require.NoError(t, os.WriteFile(modelFile, []byte(testModelSource), 0644))

			cmd := NewGenerateCommand()
			def := cmd.Definition()

			args := []string{"--models", filepath.Join(tmpDir, "app"), "--output", filepath.Join(tmpDir, "app")}
			if tc.apiFlag {
				args = append(args, "--api")
			}
			require.NoError(t, def.ParseFlags(args))

			ctx := &core.Context{
				Cmd: def,
			}

			err := cmd.Execute(ctx, []string{})
			require.NoError(t, err)

			genFile := filepath.Join(appDir, "gen.go")
			require.FileExists(t, genFile)
			genBytes, err := os.ReadFile(genFile)
			require.NoError(t, err)
			genGoOutputs = append(genGoOutputs, genBytes)

			apiFile := filepath.Join(appDir, "api_gen.go")
			if tc.expectAPI {
				assert.FileExists(t, apiFile)
			} else {
				assert.NoFileExists(t, apiFile)
			}
		})
	}

	require.Len(t, genGoOutputs, 2)
	assert.Equal(t, genGoOutputs[0], genGoOutputs[1], "gen.go should be identical in both runs")
}

func TestGenerateStrictReportsUnsupportedExpressions(t *testing.T) {
	modelsDir := t.TempDir()
	filename := filepath.Join(modelsDir, "models.go")
	source := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{defaultFields()}
}
func defaultFields() schema.Field { return schema.StringField("name") }
`
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	if output, err := executeGenerate(t, modelsDir, false); err != nil {
		t.Fatalf("non-strict generation failed: %v", err)
	} else if !strings.Contains(output, "warning: "+filename+":5:") {
		t.Fatalf("warning did not name file and line: %q", output)
	}
	if output, err := executeGenerate(t, modelsDir, true); err == nil {
		t.Fatal("strict generation succeeded")
	} else if !strings.Contains(err.Error(), "generation found 1 model expressions") || !strings.Contains(output, "warning: "+filename+":5:") {
		t.Fatalf("strict error/output = %v / %q", err, output)
	}
}

func TestGenerateStrictLeavesPreviousOutputIntact(t *testing.T) {
	modelsDir := t.TempDir()
	filename := filepath.Join(modelsDir, "models.go")
	valid := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.StringField("name")}
}
`
	if err := os.WriteFile(filename, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := executeGenerate(t, modelsDir, false); err != nil {
		t.Fatalf("initial generation failed: %v", err)
	}
	genFile := filepath.Join(modelsDir, "gen.go")
	before, err := os.ReadFile(genFile)
	if err != nil {
		t.Fatal(err)
	}
	unsupported := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{defaultFields()}
}
func defaultFields() schema.Field { return schema.StringField("name") }
`
	if err := os.WriteFile(filename, []byte(unsupported), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := executeGenerate(t, modelsDir, true); err == nil {
		t.Fatal("strict generation succeeded")
	}
	after, err := os.ReadFile(genFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("strict run truncated previously generated file")
	}
}

func TestGenerateApps_InvalidLaterAPIModelPreservesEarlierGeneratedFiles(t *testing.T) {
	root := t.TempDir()
	appsDir := filepath.Join(root, "app")
	firstDir := filepath.Join(appsDir, "a_valid")
	secondDir := filepath.Join(appsDir, "z_invalid")
	require.NoError(t, os.MkdirAll(firstDir, 0o755))
	require.NoError(t, os.MkdirAll(secondDir, 0o755))

	valid := `package valid
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema; ID int64 }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	invalid := `package invalid
import "github.com/forgego/forge/schema"
type ManualKey struct { schema.BaseSchema }
func (ManualKey) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(firstDir, "models.go"), []byte(valid), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(secondDir, "models.go"), []byte(invalid), 0o600))

	oldGen := []byte("previous generated models\n")
	oldAPI := []byte("previous generated API\n")
	require.NoError(t, os.WriteFile(filepath.Join(firstDir, "gen.go"), oldGen, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(firstDir, "api_gen.go"), oldAPI, 0o600))

	command := NewGenerateCommand().Definition()
	require.NoError(t, command.Flags().Set("models", appsDir))
	require.NoError(t, command.Flags().Set("output", appsDir))
	require.NoError(t, command.Flags().Set("api", "true"))
	err := NewGenerateCommand().Execute(&core.Context{Cmd: command}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ManualKey")

	gotGen, readErr := os.ReadFile(filepath.Join(firstDir, "gen.go"))
	require.NoError(t, readErr)
	gotAPI, readErr := os.ReadFile(filepath.Join(firstDir, "api_gen.go"))
	require.NoError(t, readErr)
	assert.Equal(t, oldGen, gotGen)
	assert.Equal(t, oldAPI, gotAPI)
}

func executeGenerate(t *testing.T, modelsDir string, strict bool) (string, error) {
	t.Helper()
	command := NewGenerateCommand().Definition()
	if err := command.Flags().Set("models", modelsDir); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("output", modelsDir); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("strict", strconv.FormatBool(strict)); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stderr
	os.Stderr = write
	err = NewGenerateCommand().Execute(&core.Context{Cmd: command}, nil)
	_ = write.Close()
	os.Stderr = original
	output, readErr := io.ReadAll(read)
	_ = read.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(output), err
}
