package cli

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

func TestCLIProjectHarness(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	workdir, cleanup := testhelpers.TempWorkdir(t, "forge-e2e-project-")
	defer cleanup()

	projectDir := filepath.Join(workdir, "example")
	stdout, _, err := testhelpers.RunCLI(ctx, workdir, nil, []string{
		"new",
		"example",
		"--path",
		projectDir,
		"--template",
		"simple",
		"--database",
		"sqlite",
		"--docker",
	}, 30*time.Second)
	require.NoError(t, err, "forge new output: %s", stdout)

	require.NoError(t, patchGeneratedProject(projectDir))

	migrationsDir := filepath.Join(projectDir, "migrations")
	writeMigrationFiles(t, migrationsDir)

	stdout, _, err = testhelpers.RunCLI(ctx, projectDir, nil, []string{"migrate", "up"}, 30*time.Second)
	require.NoError(t, err, "forge migrate output: %s", stdout)

	seedPath := filepath.Join(projectDir, "forge.db")
	seedDatabase(t, seedPath)

	port := randomPort(t)
	updateServerPort(t, filepath.Join(projectDir, "config", "config.yaml"), port)

	require.NoError(t, updateMainForAdmin(projectDir))

	serverCmd, serverLogs := startServer(t, ctx, projectDir)
	t.Cleanup(func() {
		shutdownServer(t, serverCmd, serverLogs)
	})
	waitForHealthy(t, port)
	checkAdmin(t, port)
}

func patchGeneratedProject(projectDir string) error {
	goModPath := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	repoRoot, err := repoRoot()
	if err != nil {
		return err
	}

	replaceLine := fmt.Sprintf("\nreplace github.com/forgego/forge => %s\n", filepath.Join(repoRoot, "forge"))
	if strings.Contains(string(content), "replace github.com/forgego/forge =>") {
		return nil
	}
	return os.WriteFile(goModPath, append(content, []byte(replaceLine)...), 0644)
}

func repoRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to determine current file")
	}
	return filepath.Abs(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func writeMigrationFiles(t *testing.T, migrationsDir string) {
	t.Helper()
	migrationUp := `
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	email TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
	migrationDown := `DROP TABLE users;`
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))
	testhelpers.WriteFileString(t, filepath.Join(migrationsDir, "000001_create_users.up.sql"), migrationUp)
	testhelpers.WriteFileString(t, filepath.Join(migrationsDir, "000001_create_users.down.sql"), migrationDown)
}

func seedDatabase(t *testing.T, dbPath string) {
	t.Helper()
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`INSERT INTO users (username, email) VALUES (?, ?)`, "admin", "admin@example.com")
	require.NoError(t, err)
}

func randomPort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)
	return strconv.Itoa(addr.Port)
}

func updateServerPort(t *testing.T, configPath, port string) {
	t.Helper()
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	updated := strings.Replace(string(content), "port: 8000", "port: "+port, 1)
	require.NotEqual(t, string(content), updated)
	require.NoError(t, os.WriteFile(configPath, []byte(updated), 0644))
}

func updateMainForAdmin(projectDir string) error {
	mainPath := filepath.Join(projectDir, "cmd", "server", "main.go")
	mainContent := `package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	forgelog "github.com/forgego/forge/log"
	"github.com/forgego/forge/server"
	stdlog "log"
)

func main() {
	cfg := config.NewConfig()
	settings := config.LoadSettings(cfg)

	logger, err := forgelog.NewLogger(settings.App.Debug)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer logger.Sync()

	database, err := db.NewDBFromConfig(cfg)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer database.Close()

	srv, err := server.NewServer(cfg, settings, logger)
	if err != nil {
		stdlog.Fatal(err)
	}

	srv.RegisterRoutes(func(router *server.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Welcome to example!")
		})

		if settings.Admin.Enabled {
			adminPath := strings.TrimRight(settings.Admin.Path, "/")
			site := admin.DefaultSite.WithUIConfig(admin.UIConfig{Prefix: adminPath})
			site.SetDB(database)
			router.Mount(adminPath, site.Handler())
		}
	})

	fmt.Printf("Starting server on %s:%s\n", settings.Server.Host, settings.Server.Port)
	if err := srv.Start(); err != nil {
		stdlog.Fatal(err)
	}
}
`
	return os.WriteFile(mainPath, []byte(mainContent), 0644)
}

func startServer(t *testing.T, ctx context.Context, projectDir string) (*exec.Cmd, *bytes.Buffer) {
	t.Helper()
	serverCtx, _ := context.WithCancel(ctx)
	cmd := exec.CommandContext(serverCtx, "go", "run", "cmd/server/main.go")
	cmd.Dir = projectDir

	var logs bytes.Buffer
	cmd.Stdout = &logs
	cmd.Stderr = &logs

	require.NoError(t, cmd.Start())
	return cmd, &logs
}

func waitForHealthy(t *testing.T, port string) {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%s/health", port)
	client := &http.Client{Timeout: 2 * time.Second}

	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	require.Fail(t, "server did not become healthy")
}

func checkAdmin(t *testing.T, port string) {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%s/admin/api/config", port)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func shutdownServer(t *testing.T, cmd *exec.Cmd, logs *bytes.Buffer) {
	t.Helper()
	if cmd.Process == nil {
		return
	}

	_ = cmd.Process.Signal(os.Interrupt)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	case err := <-done:
		if err != nil && logs != nil {
			t.Logf("server logs:\n%s", logs.String())
		}
	}
}
