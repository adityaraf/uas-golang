package route

import (
"crud-app/app/middleware"
"crud-app/app/service"
"database/sql"

"github.com/gofiber/fiber/v2"
"go.mongodb.org/mongo-driver/mongo"
)

func Routes(app *fiber.App, db *sql.DB, mongoDB *mongo.Database) {
// Initialize services
authService := service.NewAuthService(db)
achievementService := service.NewAchievementService(mongoDB, db)

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

	// Achievement routes (FR-003: Submit Prestasi, FR-004: Submit untuk Verifikasi, FR-006: View Prestasi Mahasiswa Bimbingan)
	// Achievement routes (FR-003: Submit Prestasi)
achievements := api.Group("/achievements")
achievements.Use(middleware.AuthRequired()) // Require authentication

	// Student endpoints
achievements.Post("/", rbac.RequirePermission("achievements.create"), achievementService.SubmitAchievement)
achievements.Get("/", rbac.RequirePermission("achievements.read"), achievementService.GetMyAchievements)
	achievements.Get("/advisees", rbac.RequirePermission("achievements.verify"), achievementService.GetAdviseeAchievements)
achievements.Get("/:id", rbac.RequirePermission("achievements.read"), achievementService.GetAchievementByID)
achievements.Put("/:id", rbac.RequirePermission("achievements.update"), achievementService.UpdateAchievement)
achievements.Delete("/:id", rbac.RequirePermission("achievements.delete"), achievementService.DeleteAchievement)
	achievements.Post("/:id/submit", rbac.RequirePermission("achievements.submit"), achievementService.SubmitForVerification)
	achievements.Post("/:id/submit", rbac.RequirePermission("achievements.create"), achievementService.SubmitForVerification)

	// Lecturer/Admin endpoints (FR-007: Verify Prestasi)
	achievements.Get("/pending", rbac.RequirePermission("achievements.verify"), achievementService.GetPendingVerification)
	achievements.Get("/:id/review", rbac.RequirePermission("achievements.verify"), achievementService.ReviewAchievementDetail)
	achievements.Post("/:id/approve", rbac.RequirePermission("achievements.verify"), achievementService.ApproveAchievement)
	achievements.Post("/:id/reject", rbac.RequirePermission("achievements.verify"), achievementService.RejectAchievement)

// Protected routes example (uncomment untuk digunakan)
// users := api.Group("/users")