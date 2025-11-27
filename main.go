package main

import (
	"crud-app/database"
	"crud-app/route"
	"fmt"
	"log"
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	database.ConnectDB()
	defer database.DB.Close()

	app := fiber.New()

	// Contoh hash password
	password := "123456"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash:", string(hash))

	// Register routes
	route.RegisterRoutes(app, database.DB)

	// 🔹 Pakai APP_PORT dari .env
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
