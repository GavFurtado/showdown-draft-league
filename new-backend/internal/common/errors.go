package common

import "errors"

var (

	// Common resource not found errors
	ErrLeagueNotFound         = errors.New("league not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrPlayerNotFound         = errors.New("player not found")
	ErrPokemonSpeciesNotFound = errors.New("species not found")
	ErrDraftedPokemonNotFound = errors.New("drafted pokemon instance not found")

	// Player creation specific errors
	ErrUserAlreadyInLeague  = errors.New("user is already a player in this league")
	ErrInLeagueNameTaken    = errors.New("the in-league name is already taken in this league")
	ErrTeamNameTaken        = errors.New("the team name is already taken in this league")
	ErrFailedToCreatePlayer = errors.New("failed to add player to league")

	// Authorization errors
	ErrUnauthorized           = errors.New("unauthorized: you do not have permission to perform this action")
	ErrInvalidUpdateForPlayer = errors.New("players cannot update score or draft details directly")

	// Business Logic Errors
	ErrMaxLeagueCreationLimitReached = errors.New("maximum league creation limit reached")
	ErrInvalidInput                  = errors.New("invalid input/request")
	ErrConflict                      = errors.New("record already exists. cannot make a duplicate")

	// Internal Service Errors
	ErrInternalService = errors.New("internal service error")
)
