package controllers

import "github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"

type LeagueController struct {
	leagueRepo         *repositories.LeagueRepository
	userRepo           *repositories.UserRepository
	playerRepo         *repositories.PlayerRepository
	leaguePokemonRepo  *repositories.LeaguePokemonRepository
	draftedPokemonRepo *repositories.DraftedPokemonRepository
	gameRepo           *repositories.GameRepository
}

func NewLeagueController(leagueRepo *repositories.LeagueRepository,
	userRepo *repositories.UserRepository,
	playerRepo *repositories.PlayerRepository,
	leaguePokemonRepo *repositories.LeaguePokemonRepository,
	draftedPokemonRepo *repositories.DraftedPokemonRepository,
	gameRepo *repositories.GameRepository) {

	return LeagueController{
		&userRepo,
		&playerRepo,
		&leaguePokemonRepo,
		&draftedPokemonRepo,
		&gameRepo,
	}
}
