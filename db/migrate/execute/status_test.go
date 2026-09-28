package execute

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeAppliedVersions_InfersAppliedHistoryForCleanState(t *testing.T) {
	merged := mergeAppliedVersions(3, false, map[uint]bool{3: true}, []uint{1, 2, 3})

	assert.True(t, merged[1])
	assert.True(t, merged[2])
	assert.True(t, merged[3])
}

func TestMergeAppliedVersions_ExcludesCurrentWhenDirty(t *testing.T) {
	merged := mergeAppliedVersions(4, true, map[uint]bool{4: true}, []uint{1, 2, 3, 4})

	assert.True(t, merged[1])
	assert.True(t, merged[2])
	assert.True(t, merged[3])
	assert.False(t, merged[4])
}

func TestMergeAppliedVersions_HandlesNoVersion(t *testing.T) {
	merged := mergeAppliedVersions(0, false, map[uint]bool{}, []uint{1, 2, 3})

	assert.Empty(t, merged)
}

func TestMergeAppliedVersions_KeepsExplicitPastVersionsWhenDirty(t *testing.T) {
	merged := mergeAppliedVersions(4, true, map[uint]bool{2: true}, []uint{1, 2, 3, 4})

	assert.True(t, merged[1])
	assert.True(t, merged[2])
	assert.True(t, merged[3])
	assert.False(t, merged[4])
}

func TestMergeAppliedVersions_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion uint
		dirty          bool
		explicit       map[uint]bool
		fileVersions   []uint
		expected       map[uint]bool
	}{
		{
			name:           "timestamp versions return immediately",
			currentVersion: 20240101120000,
			dirty:          false,
			explicit:       nil,
			fileVersions:   []uint{20231201000000, 20240101120000, 20240201000000},
			expected: map[uint]bool{
				20231201000000: true,
				20240101120000: true,
			},
		},
		{
			name:           "gap in files only includes existing versions",
			currentVersion: 3,
			dirty:          false,
			explicit:       nil,
			fileVersions:   []uint{1, 3, 5},
			expected: map[uint]bool{
				1: true,
				3: true,
			},
		},
		{
			name:           "dirty excludes current version",
			currentVersion: 3,
			dirty:          true,
			explicit:       nil,
			fileVersions:   []uint{1, 2, 3},
			expected: map[uint]bool{
				1: true,
				2: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := mergeAppliedVersions(tt.currentVersion, tt.dirty, tt.explicit, tt.fileVersions)
			assert.Equal(t, tt.expected, merged)
		})
	}
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
