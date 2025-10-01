package services

import (
	"errors"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"gorm.io/gorm"
)

type PokemonSpeciesService interface {
	GetAllPokemonSpecies() ([]common.PokemonSpeciesListDTO, error) // Updated return type
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
func (s *pokemonServiceImpl) GetAllPokemonSpecies() ([]common.PokemonSpeciesListDTO, error) { // Updated return type
	allPokemon, err := s.pokemonRepo.GetAllPokemonSpecies()
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.GetAllPokemonSpecies) - Failed to get all pokemon species: %v", err)
		return nil, common.ErrInternalService
	}

	var pokemonDTOs []common.PokemonSpeciesListDTO
	for _, pokemon := range allPokemon {
		primaryType := ""
		if len(pokemon.Types) > 0 {
			primaryType = pokemon.Types[0]
		}
		pokemonDTOs = append(pokemonDTOs, common.PokemonSpeciesListDTO{
			ID:           pokemon.ID,
			Name:         pokemon.Name,
			PrimaryType:  primaryType,
			FrontDefault: pokemon.Sprites.FrontDefault,
		})
	}

	log.Println("Success")
	return pokemonDTOs, nil
}

// lists pokemon species based on a filter.
func (s *pokemonServiceImpl) ListPokemonSpecies(filter string) ([]models.PokemonSpecies, error) {
	pokemon, err := s.pokemonRepo.FindPokemonSpecies(filter)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.ListPokemonSpecies) - Failed to find pokemon species with filter '%s': %v", filter, err)
		return nil, common.ErrInternalService
	}

	log.Println("Success")
	return pokemon, nil
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
			log.Printf("(Error: PokemonSpeciesService.GetPokemonSpeciesByID) - Pokemon species with ID %d not found", id)
			return nil, common.ErrPokemonSpeciesNotFound
		}
		log.Printf("(Error: PokemonSpeciesService.GetPokemonSpeciesByID) - Failed to get pokemon species by ID %d: %v", id, err)
		return nil, common.ErrInternalService
	}

	log.Println("Success")
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
		return nil, common.ErrInternalService
	}

	log.Println("Success")
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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Error checking existing pokemon by ID %d: %v", pokemon.ID, err)
		return common.ErrInternalService
	}

	if existingByID != nil {
		log.Printf("(Info: PokemonSpeciesService.CreatePokemonSpecies) - Pokemon species with ID %d already exists", pokemon.ID)
		return common.ErrConflict
	}

	existingByName, err := s.pokemonRepo.GetPokemonSpeciesByName(pokemon.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Error checking existing pokemon by name '%s': %v", pokemon.Name, err)
		return common.ErrInternalService
	}
	if existingByName != nil {
		log.Printf("(Info: PokemonSpeciesService.CreatePokemonSpecies) - Pokemon species with name '%s' already exists", pokemon.Name)
		return common.ErrConflict
	}

	// Create the pokemon species
	err = s.pokemonRepo.CreatePokemonSpecies(pokemon)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.CreatePokemonSpecies) - Failed to create pokemon species ID %d name '%s': %v", pokemon.ID, pokemon.Name, err)
		return common.ErrInternalService
	}

	log.Println("Success")
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
		return common.ErrInternalService
	}

	// Update the pokemon species
	err = s.pokemonRepo.UpdatePokemonSpecies(pokemon)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.UpdatePokemonSpecies) - Failed to update pokemon species ID %d: %v", pokemon.ID, err)
		return common.ErrInternalService
	}

	log.Println("Success")
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
		return common.ErrInternalService
	}

	// Delete the pokemon species
	err = s.pokemonRepo.DeletePokemonSpecies(id)
	if err != nil {
		log.Printf("(Error: PokemonSpeciesService.DeletePokemonSpecies) - Failed to delete pokemon species ID %d: %v", id, err)
		return common.ErrInternalService
	}
	return nil
}
