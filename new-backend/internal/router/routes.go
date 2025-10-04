package routes

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/app"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
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
	authMiddlewareDeps := middleware.AuthMiddlewareDependencies{
		UserRepo:    repositories.UserRepository,
		JWTService:  &services.JWTService,
		RBACService: services.RBACService,
	}
	leagueMiddlewareDeps := middleware.LeagueRBACDependencies{
		UserRepo:    repositories.UserRepository,
		RBACService: services.RBACService,
	}

	// ---- Public Routes ---
	// These do not require any authorization
	r.GET("/", HomeHandler) // eventually a landing page

	// Pokemon Species routes
	pokemonSpecies := r.Group("/api/pokemon_species")
	{
		pokemonSpecies.GET("/", controllers.PokemonSpeciesController.GetAllPokemonSpecies)
		pokemonSpecies.GET("/:id", controllers.PokemonSpeciesController.GetPokemonSpeciesByID)
		pokemonSpecies.GET("/name/:name", controllers.PokemonSpeciesController.GetPokemonSpeciesByName)
	}

	// ---- Auth Related Routes ---
	// These are routes related to Discord OAuth
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/discord/login", controllers.AuthController.Login)
		authGroup.GET("/discord/callback", controllers.AuthController.DiscCallback)
		authGroup.GET("/logout", controllers.AuthController.Logout)
	}

	// --- Protected Routes ---
	// These require authorization
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authMiddlewareDeps)) // top level logged in check
	{
		api.GET("/profile", controllers.UserController.GetMyProfile)
		leagues := api.Group("/leagues")
		{
			leagues.POST(
				"/",
				controllers.LeagueController.CreateLeague)
			leagues.GET(
				"/:leagueId",
				middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionReadLeague),
				controllers.LeagueController.GetLeague)
			leagues.GET(
				"/:leagueId/players",
				middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionReadPlayer),
				controllers.PlayerController.GetPlayersByLeague)
			leagues.POST("/:leagueId/join", controllers.PlayerController.JoinLeague)

			// not implmented yet
			// leagues.DELETE("/:id/leave", playerController.LeaveLeague)

			players := leagues.Group(":leagueId/players")
			{
				players.GET(
					"/:id",
					middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionReadPlayer),
					controllers.PlayerController.GetPlayerByID)
				players.GET(
					"/:id/roster",
					middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionReadPlayerRoster),
					controllers.PlayerController.GetPlayerWithFullRoster)
				players.PUT(
					"/:id/profile",
					middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionUpdatePlayer),
					controllers.PlayerController.UpdatePlayerProfile)
			}

			leaguePokemon := leagues.Group("/:leagueId/pokemon")
			{
				leaguePokemon.POST(
					"/single",
					middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionCreateLeaguePokemon),
					controllers.LeaguePokemonController.CreatePokemonForLeague)
				leaguePokemon.POST(
					"/batch",
					middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionCreateLeaguePokemon),
					controllers.LeaguePokemonController.BatchCreatePokemonForLeague)
				leaguePokemon.PUT(
					"/",
					middleware.LeagueRBACMiddleware(leagueMiddlewareDeps, rbac.PermissionUpdateLeaguePokemon),
					controllers.LeaguePokemonController.UpdateLeaguePokemon)
			}
		}

		users := api.Group("/users")
		{
			users.GET("/me", controllers.UserController.GetMyProfile) // same as /api/profile
			users.GET("/me/discord", controllers.UserController.GetMyDiscordDetails)
			users.GET("/me/leagues", controllers.UserController.GetMyLeagues)
			users.PUT("/profile", controllers.UserController.UpdateProfile)
			users.GET("/:id/players", controllers.PlayerController.GetPlayersByUser)
		}

	}
}

// this is temporary
func HomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}
