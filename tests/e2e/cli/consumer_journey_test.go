package cli

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/forgego/forge/tests/testhelpers"
	"github.com/stretchr/testify/require"
)

func TestConsumerJourney(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	database, dbEnv := requireJourneyPostgres(t, ctx)
	workdir := t.TempDir()
	repo, err := repoRoot()
	require.NoError(t, err)
	require.NotContains(t, workdir, repo+string(filepath.Separator), "consumer project must live outside the checkout")

	projectDir := filepath.Join(workdir, "consumer")
	app := &journeyApp{t: t, ctx: ctx, dir: projectDir, env: dbEnv, db: database}
	app.forgeIn(workdir, "new", "consumer", "--path", projectDir,
		"--template", "simple", "--database", "postgres", "--docker=false")
	require.NoError(t, patchGeneratedProject(projectDir))
	discloseLocalReplace(t, projectDir)

	modelFixture, err := os.ReadFile(filepath.Join("..", "..", "testdata", "models", "consumer", "models.go.fixture"))
	require.NoError(t, err)
	modelsSource := strings.Replace(string(modelFixture), "package models\n", "package work\n", 1)
	require.NotEqual(t, string(modelFixture), modelsSource, "fixture package clause should be adapted")
	testhelpers.WriteFileString(t, filepath.Join(projectDir, "app", "work", "models.go"), modelsSource)

	overlaySource, err := os.ReadFile(filepath.Join("..", "..", "testdata", "reference_app", "operations", "approve_order.go"))
	require.NoError(t, err)
	testhelpers.WriteFileString(t, filepath.Join(projectDir, "app", "work", "operations", "approve_order.go"), string(overlaySource))
	testhelpers.WriteFileString(t, filepath.Join(projectDir, "app", "work", "routes.go"), consumerJourneyRoutes)
	registerAppInMainForModule(t, filepath.Join(projectDir, "cmd", "server", "main.go"), "consumer")

	port := randomPort(t)
	updateServerPort(t, filepath.Join(projectDir, "config", "config.yaml"), port)
	app.forge("generate", "--api", "--models", "./app/work", "--output", "./app/work")
	app.goCmd("mod", "tidy")
	generatedHashes := hashConsumerGeneratedFiles(t, projectDir)
	app.forge("makemigrations", "initial", "--auto", "--models", "./app/work")
	app.forge("migrate", "up")
	app.goCmd("build", "./...")
	app.goCmd("build", "-o", filepath.Join("bin", serverBinaryName()), "./cmd/server")

	stop := app.startServer(port)
	checkConsumerJourneyEndpoints(t, port)
	stop()
	require.Equal(t, generatedHashes, hashConsumerGeneratedFiles(t, projectDir), "generated files changed during the run")
}

const consumerJourneyRoutes = `package work

import (
	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
)

func Register(router *server.Router, database *db.DB) {
	OrderObjects.SetDB(database)
	CustomerObjects.SetDB(database)
	if _, err := admin.Register[Order](&admin.Config[Order]{}); err != nil { panic(err) }
	if _, err := admin.Register[Customer](&admin.Config[Customer]{}); err != nil { panic(err) }
	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("orders", NewOrderViewSet())
	apiRouter.Register("customers", NewCustomerViewSet())
	apiRouter.RegisterRoutes(router)
}
`

func discloseLocalReplace(t *testing.T, projectDir string) {
	t.Helper()
	path := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	replace := "replace github.com/forgego/forge => "
	require.Contains(t, string(content), replace)
	disclosed := strings.Replace(string(content), replace, "// Disclosed local replace for this candidate checkout.\n"+replace, 1)
	require.NoError(t, os.WriteFile(path, []byte(disclosed), 0o644))
}

func serverBinaryName() string {
	if runtime.GOOS == "windows" {
		return "server.exe"
	}
	return "server"
}

func hashConsumerGeneratedFiles(t *testing.T, projectDir string) map[string][32]byte {
	t.Helper()
	hashes := make(map[string][32]byte)
	for _, name := range []string{"gen.go", "api_gen.go"} {
		path := filepath.Join(projectDir, "app", "work", name)
		content, err := os.ReadFile(path)
		require.NoError(t, err, "generated file %s", name)
		hashes[name] = sha256.Sum256(content)
	}
	require.Len(t, hashes, 2)
	return hashes
}

func checkConsumerJourneyEndpoints(t *testing.T, port string) {
	t.Helper()
	base := fmt.Sprintf("http://127.0.0.1:%s", port)
	client := &http.Client{Timeout: 5 * time.Second}
	for _, path := range []string{"/health", "/health/ready"} {
		resp, err := client.Get(base + path)
		require.NoError(t, err, "GET %s", path)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, "GET %s", path)
	}
	adminResp, err := client.Get(base + "/admin/")
	require.NoError(t, err, "GET admin entry point")
	adminResp.Body.Close()
	require.Equal(t, http.StatusOK, adminResp.StatusCode, "GET admin entry point")

	apiResp, err := client.Get(base + "/api/v1/orders/")
	require.NoError(t, err, "GET published orders list")
	apiResp.Body.Close()
	require.Equal(t, http.StatusOK, apiResp.StatusCode, "default REST list access")
}
