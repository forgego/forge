package identity

import (
	"net/http"

	forgehttp "github.com/forgego/forge/server"
	"github.com/forgego/forge/identity/handlers"
	"github.com/forgego/forge/identity/middleware"
)

// RegisterRoutes registers all user system routes
func (us *IdentitySystem) RegisterRoutes(router *forgehttp.Router) {
	// Create handlers
	userHandler := handlers.NewUserHandler(us.UserService)
	authHandler := handlers.NewAuthHandler(us.AuthService, us.UserService, us.BackendRegistry)
	authMiddleware := middleware.NewAuthenticationMiddleware(us.BackendRegistry, us.SessionRepo, us.UserRepo)

	// Auth routes (public)
	router.Route("/api/auth", func(sub *forgehttp.Router) {
		sub.Post("/register", authHandler.Register)
		sub.Post("/login", authHandler.Login)
		// Wrap handler with middleware
		sub.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Logout)).ServeHTTP(w, r)
		})
		sub.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me)).ServeHTTP(w, r)
		})
	})

	// User routes (protected)
	router.Route("/api/users", func(sub *forgehttp.Router) {
		// Apply authentication middleware to all user routes
		sub.Use(authMiddleware.RequireAuth)

		sub.Get("/", userHandler.List)
		sub.Post("/", func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.RequireSuperuser(http.HandlerFunc(userHandler.Create)).ServeHTTP(w, r)
		})
		sub.Get("/{id}", userHandler.Retrieve)
		sub.Put("/{id}", userHandler.Update)
		sub.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.RequireSuperuser(http.HandlerFunc(userHandler.Delete)).ServeHTTP(w, r)
		})
	})
}
