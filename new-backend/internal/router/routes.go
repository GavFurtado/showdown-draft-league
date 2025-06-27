package routes

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/controllers"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Initialize Repositories
	userRepo := repositories.NewUserRepository(db)

	// Initialize Services
	jwtService := services.NewJWTService(cfg.JWTSecret)

	// Initialize Controller
	authController := controllers.NewAuthController(userRepo, jwtService, cfg)

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
		// api.GET("/profile", userController.GetUserProfile)
	}
}

// this is temporary
func HomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}
