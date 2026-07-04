// Showdown Draft League API
//
//	@title			Showdown Draft League API
//	@version		1.0
//	@description	Pokemon draft league management API
//	@host			localhost:8080
//	@BasePath		/
package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/app"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	routes "github.com/GavFurtado/showdown-draft-league/new-backend/internal/router"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Swagger metadata annotations
//
//	@title			Showdown Draft League API
//	@version		1.0
//	@description	API for managing Pokemon draft leagues
//	@BasePath		/
func main() {
	// Load application configuration
	cfg := config.LoadConfig()

	// CORS config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.APP_BASE_URL}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour

	// Discord OauthConfig
	discordOauthConfig := &oauth2.Config{
		ClientID:     cfg.DISCORD_CLIENT_ID,
		ClientSecret: cfg.DISCORD_CLIENT_SECRET,
		RedirectURL:  cfg.DISCORD_REDIRECT_URI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
		Scopes: []string{"identify"},
	}

	log.SetFlags(0) // no date/time.

	// Register generic isValid custom validator on Gin's Validator engine.
	// Used for enum validation. String-based enums are uppercased first
	// so that "public" matches "PUBLIC" case-insensitively.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("isValid", func(fl validator.FieldLevel) bool {
			field := fl.Field()
			if field.Kind() == reflect.String {
				upper := reflect.New(field.Type()).Elem()
				upper.SetString(strings.ToUpper(field.String()))
				if v, ok := upper.Interface().(interface{ IsValid() bool }); ok {
					return v.IsValid()
				}
			}
			if v, ok := field.Interface().(interface{ IsValid() bool }); ok {
				return v.IsValid()
			}
			return false
		})
	}

	// Connect to PostgreSQL database
	db, err := gorm.Open(postgres.Open(cfg.DATABASE_URL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// TODO: dev only. In fact, even as far as dev only this is not ideal
	// Switch to migration files.
	err = db.AutoMigrate(
		&models.User{},
		&models.League{},
		&models.PokemonSpecies{},
		&models.Draft{},
		&models.Game{},
		&models.LeagueMember{},
		&models.PoolEntry{},
		&models.DraftPick{},
		&models.Claim{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Initialize Repositories, Servies and Controllers
	appRepositories := app.NewRepositories(db)
	appServices := app.NewServices(appRepositories, cfg, discordOauthConfig)
	appControllers := app.NewControllers(appServices, appRepositories, cfg, discordOauthConfig)

	// Start the scheduler
	if err := appServices.SchedulerService.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Start the server
	port := cfg.PORT
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
