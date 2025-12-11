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
	userService := service.NewUserService(db)

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

// Achievement routes (FR-003: Submit Prestasi)
achievements := api.Group("/achievements")
achievements.Use(middleware.AuthRequired()) // Require authentication

// Student endpoints
achievements.Post("/", rbac.RequirePermission("achievements.create"), achievementService.SubmitAchievement)
achievements.Get("/", rbac.RequirePermission("achievements.read"), achievementService.GetMyAchievements)
achievements.Get("/:id", rbac.RequirePermission("achievements.read"), achievementService.GetAchievementByID)
achievements.Put("/:id", rbac.RequirePermission("achievements.update"), achievementService.UpdateAchievement)
achievements.Delete("/:id", rbac.RequirePermission("achievements.delete"), achievementService.DeleteAchievement)
achievements.Post("/:id/submit", rbac.RequirePermission("achievements.create"), achievementService.SubmitForVerification)

// Lecturer/Admin endpoints (FR-007: Verify Prestasi)
achievements.Get("/pending", rbac.RequirePermission("achievements.verify"), achievementService.GetPendingVerification)
achievements.Get("/:id/review", rbac.RequirePermission("achievements.verify"), achievementService.ReviewAchievementDetail)
achievements.Post("/:id/approve", rbac.RequirePermission("achievements.verify"), achievementService.ApproveAchievement)
achievements.Post("/:id/reject", rbac.RequirePermission("achievements.verify"), achievementService.RejectAchievement)

	// User Management routes (FR-009: Manage Users)
	users := api.Group("/users")
	users.Use(middleware.AuthRequired()) // Require authentication
	users.Get("/", rbac.RequirePermission("users.read"), userService.GetUsers)
	users.Get("/:id", rbac.RequirePermission("users.read"), userService.GetUserByID)
	users.Post("/", rbac.RequirePermission("users.create"), userService.CreateUser)
	users.Put("/:id", rbac.RequirePermission("users.update"), userService.UpdateUser)
	users.Delete("/:id", rbac.RequirePermission("users.delete"), userService.DeleteUser)
	users.Post("/:id/role", rbac.RequirePermission("users.assign_role"), userService.AssignRole)
	users.Post("/:id/student", rbac.RequirePermission("students.create"), userService.SetStudentProfile)
	users.Post("/:id/lecturer", rbac.RequirePermission("lecturers.create"), userService.SetLecturerProfile)

	// Student Management routes
	students := api.Group("/students")
	students.Use(middleware.AuthRequired())
	students.Put("/:id", rbac.RequirePermission("students.update"), userService.UpdateStudentProfile)
	students.Post("/:id/advisor", rbac.RequirePermission("students.assign_advisor"), userService.AssignAdvisor)

	// Lecturer Management routes
	lecturers := api.Group("/lecturers")
	lecturers.Use(middleware.AuthRequired())
	lecturers.Put("/:id", rbac.RequirePermission("lecturers.update"), userService.UpdateLecturerProfile)

	// Admin routes (FR-010: View All Achievements)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired())
	admin.Get("/achievements", rbac.RequirePermission("achievements.read"), achievementService.GetAllAchievements)

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