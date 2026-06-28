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
		GameRepository:           repositories.NewGameRepository(db),
		LeaguePokemonRepository:  repositories.NewLeaguePokemonRepository(db),
		DraftedPokemonRepository: repositories.NewDraftedPokemonRepository(db),
		PokemonSpeciesRepository: repositories.NewPokemonSpeciesRepository(db),

		DraftPickRepository:    repositories.NewDraftPickRepository(db),
		ClaimRepository:        repositories.NewClaimRepository(db),
		PoolEntryRepository:    repositories.NewPoolEntryRepository(db),
		LeagueMemberRepository: repositories.NewLeagueMemberRepository(db),
	}
}

func NewServices(repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Services {
	jwtService := services.NewJWTService(cfg.JWTSecret)
	rbacService := services.NewRBACService(repos.LeagueRepository, repos.UserRepository, repos.LeagueMemberRepository)
	webhookService := services.NewWebhookService()

	draftService := services.NewDraftService(
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
		&u.TaskHeap{},
		repos.LeagueRepository,
		repos.DraftRepository,
	)

	transferService := services.NewTransferService(
		repos.DraftedPokemonRepository,
		repos.LeaguePokemonRepository,
		repos.LeagueRepository,
		repos.PlayerRepository,
		repos.LeagueMemberRepository,
	)

	transferService.SetNewRepositories(
		repos.ClaimRepository,
		repos.PoolEntryRepository,
		repos.LeagueMemberRepository,
	)

	gameService := services.NewGameService(repos.GameRepository, repos.LeagueRepository, repos.PlayerRepository, repos.LeagueMemberRepository)

	leagueService := services.NewLeagueService(repos.LeagueRepository, repos.PlayerRepository, repos.LeagueMemberRepository, repos.LeaguePokemonRepository, repos.DraftedPokemonRepository, repos.DraftRepository, repos.GameRepository)

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
		JWTService:           *jwtService,
		UserService:          services.NewUserService(repos.UserRepository),
		RBACService:          rbacService,
		WebhookService:       webhookService,
		LeagueService:        leagueService,
		AuthService:          services.NewAuthService(repos.UserRepository, jwtService, discordOauthConfig),
		DraftService:         draftService,
		PokemonSpeciesService: services.NewPokemonSpeciesService(repos.PokemonSpeciesRepository),
		SchedulerService:      schedulerService,
		GameService:           services.NewGameService(repos.GameRepository, repos.LeagueRepository, repos.PlayerRepository, repos.LeagueMemberRepository),
		TransferService:       transferService,

		PoolEntryService:    services.NewPoolEntryService(repos.PoolEntryRepository, repos.LeagueRepository, repos.UserRepository, repos.PokemonSpeciesRepository),
		LeagueMemberService: services.NewLeagueMemberService(repos.LeagueMemberRepository, repos.PlayerRepository, repos.LeagueRepository, repos.UserRepository, repos.DraftedPokemonRepository),
		DraftPickService:    services.NewDraftPickService(repos.DraftPickRepository, repos.DraftRepository),
		ClaimService:        services.NewClaimService(repos.ClaimRepository),
	}
}

func NewControllers(services *Services, repos *Repositories, cfg *config.Config, discordOauthConfig *oauth2.Config) *Controllers {
	return &Controllers{
		AuthController:           *controllers.NewAuthController(services.AuthService, cfg, discordOauthConfig),
		LeagueController:         controllers.NewLeagueController(services.LeagueService),
		UserController:           controllers.NewUserController(services.UserService),
		PokemonSpeciesController: controllers.NewPokemonSpeciesController(services.PokemonSpeciesService),
		DraftController:          controllers.NewDraftController(services.DraftService),
		GameController:           controllers.NewGameController(services.GameService, services.LeagueService),
		TransferController:       controllers.NewTransferController(services.TransferService),

		PoolEntryController:    controllers.NewPoolEntryController(services.PoolEntryService),
		LeagueMemberController: controllers.NewLeagueMemberController(services.LeagueMemberService),
		DraftPickController:    controllers.NewDraftPickController(services.DraftPickService, services.DraftService),
		ClaimController:        controllers.NewClaimController(services.ClaimService),
	}
}
