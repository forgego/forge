package internal

// CommonFlagNames contains commonly used flag names
const (
	FlagMigrationsPath = "path"
	FlagModelsDir      = "models"
	FlagOutputDir      = "output"
	FlagPort           = "port"
	FlagDryRun         = "dry-run"
	FlagAuto           = "auto"
	FlagVerbose        = "verbose"
	FlagCoverage       = "coverage"
)

// CommonFlagDefaults contains default values for common flags
const (
	DefaultMigrationsPath = "./migrations"
	DefaultModelsDir      = "./models"
	DefaultOutputDir      = "./models"
)
