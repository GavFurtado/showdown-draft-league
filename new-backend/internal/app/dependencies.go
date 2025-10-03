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
	LeaguePokemonRepository  repositories.LeaguePokemonRepository
	DraftRepository          repositories.DraftRepository
	DraftedPokemonRepository repositories.DraftedPokemonRepository
	GameRepository           repositories.GameRepository
}

type Services struct {
	AuthService           services.AuthService
	DraftService          services.DraftService
	DraftedPokemonService services.DraftedPokemonService
	JWTService            services.JWTService
	LeaguePokemonService  services.LeaguePokemonService
	LeagueService         services.LeagueService
	PlayerService         services.PlayerService
	PokemonSpeciesService services.PokemonSpeciesService
	RBACService           services.RBACService
	UserService           services.UserService
	WebhookService        services.WebhookService
}

type Controllers struct {
	AuthController           controllers.AuthController
	LeagueController         controllers.LeagueController
	PlayerController         controllers.PlayerController
	UserController           controllers.UserController
	PokemonSpeciesController controllers.PokemonSpeciesController
	LeaguePokemonController  controllers.LeaguePokemonController
}
