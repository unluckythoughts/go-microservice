// Package api provides HTTP route registration and request handling for the example application.
package api

import (
	"example/service"
	"os"

	"github.com/unluckythoughts/go-microservice/tools/web"
)

func RegisterRoutes(router web.Router, s *service.Service) {
	// Get JWT secret
	jwtSecret := os.Getenv("SESSION_SECRET")
	if jwtSecret == "" {
		panic("JWT session secret is not provided!")
	}

	// Create handlers
	h := NewHandlers(s, jwtSecret)

	// Authentication routes
	router.POST("/api/v1/login", h.Login)
	router.POST("/api/v1/signup", h.Signup)
	router.POST("/api/v1/logout", h.Authorized, h.Logout)

	// User routes
	router.GET("/api/v1/user", h.Authorized, h.GetUser)
	router.PUT("/api/v1/user", h.Authorized, h.UpdateUser)
	router.DELETE("/api/v1/user", h.Authorized, h.DeleteUser)

	// User settings routes
	router.GET("/api/v1/user/settings", h.Authorized, h.GetUserSettings)
	router.PUT("/api/v1/user/settings", h.Authorized, h.UpdateUserSettings)

	// Health check
	router.GET("/api/v1/health", func(r web.Request) (any, error) {
		return map[string]interface{}{
			"status":  "ok",
			"message": "Example API is running",
		}, nil
	})
}
