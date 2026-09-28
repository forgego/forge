package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/forgego/forge/utils"
)

// Writer writes generated code to files
type Writer struct {
	templates map[string]*template.Template
	rename    func(string, string) error
}

// NewWriter creates a new writer
func NewWriter() *Writer {
	return &Writer{
		templates: make(map[string]*template.Template),
		rename:    os.Rename,
	}
}

// WriteCombined writes all generated code to a single gen.go file
func (w *Writer) WriteCombined(definitions []*ModelDefinition, outputDir string) error {
	if len(definitions) == 0 {
		return nil
	}
	contents, err := w.renderCombined(definitions)
	if err != nil {
		return err
	}
	return w.writeBytes(contents, filepath.Join(outputDir, "gen.go"))
}

func (w *Writer) renderCombined(definitions []*ModelDefinition) ([]byte, error) {
	// Get package name from first model
	packageName := definitions[0].Package

	// Create template functions
	t := template.New("combined").Funcs(template.FuncMap{
		"ToLower":  strings.ToLower,
		"ToSnake":  utils.ToSnake,
		"ToCamel":  utils.ToCamel,
		"ToPascal": utils.ToPascal,
	})

	t, err := t.Parse(combinedTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse combined template: %w", err)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Package": packageName,
		"Models":  definitions,
	}

	return executeTemplate(t, data)
}

// WriteAPI writes all generated REST API code to an api_gen.go file
func (w *Writer) WriteAPI(definitions []*ModelDefinition, outputDir string) error {
	if len(definitions) == 0 {
		return nil
	}

	// Generated REST APIs address rows by integer IDs; reject models with
	// non-integer primary keys instead of emitting broken CRUD.
	if err := ValidateAPIModels(definitions); err != nil {
		return err
	}
	contents, err := w.renderAPI(definitions)
	if err != nil {
		return err
	}
	return w.writeBytes(contents, filepath.Join(outputDir, "api_gen.go"))
}

func (w *Writer) renderAPI(definitions []*ModelDefinition) ([]byte, error) {
	packageName := definitions[0].Package

	t := template.New("api").Funcs(template.FuncMap{
		"ToLower":       strings.ToLower,
		"ToSnake":       utils.ToSnake,
		"Pluralize":     utils.Pluralize,
		"ToKebab":       utils.ToKebab,
		"ToKebabPlural": utils.ToKebabPlural,
		"ToCamel":       utils.ToCamel,
		"ToPascal":      utils.ToPascal,
	})

	t, err := t.Parse(apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse api template: %w", err)
	}

	data := map[string]interface{}{
		"Package": packageName,
		"Models":  definitions,
	}

	return executeTemplate(t, data)
}

func executeTemplate(t *template.Template, data interface{}) ([]byte, error) {
	var output bytes.Buffer
	if err := t.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return output.Bytes(), nil
}

// WriteCombinedAndAPI renders and stages both generated files before replacing
// either destination, so preparation failures cannot leave a mixed generation.
func (w *Writer) WriteCombinedAndAPI(definitions []*ModelDefinition, outputDir string) error {
	if len(definitions) == 0 {
		return nil
	}
	if err := ValidateAPIModels(definitions); err != nil {
		return err
	}
	combined, err := w.renderCombined(definitions)
	if err != nil {
		return err
	}
	api, err := w.renderAPI(definitions)
	if err != nil {
		return err
	}

	combinedWrite, err := prepareWrite(filepath.Join(outputDir, "gen.go"), combined)
	if err != nil {
		return err
	}
	defer combinedWrite.cleanup()
	apiWrite, err := prepareWrite(filepath.Join(outputDir, "api_gen.go"), api)
	if err != nil {
		return err
	}
	defer apiWrite.cleanup()

	combinedBackup, err := backupDestination(combinedWrite.destination)
	if err != nil {
		return err
	}
	defer combinedBackup.cleanup()
	apiBackup, err := backupDestination(apiWrite.destination)
	if err != nil {
		return err
	}
	defer apiBackup.cleanup()

	if err := combinedWrite.commit(w.renameFile()); err != nil {
		return err
	}
	if err := apiWrite.commit(w.renameFile()); err != nil {
		if restoreErr := restoreDestinations(combinedBackup, apiBackup); restoreErr != nil {
			return fmt.Errorf("%w; failed to restore generated files: %v", err, restoreErr)
		}
		return err
	}
	return nil
}

// writeTemplate writes a template to a file. Replacement is atomic on Unix.
// On Windows, os.Rename is not guaranteed atomic and may fail while the
// destination is open elsewhere.
func (w *Writer) writeTemplate(t *template.Template, data interface{}, filename string) error {
	contents, err := executeTemplate(t, data)
	if err != nil {
		return err
	}
	return w.writeBytes(contents, filename)
}

func (w *Writer) writeBytes(contents []byte, filename string) error {
	prepared, err := prepareWrite(filename, contents)
	if err != nil {
		return err
	}
	defer prepared.cleanup()
	return prepared.commit(w.renameFile())
}

type preparedWrite struct {
	destination string
	tempPath    string
}

func prepareWrite(filename string, contents []byte) (*preparedWrite, error) {
	writeFilename := filename
	if fi, err := os.Lstat(filename); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(filename)
		if err != nil {
			// A dangling symlink (its target does not exist yet) cannot be
			// fully evaluated. Resolve the existing parent portion of the
			// link and keep the final target name, so generation writes
			// through the link and creates the target instead of aborting.
			resolved, err = resolveDanglingSymlink(filename)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve destination symlink %s: %w", filename, err)
			}
		}
		writeFilename = resolved
	}
	if err := ensureDestinationWritable(writeFilename); err != nil {
		return nil, err
	}
	dir := filepath.Dir(writeFilename)
	base := filepath.Base(writeFilename)

	var perm os.FileMode
	if fi, err := os.Stat(writeFilename); err == nil {
		perm = fi.Mode().Perm()
	} else {
		perm = os.FileMode(0o666 &^ processUmask())
	}

	tmpFile, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmpFile.Chmod(perm); err != nil {
		return nil, fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	if _, err := tmpFile.Write(contents); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil
	return &preparedWrite{destination: writeFilename, tempPath: tmpPath}, nil
}

func (p *preparedWrite) commit(rename func(string, string) error) error {
	if err := rename(p.tempPath, p.destination); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", p.destination, err)
	}
	p.tempPath = ""
	return nil
}

func (w *Writer) renameFile() func(string, string) error {
	if w != nil && w.rename != nil {
		return w.rename
	}
	return os.Rename
}

type destinationBackup struct {
	destination string
	backupPath  string
	existed     bool
}

func backupDestination(destination string) (*destinationBackup, error) {
	backup := &destinationBackup{destination: destination}
	info, err := os.Stat(destination)
	if err != nil {
		if os.IsNotExist(err) {
			return backup, nil
		}
		return nil, fmt.Errorf("failed to inspect destination %s: %w", destination, err)
	}
	backup.existed = true

	source, err := os.Open(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination %s for backup: %w", destination, err)
	}
	defer source.Close()

	file, err := os.CreateTemp(filepath.Dir(destination), "."+filepath.Base(destination)+".bak-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create backup for %s: %w", destination, err)
	}
	backup.backupPath = file.Name()
	failed := true
	defer func() {
		_ = file.Close()
		if failed {
			_ = os.Remove(backup.backupPath)
		}
	}()
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		return nil, fmt.Errorf("failed to preserve permissions for %s backup: %w", destination, err)
	}
	if _, err := io.Copy(file, source); err != nil {
		return nil, fmt.Errorf("failed to back up destination %s: %w", destination, err)
	}
	if err := file.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync backup for %s: %w", destination, err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close backup for %s: %w", destination, err)
	}
	failed = false
	return backup, nil
}

func restoreDestinations(backups ...*destinationBackup) error {
	var restoreErr error
	for _, backup := range backups {
		if backup == nil {
			continue
		}
		if !backup.existed {
			if err := os.Remove(backup.destination); err != nil && !os.IsNotExist(err) {
				restoreErr = fmt.Errorf("remove new destination %s: %w", backup.destination, err)
			}
			continue
		}
		if err := os.Rename(backup.backupPath, backup.destination); err != nil {
			restoreErr = fmt.Errorf("restore destination %s: %w", backup.destination, err)
			continue
		}
		backup.backupPath = ""
	}
	return restoreErr
}

func (b *destinationBackup) cleanup() {
	if b != nil && b.backupPath != "" {
		_ = os.Remove(b.backupPath)
	}
}

func (p *preparedWrite) cleanup() {
	if p != nil && p.tempPath != "" {
		_ = os.Remove(p.tempPath)
	}
}

// ensureDestinationWritable fails when an existing destination file is not
// writable by the current process, so a file deliberately marked read-only is
// never silently replaced through the temp-file-and-rename path (which only
// requires a writable parent directory). A destination that does not exist is
// left alone. Probing with O_WRONLY (no truncation, no creation) reflects the
// effective permissions of the current process, including ownership.
func ensureDestinationWritable(path string) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("refusing to overwrite read-only file %s: %w", path, err)
	}
	return f.Close()
}

// resolveDanglingSymlink resolves filename, which must be a symlink whose
// target does not exist yet, by evaluating the existing parent portion of the
// link target while keeping the final target name. The result still passes
// through the link, so writing to it creates the target.
func resolveDanglingSymlink(filename string) (string, error) {
	target, err := os.Readlink(filename)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(filename), target)
	}
	parent := filepath.Dir(target)
	resolvedParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(target)), nil
}
