package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// AuthCommand creates the "auth" command for scaffolding auth
type AuthCommand struct{}

// NewAuthCommand creates a new instance of AuthCommand
func NewAuthCommand() *AuthCommand {
	return &AuthCommand{}
}

// Definition returns the cobra command definition
func (c *AuthCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Scaffold authentication app",
		Long:  "Create a complete authentication app with users, sessions, and JWT support",
	}
	return cmd
}

// Execute runs the command logic
func (c *AuthCommand) Execute(ctx *core.Context, args []string) error {
	// Detect project root
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	// Create auth app
	appName := "auth"
	appPath := filepath.Join(projectRoot, "app", appName)

	// Check if auth app already exists
	if _, err := os.Stat(appPath); err == nil {
		return fmt.Errorf("auth app already exists")
	}

	// Validate wiring before writing any file so a wiring failure cannot
	// leave a partial scaffold behind that blocks the retry with
	// "auth app already exists".
	if err := validateAuthWiring(projectRoot); err != nil {
		return err
	}

	// Create app directory
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create auth app directory: %w", err)
	}
	// Create User model.
	// UserObjects is intentionally NOT declared here: forge generate emits
	// `var UserObjects = orm.MustNewManager[User]("users")` into gen.go and
	// a handwritten duplicate would break compilation. Run `forge generate`
	// after scaffolding; api.go and main.go wiring use the generated manager.
	userModel := `package auth

import (
	"time"

	"github.com/forgego/forge/schema"
)

// User represents a user model
type User struct {
	schema.BaseSchema
	ID int64 ` + "`json:\"id\" db:\"id\"`" + `
	Username string ` + "`json:\"username\" db:\"username\"`" + `
	Email string ` + "`json:\"email\" db:\"email\"`" + `
	Password string ` + "`json:\"password\" db:\"password\"`" + `
	IsActive bool ` + "`json:\"is_active\" db:\"is_active\"`" + `
	IsStaff bool ` + "`json:\"is_staff\" db:\"is_staff\"`" + `
	IsSuperuser bool ` + "`json:\"is_superuser\" db:\"is_superuser\"`" + `
	DateJoined time.Time ` + "`json:\"date_joined\" db:\"date_joined\"`" + `
	LastLogin *time.Time ` + "`json:\"last_login\" db:\"last_login\"`" + `
}

// Fields returns all field definitions for User
func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("username", schema.Unique(), schema.Required(), schema.MaxLength(150)),
		schema.StringField("email", schema.Unique(), schema.Required(), schema.MaxLength(255)),
		schema.StringField("password", schema.Required(), schema.MaxLength(128)),
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_staff", schema.Default(false)),
		schema.BoolField("is_superuser", schema.Default(false)),
		schema.TimeField("date_joined", schema.AutoNowAdd()),
		schema.TimeField("last_login"),
	}
}

// Meta returns model metadata
func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName:        "users",
		VerboseName:      "User",
		VerboseNamePlural: "Users",
	}
}

// Relations returns all relationship definitions
func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Hooks returns model lifecycle hooks
func (User) Hooks() *schema.ModelHooks {
	return nil
}

`

	if err := os.WriteFile(filepath.Join(appPath, "models.go"), []byte(userModel), 0644); err != nil {
		_ = os.RemoveAll(appPath)
		return fmt.Errorf("failed to create models.go: %w", err)
	}

	// Create admin.go
	adminCode := `package auth

func init() {
	// Register User model for admin
	// After code generation, uncomment and use:
	// schemaInstance := &UserSchema{} // Your schema implementation
	// manager := orm.NewManager[*User](db) // Your ORM manager
	// config := &admin.Config[*User]{
	//     VerboseName:       "User",
	//     VerboseNamePlural: "Users",
	//     ListPerPage:       20,
	// }
	// admin, err := admin.Register(schemaInstance, manager, config)
	// if err != nil {
	//     log.Fatal(err)
	// }
}
`

	if err := os.WriteFile(filepath.Join(appPath, "admin.go"), []byte(adminCode), 0644); err != nil {
		_ = os.RemoveAll(appPath)
		return fmt.Errorf("failed to create admin.go: %w", err)
	}

	// Create api.go with JWT support
	apiCode := `package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/identity"
	"github.com/forgego/forge/orm"
	httplib "github.com/forgego/forge/server"
)

func init() {
	// Auto-register auth API routes
}

// AuthTokenLifetime is the default JWT lifetime.
const AuthTokenLifetime = 24 * time.Hour

// ErrInvalidSigningKey is returned when the JWT signing key is too short.
var ErrInvalidSigningKey = errors.New("auth: signing key must be at least 32 bytes")

var findUserByUsername = func(ctx context.Context, username string) (*User, error) {
	qs, err := UserObjects.Filter(orm.F("username").Eq(username))
	if err != nil {
		return nil, err
	}
	return qs.First(ctx)
}

func authenticateUser(user *User, password string) bool {
	if user == nil {
		return false
	}
	if !user.IsActive {
		return false
	}
	return identity.CheckPasswordHash(password, user.Password)
}

// RegisterAuthAPI registers authentication API endpoints.
// The signing key must be at least 32 bytes; routes are not registered otherwise.
// Callers should read the key from config security.secret_key.
func RegisterAuthAPI(router *httplib.Router, signingKey []byte) error {
	if len(signingKey) < 32 {
		return ErrInvalidSigningKey
	}
	signingKey = append([]byte(nil), signingKey...)
	// Create viewset for users. The users endpoint requires JWT
	// authentication and staff/superuser permission: without explicit
	// authentication and permission classes an anonymous caller could list,
	// create, update and delete users.
	viewset := api.NewBaseViewSet(
		func() api.Serializer {
			return NewUserSerializer()
		},
		UserObjects,
		&User{},
	)
	viewset.Authentication = []authentication.Authentication{
		authentication.NewJWTAuthentication(signingKey, lookupUserFromClaims),
	}
	viewset.Permissions = []permissions.Permission{
		permissions.NewIsAuthenticated(),
		permissions.NewIsStaffUser(),
	}

	// Register routes
	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("users", &userViewSet{BaseViewSet: viewset})
	apiRouter.RegisterRoutes(router)

	// Register auth endpoints
	router.Post("/api/v1/auth/login", handleLogin(signingKey))
	router.Post("/api/v1/auth/logout", handleLogout)
	return nil
}

// lookupUserFromClaims resolves the JWT subject back to a user for request
// authentication. Unknown users authenticate as anonymous (nil, nil) so the
// permission layer rejects the request.
func lookupUserFromClaims(claims authentication.JWTClaims) (interface{}, error) {
	username, _ := claims["username"].(string)
	if username == "" {
		return nil, nil
	}
	user, err := findUserByUsername(context.Background(), username)
	if err != nil || user == nil {
		return nil, nil
	}
	return user, nil
}

// userViewSet wraps the users BaseViewSet to enforce write-path protections:
// privilege flags cannot be set through the public API and plaintext
// passwords are bcrypt-hashed before they reach the manager.
type userViewSet struct {
	*api.BaseViewSet
}

// Create handles POST /api/v1/users/
func (vs *userViewSet) Create(w http.ResponseWriter, r *http.Request) {
	if !prepareUserWrite(w, r) {
		return
	}
	vs.BaseViewSet.Create(w, r)
}

// Update handles PUT /api/v1/users/{id}/
func (vs *userViewSet) Update(w http.ResponseWriter, r *http.Request) {
	if !prepareUserWrite(w, r) {
		return
	}
	vs.BaseViewSet.Update(w, r)
}

// PartialUpdate handles PATCH /api/v1/users/{id}/
func (vs *userViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	if !prepareUserWrite(w, r) {
		return
	}
	vs.BaseViewSet.PartialUpdate(w, r)
}

// prepareUserWrite rewrites the request body so is_staff/is_superuser can
// never be set through the public endpoints and any supplied password is
// stored as a bcrypt hash rather than plaintext.
func prepareUserWrite(w http.ResponseWriter, r *http.Request) bool {
	if r.Body == nil {
		return true
	}
	body, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	if len(bytes.TrimSpace(body)) == 0 {
		r.Body = io.NopCloser(bytes.NewReader(body))
		return true
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		// Restore the original body and let the viewset report the 400.
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		return true
	}
	delete(data, "is_staff")
	delete(data, "is_superuser")
	if raw, ok := data["password"]; ok {
		password, _ := raw.(string)
		if strings.TrimSpace(password) == "" {
			delete(data, "password")
		} else {
			hashed, err := identity.HashPassword(password)
			if err != nil {
				http.Error(w, "could not process password", http.StatusInternalServerError)
				return false
			}
			data["password"] = hashed
		}
	}
	rewritten, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(rewritten))
	r.ContentLength = int64(len(rewritten))
	return true
}

// UserSerializer serializes User model
type UserSerializer struct {
	*api.BaseSerializer
}

// NewUserSerializer creates a new serializer
func NewUserSerializer() api.Serializer {
	return &UserSerializer{
		BaseSerializer: api.NewBaseSerializer(nil),
	}
}

// New creates a new serializer instance
func (s *UserSerializer) New() api.Serializer {
	return NewUserSerializer()
}

// Fields returns the fields to serialize. Password, is_staff and
// is_superuser are deliberately absent: the password hash must never be
// rendered and privilege flags are not part of the public representation.
func (s *UserSerializer) Fields() []string {
	return []string{"id", "username", "email", "is_active", "date_joined"}
}

// ReadOnlyFields prevents clients from setting the id, privilege flags or
// server-managed timestamps through the API.
func (s *UserSerializer) ReadOnlyFields() []string {
	return []string{"id", "is_staff", "is_superuser", "date_joined", "last_login"}
}

// WriteOnlyFields ensures passwords are accepted on write but never
// included in responses.
func (s *UserSerializer) WriteOnlyFields() []string {
	return []string{"password"}
}

func handleLogin(signingKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()
	var payload map[string]string
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(payload["username"])
	password := payload["password"]
	if username == "" || password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := findUserByUsername(req.Context(), username)
	if err != nil || !authenticateUser(user, password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWTToken(strconv.FormatInt(user.ID, 10), user.Username, signingKey)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"token_type": "Bearer",
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
	}
}

func handleLogout(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func generateJWTToken(userID, username string, signingKey []byte) (string, error) {
	if len(signingKey) < 32 {
		return "", ErrInvalidSigningKey
	}
	now := time.Now()
	headerJSON := []byte("{\"alg\":\"HS256\",\"typ\":\"JWT\"}")
	payloadJSON, err := json.Marshal(map[string]any{
		"sub":      userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      now.Add(AuthTokenLifetime).Unix(),
	})
	if err != nil {
		return "", err
	}
	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsigned := header + "." + payload
	mac := hmac.New(sha256.New, signingKey)
	_, _ = mac.Write([]byte(unsigned))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsigned + "." + signature, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
`

	if err := os.WriteFile(filepath.Join(appPath, "api.go"), []byte(apiCode), 0644); err != nil {
		_ = os.RemoveAll(appPath)
		return fmt.Errorf("failed to create api.go: %w", err)
	}
	if err := wireAuthAPI(projectRoot); err != nil {
		// Remove only the directory this run created (it did not exist
		// when we started) so the user can fix main.go and retry.
		_ = os.RemoveAll(appPath)
		return err
	}

	fmt.Printf("✓ Scaffolded auth app\n")
	fmt.Printf("  Location: %s\n", appPath)
	fmt.Printf("  Created: User model, admin config, API endpoints\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Run: forge generate --models app --output app (required: generates UserObjects)\n")
	fmt.Printf("  2. Run: forge makemigrations\n")
	fmt.Printf("  3. Run: forge migrate\n")

	return nil
}

// validateAuthWiring performs every wiring check that wireAuthAPI needs
// without writing any file, so Execute can fail fast before creating the
// auth app directory.
func validateAuthWiring(projectRoot string) error {
	_, _, _, err := checkAuthWiring(projectRoot)
	return err
}

// checkAuthWiring reads main.go and go.mod and verifies the wiring can be
// applied. It returns the module path and file content for wireAuthAPI to
// use, plus alreadyWired when no change is needed.
func checkAuthWiring(projectRoot string) (modulePath, content string, alreadyWired bool, err error) {
	mainPath := filepath.Join(projectRoot, "main.go")
	mainBytes, readErr := os.ReadFile(mainPath)
	if os.IsNotExist(readErr) {
		return "", "", true, nil
	}
	if readErr != nil {
		return "", "", false, fmt.Errorf("failed to read main.go: %w", readErr)
	}
	moduleBytes, readErr := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if readErr != nil {
		return "", "", false, fmt.Errorf("failed to read go.mod: %w", readErr)
	}
	for _, line := range strings.Split(string(moduleBytes), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			modulePath = fields[1]
			break
		}
	}
	if modulePath == "" {
		return "", "", false, fmt.Errorf("failed to wire auth API: go.mod has no module directive")
	}

	content = string(mainBytes)
	if strings.Contains(content, "auth.RegisterAuthAPI(") {
		return modulePath, content, true, nil
	}
	serverImport := `"github.com/forgego/forge/server"`
	if !strings.Contains(content, serverImport) {
		return "", "", false, fmt.Errorf("failed to wire auth API: server import not found in main.go")
	}
	adminRoutes := `		if settings.Admin.Enabled {
			router.Mount(settings.Admin.Path, adminSite.Handler())
		}`
	if !strings.Contains(content, adminRoutes) {
		return "", "", false, fmt.Errorf("failed to wire auth API: route registration block not found in main.go")
	}
	return modulePath, content, false, nil
}

func wireAuthAPI(projectRoot string) error {
	mainPath := filepath.Join(projectRoot, "main.go")
	modulePath, content, alreadyWired, err := checkAuthWiring(projectRoot)
	if err != nil {
		return err
	}
	if alreadyWired {
		// No main.go to wire, or auth routes are already registered.
		return nil
	}
	serverImport := `"github.com/forgego/forge/server"`
	content = strings.Replace(content, serverImport, fmt.Sprintf("%q\n\t%s", modulePath+"/app/auth", serverImport), 1)

	adminRoutes := `		if settings.Admin.Enabled {
			router.Mount(settings.Admin.Path, adminSite.Handler())
		}`
	authRoutes := adminRoutes + `

		auth.UserObjects.SetDB(database)
		if err := auth.RegisterAuthAPI(router, []byte(settings.Security.SecretKey)); err != nil {
			stdlog.Fatal(err)
		}`
	content = strings.Replace(content, adminRoutes, authRoutes, 1)
	if err := os.WriteFile(mainPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to update main.go: %w", err)
	}
	return nil
}
