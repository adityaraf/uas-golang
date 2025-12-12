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

	// ============================================
	// 5.1 Authentication Routes (Public)
	// ============================================
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken)
	auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
	auth.Get("/profile", middleware.AuthRequired(), authService.GetProfile)

	// ============================================
	// 5.2 Users Routes (Admin)
	// ============================================
	users := api.Group("/users")
	users.Use(middleware.AuthRequired())
	users.Get("/", rbac.RequirePermission("users.read"), userService.GetUsers)
	users.Get("/:id", rbac.RequirePermission("users.read"), userService.GetUserByID)
	users.Post("/", rbac.RequirePermission("users.create"), userService.CreateUser)
	users.Put("/:id", rbac.RequirePermission("users.update"), userService.UpdateUser)
	users.Delete("/:id", rbac.RequirePermission("users.delete"), userService.DeleteUser)
	users.Put("/:id/role", rbac.RequirePermission("users.assign_role"), userService.AssignRole)

	// ============================================
	// 5.4 Achievements Routes
	// ============================================
	achievements := api.Group("/achievements")
	achievements.Use(middleware.AuthRequired())

	// List & Detail
	achievements.Get("/", rbac.RequirePermission("achievements.read"), achievementService.GetMyAchievements)
	achievements.Get("/:id", rbac.RequirePermission("achievements.read"), achievementService.GetAchievementByID)

	// CRUD Operations (Mahasiswa)
	achievements.Post("/", rbac.RequirePermission("achievements.create"), achievementService.SubmitAchievement)
	achievements.Put("/:id", rbac.RequirePermission("achievements.update"), achievementService.UpdateAchievement)
	achievements.Delete("/:id", rbac.RequirePermission("achievements.delete"), achievementService.DeleteAchievement)

	// Workflow Operations
	achievements.Post("/:id/submit", rbac.RequirePermission("achievements.create"), achievementService.SubmitForVerification)
	achievements.Post("/:id/verify", rbac.RequirePermission("achievements.verify"), achievementService.ApproveAchievement)
	achievements.Post("/:id/reject", rbac.RequirePermission("achievements.verify"), achievementService.RejectAchievement)

	// History & Attachments
	achievements.Get("/:id/history", rbac.RequirePermission("achievements.read"), achievementService.GetAchievementHistory)
	achievements.Post("/:id/attachments", rbac.RequirePermission("achievements.create"), achievementService.UploadAttachment)

	// ============================================
	// 5.5 Students & Lecturers Routes
	// ============================================
	students := api.Group("/students")
	students.Use(middleware.AuthRequired())
	students.Get("/", rbac.RequirePermission("students.read"), userService.GetStudents)
	students.Get("/:id", rbac.RequirePermission("students.read"), userService.GetStudentByID)
	students.Get("/:id/achievements", rbac.RequirePermission("achievements.read"), achievementService.GetStudentAchievements)
	students.Put("/:id/advisor", rbac.RequirePermission("students.assign_advisor"), userService.AssignAdvisor)

	lecturers := api.Group("/lecturers")
	lecturers.Use(middleware.AuthRequired())
	lecturers.Get("/", rbac.RequirePermission("lecturers.read"), userService.GetLecturers)
	lecturers.Get("/:id/advisees", rbac.RequirePermission("lecturers.read"), userService.GetAdvisees)

	// ============================================
	// 5.8 Reports & Analytics Routes
	// ============================================
	reports := api.Group("/reports")
	reports.Use(middleware.AuthRequired())
	reports.Get("/statistics", rbac.RequirePermission("achievements.read"), achievementService.GetAllStatistics)
	reports.Get("/student/:id", rbac.RequirePermission("achievements.read"), achievementService.GetStudentReport)
}