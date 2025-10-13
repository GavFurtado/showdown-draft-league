package main

import (
	"fmt"
	"log"
	"time"

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
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour

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
		&models.Draft{},
		&models.Game{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Initialize Repositories, Servies and Controllers
	appRepositories := app.NewRepositories(db)
	appServices := app.NewServices(appRepositories, cfg, discordOauthConfig)
	appControllers := app.NewControllers(appServices, cfg, discordOauthConfig)

	// Start the scheduler
	if err := appServices.SchedulerService.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Start the server
	port := cfg.Port
	if port == "" {
		port = "8080" // Default port
	}
	server := gin.New()

	// --- Global Middlewares ---
	server.Use(gin.Recovery(), gin.Logger())
	server.Use(cors.New(corsConfig))

	routes.RegisterRoutes(server, db, cfg, appRepositories, appServices, appControllers)

	// Run server
	fmt.Printf("Server started...\n")
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
