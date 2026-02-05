package identity

import (
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/service"
)

// IdentitySystem holds all identity system components
type IdentitySystem struct {
	// Repositories
	UserRepo       repository.UserRepository
	SessionRepo    repository.SessionRepository
	TokenRepo      repository.TokenRepository
	PermissionRepo repository.PermissionRepository
	GroupRepo      repository.GroupRepository

	// Services
	UserService       service.UserService
	AuthService       service.AuthService
	PasswordService   service.PasswordService
	PermissionService service.PermissionService

	// Backends
	BackendRegistry backends.BackendRegistry

	// Config
	Config *config.IdentityConfig
}

// SetupIdentitySystem creates and configures the entire identity system
func SetupIdentitySystem(database *db.DB, userConfig *config.IdentityConfig) (*IdentitySystem, error) {
	// Use default config if not provided
	if userConfig == nil {
		userConfig = config.DefaultIdentityConfig()
	}

	// Create repositories
	userRepo := repository.NewUserRepository(database)
	sessionRepo := repository.NewSessionRepository(database)
	tokenRepo := repository.NewTokenRepository(database)
	permissionRepo := repository.NewPermissionRepository(database)
	groupRepo := repository.NewGroupRepository(database)

	// Create and register backends
	backendRegistry := backends.NewBackendRegistry()
	passwordBackend := backends.NewPasswordBackend(userRepo)
	tokenBackend := backends.NewTokenBackend(userRepo, sessionRepo)
	backendRegistry.Register(passwordBackend)
	backendRegistry.Register(tokenBackend)

	// Create services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo, backendRegistry)
	passwordService := service.NewPasswordService(userRepo, tokenRepo, userConfig.PasswordPolicy)
	permissionService := service.NewPermissionService(permissionRepo, userRepo)

	return &IdentitySystem{
		UserRepo:          userRepo,
		SessionRepo:       sessionRepo,
		TokenRepo:         tokenRepo,
		PermissionRepo:    permissionRepo,
		GroupRepo:         groupRepo,
		UserService:       userService,
		AuthService:       authService,
		PasswordService:   passwordService,
		PermissionService: permissionService,
		BackendRegistry:   backendRegistry,
		Config:            userConfig,
	}, nil
}
