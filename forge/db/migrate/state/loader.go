package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/parse"
)

// LoaderOptions configures state loader behavior
type LoaderOptions struct {
	Verbose bool // Enable verbose logging of parse errors
}

// ParseErrorInfo stores information about parse errors
type ParseErrorInfo struct {
	File    string
	Line    int
	Column  int
	SQL     string
	Message string
	Error   error
}

// LoadStateFromFiles loads state from migration files
func LoadStateFromFiles(migrationsDir string) (*SchemaState, error) {
	loader := NewFileStateLoader(migrationsDir)
	return loader.Load()
}

// FileStateLoader loads state from SQL migration files
type FileStateLoader struct {
	migrationsDir string
	state         *SchemaState
	verbose       bool
	parseErrors   []*ParseErrorInfo
	skippedFiles  []string
}

// NewFileStateLoader creates a new file-based state loader
func NewFileStateLoader(migrationsDir string) StateManager {
	return NewFileStateLoaderWithOptions(migrationsDir, LoaderOptions{Verbose: false})
}

// NewFileStateLoaderWithOptions creates a new file-based state loader with options
func NewFileStateLoaderWithOptions(migrationsDir string, opts LoaderOptions) StateManager {
	return &FileStateLoader{
		migrationsDir: migrationsDir,
		state:         nil, // Will be loaded on first Load() call
		verbose:       opts.Verbose,
		parseErrors:   []*ParseErrorInfo{},
		skippedFiles:  []string{},
	}
}

// GetParseErrors returns all parse errors collected during loading
func (l *FileStateLoader) GetParseErrors() []*ParseErrorInfo {
	return l.parseErrors
}

// GetSkippedFiles returns list of files that were skipped
func (l *FileStateLoader) GetSkippedFiles() []string {
	return l.skippedFiles
}

// Load loads state from migration files with retry mechanism
func (l *FileStateLoader) Load() (*SchemaState, error) {
	// Load state if not already loaded
	if l.state == nil {
		state := &SchemaState{
			Tables: make(map[string]*TableState),
		}

		// Check if migrations directory exists
		if _, err := os.Stat(l.migrationsDir); os.IsNotExist(err) {
			l.state = state
			return state, nil
		}

		// Find all up migration files
		pattern := filepath.Join(l.migrationsDir, "*_*.up.sql")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migrations directory: %w", err)
		}

		// Sort files by version
		sortedFiles := sortMigrationFiles(matches)

		// Create parser with verbose mode if enabled
		parserOpts := parse.ParserOptions{Verbose: l.verbose}
		parser := parse.NewSQLParserWithOptions(parserOpts)

		// Collect all changes from all files first
		type fileChanges struct {
			file    string
			changes []core.Change
			errors  []*parse.ParseError
		}
		var allFileChanges []fileChanges

		for _, file := range sortedFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				if l.verbose {
					l.parseErrors = append(l.parseErrors, &ParseErrorInfo{
						File:    file,
						Message: fmt.Sprintf("failed to read file: %v", err),
						Error:   err,
					})
				}
				l.skippedFiles = append(l.skippedFiles, file)
				continue
			}

			changes, err := parser.ParseUpSQL(string(content))
			if err != nil {
				// Store parse error
				if l.verbose {
					l.parseErrors = append(l.parseErrors, &ParseErrorInfo{
						File:    file,
						Message: fmt.Sprintf("failed to parse SQL: %v", err),
						Error:   err,
					})
				}
				l.skippedFiles = append(l.skippedFiles, file)
				continue
			}

			// Collect parse errors from parser
			parseErrs := parser.GetErrors()
			if len(parseErrs) > 0 && l.verbose {
				for _, perr := range parseErrs {
					l.parseErrors = append(l.parseErrors, &ParseErrorInfo{
						File:    file,
						Line:    perr.Line,
						Column:  perr.Column,
						SQL:     perr.SQL,
						Message: perr.Message,
					})
				}
			}

			// Log UnknownChange statements if verbose
			for _, change := range changes {
				if unknown, ok := change.(*core.UnknownChange); ok && l.verbose {
					l.parseErrors = append(l.parseErrors, &ParseErrorInfo{
						File:    file,
						SQL:     unknown.SQL,
						Message: "unparseable statement (UnknownChange)",
					})
				}
			}

			allFileChanges = append(allFileChanges, fileChanges{
				file:    file,
				changes: changes,
			})
		}

		// Three-pass retry mechanism
		// Pass 1: Process all CREATE TABLE statements
		var deferredChanges []core.Change
		for _, fc := range allFileChanges {
			for _, change := range fc.changes {
				if _, ok := change.(*core.CreateTable); ok {
					if err := applyChangeToState(state, change); err != nil {
						return nil, l.formatError(fc.file, change, err, "CREATE TABLE")
					}
				} else if _, ok := change.(*core.UnknownChange); !ok {
					// Defer non-CREATE TABLE changes
					deferredChanges = append(deferredChanges, change)
				}
			}
		}

		// Pass 2: Process constraints, indexes, foreign keys
		var retryChanges []core.Change
		for _, fc := range allFileChanges {
			for _, change := range fc.changes {
				if _, ok := change.(*core.CreateTable); ok {
					continue // Already processed
				}
				if _, ok := change.(*core.UnknownChange); ok {
					continue // Skip unknown
				}

				// Check if this is a constraint/index/FK change
				changeType := change.Type()
				if changeType == core.ChangeTypeAddIndex ||
					changeType == core.ChangeTypeAddForeignKey ||
					changeType == core.ChangeTypeAddConstraint ||
					changeType == core.ChangeTypeAddColumn {
					if err := applyChangeToState(state, change); err != nil {
						if strings.Contains(err.Error(), "does not exist") {
							// Table might not exist yet, retry in pass 3
							retryChanges = append(retryChanges, change)
							if l.verbose {
								l.parseErrors = append(l.parseErrors, &ParseErrorInfo{
									File:    fc.file,
									Message: fmt.Sprintf("deferred change (table not found): %s", change.Type()),
									Error:   err,
								})
							}
						} else {
							return nil, l.formatError(fc.file, change, err, "constraint/index")
						}
					}
				} else {
					// Other changes
					if err := applyChangeToState(state, change); err != nil {
						if strings.Contains(err.Error(), "does not exist") {
							retryChanges = append(retryChanges, change)
						} else {
							return nil, l.formatError(fc.file, change, err, "change")
						}
					}
				}
			}
		}

		// Pass 3: Retry changes that failed due to missing tables
		for _, change := range retryChanges {
			if err := applyChangeToState(state, change); err != nil {
				if l.verbose {
					l.parseErrors = append(l.parseErrors, &ParseErrorInfo{
						Message: fmt.Sprintf("failed to apply change after retry: %s - %v", change.Type(), err),
						Error:   err,
					})
				}
				// Continue - some changes might fail if tables truly don't exist
			}
		}

		l.state = state
	}

	return l.state, nil
}

// formatError formats an error with context
func (l *FileStateLoader) formatError(file string, change core.Change, err error, context string) error {
	tableName := change.TableName()
	if tableName != "" {
		return fmt.Errorf("failed to apply %s change for table %q in %s: %w", context, tableName, file, err)
	}
	return fmt.Errorf("failed to apply %s change in %s: %w", context, file, err)
}

// Apply applies changes to the state (required for StateManager interface)
func (l *FileStateLoader) Apply(changes []core.Change) error {
	// Load state first if not loaded
	if l.state == nil {
		_, err := l.Load()
		if err != nil {
			return err
		}
	}

	// Use InMemoryState to apply changes
	manager := &InMemoryState{state: l.state}
	return manager.Apply(changes)
}

// GetState returns the current state (required for StateManager interface)
func (l *FileStateLoader) GetState() *SchemaState {
	if l.state == nil {
		// Load state if not loaded
		// If Load fails, return empty state to avoid nil pointer
		state, err := l.Load()
		if err != nil {
			// Return empty state on error rather than nil
			return &SchemaState{
				Tables: make(map[string]*TableState),
			}
		}
		return state
	}
	return l.state
}

// applyChangeToState applies a single change to the state
func applyChangeToState(state *SchemaState, change core.Change) error {
	// Create a temporary manager with the state
	manager := &InMemoryState{state: state}
	return manager.Apply([]core.Change{change})
}
