package app

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/controllers"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	u "github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepository:           repositories.NewUserRepository(db),
		LeagueRepository:         repositories.NewLeagueRepository(db),
		PlayerRepository:         repositories.NewPlayerRepository(db),
		DraftRepository:          repositories.NewDraftRepository(db),
		DraftedPokemonRepository: repositories.NewDraftedPokemonRepository(db),
		GameRepository:           repositories.NewGameRepository(db),
		LeaguePokemonRepository:  repositories.NewLeaguePokemonRepository(db),
		PokemonSpeciesRepository: repositories.NewPokemonSpeciesRepository(db),
	}
}

func NewServices(repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Services {
	// Instantiate early for core Dependencies
	jwtService := services.NewJWTService(cfg.JWTSecret)
	rbacService := services.NewRBACService(repos.LeagueRepository, repos.UserRepository, repos.PlayerRepository)
	// webhooks not implemented; this is temp and does nothing
	webhookService := services.NewWebhookService()

	// Create services without circular dependencies initially
	draftService := services.NewDraftService(
		repos.LeagueRepository,
		repos.LeaguePokemonRepository,
		repos.DraftRepository,
		repos.DraftedPokemonRepository,
		repos.PlayerRepository,
		&webhookService,
	)

	schedulerService := services.NewSchedulerService(
		&u.TaskHeap{}, // Corrected: Initialize TaskHeap directly
		repos.LeagueRepository,
		repos.DraftRepository,
	)

	// Inject circular dependencies using setter methods
	draftService.SetSchedulerService(schedulerService)
	schedulerService.SetDraftService(draftService.(services.DraftService))

	return &Services{
		JWTService:           *jwtService,
		UserService:          services.NewUserService(repos.UserRepository),
		RBACService:          rbacService,
		WebhookService:       webhookService,
		LeaguePokemonService: services.NewLeaguePokemonService(repos.LeaguePokemonRepository, repos.LeagueRepository, repos.UserRepository, repos.PokemonSpeciesRepository),
		LeagueService:        services.NewLeagueService(repos.LeagueRepository, repos.PlayerRepository, repos.LeaguePokemonRepository, repos.DraftedPokemonRepository, repos.DraftRepository, repos.GameRepository),
		PlayerService:        services.NewPlayerService(repos.PlayerRepository, repos.LeagueRepository, repos.UserRepository),
		AuthService:          services.NewAuthService(repos.UserRepository, jwtService, discordOauthConfig),
		DraftService:         draftService,
		DraftedPokemonService: services.NewDraftedPokemonService(
			repos.DraftedPokemonRepository,
			repos.UserRepository,
			repos.LeagueRepository,
			repos.PlayerRepository,
			repos.PokemonSpeciesRepository,
			repos.LeaguePokemonRepository,
		),
		PokemonSpeciesService: services.NewPokemonSpeciesService(repos.PokemonSpeciesRepository),
		SchedulerService:      schedulerService,
		// GameService
	}
}

func NewControllers(services *Services, repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Controllers {
	return &Controllers{
		AuthController:           *controllers.NewAuthController(services.AuthService, cfg, discordOauthConfig),
		LeagueController:         controllers.NewLeagueController(services.LeagueService),
		PlayerController:         controllers.NewPlayerController(services.PlayerService),
		UserController:           controllers.NewUserController(services.UserService),
		PokemonSpeciesController: controllers.NewPokemonSpeciesController(services.PokemonSpeciesService),
		LeaguePokemonController:  controllers.NewLeaguePokemonSpeciesController(services.LeaguePokemonService),
		DraftedPokemonController: controllers.NewDraftedPokemonController(services.DraftedPokemonService),
		DraftController:          controllers.NewDraftController(services.DraftService),
	}
}
