// @title Alumni Management System API
// @version 1.0
// @description API untuk sistem manajemen alumni dengan fitur RBAC dan achievement management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"crud-app/app/utils"
	"crud-app/database"
	"crud-app/route"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	_ "crud-app/docs" // Import generated docs

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if os.Getenv("DB_DSN") == "" {
		log.Fatal("Set environment variable DB_DSN")
	}

	database.ConnectDB()
	defer database.DB.Close()

	mongoClient := database.MongoConnection()
	defer database.CloseDB(mongoClient)
	mongoDB := database.GetMongoDatabase()

	utils.InitCache()
	log.Println("Permission cache initialized")

	app := fiber.New()

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	route.Routes(app, database.DB, mongoDB)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server running on http://localhost:%s", port)
	log.Printf("Swagger docs: http://localhost:%s/swagger/", port)

	log.Fatal(app.Listen(":" + port))
}