package migrations

import (
	"context"
	"fmt"
	"io"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
)

// frameworkStoresEnabled reports whether server.stores selects the database
// stores, whose tables are framework-owned migrations.
func frameworkStoresEnabled(cfg *config.Config) (bool, error) {
	if cfg == nil {
		return false, nil
	}
	kind, err := stores.ParseKind(config.LoadSettings(cfg).Server.Stores)
	if err != nil {
		return false, err
	}
	return kind == stores.KindDatabase, nil
}

// applyFrameworkMigrations applies the framework store migrations when
// server.stores is database. They are embedded in Forge and tracked in
// their own table (forge_framework_migrations), so they never mix with the
// application's migrations directory or versions.
func applyFrameworkMigrations(ctx context.Context, cfg *config.Config, database *db.DB, out io.Writer) error {
	enabled, err := frameworkStoresEnabled(cfg)
	if err != nil || !enabled {
		return err
	}
	before, _, err := stores.Version(ctx, database)
	if err != nil {
		return fmt.Errorf("framework store migrations: %w", err)
	}
	if err := stores.Migrate(ctx, database); err != nil {
		return err
	}
	after, _, err := stores.Version(ctx, database)
	if err != nil {
		return fmt.Errorf("framework store migrations: %w", err)
	}
	if after != before {
		fmt.Fprintf(out, "✓ Framework store tables migrated to version %d (server.stores: database)\n", after)
	}
	return nil
}

// renderFrameworkStatus prints the framework store migration state when
// server.stores is database.
func renderFrameworkStatus(ctx context.Context, cfg *config.Config, database *db.DB, out io.Writer) {
	enabled, err := frameworkStoresEnabled(cfg)
	if err != nil {
		fmt.Fprintf(out, "\n[WARN] %v\n", err)
		return
	}
	if !enabled {
		return
	}
	latest, err := stores.LatestVersion(database.Driver)
	if err != nil {
		fmt.Fprintf(out, "\n[WARN] Framework store tables: %v\n", err)
		return
	}
	version, dirty, err := stores.Version(ctx, database)
	if err != nil {
		fmt.Fprintf(out, "\n[WARN] Framework store tables: %v\n", err)
		return
	}
	state := "up to date"
	switch {
	case dirty:
		state = "DIRTY"
	case version < latest:
		state = "pending, run 'forge migrate up'"
	}
	fmt.Fprintf(out, "\nFramework store tables (server.stores: database): version %d of %d, %s\n", version, latest, state)
}
