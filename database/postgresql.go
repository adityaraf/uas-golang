package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	dsn := os.Getenv("DB_DSN") // contoh: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}
	DB = db
}
