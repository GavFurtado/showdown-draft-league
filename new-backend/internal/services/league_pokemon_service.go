package services

import (
	"errors"
	"fmt"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
)

type LeaguePokemonService interface {
	CreatePokemonForLeague(currentUser *models.User, input *common.LeaguePokemonCreateRequest) (*models.LeaguePokemon, error)
	BatchCreatePokemonForLeague(currentUser *models.User, inputs []*common.LeaguePokemonCreateRequest) ([]*models.LeaguePokemon, error)
	UpdateLeaguePokemon(currentUser *models.User, input *common.LeaguePokemonUpdateRequest) (*models.LeaguePokemon, error)
}

type leaguePokemonServiceImpl struct {
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	leagueRepo         repositories.LeagueRepository
	userRepo           repositories.UserRepository
	pokemonSpeciesRepo repositories.PokemonSpeciesRepository
}

func NewLeaguePokemonService(
	leaguePokemonRepo repositories.LeaguePokemonRepository,
	leagueRepo repositories.LeagueRepository,
	userRepo repositories.UserRepository,
	pokemonSpeciesRepo repositories.PokemonSpeciesRepository,
) LeaguePokemonService {
	return &leaguePokemonServiceImpl{
		leaguePokemonRepo:  leaguePokemonRepo,
		leagueRepo:         leagueRepo,
		userRepo:           userRepo,
		pokemonSpeciesRepo: pokemonSpeciesRepo,
	}
}

// -- Private Helpers --
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
			log.Printf("(Service: getLeagueByID) - could not find league %s. (currentUser.ID: %s)\n", leagueID, currentUserID)
			return nil, common.ErrLeagueNotFound
		}
		// other error
		log.Printf("(Service: getLeagueByID) - could not retrieve league by leagueID %s (currentUser.ID: %s)\n", leagueID, currentUserID)
		return nil, common.ErrInternalService
	}
	return league, nil
}

func (s *leaguePokemonServiceImpl) getPokemonSpeciesByID(pokemonSpeciesID int64) (*models.PokemonSpecies, error) {
	pokemon, err := s.pokemonSpeciesRepo.GetPokemonSpeciesByID(pokemonSpeciesID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("(Service: getPokemonSpeciesByID) - pokemon %d species not found: %v\n", pokemonSpeciesID, err)
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
	if inLeague, err := s.isUserPlayerInLeague(currentUser.ID, league.ID); !inLeague && currentUser.Role != "admin" {
		if err != nil {
			return nil, nil, err
		}
		log.Printf("(Service: validateAndFetchResources) - user %s does not have a player in the league %s.", currentUser.ID, league.ID)
		return nil, nil, common.ErrUnauthorized
	}
	// check if owner or an admin using helper
	isOwner, err := s.leagueRepo.IsUserOwner(currentUser.ID, league.ID)
	if err != nil {
		return nil, nil, err
	}
	if !isOwner && currentUser.Role != "admin" {
		log.Printf("(Service: validateAndFetchResources) - user %s is not an owner for the league %s.", currentUser.ID, league.ID)
		return nil, nil, common.ErrUnauthorized
	}
	// user is an owner player in the league which exists
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
		log.Printf("(Service: CreatePokemonForLeague) - failed to create league pokemon: %v\n", err)
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
			log.Printf("(Service: CreatePokemonForLeague) - failed to create league pokemon: %v\n", err)
			return nil, common.ErrInternalService
		}

		log.Printf("(Service: CreatePokemonForLeague) - Successfully created league pokemon for league %s, species %d", input.LeagueID, input.PokemonSpeciesID)
		batchCreatedLeaguePokemon = append(batchCreatedLeaguePokemon, createdLeaguePokemon)
	}
	return batchCreatedLeaguePokemon, nil
}

func (s *leaguePokemonServiceImpl) UpdateLeaguePokemon(
	currentUser *models.User,
	input *common.LeaguePokemonUpdateRequest,
) (*models.LeaguePokemon, error) {
	existingLeaguePokemon, err := s.leaguePokemonRepo.GetLeaguePokemonByID(input.LeaguePokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: UpdateLeaguePokemon) - league pokemon %s does not exist: %v\n", input.LeaguePokemonID, err)
			return nil, common.ErrLeaguePokemonNotFound
		}
		log.Printf("(Service: UpdateLeaguePokemon) - could not fetch league pokemon: %s\n", err.Error())
		return nil, common.ErrInternalService
	}

	league, err := s.leagueRepo.GetLeagueByID(existingLeaguePokemon.LeagueID)
	// idk how this could happen ngl
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: UpdateLeaguePokemon) - league %s does not exist: %s\n", existingLeaguePokemon.LeagueID, err.Error())
			return nil, common.ErrLeagueNotFound
		}
		log.Printf("(Service: UpdateLeaguePokemon) - could not fetch league %s: %v\n", existingLeaguePokemon.LeagueID, err)
		return nil, common.ErrInternalService
	}

	isUserPlayer, err := s.isUserPlayerInLeague(currentUser.ID, existingLeaguePokemon.LeagueID)
	if err != nil {
		return nil, err
	}

	isOwner, err := s.leagueRepo.IsUserOwner(currentUser.ID, existingLeaguePokemon.LeagueID)
	if err != nil {
		return nil, err
	}

	if !isUserPlayer && !isOwner && currentUser.Role != "admin" {
		log.Printf("(Service: UpdateLeaguePokemon) - user %s is not authorized to update league pokemon in league %s.\n", currentUser.ID, existingLeaguePokemon.LeagueID)
		return nil, common.ErrUnauthorized
	}

	if league.Status != models.LeagueStatusSetup && league.Status != models.LeagueStatusDrafting {
		log.Printf("(Service: UpdateLeaguePokemon) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
		return nil, common.ErrUnauthorized
	}

	// Update fields if provided in the input
	if input.Cost != nil && *input.Cost != *existingLeaguePokemon.Cost {
		existingLeaguePokemon.Cost = input.Cost
	}
	// Check if IsAvailable was explicitly provided and different from existing
	if input.IsAvailable != existingLeaguePokemon.IsAvailable {
		existingLeaguePokemon.IsAvailable = input.IsAvailable
	}

	updatedLeaguePokemon, err := s.leaguePokemonRepo.UpdateLeaguePokemon(existingLeaguePokemon)
	if err != nil {
		log.Printf("(Service: UpdateLeaguePokemon) - failed to update league pokemon: %s\n", err.Error())
		return nil, common.ErrInternalService
	}

	log.Printf("(Service: UpdateLeaguePokemon) - Successfully updated league pokemon %s", updatedLeaguePokemon.ID)
	return updatedLeaguePokemon, nil
}
