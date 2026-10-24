package migrations

import (
	"context"
	"fmt"
	"io"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
)

// applyFrameworkMigrations applies the framework store migrations.
// They are embedded in Forge and tracked in their own table
// (forge_framework_migrations), so they never mix with the
// application's migrations directory or versions.
func applyFrameworkMigrations(ctx context.Context, _ *config.Config, database *db.DB, out io.Writer) error {
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
		fmt.Fprintf(out, "✓ Framework store tables migrated to version %d\n", after)
	}
	return nil
}

// renderFrameworkStatus prints the framework store migration state.
func renderFrameworkStatus(ctx context.Context, _ *config.Config, database *db.DB, out io.Writer) {
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
	fmt.Fprintf(out, "\nFramework store tables: version %d of %d, %s\n", version, latest, state)
}
