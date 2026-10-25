package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

func TestUpgradeFromPreviousRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping upgrade e2e test in -short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	root, err := repoRoot()
	require.NoError(t, err)

	// Step 1: Find previous tag
	tag := os.Getenv("FORGE_UPGRADE_FROM")
	if tag == "" {
		cmd := exec.CommandContext(ctx, "git", "-C", root, "tag", "--list", "v0.1.*", "--sort=-v:refname")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("git tag list unavailable: %v (%s)", err, strings.TrimSpace(string(out)))
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			candidate := strings.TrimSpace(line)
			if candidate != "" && strings.HasPrefix(candidate, "v0.1.") {
				tag = candidate
				break
			}
		}
		if tag == "" {
			t.Skip("no v0.1.* tag found in git repository")
		}
	}
	t.Logf("upgrade test using tag: %s", tag)

	// Step 2: Extract tag and build old CLI
	tmpDir := t.TempDir()
	oldSrcDir := filepath.Join(tmpDir, "old")
	require.NoError(t, os.MkdirAll(oldSrcDir, 0755))

	archiveCmd := exec.CommandContext(ctx, "git", "-C", root, "archive", tag)
	tarCmd := exec.CommandContext(ctx, "tar", "-x", "-C", oldSrcDir)

	pipe, err := archiveCmd.StdoutPipe()
	require.NoError(t, err)
	tarCmd.Stdin = pipe

	var archiveErr, tarErr bytes.Buffer
	archiveCmd.Stderr = &archiveErr
	tarCmd.Stderr = &tarErr

	if err := archiveCmd.Start(); err != nil {
		t.Skipf("git archive start failed for tag %s: %v", tag, err)
	}
	if err := tarCmd.Start(); err != nil {
		_ = archiveCmd.Process.Kill()
		t.Fatalf("tar start failed: %v", err)
	}

	errArchive := archiveCmd.Wait()
	errTar := tarCmd.Wait()
	if errArchive != nil {
		t.Skipf("git archive failed for tag %s (shallow clone or missing ref?): %v (%s)", tag, errArchive, strings.TrimSpace(archiveErr.String()))
	}
	require.NoError(t, errTar, "tar extraction failed: %s", tarErr.String())

	oldCLIBin := filepath.Join(tmpDir, "forge-old")
	if runtime.GOOS == "windows" {
		oldCLIBin += ".exe"
	}

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", oldCLIBin, "./cmd/forge")
	buildCmd.Dir = oldSrcDir
	buildCmd.Env = append(os.Environ(), "GOFLAGS=-buildvcs=false")
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build old forge CLI from tag %s:\n%s", tag, string(buildOut))

	// Step 3: Run old CLI new, point replace at old source, write migrations, migrate up, seed data
	projectDir := filepath.Join(tmpDir, "example")
	newCmd := exec.CommandContext(ctx, oldCLIBin, "new", "example", "--path", projectDir, "--template", "simple", "--database", "sqlite", "--docker=false")
	newCmd.Dir = tmpDir
	newCmd.Env = os.Environ()
	newOut, err := newCmd.CombinedOutput()
	require.NoError(t, err, "old forge new failed:\n%s", string(newOut))

	require.NoError(t, patchProjectReplace(projectDir, oldSrcDir))

	migrationsDir := filepath.Join(projectDir, "migrations")
	writeMigrationFiles(t, migrationsDir)

	oldMigrateCmd := exec.CommandContext(ctx, oldCLIBin, "migrate", "up")
	oldMigrateCmd.Dir = projectDir
	oldMigrateCmd.Env = os.Environ()
	oldMigrateOut, err := oldMigrateCmd.CombinedOutput()
	require.NoError(t, err, "old forge migrate up failed:\n%s", string(oldMigrateOut))

	dbPath := filepath.Join(projectDir, "forge.db")
	seedDatabase(t, dbPath)

	dbBefore, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	var countBefore int
	require.NoError(t, dbBefore.QueryRow("SELECT COUNT(*) FROM users").Scan(&countBefore))
	var userBefore, emailBefore string
	require.NoError(t, dbBefore.QueryRow("SELECT username, email FROM users WHERE id = 1").Scan(&userBefore, &emailBefore))
	require.NoError(t, dbBefore.Close())
	t.Logf("seeded database before upgrade: %d row(s), sample: username=%s, email=%s", countBefore, userBefore, emailBefore)

	// Step 4: Switch to current source and run current CLI migrate up
	require.NoError(t, patchProjectReplace(projectDir, root))

	stdout, stderr, err := testhelpers.RunCLI(ctx, projectDir, nil, []string{"migrate", "up"}, 60*time.Second)
	require.NoError(t, err, "current CLI migrate up failed (stdout: %s, stderr: %s)", stdout, stderr)

	// Step 5: Assert seeded rows unchanged and framework tables exist
	dbAfter, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer dbAfter.Close()

	var countAfter int
	require.NoError(t, dbAfter.QueryRow("SELECT COUNT(*) FROM users").Scan(&countAfter))
	require.Equal(t, countBefore, countAfter, "user row count must be unchanged after upgrade")

	var userAfter, emailAfter string
	require.NoError(t, dbAfter.QueryRow("SELECT username, email FROM users WHERE id = 1").Scan(&userAfter, &emailAfter))
	require.Equal(t, userBefore, userAfter, "username must be unchanged after upgrade")
	require.Equal(t, emailBefore, emailAfter, "email must be unchanged after upgrade")

	frameworkTables := []string{
		"forge_framework_migrations",
		"forge_sessions",
		"forge_rate_limits",
		"forge_admin_tokens",
	}
	for _, table := range frameworkTables {
		var tableName string
		err := dbAfter.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		require.NoError(t, err, "framework table %s must exist in sqlite_master", table)
		require.Equal(t, table, tableName)
	}

	var fmMigrations int
	require.NoError(t, dbAfter.QueryRow("SELECT count(*) FROM forge_framework_migrations").Scan(&fmMigrations))
	require.Greater(t, fmMigrations, 0, "forge_framework_migrations must contain applied migrations")

	// Step 6: OpenAPI diff
	runOpenAPIDiff(t, ctx, oldCLIBin, projectDir)
}

func patchProjectReplace(projectDir, targetPath string) error {
	goModPath := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	replaceDirective := fmt.Sprintf("replace github.com/forgego/forge => %s", targetPath)
	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "replace github.com/forgego/forge =>") {
			lines[i] = replaceDirective
			found = true
			break
		}
	}
	var newContent string
	if found {
		newContent = strings.Join(lines, "\n")
	} else {
		newContent = strings.TrimRight(string(content), "\n") + "\n" + replaceDirective + "\n"
	}

	if err := os.WriteFile(goModPath, []byte(newContent), 0644); err != nil {
		return err
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s: %w", string(out), err)
	}
	return nil
}

func runOpenAPIDiff(t *testing.T, ctx context.Context, oldCLIBin, projectDir string) {
	t.Helper()

	// Check if current CLI supports forge generate --api
	currentHelp, _, err := testhelpers.RunCLI(ctx, projectDir, nil, []string{"generate", "--help"}, 15*time.Second)
	if err != nil || !strings.Contains(currentHelp, "--api") {
		t.Log("current CLI does not support 'forge generate --api'; skipping OpenAPI diff")
		return
	}

	// Check if old CLI supports forge generate --api
	oldHelpCmd := exec.CommandContext(ctx, oldCLIBin, "generate", "--help")
	oldHelpCmd.Dir = projectDir
	oldHelpCmd.Env = os.Environ()
	oldHelpOut, err := oldHelpCmd.CombinedOutput()
	if err != nil || !strings.Contains(string(oldHelpOut), "--api") {
		t.Log("old CLI lacks 'generate --api' command/flag; skipping OpenAPI diff")
		return
	}

	// Both CLIs support forge generate --api; attempt generation and compare specs
	t.Log("both CLIs support 'forge generate --api'; running OpenAPI comparison")

	// Ensure models dir exists so old CLI (which defaults to ./models) can run without directory missing error
	modelsDir := filepath.Join(projectDir, "models")
	if _, statErr := os.Stat(modelsDir); os.IsNotExist(statErr) {
		_ = os.MkdirAll(modelsDir, 0755)
	}

	oldGenCmd := exec.CommandContext(ctx, oldCLIBin, "generate", "--api")
	oldGenCmd.Dir = projectDir
	oldGenCmd.Env = os.Environ()
	oldOut, oldGenErr := oldGenCmd.CombinedOutput()

	currOut, _, currGenErr := testhelpers.RunCLI(ctx, projectDir, nil, []string{"generate", "--api"}, 30*time.Second)

	// Helper to find and extract paths from any generated OpenAPI spec file or generated routes
	findPaths := func(dir string) []string {
		candidates := []string{
			filepath.Join(dir, "openapi.json"),
			filepath.Join(dir, "openapi.yaml"),
			filepath.Join(dir, "api", "openapi.json"),
			filepath.Join(dir, "app", "openapi.json"),
		}
		for _, path := range candidates {
			if data, err := os.ReadFile(path); err == nil {
				var doc struct {
					Paths map[string]interface{} `json:"paths"`
				}
				if jsonErr := json.Unmarshal(data, &doc); jsonErr == nil && len(doc.Paths) > 0 {
					var p []string
					for k := range doc.Paths {
						p = append(p, k)
					}
					sort.Strings(p)
					return p
				}
			}
		}
		// Also inspect api_gen.go for registered endpoints if present
		var registeredPaths []string
		_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && info.Name() == "api_gen.go" {
				if content, rErr := os.ReadFile(p); rErr == nil {
					for _, line := range strings.Split(string(content), "\n") {
						if strings.Contains(line, "apiRouter.Register(") {
							registeredPaths = append(registeredPaths, strings.TrimSpace(line))
						}
					}
				}
			}
			return nil
		})
		sort.Strings(registeredPaths)
		return registeredPaths
	}

	oldPaths := findPaths(projectDir)
	currPaths := findPaths(projectDir)

	var diffSummary string
	if len(oldPaths) == 0 && len(currPaths) == 0 {
		diffSummary = "no OpenAPI paths or generated API routes found in either release"
	} else {
		oldSet := make(map[string]bool)
		for _, p := range oldPaths {
			oldSet[p] = true
		}
		currSet := make(map[string]bool)
		for _, p := range currPaths {
			currSet[p] = true
		}
		var added, removed []string
		for _, p := range currPaths {
			if !oldSet[p] {
				added = append(added, p)
			}
		}
		for _, p := range oldPaths {
			if !currSet[p] {
				removed = append(removed, p)
			}
		}
		diffSummary = fmt.Sprintf("added paths: %v, removed paths: %v, unchanged: %d", added, removed, len(currPaths)-len(added))
	}

	t.Logf("OpenAPI diff summary: %s (old CLI out: %q, current CLI out: %q, old err: %v, curr err: %v)",
		diffSummary, strings.TrimSpace(string(oldOut)), strings.TrimSpace(currOut), oldGenErr, currGenErr)
}
