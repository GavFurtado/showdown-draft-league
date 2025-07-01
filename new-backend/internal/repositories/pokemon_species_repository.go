package repositories

import (
	"fmt"
	"strings"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"gorm.io/gorm"
)

// handles operations related to PokemonSpecies.
type PokemonSpeciesRepository struct {
	db *gorm.DB
}

// creates a new instance of PokemonSpeciesRepository.
func NewPokemonSpeciesRepository(db *gorm.DB) *PokemonSpeciesRepository {
	return &PokemonSpeciesRepository{
		db: db,
	}
}

// retrieves all pokemon species from the database.
func (r *PokemonSpeciesRepository) GetAllPokemonSpecies() ([]models.PokemonSpecies, error) {
	var pokemon []models.PokemonSpecies
	result := r.db.Find(&pokemon)
	return pokemon, result.Error
}

// searches for pokemon species in the database by name.
func (r *PokemonSpeciesRepository) FindPokemonSpecies(filter string) ([]models.PokemonSpecies, error) {
	var pokemon []models.PokemonSpecies
	result := r.db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(filter)+"%").Find(&pokemon)
	return pokemon, result.Error
}

// retrieves a single pokemon species from the database by its ID.
func (r *PokemonSpeciesRepository) GetPokemonSpeciesByID(id int) (*models.PokemonSpecies, error) {
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

// retrieves a single pokemon species from the database by its exact name.
func (r *PokemonSpeciesRepository) GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error) {
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
func (r *PokemonSpeciesRepository) CreatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	result := r.db.Create(pokemon)
	return result.Error
}

// updates an existing pokemon species record in the database.
func (r *PokemonSpeciesRepository) UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	result := r.db.Save(pokemon)
	return result.Error
}

// deletes a pokemon species record from the database by its ID.
func (r *PokemonSpeciesRepository) DeletePokemonSpecies(id int) error {
	result := r.db.Delete(&models.PokemonSpecies{}, id)
	return result.Error
}
