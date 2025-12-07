package route

import (
	"crud-app/app/middleware"
	"crud-app/app/service"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App, db *sql.DB) {
	// Initialize services
	authService := service.NewAuthService(db)

	// Initialize RBAC middleware
	rbac := middleware.NewRBACMiddleware(db)

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)

	// Test endpoint untuk RBAC
	api.Get("/test-permission",
		middleware.AuthRequired(),
		rbac.RequirePermission("users.read"),
		func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status":  "success",
				"message": "Anda memiliki permission users.read",
			})
		})

	// Protected routes example (uncomment untuk digunakan)
	// users := api.Group("/users")
	// users.Use(middleware.AuthRequired()) // Require authentication
	// users.Get("/", rbac.RequirePermission("users.read"), getUsersHandler)
	// users.Post("/", rbac.RequirePermission("users.create"), createUserHandler)
	// users.Put("/:id", rbac.RequirePermission("users.update"), updateUserHandler)
	// users.Delete("/:id", rbac.RequirePermission("users.delete"), deleteUserHandler)

	// Example: Multiple permissions
	// admin := api.Group("/admin")
	// admin.Use(middleware.AuthRequired())
	// admin.Get("/dashboard", rbac.RequireAnyPermission("admin.dashboard", "superadmin.access"), dashboardHandler)
	// admin.Post("/settings", rbac.RequireAllPermissions("admin.settings.read", "admin.settings.write"), settingsHandler)
}