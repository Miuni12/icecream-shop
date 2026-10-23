package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/miuni12/icecream-shop/internal/database"
	"github.com/miuni12/icecream-shop/internal/routes"
)

func main() {
	_ = godotenv.Load() // .env is optional if variables are set in the environment

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set (copy .env.example to .env)")
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "shop.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	db, err := database.Connect(dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	app := routes.NewApp(db, secret)
	log.Fatal(app.Listen(":" + port))
}