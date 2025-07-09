package services

import (
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaguePokemonService interface {
	CreatePokemonForLeague(currentUser *models.User, input *common.LeaguePokemonCreateRequest) (*models.LeaguePokemon, error)
	BatchCreatePokemonForLeague(currentUser *models.User, inputs []*common.LeaguePokemonCreateRequest) ([]*models.LeaguePokemon, error)
}

type leaguePokemonServiceImpl struct {
	leaguePokemonRepo  *repositories.LeaguePokemonRepository
	leagueRepo         *repositories.LeagueRepository
	userRepo           *repositories.UserRepository
	pokemonSpeciesRepo *repositories.PokemonSpeciesRepository
}

func NewLeaguePokemonService(
	leaguePokemonRepo *repositories.LeaguePokemonRepository,
	leagueRepo *repositories.LeagueRepository,
	userRepo *repositories.UserRepository,
	pokemonSpeciesRepo *repositories.PokemonSpeciesRepository,
) LeaguePokemonService {
	return &leaguePokemonServiceImpl{
		leaguePokemonRepo:  leaguePokemonRepo,
		leagueRepo:         leagueRepo,
		userRepo:           userRepo,
		pokemonSpeciesRepo: pokemonSpeciesRepo,
	}
}

// -- Private Helpers --
func (s *leaguePokemonServiceImpl) isUserCommissioner(userID, leagueID uuid.UUID) (bool, error) {
	isComm, err := s.leagueRepo.IsUserCommissioner(userID, leagueID)
	if err != nil {
		log.Printf("(Service: isUserCommissioner) - Failed to check commissioner status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check commissioner status: %w", err)
	}
	return isComm, nil
}

func (s *leaguePokemonServiceImpl) isUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {

	isPlayer, err := s.leagueRepo.IsUserPlayerInLeague(userID, leagueID)
	if err != nil {
		log.Printf("(Service: isUserPlayerInLeague) - Failed to check player status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check player status: %w", err)
	}
	return isPlayer, nil
}

func (s *leaguePokemonServiceImpl) getLeagueByID(leagueID, currentUserID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("(Service: getLeagueByID) - could not find league %d. (currentUser.ID: %d)\n", leagueID, currentUserID)
			return nil, common.ErrLeagueNotFound
		}
		// other error
		log.Printf("(Service: getLeagueByID) - could not retrieve league by leagueID %d (currentUser.ID: %d)\n", leagueID, currentUserID)
		return nil, common.ErrInternalService
	}
	return league, nil
}

func (s *leaguePokemonServiceImpl) getPokemonSpeciesByID(pokemonSpeciesID int64) (*models.PokemonSpecies, error) {
	pokemon, err := s.pokemonSpeciesRepo.GetPokemonSpeciesByID(pokemonSpeciesID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("(Service: getPokemonSpeciesByID) - pokemon %d species not found: %w\n", pokemonSpeciesID, err)
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Service: getPokemonSpeciesByID) - could not retrieve pokemon species %d.\n", pokemonSpeciesID)
		return nil, common.ErrInternalService
	}
	return pokemon, nil
}

// -- Validation and Resource Fetching Helper --
// This helper consolidates the initial checks and resource fetching for creating pokemon
func (s *leaguePokemonServiceImpl) validateAndFetchResourcesToCreatePokemon(
	currentUser *models.User,
	input *common.LeaguePokemonCreateRequest,
) (*models.League, *models.PokemonSpecies, error) {
	league, err := s.getLeagueByID(input.LeagueID, currentUser.ID)
	if err != nil {
		return nil, nil, err
	}
	// check if user has a player in league or if user is an admin using helper
	if inLeague, err := s.isUserPlayerInLeague(currentUser.ID, league.ID); !inLeague && !currentUser.IsAdmin {
		if err != nil {
			return nil, nil, err
		}
		log.Printf("(Service: validateAndFetchResources) - user %s does not have a player in the league %s.", currentUser.ID, league.ID)
		return nil, nil, common.ErrUnauthorized
	}
	// check if commissioner or an admin using helper
	if isComm, err := s.isUserCommissioner(currentUser.ID, league.ID); !isComm && !currentUser.IsAdmin {
		if err != nil {
			return nil, nil, err
		}
		log.Printf("(Service: validateAndFetchResources) - user %s is not a commissioner for the league %s.", currentUser.ID, league.ID)
		return nil, nil, common.ErrUnauthorized
	}
	// user is a commissioner player in the league which exists
	// check league status
	if league.Status != models.LeagueStatusSetup {
		log.Printf("(Service: validateAndFetchResources) - unauthorized: league %s status is not SETUP for user %s", league.ID, currentUser.ID)
		return nil, nil, common.ErrUnauthorized
	}
	// fetch pokemon species using helper
	pokemon, err := s.getPokemonSpeciesByID(input.PokemonSpeciesID)
	if err != nil {
		return nil, nil, err
	}
	return league, pokemon, nil
}

// CreatePokemonForLeague now uses the helper function
func (s *leaguePokemonServiceImpl) CreatePokemonForLeague(
	currentUser *models.User,
	input *common.LeaguePokemonCreateRequest,
) (*models.LeaguePokemon, error) {
	// Validate input and fetch necessary resources using the helper
	_, _, err := s.validateAndFetchResourcesToCreatePokemon(currentUser, input)
	if err != nil {
		log.Printf("(Service: CreatePokemonForLeague) - Validation and resource fetching failed: %v", err)
		return nil, err
	}

	// If helper succeeded, proceed to create the LeaguePokemon model
	leaguePokemon := &models.LeaguePokemon{
		LeagueID:         input.LeagueID,
		PokemonSpeciesID: input.PokemonSpeciesID,
		Cost:             input.Cost,
		IsAvailable:      true,
	}

	// Create the LeaguePokemon using the repository
	createdLeaguePokemon, err := s.leaguePokemonRepo.CreateLeaguePokemon(leaguePokemon)
	if err != nil {
		log.Printf("(Service: CreatePokemonForLeague) - failed to create league pokemon: %s", err.Error())
		return nil, common.ErrInternalService
	}

	log.Printf("(Service: CreatePokemonForLeague) - Successfully created league pokemon for league %s, species %d", input.LeagueID, input.PokemonSpeciesID)
	return createdLeaguePokemon, nil
}

func (s *leaguePokemonServiceImpl) BatchCreatePokemonForLeague(
	currentUser *models.User,
	inputs []*common.LeaguePokemonCreateRequest,
) ([]*models.LeaguePokemon, error) {
	var batchCreatedLeaguePokemon []*models.LeaguePokemon
	for _, input := range inputs {
		// Validate input and fetch necessary resources using the helper
		_, _, err := s.validateAndFetchResourcesToCreatePokemon(currentUser, input)
		if err != nil {
			log.Printf("(Service: CreatePokemonForLeague) - Validation and resource fetching failed: %v", err)
			return nil, err
		}

		// If helper succeeded, proceed to create the LeaguePokemon model
		leaguePokemon := &models.LeaguePokemon{
			LeagueID:         input.LeagueID,
			PokemonSpeciesID: input.PokemonSpeciesID,
			Cost:             input.Cost,
			IsAvailable:      true,
		}

		// Create the LeaguePokemon using the repository
		createdLeaguePokemon, err := s.leaguePokemonRepo.CreateLeaguePokemon(leaguePokemon)
		if err != nil {
			log.Printf("(Service: CreatePokemonForLeague) - failed to create league pokemon: %w", err)
			return nil, common.ErrInternalService
		}

		log.Printf("(Service: CreatePokemonForLeague) - Successfully created league pokemon for league %s, species %d", input.LeagueID, input.PokemonSpeciesID)
		batchCreatedLeaguePokemon = append(batchCreatedLeaguePokemon, createdLeaguePokemon)
	}
	return batchCreatedLeaguePokemon, nil
}

// func (s *leaguePokemonServiceImpl) UpdateLeaguePokemon(
// 	currentUser *models.User,
// 	input *common.LeaguePokemonUpdateRequest,
// ) (models.LeaguePokemon, error) {
// 	s.isUserCommissioner()
//
// 	return nil, nil
// }
