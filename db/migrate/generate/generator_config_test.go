package generate

import (
	"testing"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/sql"
	"github.com/forgego/forge/db/migrate/state"
	"github.com/stretchr/testify/require"
)

func TestNewMigrationGeneratorPreservesCallerDriver(t *testing.T) {
	driver := core.DriverSQLite
	builder, err := sql.NewSQLBuilder(driver)
	require.NoError(t, err)

	gen, err := NewMigrationGeneratorWithDriver("models", "migrations", driver, NewDetector(), builder, state.NewInMemoryState())

	require.NoError(t, err)
	require.Equal(t, driver, gen.driver)
}

func TestNewMigrationGeneratorForDriverUsesCallerDriver(t *testing.T) {
	gen, err := NewMigrationGeneratorForDriver("models", "migrations", core.DriverSQLite)

	require.NoError(t, err)
	require.Equal(t, core.DriverSQLite, gen.driver)
}

func TestNewMigrationGeneratorForDriverRejectsUnsupportedCallerDriver(t *testing.T) {
	gen, err := NewMigrationGeneratorForDriver("models", "migrations", core.Driver("unsupported"))

	require.Nil(t, gen)
	require.ErrorContains(t, err, "unsupported database driver: unsupported")
}

func TestNewMigrationGeneratorWithDefaults_MatchesDefaultDriver(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()

	gen, err := NewMigrationGeneratorWithDefaults(modelsDir, migrationsDir)
	require.NoError(t, err)
	require.NotNil(t, gen)

	defaultDriver := core.Driver(config.NewConfig().GetDriver())
	driverGen, err := NewMigrationGeneratorForDriver(modelsDir, migrationsDir, defaultDriver)
	require.NoError(t, err)
	require.NotNil(t, driverGen)

	require.Equal(t, driverGen.driver, gen.driver)
}
