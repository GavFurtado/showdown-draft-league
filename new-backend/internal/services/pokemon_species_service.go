package services

import (
	"errors"
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"gorm.io/gorm"
)

type PokemonSpeciesService interface {
	GetAllPokemonSpecies() ([]models.PokemonSpecies, error)
	ListPokemonSpecies(filter string) ([]models.PokemonSpecies, error)
	GetPokemonSpeciesByID(id int) (*models.PokemonSpecies, error)
	GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error)
	CreatePokemonSpecies(pokemon *models.PokemonSpecies) error
	UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error
	DeletePokemonSpecies(id int) error
}

type pokemonServiceImpl struct {
	pokemonRepo *repositories.PokemonSpeciesRepository
}

func NewPokemonSpeciesService(repo *repositories.PokemonSpeciesRepository) PokemonSpeciesService {
	return &pokemonServiceImpl{
		pokemonRepo: repo,
	}
}

// retrieves all pokemon species.
func (s *pokemonServiceImpl) GetAllPokemonSpecies() ([]models.PokemonSpecies, error) {
	return s.pokemonRepo.GetAllPokemonSpecies()
}

// lists pokemon species based on a filter.
func (s *pokemonServiceImpl) ListPokemonSpecies(filter string) ([]models.PokemonSpecies, error) {
	return s.pokemonRepo.FindPokemonSpecies(filter)
}

// retrieves a single pokemon species by its ID.
func (s *pokemonServiceImpl) GetPokemonSpeciesByID(id int) (*models.PokemonSpecies, error) {
	if id <= 0 {
		return nil, common.ErrInvalidInput
	}

	pokemon, err := s.pokemonRepo.GetPokemonSpeciesByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPokemonSpeciesNotFound
		}
		return nil, fmt.Errorf("failed to get pokemon species by ID: %w", err)
	}
	return pokemon, nil
}

// retrieves a single pokemon species by exact name.
func (s *pokemonServiceImpl) GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error) {
	if name == "" {
		return nil, common.ErrBadRequest
	}

	pokemon, err := s.pokemonRepo.GetPokemonSpeciesByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPokemonSpeciesNotFound
		}
		return nil, fmt.Errorf("failed to get pokemon species by name: %w", err)
	}
	return pokemon, nil
}

// creates a new pokemon species record.
func (s *pokemonServiceImpl) CreatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	if pokemon == nil || pokemon.ID == 0 || pokemon.Name == "" {
		return common.ErrInvalidInput
	}

	// Check if pokemon with the same ID or name already exists
	existingByID, err := s.pokemonRepo.GetPokemonSpeciesByID(int(pokemon.ID))
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("error checking existing pokemon by ID: %w", err)
	}
	if existingByID != nil {
		return common.ErrConflict
	}

	existingByName, err := s.pokemonRepo.GetPokemonSpeciesByName(pokemon.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("error checking existing pokemon by name: %w", err)
	}
	if existingByName != nil {
		return common.ErrConflict
	}

	// Create the pokemon species
	err = s.pokemonRepo.CreatePokemonSpecies(pokemon)
	if err != nil {
		return fmt.Errorf("failed to create pokemon species: %w", err)
	}

	return nil
}

// updates a pokemon species record.
func (s *pokemonServiceImpl) UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	if pokemon == nil || pokemon.ID == 0 {
		return common.ErrInvalidInput
	}

	// Check if the pokemon exists
	_, err := s.pokemonRepo.GetPokemonSpeciesByID(int(pokemon.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrPokemonSpeciesNotFound // Use specific not found error from common
		}
		return fmt.Errorf("error checking existing pokemon for update: %w", err)
	}

	// Update the pokemon species
	err = s.pokemonRepo.UpdatePokemonSpecies(pokemon)
	if err != nil {
		return fmt.Errorf("failed to update pokemon species: %w", err)
	}

	return nil
}

// deletes a pokemon species record.
func (s *pokemonServiceImpl) DeletePokemonSpecies(id int) error {
	if id <= 0 {
		return common.ErrInvalidInput
	}

	// Check if the pokemon exists
	_, err := s.pokemonRepo.GetPokemonSpeciesByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrPokemonSpeciesNotFound
		}
		return fmt.Errorf("error checking existing pokemon for deletion: %w", err)
	}

	// Delete the pokemon species
	err = s.pokemonRepo.DeletePokemonSpecies(id)
	if err != nil {
		return fmt.Errorf("failed to delete pokemon species: %w", err)
	}

	return nil
}
