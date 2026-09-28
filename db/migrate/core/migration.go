package core

// Migration represents a migration with its metadata
type Migration struct {
	Version       string
	Name          string
	UpSQL         string
	DownSQL       string
	Checksum      string
	Reversible    bool
	Transactional bool
}

// MigrationPlan represents a complete migration plan
type MigrationPlan struct {
	Version       string
	Name          string
	Changes       []Change
	UpSQL         string
	DownSQL       string
	Reversible    bool
	Transactional bool
	Dependencies  []Dependency // Dependencies on other migrations
	Replaces      []string     // Migrations that this migration replaces (for squashing)
	Initial       bool         // Whether this is an initial migration
}

// Dependency represents a dependency on another migration
type Dependency struct {
	App     string // App/module name (optional)
	Version string // Migration version (e.g., "000001")
}

// NewMigrationPlan creates a new migration plan
func NewMigrationPlan(version, name string, changes []Change) *MigrationPlan {
	return &MigrationPlan{
		Version:       version,
		Name:          name,
		Changes:       changes,
		Reversible:    allReversible(changes),
		Transactional: true, // Most migrations are transactional
	}
}

// allReversible checks if all changes are reversible
func allReversible(changes []Change) bool {
	for _, change := range changes {
		if !change.Reversible() {
			return false
		}
	}
	return true
}
