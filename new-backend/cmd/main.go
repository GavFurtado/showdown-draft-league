package main

import (
	"log"
	"os"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/app"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	routes "github.com/GavFurtado/showdown-draft-league/new-backend/internal/router"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load application configuration (compile time)
	cfg := config.LoadConfig()

	// CORS config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.AppBaseURL}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}

	// discord OauthConfig
	discordOauthConfig := &oauth2.Config{
		ClientID:     cfg.DiscordClientID,
		ClientSecret: cfg.DiscordClientSecret,
		RedirectURL:  cfg.DiscordRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
		Scopes: []string{"identify"},
	}

	log.SetFlags(0) // no date/time.

	// Connect to PostgreSQL database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database models
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

	// Initialize Repositories, Servies and Controllers
	appRepositories := app.NewRepositories(db)
	appServices := app.NewServices(appRepositories, cfg, discordOauthConfig)
	appControllers := app.NewControllers(appServices, cfg, discordOauthConfig)

	// Set Port and Initialize Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := gin.New()

	// --- Global Middlewares ---
	server.Use(gin.Recovery(), gin.Logger())
	server.Use(cors.New(corsConfig))

	routes.RegisterRoutes(server, db, cfg, appRepositories, appServices, appControllers)

	// Run server
	log.Printf("Server starting...\n")
	log.Printf("Server running on port: %s\n", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
