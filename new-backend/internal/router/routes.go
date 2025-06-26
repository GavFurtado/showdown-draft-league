package routes

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/controllers"
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

}

func HomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}
