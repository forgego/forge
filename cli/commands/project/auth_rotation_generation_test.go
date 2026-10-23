package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	generator "github.com/forgego/forge/codegen"
	"github.com/stretchr/testify/require"
)

func TestGeneratedAuthProjectRotatesJWTKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated-project runtime test in short mode")
	}

	forgePath, err := filepath.Abs("../../..")
	require.NoError(t, err)
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module example.com/authrotation\n\ngo 1.26.0\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => "+forgePath+"\n"), 0644))

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	require.NoError(t, os.Chdir(projectPath))
	require.NoError(t, NewAuthCommand().Execute(&core.Context{}, nil))

	appPath := filepath.Join(projectPath, "app", "auth")
	require.NoError(t, generator.NewGenerator(appPath, appPath).Generate())

	runtimeTest := `package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/config"
)

func TestGeneratedJWTKeyRotationRuntime(t *testing.T) {
	currentKey := []byte("current-signing-key-must-be-at-least-32-bytes")
	previousKey := []byte("previous-signing-key-must-be-at-least-32-bytes")
	claims := authentication.JWTClaims{"sub": "1", "username": "alice", "exp": time.Now().Add(time.Hour).Unix()}
	lookup := func(authentication.JWTClaims) (interface{}, error) { return &User{IsActive: true}, nil }
	legacyAuth, err := newJWTAuthentication(config.SecuritySettings{SecretKey: string(currentKey)}, lookup)
	if err != nil { t.Fatal(err) }
	legacyToken, err := legacyAuth.SignToken(claims)
	if err != nil { t.Fatal(err) }
	legacyHeader, err := base64.RawURLEncoding.DecodeString(strings.Split(legacyToken, ".")[0])
	if err != nil { t.Fatal(err) }
	var legacyFields map[string]interface{}
	if err := json.Unmarshal(legacyHeader, &legacyFields); err != nil { t.Fatal(err) }
	if _, ok := legacyFields["kid"]; ok { t.Fatalf("legacy token unexpectedly has kid: %v", legacyFields["kid"]) }
	legacyRequest := httptest.NewRequest("GET", "/", nil)
	legacyRequest.Header.Set("Authorization", "Bearer "+legacyToken)
	if _, err := legacyAuth.Authenticate(legacyRequest); err != nil { t.Fatalf("legacy token rejected: %v", err) }

	missingID := config.SecuritySettings{
		SecretKey: string(currentKey),
		PreviousSecretKeys: []config.PreviousSecretKeySettings{{ID: "previous", Key: string(previousKey), VerifyUntil: time.Now().Add(time.Hour)}},
	}
	if _, err := newJWTAuthentication(missingID, lookup); err == nil { t.Fatal("rotation settings without a current key ID were accepted") }

	configured := config.SecuritySettings{
		SecretKey: string(currentKey), SecretKeyID: "current",
		PreviousSecretKeys: []config.PreviousSecretKeySettings{{ID: "previous", Key: string(previousKey), VerifyUntil: time.Now().Add(time.Hour)}},
	}
	currentAuth, err := newJWTAuthentication(configured, lookup)
	if err != nil { t.Fatal(err) }
	token, err := currentAuth.SignToken(claims)
	if err != nil { t.Fatal(err) }
	parts := strings.Split(token, ".")
	if len(parts) != 3 { t.Fatalf("malformed token: %q", token) }
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil { t.Fatal(err) }
	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil { t.Fatal(err) }
	if header["kid"] != "current" { t.Fatalf("signed kid = %v, want current", header["kid"]) }

	previousAuth := authentication.NewJWTAuthenticationWithKeys("previous", previousKey, nil, nil)
	previousToken, err := previousAuth.SignToken(claims)
	if err != nil { t.Fatal(err) }
	request := func(value string) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+value)
		return req
	}
	if _, err := currentAuth.Authenticate(request(previousToken)); err != nil { t.Fatalf("previous key rejected before deadline: %v", err) }

	configured.PreviousSecretKeys[0].VerifyUntil = time.Now().Add(-time.Second)
	expiredAuth, err := newJWTAuthentication(configured, lookup)
	if err != nil { t.Fatal(err) }
	if _, err := expiredAuth.Authenticate(request(previousToken)); err == nil { t.Fatal("previous key accepted after deadline") }
}
`
	require.NoError(t, os.WriteFile(filepath.Join(appPath, "auth_rotation_test.go"), []byte(runtimeTest), 0644))

	cmd := exec.Command("go", "test", "-mod=mod", "./app/auth", "-run", "^TestGeneratedJWTKeyRotationRuntime$")
	cmd.Dir = projectPath
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=go1.26.10", "GOFLAGS=-buildvcs=false")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated auth project key rotation test failed:\n%s", output)
}
