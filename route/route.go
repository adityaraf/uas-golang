package route

import (
	"crud-app/app/service"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App, db *sql.DB) {
	// Initialize services
	authService := service.NewAuthService(db)

	// API routes
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
}