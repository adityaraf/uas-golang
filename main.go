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
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	_ "crud-app/docs" // Import generated docs

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func main() {
	// 🔹 Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if os.Getenv("DB_DSN") == "" {
		log.Fatal("Set environment variable DB_DSN")
	}

	// Connect to PostgreSQL
	database.ConnectDB()
	defer database.DB.Close()

	// Connect to MongoDB
	mongoClient := database.MongoConnection()
	defer database.CloseDB(mongoClient)
	mongoDB := database.GetMongoDatabase()

	// 🔹 Initialize permission cache
	utils.InitCache()
	log.Println("Permission cache initialized")

	app := fiber.New()

	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Contoh hash password
	password := "123456"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash:", string(hash))

	// Register routes
	route.Routes(app, database.DB, mongoDB)

	// 🔹 Pakai APP_PORT dari .env
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Printf("📚 Swagger docs: http://localhost:%s/swagger/", port)

	log.Fatal(app.Listen(":" + port))
}