package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaguePokemonService interface {
	GetAllPokemonInLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.LeaguePokemon, error)
	CreatePokemonForLeague(currentUser *models.User, input *common.LeaguePokemonCreateRequestDTO) (*models.LeaguePokemon, error)
	BatchCreatePokemonForLeague(currentUser *models.User, inputs []common.LeaguePokemonCreateRequestDTO) ([]models.LeaguePokemon, error)
	UpdateLeaguePokemon(currentUser *models.User, input *common.LeaguePokemonUpdateRequest) (*models.LeaguePokemon, error)

	// Consider implementing service methods for GetAvailablePokemonByLeague
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
func (s *leaguePokemonServiceImpl) getLeagueByID(leagueID, currentUserID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Service: getPokemonSpeciesByID) - pokemon %d species not found: %v\n", pokemonSpeciesID, err)
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Service: getPokemonSpeciesByID) - could not retrieve pokemon species %d.\n", pokemonSpeciesID)
		return nil, common.ErrInternalService
	}
	return pokemon, nil
}

// GetAllPokemonInLeague returns all pokemon in a particular league
// Player permissions required: rbac.PermissionReadLeaguePokemon (consider making it unprotected)
func (s *leaguePokemonServiceImpl) GetAllPokemonInLeague(currentUser *models.User, leagueID uuid.UUID) ([]models.LeaguePokemon, error) {
	// leagueID should already be validated by middleware

	// // commented in case this ever becomes an unprotected route
	// league, err := s.getLeagueByID(leagueID)
	// if err != nil {
	// 	return nil, err // correct error already returned by helper
	// }

	leaguePokemon, err := s.leaguePokemonRepo.GetAllPokemonByLeague(leagueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound { // should be impossible
			log.Printf("LOG: (Service: CreatePokemonForLeague) - league %s not found (user %s): %v", leagueID, currentUser.ID, err)
			return nil, common.ErrLeagueNotFound
		}
		log.Printf("LOG: (Service: CreatePokemonForLeague) - error fetching all league pokemon for league %s (user %s): %v", leagueID, currentUser.ID, err)
		return nil, common.ErrInternalService
	}
	return leaguePokemon, nil
}

// handles creating a single LeaguePokemon entry.
// Player Permission required: rbac.PermissionCreateLeaguePokemon
func (s *leaguePokemonServiceImpl) CreatePokemonForLeague(
	currentUser *models.User,
	input *common.LeaguePokemonCreateRequestDTO,
) (*models.LeaguePokemon, error) {
	league, err := s.getLeagueByID(input.LeagueID, currentUser.ID)
	if err != nil {
		return nil, common.ErrLeagueNotFound
	}

	// League must be in Setup status to add new pokemon
	if league.Status != enums.LeagueStatusSetup {
		log.Printf("LOG: (Service: CreatePokemonForLeague) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
		return nil, common.ErrInvalidState
	}
	// Ensure PokemonSpeciesID is valid
	_, err = s.getPokemonSpeciesByID(input.PokemonSpeciesID)
	if err != nil {
		return nil, common.ErrPokemonSpeciesNotFound
	}

	leaguePokemon := &models.LeaguePokemon{
		LeagueID:         input.LeagueID,
		PokemonSpeciesID: input.PokemonSpeciesID,
		Cost:             input.Cost,
		IsAvailable:      true,
	}

	createdLeaguePokemon, err := s.leaguePokemonRepo.CreateLeaguePokemon(leaguePokemon)
	if err != nil {
		log.Printf("LOG: (Service: CreatePokemonForLeague) - failed to create league pokemon: %v\n", err)
		return nil, common.ErrInternalService
	}

	log.Printf("LOG: (Service: CreatePokemonForLeague) - Successfully created league pokemon for league %s, species %d", input.LeagueID, input.PokemonSpeciesID)
	return createdLeaguePokemon, nil
}

// BatchCreatePokemonForLeague handles creating multiple LeaguePokemon entries.
// Player Permission required: rbac.PermissionCreateLeaguePokemon
func (s *leaguePokemonServiceImpl) BatchCreatePokemonForLeague(
	currentUser *models.User,
	inputs []common.LeaguePokemonCreateRequestDTO,
) ([]models.LeaguePokemon, error) {
	if len(inputs) == 0 {
		return []models.LeaguePokemon{}, nil
	}

	// Pre-validate all inputs before making any database changes
	leagueCache := make(map[uuid.UUID]*models.League)
	var leaguePokemonToCreate []models.LeaguePokemon

	for _, input := range inputs {
		// Check if we've already validated this league
		league, exists := leagueCache[input.LeagueID]
		if !exists {
			var err error
			league, err = s.getLeagueByID(input.LeagueID, currentUser.ID)
			if err != nil {
				return nil, err
			}
			leagueCache[input.LeagueID] = league
		}

		// League must be in Setup status to add new pokemon
		if league.Status != enums.LeagueStatusSetup {
			log.Printf("LOG: (Service: BatchCreatePokemonForLeague) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
			return nil, common.ErrInvalidState
		}

		// Ensure PokemonSpeciesID is valid
		_, err := s.getPokemonSpeciesByID(input.PokemonSpeciesID)
		if err != nil {
			return nil, err
		}

		leaguePokemon := models.LeaguePokemon{
			LeagueID:         input.LeagueID,
			PokemonSpeciesID: input.PokemonSpeciesID,
			Cost:             input.Cost,
			IsAvailable:      true,
		}
		leaguePokemonToCreate = append(leaguePokemonToCreate, leaguePokemon)
	}

	// All validations passed, now perform the batch creation
	createdLeaguePokemon, err := s.leaguePokemonRepo.CreateLeaguePokemonBatch(leaguePokemonToCreate)
	if err != nil {
		log.Printf("LOG: (Service: BatchCreatePokemonForLeague) - failed to batch create league pokemon: %v\n", err)
		return nil, common.ErrInternalService
	}

	log.Printf("LOG: (Service: BatchCreatePokemonForLeague) - Successfully batch created %d league pokemon", len(createdLeaguePokemon))
	return createdLeaguePokemon, nil
}

// UpdateLeaguePokemon handles updating an existing LeaguePokemon entry.
// Player Permission required: rbac.PermissionUpdateLeaguePokemon
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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // unreachable code (right??)
			log.Printf("(Service: UpdateLeaguePokemon) - league %s does not exist: %s\n", existingLeaguePokemon.LeagueID, err.Error())
			return nil, common.ErrLeagueNotFound
		}
		log.Printf("(Service: UpdateLeaguePokemon) - could not fetch league %s: %v\n", existingLeaguePokemon.LeagueID, err)
		return nil, common.ErrInternalService
	}

	// Operation allowed only during Setup or Drafting status
	// NOTE: COULD CHANGE WITH THE TRANSFER CREDIT SYSTEM. here updates would happen during the league
	if currentUser.Role != "admin" &&
		(league.Status != enums.LeagueStatusSetup && league.Status != enums.LeagueStatusDrafting) {
		log.Printf("(Service: UpdateLeaguePokemon) - operation not allowed for current league status: %s for user %s", league.Status, currentUser.ID)
		return nil, common.ErrInvalidState
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
