package main

import (
	"log"
	"os"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	routes "github.com/GavFurtado/showdown-draft-league/new-backend/internal/router"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load application configuration
	cfg := config.LoadConfig()

	// 2. Connect to PostgreSQL database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Auto-migrate database models
	// Ensure the order respects foreign key dependencies.
	err = db.AutoMigrate(
		&models.User{},
		&models.League{},
		&models.Player{},
		&models.PokemonSpecies{},
		&models.LeaguePokemon{},
		&models.DraftedPokemon{},
		&models.Game{},
		&models.PlayerRoster{},
	)

	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := gin.New()

	// Middlewares
	server.Use(gin.Recovery(), gin.Logger())

	routes.RegisterRoutes(server, db, cfg)

	// Run server
	log.Printf("Server starting on :%s", port) // Use :port for Gin.Run
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
