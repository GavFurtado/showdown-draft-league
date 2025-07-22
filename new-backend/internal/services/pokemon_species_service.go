package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"gorm.io/gorm"
)

type PokemonSpeciesService interface {
	GetAllPokemonSpecies() ([]models.PokemonSpecies, error)
	ListPokemonSpecies(filter string) ([]models.PokemonSpecies, error)
	GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error)
	GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error)
	CreatePokemonSpecies(pokemon *models.PokemonSpecies) error
	UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error
	DeletePokemonSpecies(id int64) error
}

type pokemonServiceImpl struct {
	pokemonRepo       repositories.PokemonSpeciesRepository
	leaguePokemonRepo repositories.LeaguePokemonRepository
}

func NewPokemonSpeciesService(
	pokemonRepo repositories.PokemonSpeciesRepository,
	leaguePokemonRepo repositories.LeaguePokemonRepository,
) PokemonSpeciesService {
	return &pokemonServiceImpl{
		pokemonRepo:       pokemonRepo,
		leaguePokemonRepo: leaguePokemonRepo,
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
func (s *pokemonServiceImpl) GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error) {
	if id <= 0 {
		log.Printf("(Error: PokemonSpeciesService.GetPokemonSpeciesByID) - Invalid input ID: %d", id)
		return nil, common.ErrInvalidInput
	}

	pokemon, err := s.pokemonRepo.GetPokemonSpeciesByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Info: PokemonSpeciesService.GetPokemonSpeciesByID) - Pokemon species with ID %d not found", id)
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Error: PokemonSpeciesService.GetPokemonSpeciesByID) - Failed to get pokemon species by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to get pokemon species by ID: %w", err)
	}
	return pokemon, nil
}

// retrieves a single pokemon species by exact name.
func (s *pokemonServiceImpl) GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error) {
	if name == "" {
		log.Println("(Error: PokemonSpeciesService.GetPokemonSpeciesByName) - Invalid input: empty name")
		return nil, common.ErrInvalidInput
	}

	pokemon, err := s.pokemonRepo.GetPokemonSpeciesByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Info: PokemonSpeciesService.GetPokemonSpeciesByName) - Pokemon species with name '%s' not found", name)
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Error: PokemonSpeciesService.GetPokemonSpeciesByName) - Failed to get pokemon species by name '%s': %v", name, err)
		return nil, fmt.Errorf("failed to get pokemon species by name: %w", err)
	}
	return pokemon, nil
}

// creates a new pokemon species record.
func (s *pokemonServiceImpl) CreatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	if pokemon == nil || pokemon.ID == 0 || pokemon.Name == "" {
		log.Println("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Invalid input: pokemon is nil, ID is zero, or Name is empty")
		return common.ErrInvalidInput
	}

	// Check if pokemon with the same ID or name already exists
	existingByID, err := s.pokemonRepo.GetPokemonSpeciesByID(pokemon.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Error checking existing pokemon by ID %d: %v", pokemon.ID, err)
		return fmt.Errorf("error checking existing pokemon by ID: %w", err)
	}

	if existingByID != nil {
		log.Printf("(Info: PokemonSpeciesService.CreatePokemonSpecies) - Pokemon species with ID %d already exists", pokemon.ID)
		return common.ErrConflict
	}

	existingByName, err := s.pokemonRepo.GetPokemonSpeciesByName(pokemon.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Error checking existing pokemon by name '%s': %v", pokemon.Name, err)
		return fmt.Errorf("error checking existing pokemon by name: %w", err)
	}
	if existingByName != nil {
		log.Printf("(Info: PokemonSpeciesService.CreatePokemonSpecies) - Pokemon species with name '%s' already exists", pokemon.Name)
		return common.ErrConflict
	}

	// Create the pokemon species
	err = s.pokemonRepo.CreatePokemonSpecies(pokemon)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Failed to create pokemon species ID %d name '%s': %v", pokemon.ID, pokemon.Name, err)
		return fmt.Errorf("failed to create pokemon species: %w", err)
	}

	return nil
}

// updates a pokemon species record.
func (s *pokemonServiceImpl) UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	if pokemon == nil || pokemon.ID == 0 {
		log.Println("(Error: PokemonSpeciesService.UpdatePokemonSpecies) - Invalid input: pokemon is nil or ID is zero")
		return common.ErrInvalidInput
	}

	// Check if the pokemon exists
	_, err := s.pokemonRepo.GetPokemonSpeciesByID(pokemon.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Info: PokemonSpeciesService.UpdatePokemonSpecies) - Pokemon species with ID %d not found for update", pokemon.ID)
			return common.ErrPokemonSpeciesNotFound // Use specific not found error from common
		}
		log.Printf("(Error: PokemonSpeciesService.UpdatePokemonSpecies) - Error checking existing pokemon for update ID %d: %v", pokemon.ID, err)
		return fmt.Errorf("error checking existing pokemon for update: %w", err)
	}

	// Update the pokemon species
	err = s.pokemonRepo.UpdatePokemonSpecies(pokemon)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.UpdatePokemonSpecies) - Failed to update pokemon species ID %d: %v", pokemon.ID, err)
		return fmt.Errorf("failed to update pokemon species: %w", err)
	}

	return nil
}

// deletes a pokemon species record.
func (s *pokemonServiceImpl) DeletePokemonSpecies(id int64) error {
	if id <= 0 {
		log.Printf("(Error: PokemonSpeciesService.DeletePokemonSpecies) - Invalid input ID: %d", id)
		return common.ErrInvalidInput
	}

	// Check if the pokemon exists
	_, err := s.pokemonRepo.GetPokemonSpeciesByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("(Info: PokemonSpeciesService.DeletePokemonSpecies) - Pokemon species with ID %d not found for deletion", id)
			return common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Error: PokemonSpeciesService.DeletePokemonSpecies) - Error checking existing pokemon for deletion ID %d: %v", id, err)
		return fmt.Errorf("error checking existing pokemon for deletion: %w", err)
	}

	// Delete the pokemon species
	err = s.pokemonRepo.DeletePokemonSpecies(id)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.DeletePokemonSpecies) - Failed to delete pokemon species ID %d: %v", id, err)
		return fmt.Errorf("failed to delete pokemon species: %w", err)
	}
	return nil
}
