package routes

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/controllers"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// --- Initialize Repositories ---
	userRepo := repositories.NewUserRepository(db)
	// playerRepo := repositories.NewPlayerRepository(db)

	//  --- Initialize Services ---
	jwtService := services.NewJWTService(cfg.JWTSecret)

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

	authService := services.NewAuthService(userRepo, jwtService, discordOauthConfig)

	//  --- Initialize Controller  ---
	authController := controllers.NewAuthController(authService, cfg, discordOauthConfig)

	println(authController)
	// ---- Public Routes ---
	// These do not require any authorization
	r.GET("/", HomeHandler)

	// ---- Auth Related Routes ---
	// These are routes related to Discord OAuth
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/discord/login", authController.Login)
		authGroup.GET("/discord/callback", authController.DiscCallback)
	}

	// --- Protected Routes ---
	// These require authorization
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtService, userRepo))
	{
		// api.GET("/leagues/:id")
		// api.GET("/leagues/:id")
		// api.GET("/leagues/:id")
		// api.GET("/profile", userController.GetUserProfile)
	}
	api.Use(middleware.AuthMiddleware(jwtService, userRepo))
	{
		// api.GET("/profile", userController.GetUserProfile)
	}
}

// this is temporary
func HomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}
