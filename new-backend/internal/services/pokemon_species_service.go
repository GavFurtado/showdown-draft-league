package services

import (
	"errors"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"gorm.io/gorm"
)

type PokemonSpeciesService interface {
	GetAllPokemonSpecies() ([]responses.PokemonSpeciesListResponseDTO, error) // Updated return type
	ListPokemonSpecies(filter string) ([]models.PokemonSpecies, error)
	GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error)
	GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error)
	CreatePokemonSpecies(pokemon *models.PokemonSpecies) error
	UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error
	DeletePokemonSpecies(id int64) error
}

type pokemonServiceImpl struct {
	pokemonRepo repositories.PokemonSpeciesRepository
	logger      *slog.Logger
}

func NewPokemonSpeciesService(
	logger *slog.Logger,
	pokemonRepo repositories.PokemonSpeciesRepository,
) PokemonSpeciesService {
	return &pokemonServiceImpl{
		pokemonRepo: pokemonRepo,
		logger:      utils.LoggerWithService(logger, "PokemonSpeciesService"),
	}
}

// retrieves all pokemon species.
func (s *pokemonServiceImpl) GetAllPokemonSpecies() ([]responses.PokemonSpeciesListResponseDTO, error) { // Updated return type
	allPokemon, err := s.pokemonRepo.GetAllPokemonSpecies()
	if err != nil {
		s.logger.Error("GetAllPokemonSpecies - failed to get all pokemon species", "error", err)
		return nil, types.ErrInternalService
	}

	var pokemonDTOs []responses.PokemonSpeciesListResponseDTO
	for _, pokemon := range allPokemon {
		pokemonDTOs = append(pokemonDTOs, responses.PokemonSpeciesListResponseDTO{
			ID:           pokemon.ID,
			Name:         pokemon.Name,
			Types:        pokemon.Types,
			FrontDefault: pokemon.Sprites.FrontDefault,
		})
	}

	return pokemonDTOs, nil
}

// lists pokemon species based on a filter.
func (s *pokemonServiceImpl) ListPokemonSpecies(filter string) ([]models.PokemonSpecies, error) {
	pokemon, err := s.pokemonRepo.FindPokemonSpecies(filter)
	if err != nil {
		s.logger.Error("ListPokemonSpecies - failed to find pokemon species with filter", "filter", filter, "error", err)
		return nil, types.ErrInternalService
	}

	return pokemon, nil
}

// retrieves a single pokemon species by its ID.
func (s *pokemonServiceImpl) GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error) {
	if id <= 0 {
		s.logger.Error("GetPokemonSpeciesByID - invalid input ID", "id", id)
		return nil, types.ErrInvalidInput
	}

	pokemon, err := s.pokemonRepo.GetPokemonSpeciesByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Info("GetPokemonSpeciesByID - pokemon species not found", "id", id)
			return nil, types.ErrPokemonSpeciesNotFound
		}
		s.logger.Error("GetPokemonSpeciesByID - failed to get pokemon species by ID", "id", id, "error", err)
		return nil, types.ErrInternalService
	}

	return pokemon, nil
}

// retrieves a single pokemon species by exact name.
func (s *pokemonServiceImpl) GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error) {
	if name == "" {
		s.logger.Error("GetPokemonSpeciesByName - invalid input: empty name")
		return nil, types.ErrInvalidInput
	}

	pokemon, err := s.pokemonRepo.GetPokemonSpeciesByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Info("GetPokemonSpeciesByName - pokemon species not found", "name", name)
			return nil, types.ErrPokemonSpeciesNotFound
		}
		s.logger.Error("GetPokemonSpeciesByName - failed", "name", name, "error", err)
		return nil, types.ErrInternalService
	}

	return pokemon, nil
}

// creates a new pokemon species record.
func (s *pokemonServiceImpl) CreatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	if pokemon == nil || pokemon.ID == 0 || pokemon.Name == "" {
		s.logger.Error("CreatePokemonSpecies - invalid input: pokemon is nil, ID is zero, or Name is empty")
		return types.ErrInvalidInput
	}

	// Check if pokemon with the same ID or name already exists
	existingByID, err := s.pokemonRepo.GetPokemonSpeciesByID(pokemon.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("CreatePokemonSpecies - error checking existing pokemon by ID", "pokemon_id", pokemon.ID, "error", err)
		return types.ErrInternalService
	}

	if existingByID != nil {
		s.logger.Info("CreatePokemonSpecies - pokemon species already exists", "pokemon_id", pokemon.ID)
		return types.ErrConflict
	}

	existingByName, err := s.pokemonRepo.GetPokemonSpeciesByName(pokemon.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("CreatePokemonSpecies - error checking existing pokemon by name", "name", pokemon.Name, "error", err)
		return types.ErrInternalService
	}
	if existingByName != nil {
		s.logger.Info("CreatePokemonSpecies - pokemon species with name already exists", "name", pokemon.Name)
		return types.ErrConflict
	}

	// Create the pokemon species
	err = s.pokemonRepo.CreatePokemonSpecies(pokemon)
	if err != nil {
		s.logger.Error("CreatePokemonSpecies - failed to create", "pokemon_id", pokemon.ID, "name", pokemon.Name, "error", err)
		return types.ErrInternalService
	}

	return nil
}

// updates a pokemon species record.
func (s *pokemonServiceImpl) UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	if pokemon == nil || pokemon.ID == 0 {
		s.logger.Error("UpdatePokemonSpecies - invalid input: pokemon is nil or ID is zero")
		return types.ErrInvalidInput
	}

	// Check if the pokemon exists
	_, err := s.pokemonRepo.GetPokemonSpeciesByID(pokemon.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Info("UpdatePokemonSpecies - pokemon species not found for update", "pokemon_id", pokemon.ID)
			return types.ErrPokemonSpeciesNotFound
		}
		s.logger.Error("UpdatePokemonSpecies - error checking existing pokemon for update", "pokemon_id", pokemon.ID, "error", err)
		return types.ErrInternalService
	}

	// Update the pokemon species
	err = s.pokemonRepo.UpdatePokemonSpecies(pokemon)
	if err != nil {
		s.logger.Error("UpdatePokemonSpecies - failed to update", "pokemon_id", pokemon.ID, "error", err)
		return types.ErrInternalService
	}

	return nil
}

// deletes a pokemon species record.
func (s *pokemonServiceImpl) DeletePokemonSpecies(id int64) error {
	if id <= 0 {
		s.logger.Error("DeletePokemonSpecies - invalid input ID", "id", id)
		return types.ErrInvalidInput
	}

	// Check if the pokemon exists
	_, err := s.pokemonRepo.GetPokemonSpeciesByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Info("DeletePokemonSpecies - pokemon species not found for deletion", "id", id)
			return types.ErrPokemonSpeciesNotFound
		}
		s.logger.Error("DeletePokemonSpecies - error checking existing pokemon for deletion", "id", id, "error", err)
		return types.ErrInternalService
	}

	// Delete the pokemon species
	err = s.pokemonRepo.DeletePokemonSpecies(id)
	if err != nil {
		s.logger.Error("DeletePokemonSpecies - failed to delete", "id", id, "error", err)
		return types.ErrInternalService
	}
	return nil
}
