package app

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/controllers"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
)

type Repositories struct {
	LeagueRepository         repositories.LeagueRepository
	UserRepository           repositories.UserRepository
	PlayerRepository         repositories.PlayerRepository
	PokemonSpeciesRepository repositories.PokemonSpeciesRepository
	DraftRepository          repositories.DraftRepository
	GameRepository           repositories.GameRepository
	LeaguePokemonRepository  repositories.LeaguePokemonRepository
	DraftedPokemonRepository repositories.DraftedPokemonRepository

	DraftPickRepository      repositories.DraftPickRepository
	ClaimRepository          repositories.ClaimRepository
	PoolEntryRepository      repositories.PoolEntryRepository
	LeagueMemberRepository   repositories.LeagueMemberRepository
}

type Services struct {
	AuthService           services.AuthService
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

	PoolEntryService    services.PoolEntryService
	LeagueMemberService services.LeagueMemberService
	DraftPickService    services.DraftPickService
	ClaimService        services.ClaimService
}

type Controllers struct {
	AuthController           controllers.AuthController
	LeagueController         controllers.LeagueController
	UserController           controllers.UserController
	PokemonSpeciesController controllers.PokemonSpeciesController
	DraftController          controllers.DraftController
	GameController           controllers.GameController
	TransferController       controllers.TransferController

	PoolEntryController    controllers.PoolEntryController
	LeagueMemberController controllers.LeagueMemberController
	DraftPickController    controllers.DraftPickController
	ClaimController        controllers.ClaimController
}
