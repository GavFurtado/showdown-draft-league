package app

import (
	"log/slog"

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
		DraftRepository:          repositories.NewDraftRepository(db),
		GameRepository:           repositories.NewGameRepository(db),
		PokemonSpeciesRepository: repositories.NewPokemonSpeciesRepository(db),

		DraftPickRepository: repositories.NewDraftPickRepository(db),
		ClaimRepository:     repositories.NewClaimRepository(db),
		PoolEntryRepository: repositories.NewPoolEntryRepository(db),
		PlayerRepository:    repositories.NewPlayerRepository(db),
		SpectatorRepository: repositories.NewSpectatorRepository(db),
	}
}

func NewServices(logger *slog.Logger, repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Services {
	jwtService := services.NewJWTService(cfg.JWT_SECRET)
	rbacService := services.NewRBACService(logger, repos.LeagueRepository, repos.UserRepository, repos.PlayerRepository, repos.SpectatorRepository)
	webhookService := services.NewWebhookService(logger)

	draftService := services.NewDraftService(
		logger,
		repos.LeagueRepository,
		repos.DraftRepository,
		repos.PlayerRepository,
		&webhookService,
	)

	draftService.SetNewRepositories(
		repos.DraftPickRepository,
		repos.ClaimRepository,
		repos.PoolEntryRepository,
	)

	schedulerService := services.NewSchedulerService(
		logger,
		&u.TaskHeap{},
		repos.LeagueRepository,
		repos.DraftRepository,
	)

	transferService := services.NewTransferService(
		logger,
		repos.LeagueRepository,
		repos.PlayerRepository,
	)

	transferService.SetNewRepositories(
		repos.ClaimRepository,
		repos.PoolEntryRepository,
		repos.PlayerRepository,
	)

	gameService := services.NewGameService(logger, repos.GameRepository, repos.LeagueRepository, repos.PlayerRepository)

	leagueService := services.NewLeagueService(logger, repos.LeagueRepository, repos.PlayerRepository, repos.DraftRepository, repos.GameRepository)

	draftService.SetSchedulerService(schedulerService)
	schedulerService.SetDraftService(draftService.(services.DraftService))

	transferService.SetSchedulerService(schedulerService)
	schedulerService.SetTransferService(transferService.(services.TransferService))

	schedulerService.SetLeagueService(leagueService)
	leagueService.SetSchedulerService(schedulerService)

	gameService.SetLeagueService(leagueService)
	leagueService.SetGameService(gameService)

	leagueService.SetTransferService(transferService)

	return &Services{
		Logger:                logger,
		JWTService:            *jwtService,
		UserService:           services.NewUserService(logger, repos.UserRepository),
		RBACService:           rbacService,
		WebhookService:        webhookService,
		LeagueService:         leagueService,
		AuthService:           services.NewAuthService(logger, repos.UserRepository, jwtService, discordOauthConfig),
		DevAuthService:        services.NewDevAuthService(logger, repos.UserRepository, repos.LeagueRepository, repos.PlayerRepository, repos.SpectatorRepository, jwtService),
		DraftService:          draftService,
		PokemonSpeciesService: services.NewPokemonSpeciesService(logger, repos.PokemonSpeciesRepository),
		SchedulerService:      schedulerService,
		GameService:           gameService,
		TransferService:       transferService,

		PoolEntryService: services.NewPoolEntryService(logger, repos.PoolEntryRepository, repos.LeagueRepository, repos.UserRepository, repos.PokemonSpeciesRepository),
		PlayerService:    services.NewPlayerService(logger, repos.PlayerRepository, repos.LeagueRepository, repos.UserRepository, repos.SpectatorRepository),
		SpectatorService: services.NewSpectatorService(logger, repos.SpectatorRepository, repos.PlayerRepository, repos.LeagueRepository),
		DraftPickService: services.NewDraftPickService(logger, repos.DraftPickRepository, repos.DraftRepository),
		ClaimService:     services.NewClaimService(logger, repos.ClaimRepository),
	}
}

func NewControllers(logger *slog.Logger, services *Services, repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Controllers {
	return &Controllers{
		AuthController:           controllers.NewAuthController(logger, services.AuthService, cfg, discordOauthConfig),
		DevAuthController:        controllers.NewDevAuthController(logger, services.DevAuthService),
		LeagueController:         controllers.NewLeagueController(logger, services.LeagueService),
		UserController:           controllers.NewUserController(logger, services.UserService),
		PokemonSpeciesController: controllers.NewPokemonSpeciesController(logger, services.PokemonSpeciesService),
		DraftController:          controllers.NewDraftController(logger, services.DraftService),
		GameController:           controllers.NewGameController(logger, services.GameService, services.LeagueService),
		TransferController:       controllers.NewTransferController(logger, services.TransferService),

		PoolEntryController: controllers.NewPoolEntryController(logger, services.PoolEntryService),
		PlayerController:    controllers.NewPlayerController(logger, services.PlayerService),
		SpectatorController: controllers.NewSpectatorController(logger, services.SpectatorService),
		DraftPickController: controllers.NewDraftPickController(logger, services.DraftPickService, services.DraftService),
		ClaimController:     controllers.NewClaimController(logger, services.ClaimService),
	}
}
