package routes

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/app"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	// "github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(
	r *gin.Engine,
	db *gorm.DB,
	cfg *config.Config,
	repositories *app.Repositories,
	services *app.Services,
	controllers *app.Controllers,
) {
	// // ---- Public Routes ---
	// // These do not require any authorization
	// r.GET("/", HomeHandler)
	//
	// // ---- Auth Related Routes ---
	// // These are routes related to Discord OAuth
	// authGroup := r.Group("/auth")
	// {
	// 	authGroup.GET("/discord/login", authController.Login)
	// 	authGroup.GET("/discord/callback", authController.DiscCallback)
	// }
	//
	// // --- Protected Routes ---
	// // These require authorization
	// api := r.Group("/api")
	// api.Use(middleware.AuthMiddleware(jwtService, userRepo))
	// api.Use(middleware.LeagueRBACMiddleware(rbacService, leagueService, leagueRepo, rbacService))
	// {
	// 	api.GET("/profile", userController.GetMyProfile)
	//
	// 	leagues := api.Group("/leagues")
	// 	{
	// 		leagues.POST("/", leagueController.CreateLeague)
	// 		leagues.GET("/:id", leagueController.GetLeague)
	// 		leagues.GET("/:id/players", playerController.GetPlayersByLeague)
	// 		leagues.POST("/:id/join", playerController.JoinLeague)
	// 		// not implmented yet
	// 		// leagues.DELETE("/:id/leave", playerController.LeaveLeague)
	// 	}
	//
	// 	users := api.Group("/users")
	// 	{
	// 		users.GET("/me", userController.GetMyProfile)
	// 		users.GET("/me/discord", userController.GetMyDiscordDetails)
	// 		users.GET("/me/leagues", userController.GetMyLeagues)
	// 		users.PUT("/profile", userController.UpdateProfile)
	// 		users.GET("/:id/players", playerController.GetPlayersByUser)
	// 	}
	//
	// 	players := api.Group("/players")
	// 	{
	// 		players.GET("/:id", playerController.GetPlayerByID)
	// 		players.GET("/:id/roster", playerController.GetPlayerWithFullRoster)
	// 		players.PUT("/:id/profile", playerController.UpdatePlayerProfile)
	// 	}
	// }
	//
}

// this is temporary
func HomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}
