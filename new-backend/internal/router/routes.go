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
	playerRepo := repositories.NewPlayerRepository(db)
	leagueRepo := repositories.NewLeagueRepository(db)
	// pokemonSpeciesRepo := repositories.NewPokemonSpeciesRepository(db)
	leaguePokemonRepo := repositories.NewLeaguePokemonRepository(db)
	draftedPokemonRepo := repositories.NewDraftedPokemonRepository(db)
	// draftRepo := repositories.NewDraftRepository(db)
	gameRepo := repositories.NewGameRepository(db)

	//  --- Initialize Services ---
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

	jwtService := services.NewJWTService(cfg.JWTSecret)
	authService := services.NewAuthService(userRepo, jwtService, discordOauthConfig)
	userService := services.NewUserService(userRepo)
	playerService := services.NewPlayerService(playerRepo, leagueRepo, userRepo)
	leagueService := services.NewLeagueService(leagueRepo, playerRepo, leaguePokemonRepo, draftedPokemonRepo, gameRepo)
	// webhookService := services.NewWebhookService()
	// draftService := services.NewDraftService(leagueRepo, leaguePokemonRepo, draftRepo, draftedPokemonRepo, playerRepo, &webhookService)
	// draftedPokemonService := services.NewDraftedPokemonService(draftedPokemonRepo, userRepo, leagueRepo, playerRepo)

	//  --- Initialize Controller  ---
	authController := controllers.NewAuthController(authService, cfg, discordOauthConfig)
	userController := controllers.NewUserController(userService)
	leagueController := controllers.NewLeagueController(leagueService)
	playerController := controllers.NewPlayerController(playerService)

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
		api.GET("/profile", userController.GetMyProfile)

		leagues := api.Group("/leagues")
		{
			leagues.POST("/", leagueController.CreateLeague)
			leagues.GET("/:id", leagueController.GetLeague)
			leagues.GET("/:id/players", playerController.GetPlayersByLeague)
			leagues.POST("/:id/join", playerController.JoinLeague)
			// not implmented yet
			// leagues.DELETE("/:id/leave", playerController.LeaveLeague)
		}

		users := api.Group("/users")
		{
			users.GET("/me", userController.GetMyProfile)
			users.GET("/me/discord", userController.GetMyDiscordDetails)
			users.GET("/me/leagues", userController.GetMyLeagues)
			users.PUT("/profile", userController.UpdateProfile)
			users.GET("/:id/players", playerController.GetPlayersByLeague)
		}

		players := api.Group("/players")
		{
			players.GET("/:id", playerController.GetPlayerByID)
			players.GET("/:id/roster", playerController.GetPlayerWithFullRoster)
			players.PUT("/:id/profile", playerController.UpdatePlayerProfile)
		}
	}

}

// this is temporary
func HomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}
