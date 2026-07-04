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

		DraftPickRepository:    repositories.NewDraftPickRepository(db),
		ClaimRepository:        repositories.NewClaimRepository(db),
		PoolEntryRepository:    repositories.NewPoolEntryRepository(db),
		LeagueMemberRepository: repositories.NewLeagueMemberRepository(db),
	}
}

func NewServices(logger *slog.Logger, repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Services {
	jwtService := services.NewJWTService(cfg.JWT_SECRET)
	rbacService := services.NewRBACService(logger, repos.LeagueRepository, repos.UserRepository, repos.LeagueMemberRepository)
	webhookService := services.NewWebhookService(logger)

	draftService := services.NewDraftService(
		logger,
		repos.LeagueRepository,
		repos.DraftRepository,
		repos.LeagueMemberRepository,
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
		repos.LeagueMemberRepository,
	)

	transferService.SetNewRepositories(
		repos.ClaimRepository,
		repos.PoolEntryRepository,
		repos.LeagueMemberRepository,
	)

	gameService := services.NewGameService(logger, repos.GameRepository, repos.LeagueRepository, repos.LeagueMemberRepository)

	leagueService := services.NewLeagueService(logger, repos.LeagueRepository, repos.LeagueMemberRepository, repos.DraftRepository, repos.GameRepository)

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
		DraftService:          draftService,
		PokemonSpeciesService: services.NewPokemonSpeciesService(logger, repos.PokemonSpeciesRepository),
		SchedulerService:      schedulerService,
		GameService:           gameService,
		TransferService:       transferService,

		PoolEntryService:    services.NewPoolEntryService(logger, repos.PoolEntryRepository, repos.LeagueRepository, repos.UserRepository, repos.PokemonSpeciesRepository),
		LeagueMemberService: services.NewLeagueMemberService(logger, repos.LeagueMemberRepository, repos.LeagueRepository, repos.UserRepository),
		DraftPickService:    services.NewDraftPickService(logger, repos.DraftPickRepository, repos.DraftRepository),
		ClaimService:        services.NewClaimService(logger, repos.ClaimRepository),
	}
}

func NewControllers(logger *slog.Logger, services *Services, repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Controllers {
	return &Controllers{
		AuthController:           controllers.NewAuthController(logger, services.AuthService, cfg, discordOauthConfig),
		LeagueController:         controllers.NewLeagueController(logger, services.LeagueService),
		UserController:           controllers.NewUserController(logger, services.UserService),
		PokemonSpeciesController: controllers.NewPokemonSpeciesController(logger, services.PokemonSpeciesService),
		DraftController:          controllers.NewDraftController(logger, services.DraftService),
		GameController:           controllers.NewGameController(logger, services.GameService, services.LeagueService),
		TransferController:       controllers.NewTransferController(logger, services.TransferService),

		PoolEntryController:    controllers.NewPoolEntryController(logger, services.PoolEntryService),
		LeagueMemberController: controllers.NewLeagueMemberController(logger, services.LeagueMemberService),
		DraftPickController:    controllers.NewDraftPickController(logger, services.DraftPickService, services.DraftService),
		ClaimController:        controllers.NewClaimController(logger, services.ClaimService),
	}
}
