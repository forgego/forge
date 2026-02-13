package execute

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeAppliedVersions_InfersAppliedHistoryForCleanState(t *testing.T) {
	merged := mergeAppliedVersions(3, false, map[uint]bool{3: true})

	assert.True(t, merged[1])
	assert.True(t, merged[2])
	assert.True(t, merged[3])
}

func TestMergeAppliedVersions_ExcludesCurrentWhenDirty(t *testing.T) {
	merged := mergeAppliedVersions(4, true, map[uint]bool{4: true})

	assert.True(t, merged[1])
	assert.True(t, merged[2])
	assert.True(t, merged[3])
	assert.False(t, merged[4])
}

func TestMergeAppliedVersions_HandlesNoVersion(t *testing.T) {
	merged := mergeAppliedVersions(0, false, map[uint]bool{})

	assert.Empty(t, merged)
}

func TestMergeAppliedVersions_KeepsExplicitPastVersionsWhenDirty(t *testing.T) {
	merged := mergeAppliedVersions(4, true, map[uint]bool{2: true})

	assert.True(t, merged[1])
	assert.True(t, merged[2])
	assert.True(t, merged[3])
	assert.False(t, merged[4])
}

func TestGetAllMigrations_IgnoresMalformedVersions(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "abc_bad.up.sql")
	writeMigrationFile(t, dir, "000002_add_users.up.sql")

	reporter := NewStatusReporter(dir, nil, nil)
	migrations, err := reporter.getAllMigrations()

	assert.NoError(t, err)
	assert.Len(t, migrations, 1)
	assert.Equal(t, "000002", migrations[0].Version)
	assert.Equal(t, "add_users", migrations[0].Name)
}

func TestGetAllMigrations_SortsByNumericVersion(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "000010_ten.up.sql")
	writeMigrationFile(t, dir, "000002_two.up.sql")
	writeMigrationFile(t, dir, "000001_one.up.sql")

	reporter := NewStatusReporter(dir, nil, nil)
	migrations, err := reporter.getAllMigrations()

	assert.NoError(t, err)
	assert.Len(t, migrations, 3)
	assert.Equal(t, "000001", migrations[0].Version)
	assert.Equal(t, "000002", migrations[1].Version)
	assert.Equal(t, "000010", migrations[2].Version)
}

func TestGetDetailedStatus_WithoutMigrationEngine(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "000001_init.up.sql")
	writeMigrationFile(t, dir, "000002_add_users.up.sql")

	reporter := NewStatusReporter(dir, nil, nil)
	status, err := reporter.GetDetailedStatus(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "PENDING", status.Status)
	assert.Equal(t, "Unknown (migration engine unavailable)", status.Current)
	assert.Contains(t, status.Error, "Migration engine unavailable")
	assert.Equal(t, "000001", status.Next)
	assert.Empty(t, status.Applied)
	assert.Len(t, status.Pending, 2)
}

func TestGetAppliedVersions_WithoutDBAndMigrationEngine(t *testing.T) {
	reporter := NewStatusReporter(t.TempDir(), nil, nil)
	applied, err := reporter.getAppliedVersions(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, applied)
}

func writeMigrationFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("-- migration"), 0o644); err != nil {
		t.Fatalf("failed to write migration file %s: %v", name, err)
	}
}
