package route

import (
    "database/sql"

    "github.com/gofiber/fiber/v2"
    "crud-app/app/service"
    "crud-app/app/middleware"
)

func RegisterRoutes(app *fiber.App, db *sql.DB) {
    api := app.Group("/api")

    // === Auth ===
    api.Post("/login", service.Login)
    api.Get("/profile", middleware.AuthRequired(), service.GetProfile)

    // === Users ===
    users := api.Group("/users", middleware.AuthRequired(), middleware.AdminOnly())
    users.Get("/", service.GetAllUsers)
    users.Get("/:id", service.GetUserByID)
    users.Delete("/soft/:id", service.SoftDeleteUser)
    users.Delete("/permanent/:id", service.DeleteUser)

    // === Alumni ===
    alumni := api.Group("/alumni", middleware.AuthRequired())
    alumni.Get("/", service.GetAllAlumni)
    alumni.Get("/:id", service.GetAlumniByID)
    alumni.Get("/pekerjaan/:id", service.GetAlumniWithMultipleJobsService(db))
    alumni.Post("/", middleware.AdminOnly(), service.CreateAlumni)
    alumni.Put("/:id", middleware.AdminOnly(), service.UpdateAlumni)
    alumni.Delete("/:id", middleware.AdminOnly(), service.DeleteAlumni)
    alumni.Delete("/soft/:id", middleware.AdminOnly(), service.SoftDeleteAlumni)

    // === Pekerjaan ===
    pekerjaan := api.Group("/pekerjaan", middleware.AuthRequired())
    pekerjaan.Get("/", func(c *fiber.Ctx) error { return service.GetAllPekerjaanService(c, db) })
    pekerjaan.Get("/trash", func(c *fiber.Ctx) error { return service.GetTrashedPekerjaanService(c, db) })
    pekerjaan.Get("/:id", func(c *fiber.Ctx) error { return service.GetPekerjaanByIDService(c, db) })
    pekerjaan.Get("/alumni/:alumni_id", func(c *fiber.Ctx) error { return service.GetPekerjaanByAlumniIDService(c, db) })
    pekerjaan.Get("/user/my", func(c *fiber.Ctx) error { return service.GetUserPekerjaanService(c, db) })
    pekerjaan.Put("/:id", middleware.AdminOnly(), func(c *fiber.Ctx) error { return service.UpdatePekerjaanService(c, db) })
    pekerjaan.Delete("/:id", middleware.AdminOnly(), func(c *fiber.Ctx) error { return service.DeletePekerjaanService(c, db) })
    pekerjaan.Delete("/soft/:id", func(c *fiber.Ctx) error { return service.SoftDeletePekerjaanService(c, db) })
    pekerjaan.Post("/trash/:id/restore", func(c *fiber.Ctx) error { return service.RestorePekerjaanService(c, db) })
    pekerjaan.Delete("/trash/:id", func(c *fiber.Ctx) error { return service.HardDeletePekerjaanService(c, db) })
}
