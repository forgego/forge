package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/forgego/forge/identity/utils"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthServiceTest(t *testing.T) (AuthService, *db.DB, context.Context) {
	testDB := setupTestDB(t)
	userRepo := repository.NewUserRepository(testDB)
	sessionRepo := repository.NewSessionRepository(testDB)

	// Create a simple backend registry mock
	backendRegistry := &mockBackendRegistry{
		userRepo: userRepo,
	}

	service := NewAuthService(userRepo, sessionRepo, backendRegistry)
	ctx := context.Background()
	return service, testDB, ctx
}

type mockBackendRegistry struct {
	userRepo repository.UserRepository
	backends []backends.AuthenticationBackend
	byName   map[string]backends.AuthenticationBackend
}

func (m *mockBackendRegistry) Register(backend backends.AuthenticationBackend) {
	if m.byName == nil {
		m.byName = make(map[string]backends.AuthenticationBackend)
	}
	name := backend.Name()
	if _, exists := m.byName[name]; !exists {
		m.backends = append(m.backends, backend)
		m.byName[name] = backend
	}
}

func (m *mockBackendRegistry) Get(name string) (backends.AuthenticationBackend, error) {
	if m.byName == nil {
		return nil, fmt.Errorf("backend %s not found", name)
	}
	backend, ok := m.byName[name]
	if !ok {
		return nil, fmt.Errorf("backend %s not found", name)
	}
	return backend, nil
}

func (m *mockBackendRegistry) GetByCredentialType(credentialType string) []backends.AuthenticationBackend {
	var supported []backends.AuthenticationBackend
	for _, backend := range m.backends {
		if backend.Supports(credentialType) {
			supported = append(supported, backend)
		}
	}
	return supported
}

func (m *mockBackendRegistry) Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error) {
	username := credentials["username"]
	if username == "" {
		username = credentials["email"]
	}

	user, err := m.userRepo.GetByUsername(ctx, username)
	if err != nil {
		user, err = m.userRepo.GetByEmail(ctx, username)
		if err != nil {
			return nil, err
		}
	}

	password := credentials["password"]
	if !utils.CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func (m *mockBackendRegistry) GetUser(ctx context.Context, identifier string) (*models.User, error) {
	return nil, nil
}

func TestAuthService_Authenticate(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	// Create test user
	hashedPassword, _ := utils.HashPassword("correctpassword")
	userRepo := repository.NewUserRepository(testDB)
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("authenticates with correct credentials", func(t *testing.T) {
		req := &AuthenticateRequest{
			UsernameOrEmail: "testuser",
			Password:        "correctpassword",
		}

		authenticated, err := service.Authenticate(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, user.ID, authenticated.ID)
		assert.NotNil(t, authenticated.LastLogin)
	})

	t.Run("authenticates with email", func(t *testing.T) {
		req := &AuthenticateRequest{
			UsernameOrEmail: "test@example.com",
			Password:        "correctpassword",
		}

		authenticated, err := service.Authenticate(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, user.ID, authenticated.ID)
	})

	t.Run("fails with incorrect password", func(t *testing.T) {
		req := &AuthenticateRequest{
			UsernameOrEmail: "testuser",
			Password:        "wrongpassword",
		}

		_, err := service.Authenticate(ctx, req)
		assert.Error(t, err)
	})

	t.Run("fails with inactive user", func(t *testing.T) {
		// Deactivate user
		user.IsActive = false
		userRepo.Update(ctx, user)

		req := &AuthenticateRequest{
			UsernameOrEmail: "testuser",
			Password:        "correctpassword",
		}

		_, err := service.Authenticate(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, ErrUserInactive, err)
	})

	t.Run("fails with locked user", func(t *testing.T) {
		// Reactivate and lock user
		user.IsActive = true
		user.IsLocked = true
		userRepo.Update(ctx, user)

		req := &AuthenticateRequest{
			UsernameOrEmail: "testuser",
			Password:        "correctpassword",
		}

		_, err := service.Authenticate(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, ErrUserLocked, err)
	})
}

func TestAuthService_CreateSession(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	// Create test user
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("password123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("creates session successfully", func(t *testing.T) {
		req := &CreateSessionRequest{
			UserID:     user.ID,
			IPAddress:  "127.0.0.1",
			UserAgent:  "test-agent",
			RememberMe: false,
		}

		session, err := service.CreateSession(ctx, user.ID, req)
		require.NoError(t, err)
		assert.NotEmpty(t, session.SessionKey)
		assert.Equal(t, user.ID, session.UserID)
		assert.Equal(t, "127.0.0.1", session.IPAddress)
		assert.False(t, session.IsRememberMe)
	})

	t.Run("creates remember me session", func(t *testing.T) {
		req := &CreateSessionRequest{
			UserID:     user.ID,
			IPAddress:  "127.0.0.1",
			UserAgent:  "test-agent",
			RememberMe: true,
		}

		session, err := service.CreateSession(ctx, user.ID, req)
		require.NoError(t, err)
		assert.True(t, session.IsRememberMe)
		assert.NotNil(t, session.ExpiresAt)
		// Should expire in ~30 days
		assert.True(t, session.ExpiresAt.After(time.Now().Add(29*24*time.Hour)))
	})
}

func TestAuthService_GetSession(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	// Create test user and session
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("password123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	req := &CreateSessionRequest{
		UserID:    user.ID,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}
	session, err := service.CreateSession(ctx, user.ID, req)
	require.NoError(t, err)

	t.Run("retrieves valid session", func(t *testing.T) {
		retrieved, err := service.GetSession(ctx, session.SessionKey)
		require.NoError(t, err)
		assert.Equal(t, session.SessionKey, retrieved.SessionKey)
		assert.Equal(t, user.ID, retrieved.UserID)
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		_, err := service.GetSession(ctx, "nonexistent-key")
		assert.Error(t, err)
	})
}

func TestAuthService_Logout(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	// Create test user and session
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("password123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	req := &CreateSessionRequest{
		UserID:    user.ID,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}
	session, err := service.CreateSession(ctx, user.ID, req)
	require.NoError(t, err)

	t.Run("logs out successfully", func(t *testing.T) {
		err := service.Logout(ctx, session.SessionKey)
		require.NoError(t, err)

		// Session should no longer exist
		_, err = service.GetSession(ctx, session.SessionKey)
		assert.Error(t, err)
	})
}

func TestAuthService_LogoutAll(t *testing.T) {
	service, testDB, ctx := setupAuthServiceTest(t)
	defer testDB.Close()

	// Create test user
	userRepo := repository.NewUserRepository(testDB)
	hashedPassword, _ := utils.HashPassword("password123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		req := &CreateSessionRequest{
			UserID:    user.ID,
			IPAddress: "127.0.0.1",
			UserAgent: "test-agent",
		}
		_, err := service.CreateSession(ctx, user.ID, req)
		require.NoError(t, err)
	}

	t.Run("logs out all sessions", func(t *testing.T) {
		err := service.LogoutAll(ctx, user.ID)
		require.NoError(t, err)

		// All sessions should be deleted
		sessions, err := service.ListSessions(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, sessions, 0)
	})
}
