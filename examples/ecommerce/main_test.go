package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/admin/components"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{name: "empty_uses_fallback", value: "", fallback: "/admin", want: "/admin"},
		{name: "adds_leading_slash", value: "api/v1", fallback: "/api", want: "/api/v1"},
		{name: "trims_trailing_slash", value: "/admin/", fallback: "/admin", want: "/admin"},
		{name: "single_slash_root", value: "/", fallback: "/admin", want: "/"},
		{name: "whitespace_uses_fallback", value: "   ", fallback: "/api/v1", want: "/api/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePath(tt.value, tt.fallback)
			if got != tt.want {
				t.Fatalf("normalizePath(%q, %q) = %q, want %q", tt.value, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestSetupDashboardRegistersDashboard(t *testing.T) {
	SetupDashboard()

	cfg := core.GetDashboard(context.Background())
	if cfg == nil {
		t.Fatal("expected dashboard config, got nil")
	}
	if cfg.Title != "Ecommerce Dashboard" {
		t.Fatalf("expected dashboard title %q, got %q", "Ecommerce Dashboard", cfg.Title)
	}
	if cfg.Layout.Type != components.TypeGrid {
		t.Fatalf("expected dashboard layout type %q, got %q", components.TypeGrid, cfg.Layout.Type)
	}
	if len(cfg.Layout.Children) == 0 {
		t.Fatal("expected dashboard layout children to be populated")
	}
}

func TestReportsPluginMenuAndPages(t *testing.T) {
	plugin := &ReportsPlugin{}

	if plugin.ID() != "reports" {
		t.Fatalf("expected plugin id %q, got %q", "reports", plugin.ID())
	}
	if plugin.Name() != "Reports" {
		t.Fatalf("expected plugin name %q, got %q", "Reports", plugin.Name())
	}

	items := plugin.GetMenuItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 top-level menu item, got %d", len(items))
	}
	if items[0].Path != "/plugins/reports/pages/overview" {
		t.Fatalf("unexpected menu path %q", items[0].Path)
	}
	if len(items[0].Children) == 0 {
		t.Fatal("expected reports menu to include child pages")
	}

	pages := plugin.GetPages()
	if len(pages) != 2 {
		t.Fatalf("expected 2 plugin pages, got %d", len(pages))
	}
	if pages["overview"].Type != components.TypePage {
		t.Fatalf("expected overview page component type %q, got %q", components.TypePage, pages["overview"].Type)
	}
	if pages["sales"].Type != components.TypePage {
		t.Fatalf("expected sales page component type %q, got %q", components.TypePage, pages["sales"].Type)
	}
}

func TestBuildEcommerceRouter_HTTPReachability(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping HTTP reachability test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	for _, tc := range []struct {
		name  string
		path  string
		check func(int, string) error
	}{
		{
			name: "health endpoint",
			path: "/health",
			check: func(status int, body string) error {
				if status != http.StatusOK {
					return newTestErr("expected /health status 200, got %d", status)
				}
				if !strings.Contains(body, `"status":"healthy"`) {
					return newTestErr("expected healthy body, got %q", body)
				}
				return nil
			},
		},
		{
			name: "admin reachability",
			path: "/admin/",
			check: func(status int, _ string) error {
				if status == http.StatusNotFound || status >= http.StatusInternalServerError {
					return newTestErr("expected /admin/ to be reachable, got status %d", status)
				}
				return nil
			},
		},
		{
			name: "api list reachability",
			path: "/api/v1/products/",
			check: func(status int, body string) error {
				if status != http.StatusOK {
					return newTestErr("expected /api/v1/products/ status 200, got %d", status)
				}
				if !strings.Contains(body, `"results"`) {
					return newTestErr("expected products list payload, got %q", body)
				}
				return nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if err := tc.check(rr.Code, rr.Body.String()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBuildEcommerceRouter_APICategoryCRUDFlow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-crud-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping ecommerce CRUD flow test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	createdCategory := performJSONRequest(t, router, http.MethodPost, "/api/v1/categories/", map[string]interface{}{
		"name":        "Electronics",
		"slug":        "electronics",
		"description": "Initial category",
	}, http.StatusCreated)

	categoryID := int64(createdCategory["id"].(float64))

	retrievedCategory := performJSONRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/categories/%d", categoryID), nil, http.StatusOK)
	if got, _ := retrievedCategory["slug"].(string); got != "electronics" {
		t.Fatalf("expected slug %q, got %q", "electronics", got)
	}

	updatedCategory := performJSONRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/v1/categories/%d", categoryID), map[string]interface{}{
		"name":        "Electronics and Gadgets",
		"slug":        "electronics",
		"description": "Updated category",
		"is_active":   false,
	}, http.StatusOK)
	if got, _ := updatedCategory["name"].(string); got != "Electronics and Gadgets" {
		t.Fatalf("expected updated name %q, got %q", "Electronics and Gadgets", got)
	}

	filteredList := performJSONRequest(t, router, http.MethodGet, "/api/v1/categories/?name__contains=Gadgets", nil, http.StatusOK)
	results, ok := filteredList["results"].([]interface{})
	if !ok || len(results) == 0 {
		t.Fatalf("expected non-empty filtered results, got %v", filteredList["results"])
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/categories/%d", categoryID), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d with body %q", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/categories/%d", categoryID), nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected deleted category to return 404, got %d with body %q", rr.Code, rr.Body.String())
	}
}

func TestBuildEcommerceRouter_AdminAPIAuthFlow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-auth-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin auth flow test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	unauthBody, err := json.Marshal(map[string]interface{}{
		"name":        "Unauthorized",
		"slug":        "unauthorized-category",
		"description": "should fail without token",
	})
	if err != nil {
		t.Fatalf("failed to marshal unauth payload: %v", err)
	}

	unauthReq := httptest.NewRequest(http.MethodPost, "/admin/api/categories/", bytes.NewReader(unauthBody))
	unauthReq.Header.Set("Content-Type", "application/json")
	unauthRec := httptest.NewRecorder()
	router.ServeHTTP(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated admin create to return 401, got %d with body %q", unauthRec.Code, unauthRec.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	authBody, err := json.Marshal(map[string]interface{}{
		"name":        "Authorized",
		"slug":        "authorized-category",
		"description": "created with token",
	})
	if err != nil {
		t.Fatalf("failed to marshal auth payload: %v", err)
	}

	authReq := httptest.NewRequest(http.MethodPost, "/admin/api/categories/", bytes.NewReader(authBody))
	authReq.Header.Set("Content-Type", "application/json")
	authReq.Header.Set("Authorization", "Bearer "+token)
	authRec := httptest.NewRecorder()
	router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusCreated {
		t.Fatalf("expected authenticated admin create status 201, got %d with body %q", authRec.Code, authRec.Body.String())
	}

	var createdCategory map[string]interface{}
	if err := json.Unmarshal(authRec.Body.Bytes(), &createdCategory); err != nil {
		t.Fatalf("failed to decode authenticated create response: %v", err)
	}
	rawID, ok := createdCategory["id"].(float64)
	if !ok {
		t.Fatalf("expected created category id in response, got %v", createdCategory["id"])
	}
	categoryID := int64(rawID)

	invalidUpdateReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/admin/api/categories/%d/", categoryID),
		bytes.NewBufferString(`{"description":"blocked by invalid token"}`),
	)
	invalidUpdateReq.Header.Set("Content-Type", "application/json")
	invalidUpdateReq.Header.Set("Authorization", "Bearer invalid-token")
	invalidUpdateRec := httptest.NewRecorder()
	router.ServeHTTP(invalidUpdateRec, invalidUpdateReq)
	if invalidUpdateRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid-token update to return 401, got %d with body %q", invalidUpdateRec.Code, invalidUpdateRec.Body.String())
	}

	validUpdateReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/admin/api/categories/%d/", categoryID),
		bytes.NewBufferString(`{"description":"updated with valid token"}`),
	)
	validUpdateReq.Header.Set("Content-Type", "application/json")
	validUpdateReq.Header.Set("Authorization", "Bearer "+token)
	validUpdateRec := httptest.NewRecorder()
	router.ServeHTTP(validUpdateRec, validUpdateReq)
	if validUpdateRec.Code != http.StatusOK {
		t.Fatalf("expected authenticated admin update status 200, got %d with body %q", validUpdateRec.Code, validUpdateRec.Body.String())
	}

	authListReq := httptest.NewRequest(http.MethodGet, "/admin/api/categories/", nil)
	authListReq.Header.Set("Authorization", "Bearer "+token)
	authListRec := httptest.NewRecorder()
	router.ServeHTTP(authListRec, authListReq)
	if authListRec.Code != http.StatusOK {
		t.Fatalf("expected authenticated admin list status 200, got %d with body %q", authListRec.Code, authListRec.Body.String())
	}

	invalidDeleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/api/categories/%d/", categoryID), nil)
	invalidDeleteReq.Header.Set("Authorization", "Bearer invalid-token")
	invalidDeleteRec := httptest.NewRecorder()
	router.ServeHTTP(invalidDeleteRec, invalidDeleteReq)
	if invalidDeleteRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid-token delete to return 401, got %d with body %q", invalidDeleteRec.Code, invalidDeleteRec.Body.String())
	}

	validDeleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/api/categories/%d/", categoryID), nil)
	validDeleteReq.Header.Set("Authorization", "Bearer "+token)
	validDeleteRec := httptest.NewRecorder()
	router.ServeHTTP(validDeleteRec, validDeleteReq)
	if validDeleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected authenticated admin delete status 204, got %d with body %q", validDeleteRec.Code, validDeleteRec.Body.String())
	}

	authGetDeletedReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/categories/%d/", categoryID), nil)
	authGetDeletedReq.Header.Set("Authorization", "Bearer "+token)
	authGetDeletedRec := httptest.NewRecorder()
	router.ServeHTTP(authGetDeletedRec, authGetDeletedReq)
	if authGetDeletedRec.Code != http.StatusNotFound {
		t.Fatalf("expected deleted category to return 404 through admin API, got %d with body %q", authGetDeletedRec.Code, authGetDeletedRec.Body.String())
	}
}

func TestBuildEcommerceRouter_AdminAPILoginConfiguredCredentials(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "ecommerce-admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "ecommerce-secret")

	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-login-config-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin configured-credentials test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	defaultLoginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	defaultLoginReq.Header.Set("Content-Type", "application/json")
	defaultLoginRec := httptest.NewRecorder()
	router.ServeHTTP(defaultLoginRec, defaultLoginReq)
	if defaultLoginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected default credentials to be rejected with configured credentials, got %d with body %q", defaultLoginRec.Code, defaultLoginRec.Body.String())
	}

	configuredLoginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"ecommerce-admin","password":"ecommerce-secret"}`))
	configuredLoginReq.Header.Set("Content-Type", "application/json")
	configuredLoginRec := httptest.NewRecorder()
	router.ServeHTTP(configuredLoginRec, configuredLoginReq)
	if configuredLoginRec.Code != http.StatusOK {
		t.Fatalf("expected configured credentials login status 200, got %d with body %q", configuredLoginRec.Code, configuredLoginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(configuredLoginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode configured credentials login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	authReq := httptest.NewRequest(http.MethodGet, "/admin/api/categories/", nil)
	authReq.Header.Set("Authorization", "Bearer "+token)
	authRec := httptest.NewRecorder()
	router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusOK {
		t.Fatalf("expected authenticated admin list status 200, got %d with body %q", authRec.Code, authRec.Body.String())
	}
}

func TestBuildEcommerceRouter_AdminAPIDeletePermissionDeniedFlow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin permission deny flow test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{
			name: "order delete is denied",
			path: "/admin/api/orders/1/",
		},
		{
			name: "payment delete is denied",
			path: "/admin/api/payments/1/",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected %s to return 403, got %d with body %q", tc.path, rec.Code, rec.Body.String())
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("failed to decode permission denied payload for %s: %v", tc.path, err)
			}
			errorPayload, ok := payload["error"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected error payload for %s, got %v", tc.path, payload)
			}
			if gotCode, _ := errorPayload["code"].(string); gotCode != "permission_denied" {
				t.Fatalf("expected permission_denied code for %s, got %q (payload=%v)", tc.path, gotCode, payload)
			}
		})
	}
}

func TestBuildEcommerceRouter_AdminAPIObjectSpecificPaymentChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-payment-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	ctx := context.Background()
	res, err := database.ExecContext(ctx, `
		INSERT INTO customers (email, password_hash, first_name, last_name)
		VALUES (?, ?, ?, ?)
	`, "perm-test@example.com", "hashed", "Perm", "Test")
	if err != nil {
		t.Fatalf("failed to insert customer fixture: %v", err)
	}
	customerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve customer fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-PERM-001", customerID, "perm-test@example.com", "Perm", "Test", 100.00, 8.00, 5.00, 113.00)
	if err != nil {
		t.Fatalf("failed to insert order fixture: %v", err)
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve order fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO payments (order_id, amount, payment_method, status)
		VALUES (?, ?, ?, ?)
	`, orderID, 113.00, "card", "completed")
	if err != nil {
		t.Fatalf("failed to insert completed payment fixture: %v", err)
	}
	completedPaymentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve completed payment id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO payments (order_id, amount, payment_method, status)
		VALUES (?, ?, ?, ?)
	`, orderID, 113.00, "card", "pending")
	if err != nil {
		t.Fatalf("failed to insert pending payment fixture: %v", err)
	}
	pendingPaymentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve pending payment id: %v", err)
	}

	completedReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/admin/api/payments/%d/", completedPaymentID),
		bytes.NewBufferString(`{"gateway_response":"manual note"}`),
	)
	completedReq.Header.Set("Content-Type", "application/json")
	completedReq.Header.Set("Authorization", "Bearer "+token)
	completedRec := httptest.NewRecorder()
	router.ServeHTTP(completedRec, completedReq)
	if completedRec.Code != http.StatusForbidden {
		t.Fatalf("expected completed payment update to return 403, got %d with body %q", completedRec.Code, completedRec.Body.String())
	}

	var completedPayload map[string]interface{}
	if err := json.Unmarshal(completedRec.Body.Bytes(), &completedPayload); err != nil {
		t.Fatalf("failed to decode completed payment denial response: %v", err)
	}
	completedError, ok := completedPayload["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error payload for completed payment denial, got %v", completedPayload)
	}
	if gotCode, _ := completedError["code"].(string); gotCode != "permission_denied" {
		t.Fatalf("expected permission_denied code for completed payment update, got %q (payload=%v)", gotCode, completedPayload)
	}

	pendingReq := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/admin/api/payments/%d/", pendingPaymentID),
		bytes.NewBufferString(`{"gateway_response":"manual note"}`),
	)
	pendingReq.Header.Set("Content-Type", "application/json")
	pendingReq.Header.Set("Authorization", "Bearer "+token)
	pendingRec := httptest.NewRecorder()
	router.ServeHTTP(pendingRec, pendingReq)
	if pendingRec.Code != http.StatusOK {
		t.Fatalf("expected pending payment update to return 200, got %d with body %q", pendingRec.Code, pendingRec.Body.String())
	}
}

func TestBuildEcommerceRouter_AdminAPIBulkUpdateObjectSpecificPaymentChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-payment-bulk-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin bulk object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	ctx := context.Background()
	res, err := database.ExecContext(ctx, `
		INSERT INTO customers (email, password_hash, first_name, last_name)
		VALUES (?, ?, ?, ?)
	`, "bulk-perm-test@example.com", "hashed", "Bulk", "Perm")
	if err != nil {
		t.Fatalf("failed to insert customer fixture: %v", err)
	}
	customerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve customer fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-BULK-PERM-001", customerID, "bulk-perm-test@example.com", "Bulk", "Perm", 200.00, 16.00, 10.00, 226.00)
	if err != nil {
		t.Fatalf("failed to insert order fixture: %v", err)
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve order fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO payments (order_id, amount, payment_method, status)
		VALUES (?, ?, ?, ?)
	`, orderID, 226.00, "card", "completed")
	if err != nil {
		t.Fatalf("failed to insert completed payment fixture: %v", err)
	}
	completedPaymentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve completed payment id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO payments (order_id, amount, payment_method, status)
		VALUES (?, ?, ?, ?)
	`, orderID, 226.00, "card", "pending")
	if err != nil {
		t.Fatalf("failed to insert pending payment fixture: %v", err)
	}
	pendingPaymentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve pending payment id: %v", err)
	}

	bulkPayload := fmt.Sprintf(`{"ids":[%d,%d],"data":{"gateway_response":"bulk note"}}`, completedPaymentID, pendingPaymentID)
	bulkReq := httptest.NewRequest(http.MethodPost, "/admin/api/payments/bulk-update", bytes.NewBufferString(bulkPayload))
	bulkReq.Header.Set("Content-Type", "application/json")
	bulkReq.Header.Set("Authorization", "Bearer "+token)
	bulkRec := httptest.NewRecorder()
	router.ServeHTTP(bulkRec, bulkReq)
	if bulkRec.Code != http.StatusMultiStatus {
		t.Fatalf("expected payment bulk update to return 207, got %d with body %q", bulkRec.Code, bulkRec.Body.String())
	}

	var bulkResp map[string]interface{}
	if err := json.Unmarshal(bulkRec.Body.Bytes(), &bulkResp); err != nil {
		t.Fatalf("failed to decode bulk update response: %v", err)
	}
	if gotUpdated, _ := bulkResp["updated"].(float64); int(gotUpdated) != 1 {
		t.Fatalf("expected 1 updated object, got %v (response=%v)", bulkResp["updated"], bulkResp)
	}
	errorsRaw, ok := bulkResp["errors"].([]interface{})
	if !ok || len(errorsRaw) != 1 {
		t.Fatalf("expected one bulk-update error entry, got %v", bulkResp["errors"])
	}
	errorEntry, ok := errorsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected object error entry, got %T", errorsRaw[0])
	}
	if gotCode, _ := errorEntry["code"].(string); gotCode != "permission_denied" {
		t.Fatalf("expected permission_denied bulk error code, got %q (response=%v)", gotCode, bulkResp)
	}

	pendingGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/payments/%d/", pendingPaymentID), nil)
	pendingGetReq.Header.Set("Authorization", "Bearer "+token)
	pendingGetRec := httptest.NewRecorder()
	router.ServeHTTP(pendingGetRec, pendingGetReq)
	if pendingGetRec.Code != http.StatusOK {
		t.Fatalf("expected pending payment detail to return 200, got %d with body %q", pendingGetRec.Code, pendingGetRec.Body.String())
	}
	var pendingPayload map[string]interface{}
	if err := json.Unmarshal(pendingGetRec.Body.Bytes(), &pendingPayload); err != nil {
		t.Fatalf("failed to decode pending payment detail response: %v", err)
	}
	if gotResponse, _ := pendingPayload["gateway_response"].(string); gotResponse != "bulk note" {
		t.Fatalf("expected pending payment gateway_response to be updated, got %q", gotResponse)
	}

	completedGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/payments/%d/", completedPaymentID), nil)
	completedGetReq.Header.Set("Authorization", "Bearer "+token)
	completedGetRec := httptest.NewRecorder()
	router.ServeHTTP(completedGetRec, completedGetReq)
	if completedGetRec.Code != http.StatusOK {
		t.Fatalf("expected completed payment detail to return 200, got %d with body %q", completedGetRec.Code, completedGetRec.Body.String())
	}
	var completedPayload map[string]interface{}
	if err := json.Unmarshal(completedGetRec.Body.Bytes(), &completedPayload); err != nil {
		t.Fatalf("failed to decode completed payment detail response: %v", err)
	}
	if gotResponse, _ := completedPayload["gateway_response"].(string); gotResponse == "bulk note" {
		t.Fatalf("expected completed payment gateway_response to remain unchanged, got %q", gotResponse)
	}
}

func TestBuildEcommerceRouter_AdminAPIBulkUpdateObjectSpecificOrderChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-order-bulk-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin bulk order object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	ctx := context.Background()
	res, err := database.ExecContext(ctx, `
		INSERT INTO customers (email, password_hash, first_name, last_name)
		VALUES (?, ?, ?, ?)
	`, "order-bulk-perm-test@example.com", "hashed", "Order", "Perm")
	if err != nil {
		t.Fatalf("failed to insert customer fixture: %v", err)
	}
	customerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve customer fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total, status, admin_notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-BULK-ORDER-PERM-DELIVERED", customerID, "order-bulk-perm-test@example.com", "Order", "Perm", 300.00, 24.00, 15.00, 339.00, "delivered", "finalized")
	if err != nil {
		t.Fatalf("failed to insert delivered order fixture: %v", err)
	}
	deliveredOrderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve delivered order id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total, status, admin_notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-BULK-ORDER-PERM-PROCESSING", customerID, "order-bulk-perm-test@example.com", "Order", "Perm", 180.00, 14.40, 9.00, 203.40, "processing", "mutable")
	if err != nil {
		t.Fatalf("failed to insert processing order fixture: %v", err)
	}
	processingOrderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve processing order id: %v", err)
	}

	bulkPayload := fmt.Sprintf(`{"ids":[%d,%d],"data":{"admin_notes":"bulk order note"}}`, deliveredOrderID, processingOrderID)
	bulkReq := httptest.NewRequest(http.MethodPost, "/admin/api/orders/bulk-update", bytes.NewBufferString(bulkPayload))
	bulkReq.Header.Set("Content-Type", "application/json")
	bulkReq.Header.Set("Authorization", "Bearer "+token)
	bulkRec := httptest.NewRecorder()
	router.ServeHTTP(bulkRec, bulkReq)
	if bulkRec.Code != http.StatusMultiStatus {
		t.Fatalf("expected order bulk update to return 207, got %d with body %q", bulkRec.Code, bulkRec.Body.String())
	}

	var bulkResp map[string]interface{}
	if err := json.Unmarshal(bulkRec.Body.Bytes(), &bulkResp); err != nil {
		t.Fatalf("failed to decode bulk update response: %v", err)
	}
	if gotUpdated, _ := bulkResp["updated"].(float64); int(gotUpdated) != 1 {
		t.Fatalf("expected 1 updated object, got %v (response=%v)", bulkResp["updated"], bulkResp)
	}
	errorsRaw, ok := bulkResp["errors"].([]interface{})
	if !ok || len(errorsRaw) != 1 {
		t.Fatalf("expected one bulk-update error entry, got %v", bulkResp["errors"])
	}
	errorEntry, ok := errorsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected object error entry, got %T", errorsRaw[0])
	}
	if gotCode, _ := errorEntry["code"].(string); gotCode != "permission_denied" {
		t.Fatalf("expected permission_denied bulk error code, got %q (response=%v)", gotCode, bulkResp)
	}

	processingGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/orders/%d/", processingOrderID), nil)
	processingGetReq.Header.Set("Authorization", "Bearer "+token)
	processingGetRec := httptest.NewRecorder()
	router.ServeHTTP(processingGetRec, processingGetReq)
	if processingGetRec.Code != http.StatusOK {
		t.Fatalf("expected processing order detail to return 200, got %d with body %q", processingGetRec.Code, processingGetRec.Body.String())
	}
	var processingPayload map[string]interface{}
	if err := json.Unmarshal(processingGetRec.Body.Bytes(), &processingPayload); err != nil {
		t.Fatalf("failed to decode processing order detail response: %v", err)
	}
	if gotNotes, _ := processingPayload["admin_notes"].(string); gotNotes != "bulk order note" {
		t.Fatalf("expected processing order admin_notes to be updated, got %q", gotNotes)
	}

	deliveredGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/orders/%d/", deliveredOrderID), nil)
	deliveredGetReq.Header.Set("Authorization", "Bearer "+token)
	deliveredGetRec := httptest.NewRecorder()
	router.ServeHTTP(deliveredGetRec, deliveredGetReq)
	if deliveredGetRec.Code != http.StatusOK {
		t.Fatalf("expected delivered order detail to return 200, got %d with body %q", deliveredGetRec.Code, deliveredGetRec.Body.String())
	}
	var deliveredPayload map[string]interface{}
	if err := json.Unmarshal(deliveredGetRec.Body.Bytes(), &deliveredPayload); err != nil {
		t.Fatalf("failed to decode delivered order detail response: %v", err)
	}
	if gotNotes, _ := deliveredPayload["admin_notes"].(string); gotNotes == "bulk order note" {
		t.Fatalf("expected delivered order admin_notes to remain unchanged, got %q", gotNotes)
	}
}

func TestBuildEcommerceRouter_AdminAPIActionObjectSpecificOrderChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-order-action-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin action object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	ctx := context.Background()
	res, err := database.ExecContext(ctx, `
		INSERT INTO customers (email, password_hash, first_name, last_name)
		VALUES (?, ?, ?, ?)
	`, "order-action-perm-test@example.com", "hashed", "Action", "Perm")
	if err != nil {
		t.Fatalf("failed to insert customer fixture: %v", err)
	}
	customerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve customer fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-ACTION-ORDER-PERM-DELIVERED", customerID, "order-action-perm-test@example.com", "Action", "Perm", 120.00, 9.60, 6.00, 135.60, "delivered")
	if err != nil {
		t.Fatalf("failed to insert delivered order fixture: %v", err)
	}
	deliveredOrderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve delivered order id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-ACTION-ORDER-PERM-PROCESSING", customerID, "order-action-perm-test@example.com", "Action", "Perm", 95.00, 7.60, 4.75, 107.35, "processing")
	if err != nil {
		t.Fatalf("failed to insert processing order fixture: %v", err)
	}
	processingOrderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve processing order id: %v", err)
	}

	actionPayload := fmt.Sprintf(`{"ids":[%d,%d]}`, deliveredOrderID, processingOrderID)
	actionReq := httptest.NewRequest(http.MethodPost, "/admin/api/orders/action/mark_shipped", bytes.NewBufferString(actionPayload))
	actionReq.Header.Set("Content-Type", "application/json")
	actionReq.Header.Set("Authorization", "Bearer "+token)
	actionRec := httptest.NewRecorder()
	router.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusMultiStatus {
		t.Fatalf("expected order action endpoint to return 207, got %d with body %q", actionRec.Code, actionRec.Body.String())
	}

	var actionResp map[string]interface{}
	if err := json.Unmarshal(actionRec.Body.Bytes(), &actionResp); err != nil {
		t.Fatalf("failed to decode action response: %v", err)
	}
	if gotAffected, _ := actionResp["affected"].(float64); int(gotAffected) != 1 {
		t.Fatalf("expected 1 affected object, got %v (response=%v)", actionResp["affected"], actionResp)
	}
	errorsRaw, ok := actionResp["errors"].([]interface{})
	if !ok || len(errorsRaw) != 1 {
		t.Fatalf("expected one action error entry, got %v", actionResp["errors"])
	}

	processingGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/orders/%d/", processingOrderID), nil)
	processingGetReq.Header.Set("Authorization", "Bearer "+token)
	processingGetRec := httptest.NewRecorder()
	router.ServeHTTP(processingGetRec, processingGetReq)
	if processingGetRec.Code != http.StatusOK {
		t.Fatalf("expected processing order detail to return 200, got %d with body %q", processingGetRec.Code, processingGetRec.Body.String())
	}
	var processingPayload map[string]interface{}
	if err := json.Unmarshal(processingGetRec.Body.Bytes(), &processingPayload); err != nil {
		t.Fatalf("failed to decode processing order detail response: %v", err)
	}
	if gotStatus, _ := processingPayload["status"].(string); gotStatus != "shipped" {
		t.Fatalf("expected processing order status to be updated by action, got %q", gotStatus)
	}

	deliveredGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/orders/%d/", deliveredOrderID), nil)
	deliveredGetReq.Header.Set("Authorization", "Bearer "+token)
	deliveredGetRec := httptest.NewRecorder()
	router.ServeHTTP(deliveredGetRec, deliveredGetReq)
	if deliveredGetRec.Code != http.StatusOK {
		t.Fatalf("expected delivered order detail to return 200, got %d with body %q", deliveredGetRec.Code, deliveredGetRec.Body.String())
	}
	var deliveredPayload map[string]interface{}
	if err := json.Unmarshal(deliveredGetRec.Body.Bytes(), &deliveredPayload); err != nil {
		t.Fatalf("failed to decode delivered order detail response: %v", err)
	}
	if gotStatus, _ := deliveredPayload["status"].(string); gotStatus != "delivered" {
		t.Fatalf("expected delivered order status to remain unchanged, got %q", gotStatus)
	}
}

func TestBuildEcommerceRouter_AdminAPIActionObjectSpecificPaymentChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-payment-action-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin payment action object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	ctx := context.Background()
	res, err := database.ExecContext(ctx, `
		INSERT INTO customers (email, password_hash, first_name, last_name)
		VALUES (?, ?, ?, ?)
	`, "payment-action-perm-test@example.com", "hashed", "Payment", "Perm")
	if err != nil {
		t.Fatalf("failed to insert customer fixture: %v", err)
	}
	customerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve customer fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO orders (
			order_number, customer_id, customer_email, customer_first_name, customer_last_name,
			subtotal, tax_amount, shipping_amount, total
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ORD-ACTION-PAYMENT-PERM-001", customerID, "payment-action-perm-test@example.com", "Payment", "Perm", 140.00, 11.20, 7.00, 158.20)
	if err != nil {
		t.Fatalf("failed to insert order fixture: %v", err)
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve order fixture id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO payments (order_id, amount, payment_method, status)
		VALUES (?, ?, ?, ?)
	`, orderID, 158.20, "card", "completed")
	if err != nil {
		t.Fatalf("failed to insert completed payment fixture: %v", err)
	}
	completedPaymentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve completed payment id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO payments (order_id, amount, payment_method, status)
		VALUES (?, ?, ?, ?)
	`, orderID, 158.20, "card", "pending")
	if err != nil {
		t.Fatalf("failed to insert pending payment fixture: %v", err)
	}
	pendingPaymentID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve pending payment id: %v", err)
	}

	actionPayload := fmt.Sprintf(`{"ids":[%d,%d]}`, completedPaymentID, pendingPaymentID)
	actionReq := httptest.NewRequest(http.MethodPost, "/admin/api/payments/action/mark_failed", bytes.NewBufferString(actionPayload))
	actionReq.Header.Set("Content-Type", "application/json")
	actionReq.Header.Set("Authorization", "Bearer "+token)
	actionRec := httptest.NewRecorder()
	router.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusMultiStatus {
		t.Fatalf("expected payment action endpoint to return 207, got %d with body %q", actionRec.Code, actionRec.Body.String())
	}

	var actionResp map[string]interface{}
	if err := json.Unmarshal(actionRec.Body.Bytes(), &actionResp); err != nil {
		t.Fatalf("failed to decode action response: %v", err)
	}
	if gotAffected, _ := actionResp["affected"].(float64); int(gotAffected) != 1 {
		t.Fatalf("expected 1 affected object, got %v (response=%v)", actionResp["affected"], actionResp)
	}
	errorsRaw, ok := actionResp["errors"].([]interface{})
	if !ok || len(errorsRaw) != 1 {
		t.Fatalf("expected one action error entry, got %v", actionResp["errors"])
	}
	errorEntry, ok := errorsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected object error entry, got %T", errorsRaw[0])
	}
	if gotCode, _ := errorEntry["code"].(string); gotCode != "permission_denied" {
		t.Fatalf("expected permission_denied action error code, got %q (response=%v)", gotCode, actionResp)
	}

	pendingGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/payments/%d/", pendingPaymentID), nil)
	pendingGetReq.Header.Set("Authorization", "Bearer "+token)
	pendingGetRec := httptest.NewRecorder()
	router.ServeHTTP(pendingGetRec, pendingGetReq)
	if pendingGetRec.Code != http.StatusOK {
		t.Fatalf("expected pending payment detail to return 200, got %d with body %q", pendingGetRec.Code, pendingGetRec.Body.String())
	}
	var pendingPayload map[string]interface{}
	if err := json.Unmarshal(pendingGetRec.Body.Bytes(), &pendingPayload); err != nil {
		t.Fatalf("failed to decode pending payment detail response: %v", err)
	}
	if gotStatus, _ := pendingPayload["status"].(string); gotStatus != "failed" {
		t.Fatalf("expected pending payment status to be updated by action, got %q", gotStatus)
	}

	completedGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/payments/%d/", completedPaymentID), nil)
	completedGetReq.Header.Set("Authorization", "Bearer "+token)
	completedGetRec := httptest.NewRecorder()
	router.ServeHTTP(completedGetRec, completedGetReq)
	if completedGetRec.Code != http.StatusOK {
		t.Fatalf("expected completed payment detail to return 200, got %d with body %q", completedGetRec.Code, completedGetRec.Body.String())
	}
	var completedPayload map[string]interface{}
	if err := json.Unmarshal(completedGetRec.Body.Bytes(), &completedPayload); err != nil {
		t.Fatalf("failed to decode completed payment detail response: %v", err)
	}
	if gotStatus, _ := completedPayload["status"].(string); gotStatus != "completed" {
		t.Fatalf("expected completed payment status to remain unchanged, got %q", gotStatus)
	}
}

func TestBuildEcommerceRouter_AdminAPIActionObjectSpecificWarehouseChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-warehouse-action-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin warehouse action object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	ctx := context.Background()
	res, err := database.ExecContext(ctx, `
		INSERT INTO warehouses (
			name, code, address_line1, city, postal_code, country_code, country_name, is_active, is_primary
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "Primary Warehouse", "WH-PRIMARY-PERM", "10 Main St", "Boston", "02108", "US", "United States", true, true)
	if err != nil {
		t.Fatalf("failed to insert primary warehouse fixture: %v", err)
	}
	primaryWarehouseID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve primary warehouse id: %v", err)
	}

	res, err = database.ExecContext(ctx, `
		INSERT INTO warehouses (
			name, code, address_line1, city, postal_code, country_code, country_name, is_active, is_primary
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "Secondary Warehouse", "WH-SECONDARY-PERM", "22 Side St", "Cambridge", "02139", "US", "United States", true, false)
	if err != nil {
		t.Fatalf("failed to insert secondary warehouse fixture: %v", err)
	}
	secondaryWarehouseID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to resolve secondary warehouse id: %v", err)
	}

	actionPayload := fmt.Sprintf(`{"ids":[%d,%d]}`, primaryWarehouseID, secondaryWarehouseID)
	actionReq := httptest.NewRequest(http.MethodPost, "/admin/api/warehouses/action/deactivate", bytes.NewBufferString(actionPayload))
	actionReq.Header.Set("Content-Type", "application/json")
	actionReq.Header.Set("Authorization", "Bearer "+token)
	actionRec := httptest.NewRecorder()
	router.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusMultiStatus {
		t.Fatalf("expected warehouse action endpoint to return 207, got %d with body %q", actionRec.Code, actionRec.Body.String())
	}

	var actionResp map[string]interface{}
	if err := json.Unmarshal(actionRec.Body.Bytes(), &actionResp); err != nil {
		t.Fatalf("failed to decode warehouse action response: %v", err)
	}
	if gotAffected, _ := actionResp["affected"].(float64); int(gotAffected) != 1 {
		t.Fatalf("expected 1 affected object, got %v (response=%v)", actionResp["affected"], actionResp)
	}
	errorsRaw, ok := actionResp["errors"].([]interface{})
	if !ok || len(errorsRaw) != 1 {
		t.Fatalf("expected one action error entry, got %v", actionResp["errors"])
	}
	errorEntry, ok := errorsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected object error entry, got %T", errorsRaw[0])
	}
	if gotCode, _ := errorEntry["code"].(string); gotCode != "permission_denied" {
		t.Fatalf("expected permission_denied action error code, got %q (response=%v)", gotCode, actionResp)
	}

	secondaryGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/warehouses/%d/", secondaryWarehouseID), nil)
	secondaryGetReq.Header.Set("Authorization", "Bearer "+token)
	secondaryGetRec := httptest.NewRecorder()
	router.ServeHTTP(secondaryGetRec, secondaryGetReq)
	if secondaryGetRec.Code != http.StatusOK {
		t.Fatalf("expected secondary warehouse detail to return 200, got %d with body %q", secondaryGetRec.Code, secondaryGetRec.Body.String())
	}
	var secondaryPayload map[string]interface{}
	if err := json.Unmarshal(secondaryGetRec.Body.Bytes(), &secondaryPayload); err != nil {
		t.Fatalf("failed to decode secondary warehouse detail response: %v", err)
	}
	if gotActive, _ := secondaryPayload["is_active"].(bool); gotActive {
		t.Fatalf("expected secondary warehouse to be deactivated, got is_active=%v", secondaryPayload["is_active"])
	}

	primaryGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/warehouses/%d/", primaryWarehouseID), nil)
	primaryGetReq.Header.Set("Authorization", "Bearer "+token)
	primaryGetRec := httptest.NewRecorder()
	router.ServeHTTP(primaryGetRec, primaryGetReq)
	if primaryGetRec.Code != http.StatusOK {
		t.Fatalf("expected primary warehouse detail to return 200, got %d with body %q", primaryGetRec.Code, primaryGetRec.Body.String())
	}
	var primaryPayload map[string]interface{}
	if err := json.Unmarshal(primaryGetRec.Body.Bytes(), &primaryPayload); err != nil {
		t.Fatalf("failed to decode primary warehouse detail response: %v", err)
	}
	if gotActive, _ := primaryPayload["is_active"].(bool); !gotActive {
		t.Fatalf("expected primary warehouse to remain active, got is_active=%v", primaryPayload["is_active"])
	}
}

func TestBuildEcommerceRouter_AdminAPIActionObjectSpecificCouponChangePermission(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ecommerce-admin-coupon-action-object-permissions-test.sqlite")
	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite3")
	cfg.Set("database.sqlite_path", dbPath)
	cfg.Set("admin.path", "/admin")
	cfg.Set("api.path", "/api/v1")
	cfg.Set("api.enabled", true)
	cfg.Set("cors.allowed_origins", []string{"http://localhost"})
	cfg.Set("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	cfg.Set("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	cfg.Set("cors.exposed_headers", []string{"X-Total-Count"})

	database, err := db.NewDB(dbPath)
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, "go-sqlite3 requires cgo") || strings.Contains(errText, "gcc") {
			t.Skipf("skipping admin coupon action object permission test: sqlite driver unavailable in this environment (%v)", err)
		}
		t.Fatalf("failed to create sqlite db: %v", err)
	}
	defer database.Close()

	admin.DefaultSite = admin.NewSite("default")
	router := buildEcommerceRouter(context.Background(), cfg, database)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected admin login status 200, got %d with body %q", loginRec.Code, loginRec.Body.String())
	}

	var loginPayload map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	token, ok := loginPayload["token"].(string)
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected login token in response, got %v", loginPayload["token"])
	}

	createCoupon := func(code string) int64 {
		payload := fmt.Sprintf(`{
			"code":"%s",
			"name":"%s",
			"discount_type":"percentage",
			"discount_value":10,
			"valid_from":"2026-01-01T00:00:00Z",
			"valid_until":"2027-01-01T00:00:00Z",
			"is_active":true
		}`, code, code)
		req := httptest.NewRequest(http.MethodPost, "/admin/api/coupons/", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected coupon create status 201, got %d with body %q", rec.Code, rec.Body.String())
		}

		var created map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to decode coupon create response: %v", err)
		}
		rawID, ok := created["id"].(float64)
		if !ok {
			t.Fatalf("expected created coupon id in response, got %v", created["id"])
		}
		return int64(rawID)
	}

	usedCouponID := createCoupon("USED-ACTION-PERM")
	unusedCouponID := createCoupon("UNUSED-ACTION-PERM")

	if _, err := database.ExecContext(context.Background(), `UPDATE coupons SET usage_count = ? WHERE id = ?`, 3, usedCouponID); err != nil {
		t.Fatalf("failed to update used coupon usage_count fixture: %v", err)
	}

	actionPayload := fmt.Sprintf(`{"ids":[%d,%d]}`, usedCouponID, unusedCouponID)
	actionReq := httptest.NewRequest(http.MethodPost, "/admin/api/coupons/action/deactivate", bytes.NewBufferString(actionPayload))
	actionReq.Header.Set("Content-Type", "application/json")
	actionReq.Header.Set("Authorization", "Bearer "+token)
	actionRec := httptest.NewRecorder()
	router.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusMultiStatus {
		t.Fatalf("expected coupon action endpoint to return 207, got %d with body %q", actionRec.Code, actionRec.Body.String())
	}

	var actionResp map[string]interface{}
	if err := json.Unmarshal(actionRec.Body.Bytes(), &actionResp); err != nil {
		t.Fatalf("failed to decode coupon action response: %v", err)
	}
	if gotAffected, _ := actionResp["affected"].(float64); int(gotAffected) != 1 {
		t.Fatalf("expected 1 affected object, got %v (response=%v)", actionResp["affected"], actionResp)
	}
	errorsRaw, ok := actionResp["errors"].([]interface{})
	if !ok || len(errorsRaw) != 1 {
		t.Fatalf("expected one action error entry, got %v", actionResp["errors"])
	}
	errorEntry, ok := errorsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected object error entry, got %T", errorsRaw[0])
	}
	if gotCode, _ := errorEntry["code"].(string); gotCode != "permission_denied" {
		t.Fatalf("expected permission_denied action error code, got %q (response=%v)", gotCode, actionResp)
	}

	unusedGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/coupons/%d/", unusedCouponID), nil)
	unusedGetReq.Header.Set("Authorization", "Bearer "+token)
	unusedGetRec := httptest.NewRecorder()
	router.ServeHTTP(unusedGetRec, unusedGetReq)
	if unusedGetRec.Code != http.StatusOK {
		t.Fatalf("expected unused coupon detail to return 200, got %d with body %q", unusedGetRec.Code, unusedGetRec.Body.String())
	}
	var unusedPayload map[string]interface{}
	if err := json.Unmarshal(unusedGetRec.Body.Bytes(), &unusedPayload); err != nil {
		t.Fatalf("failed to decode unused coupon detail response: %v", err)
	}
	if gotActive, _ := unusedPayload["is_active"].(bool); gotActive {
		t.Fatalf("expected unused coupon to be deactivated, got is_active=%v", unusedPayload["is_active"])
	}

	usedGetReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/api/coupons/%d/", usedCouponID), nil)
	usedGetReq.Header.Set("Authorization", "Bearer "+token)
	usedGetRec := httptest.NewRecorder()
	router.ServeHTTP(usedGetRec, usedGetReq)
	if usedGetRec.Code != http.StatusOK {
		t.Fatalf("expected used coupon detail to return 200, got %d with body %q", usedGetRec.Code, usedGetRec.Body.String())
	}
	var usedPayload map[string]interface{}
	if err := json.Unmarshal(usedGetRec.Body.Bytes(), &usedPayload); err != nil {
		t.Fatalf("failed to decode used coupon detail response: %v", err)
	}
	if gotActive, _ := usedPayload["is_active"].(bool); !gotActive {
		t.Fatalf("expected used coupon to remain active, got is_active=%v", usedPayload["is_active"])
	}
}

func performJSONRequest(
	t *testing.T,
	router *server.Router,
	method string,
	path string,
	payload map[string]interface{},
	expectedStatus int,
) map[string]interface{} {
	t.Helper()

	var bodyReader *bytes.Reader
	if payload == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d for %s %s, got %d with body %q", expectedStatus, method, path, rr.Code, rr.Body.String())
	}

	if rr.Body.Len() == 0 {
		return map[string]interface{}{}
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON response for %s %s: %v (body=%q)", method, path, err, rr.Body.String())
	}
	return decoded
}

type testErr string

func (e testErr) Error() string { return string(e) }

func newTestErr(format string, args ...any) error {
	return testErr(fmt.Sprintf(format, args...))
}
