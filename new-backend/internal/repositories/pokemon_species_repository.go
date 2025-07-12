package repositories

import (
	"fmt"
	"strings"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"gorm.io/gorm"
)

type PokemonSpeciesRepository interface {
	// retrieves all pokemon species from the database.
	GetAllPokemonSpecies() ([]models.PokemonSpecies, error)
	// retrieves a single pokemon species from the database by its ID.
	GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error)
	// retrieves a single pokemon species from the database by its exact name.
	GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error)
	// searches for pokemon species in the database by name.
	FindPokemonSpecies(filter string) ([]models.PokemonSpecies, error)
	// creates a new pokemon species record in the database.
	CreatePokemonSpecies(pokemon *models.PokemonSpecies) error
	// updates an existing pokemon species record in the database.
	UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error
	// deletes a pokemon species record from the database by its ID.
	DeletePokemonSpecies(id int64) error
}

type pokemonSpeciesRepositoryImpl struct {
	db *gorm.DB
}

// creates a new instance of PokemonSpeciesRepository.
func NewPokemonSpeciesRepository(db *gorm.DB) PokemonSpeciesRepository {
	return &pokemonSpeciesRepositoryImpl{
		db: db,
	}
}

// retrieves all pokemon species from the database.
func (r *pokemonSpeciesRepositoryImpl) GetAllPokemonSpecies() ([]models.PokemonSpecies, error) {
	var pokemon []models.PokemonSpecies
	result := r.db.Find(&pokemon)
	return pokemon, result.Error
}

// retrieves a single pokemon species from the database by its ID.
func (r *pokemonSpeciesRepositoryImpl) GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error) {
	var pokemon models.PokemonSpecies
	result := r.db.First(&pokemon, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pokemon species with ID %d not found", id)
		}
		return nil, result.Error
	}
	return &pokemon, nil
}

// searches for pokemon species in the database by name.
func (r *pokemonSpeciesRepositoryImpl) FindPokemonSpecies(filter string) ([]models.PokemonSpecies, error) {
	var pokemon []models.PokemonSpecies
	result := r.db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(filter)+"%").Find(&pokemon)
	return pokemon, result.Error
}

// retrieves a single pokemon species from the database by its exact name.
func (r *pokemonSpeciesRepositoryImpl) GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error) {
	var pokemon models.PokemonSpecies
	result := r.db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&pokemon)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pokemon species with name %s not found", name)
		}
		return nil, result.Error
	}
	return &pokemon, nil
}

// creates a new pokemon species record in the database.
func (r *pokemonSpeciesRepositoryImpl) CreatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	result := r.db.Create(pokemon)
	return result.Error
}

// updates an existing pokemon species record in the database.
func (r *pokemonSpeciesRepositoryImpl) UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	result := r.db.Save(pokemon)
	return result.Error
}

// deletes a pokemon species record from the database by its ID.
func (r *pokemonSpeciesRepositoryImpl) DeletePokemonSpecies(id int64) error {
	result := r.db.Delete(&models.PokemonSpecies{}, id)
	return result.Error
}
