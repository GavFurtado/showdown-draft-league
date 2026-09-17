package app

import (
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/controllers"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
)

type Repositories struct {
	LeagueRepository         repositories.LeagueRepository
	UserRepository           repositories.UserRepository
	PokemonSpeciesRepository repositories.PokemonSpeciesRepository
	DraftRepository          repositories.DraftRepository
	GameRepository           repositories.GameRepository

	DraftPickRepository repositories.DraftPickRepository
	ClaimRepository     repositories.ClaimRepository
	PoolEntryRepository repositories.PoolEntryRepository
	PlayerRepository    repositories.PlayerRepository
	SpectatorRepository repositories.SpectatorRepository
}

type Services struct {
	Logger                *slog.Logger
	AuthService           services.AuthService
	DevAuthService        services.DevAuthService
	DraftService          services.DraftService
	JWTService            services.JWTService
	LeagueService         services.LeagueService
	PokemonSpeciesService services.PokemonSpeciesService
	RBACService           services.RBACService
	UserService           services.UserService
	WebhookService        services.WebhookService
	SchedulerService      services.SchedulerService
	GameService           services.GameService
	TransferService       services.TransferService

	PoolEntryService services.PoolEntryService
	PlayerService    services.PlayerService
	SpectatorService services.SpectatorService
	DraftPickService services.DraftPickService
	ClaimService     services.ClaimService
}

type Controllers struct {
	AuthController           controllers.AuthController
	DevAuthController        controllers.DevAuthController
	LeagueController         controllers.LeagueController
	UserController           controllers.UserController
	PokemonSpeciesController controllers.PokemonSpeciesController
	DraftController          controllers.DraftController
	GameController           controllers.GameController
	TransferController       controllers.TransferController

	PoolEntryController controllers.PoolEntryController
	PlayerController    controllers.PlayerController
	SpectatorController controllers.SpectatorController
	DraftPickController controllers.DraftPickController
	ClaimController     controllers.ClaimController
}
