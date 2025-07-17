package repositories

import (
	"fmt"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaguePokemonRepository interface {
	// adds a Pokemon species to a league's draft pool
	CreateLeaguePokemon(leaguePokemon *models.LeaguePokemon) (*models.LeaguePokemon, error)
	// adds multiple Pokemon species to a league's draft pool in a transaction
	CreateLeaguePokemonBatch(leaguePokemon []models.LeaguePokemon) error
	// gets all Pokemon in a league's draft pool (available and unavailable)
	GetAllPokemonByLeague(leagueID uuid.UUID) ([]models.LeaguePokemon, error)
	// gets all available Pokemon for a specific league's draft pool
	GetAvailablePokemonByLeague(leagueID uuid.UUID) ([]models.LeaguePokemon, error)
	// gets a specific Pokemon from a league's draft pool by its ID
	GetLeaguePokemonByID(leaguePokemonID uuid.UUID) (*models.LeaguePokemon, error)
	// updates a Pokemon's availability or cost in a league
	UpdateLeaguePokemon(leaguePokemon *models.LeaguePokemon) (*models.LeaguePokemon, error)
	// marks a Pokemon as unavailable (typically after being drafted)
	MarkPokemonUnavailable(leagueID, pokemonSpeciesID uuid.UUID) error
	// gets Pokemon filtered by cost range
	GetPokemonByCostRange(leagueID uuid.UUID, minCost, maxCost int) ([]models.LeaguePokemon, error)
	// checks if a Pokemon species is available in a league's draft pool
	IsPokemonAvailable(leagueID, pokemonSpeciesID uuid.UUID) (bool, error)
	// gets the cost of a specific Pokemon in a league
	GetPokemonCost(leagueID, pokemonSpeciesID uuid.UUID) (*int, error)
	// removes a Pokemon species from a league's draft pool (soft delete)
	DeleteLeaguePokemon(leagueID, pokemonSpeciesID uuid.UUID) error
	// gets count of available Pokemon in a league
	GetAvailablePokemonCount(leagueID uuid.UUID) (int64, error)
	// removes all Pokemon from a league's draft pool (used when deleting a league)
	DeleteAllLeaguePokemon(leagueID uuid.UUID) error
}

type leaguePokemonRepositoryImpl struct {
	db *gorm.DB
}

func NewLeaguePokemonRepository(db *gorm.DB) LeaguePokemonRepository {
	return &leaguePokemonRepositoryImpl{
		db: db,
	}
}

// adds a Pokemon species to a league's draft pool
func (r *leaguePokemonRepositoryImpl) CreateLeaguePokemon(leaguePokemon *models.LeaguePokemon) (*models.LeaguePokemon, error) {
	err := r.db.Create(leaguePokemon).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreateLeaguePokemon) - failed to create league pokemon: %w", err)
	}
	return leaguePokemon, nil
}

// adds multiple Pokemon species to a league's draft pool in a transaction
func (r *leaguePokemonRepositoryImpl) CreateLeaguePokemonBatch(leaguePokemon []models.LeaguePokemon) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: CreateLeaguePokemonBatch) - failed to start transaction: %w", tx.Error)
	}

	// if fails at any point due to panic, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create all league pokemon entries in batch
	if err := tx.CreateInBatches(leaguePokemon, 100).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: CreateLeaguePokemonBatch) - failed to create league pokemon batch: %w", err)
	}

	return tx.Commit().Error
}

// gets all available Pokemon for a specific league's draft pool
func (r *leaguePokemonRepositoryImpl) GetAvailablePokemonByLeague(leagueID uuid.UUID) ([]models.LeaguePokemon, error) {
	var leaguePokemon []models.LeaguePokemon
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND is_available = ?", leagueID, true).
		Order("cost DESC, created_at ASC").
		Find(&leaguePokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetAvailablePokemonByLeague) - failed to get available pokemon: %w", err)
	}
	return leaguePokemon, nil
}

// gets all Pokemon in a league's draft pool (available and unavailable)
func (r *leaguePokemonRepositoryImpl) GetAllPokemonByLeague(leagueID uuid.UUID) ([]models.LeaguePokemon, error) {
	var leaguePokemon []models.LeaguePokemon
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ?", leagueID).
		Order("cost ASC, pokemon_species_id ASC").
		Find(&leaguePokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetAllPokemonByLeague) - failed to get league pokemon: %w", err)
	}
	return leaguePokemon, nil
}

// gets a specific Pokemon from a league's draft pool
func (r *leaguePokemonRepositoryImpl) GetLeaguePokemonBySpecies(leagueID, pokemonSpeciesID uuid.UUID) (*models.LeaguePokemon, error) {
	var leaguePokemon models.LeaguePokemon
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND pokemon_species_id = ?", leagueID, pokemonSpeciesID).
		First(&leaguePokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetLeaguePokemonBySpecies) - failed to get league pokemon: %w", err)
	}
	return &leaguePokemon, nil
}

// gets a specific Pokemon from a league's draft pool by its ID
func (r *leaguePokemonRepositoryImpl) GetLeaguePokemonByID(leaguePokemonID uuid.UUID) (*models.LeaguePokemon, error) {
	var leaguePokemon models.LeaguePokemon
	err := r.db.Preload("League").
		Preload("PokemonSpecies").
		First(&leaguePokemon, "id = ?", leaguePokemonID).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetLeaguePokemonByID) - failed to get league pokemon by ID: %w", err)
	}
	return &leaguePokemon, nil
}

// updates a Pokemon's availability or cost in a league
func (r *leaguePokemonRepositoryImpl) UpdateLeaguePokemon(leaguePokemon *models.LeaguePokemon) (*models.LeaguePokemon, error) {
	err := r.db.Select("cost", "is_available", "updated_at").
		Updates(leaguePokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: UpdateLeaguePokemon) - failed to update league pokemon: %w", err)
	}
	return leaguePokemon, nil
}

// marks a Pokemon as unavailable (typically after being drafted)
func (r *leaguePokemonRepositoryImpl) MarkPokemonUnavailable(leagueID, pokemonSpeciesID uuid.UUID) error {
	err := r.db.Model(&models.LeaguePokemon{}).
		Where("league_id = ? AND pokemon_species_id = ?", leagueID, pokemonSpeciesID).
		Update("is_available", false).Error

	if err != nil {
		return fmt.Errorf("(Error: MarkPokemonUnavailable) - failed to mark pokemon unavailable: %w", err)
	}
	return nil
}

// marks a Pokemon as available (typically when released back to free agents)
func (r *leaguePokemonRepositoryImpl) MarkPokemonAvailable(leagueID, pokemonSpeciesID uuid.UUID) error {
	err := r.db.Model(&models.LeaguePokemon{}).
		Where("league_id = ? AND pokemon_species_id = ?", leagueID, pokemonSpeciesID).
		Update("is_available", true).Error

	if err != nil {
		return fmt.Errorf("(Error: MarkPokemonAvailable) - failed to mark pokemon available: %w", err)
	}
	return nil
}

// gets Pokemon filtered by cost range
func (r *leaguePokemonRepositoryImpl) GetPokemonByCostRange(leagueID uuid.UUID, minCost, maxCost int) ([]models.LeaguePokemon, error) {
	var leaguePokemon []models.LeaguePokemon
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND cost BETWEEN ? AND ? AND is_available = ?", leagueID, minCost, maxCost, true).
		Order("cost ASC").
		Find(&leaguePokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPokemonByCostRange) - failed to get pokemon by cost range: %w", err)
	}
	return leaguePokemon, nil
}

// checks if a Pokemon species is available in a league's draft pool
func (r *leaguePokemonRepositoryImpl) IsPokemonAvailable(leagueID, pokemonSpeciesID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.LeaguePokemon{}).
		Where("league_id = ? AND pokemon_species_id = ? AND is_available = ?", leagueID, pokemonSpeciesID, true).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("(Error: IsPokemonAvailable) - failed to check pokemon availability: %w", err)
	}
	return count > 0, nil
}

// gets the cost of a specific Pokemon in a league
func (r *leaguePokemonRepositoryImpl) GetPokemonCost(leagueID, pokemonSpeciesID uuid.UUID) (*int, error) {
	var leaguePokemon models.LeaguePokemon
	err := r.db.Select("cost").
		Where("league_id = ? AND pokemon_species_id = ?", leagueID, pokemonSpeciesID).
		First(&leaguePokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPokemonCost) - failed to get pokemon cost: %w", err)
	}
	return leaguePokemon.Cost, nil
}

// removes a Pokemon species from a league's draft pool (soft delete)
func (r *leaguePokemonRepositoryImpl) DeleteLeaguePokemon(leagueID, pokemonSpeciesID uuid.UUID) error {
	err := r.db.Where("league_id = ? AND pokemon_species_id = ?", leagueID, pokemonSpeciesID).
		Delete(&models.LeaguePokemon{}).Error

	if err != nil {
		return fmt.Errorf("(Error: DeleteLeaguePokemon) - failed to delete league pokemon: %w", err)
	}
	return nil
}

// gets count of available Pokemon in a league
func (r *leaguePokemonRepositoryImpl) GetAvailablePokemonCount(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.LeaguePokemon{}).
		Where("league_id = ? AND is_available = ?", leagueID, true).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("(Error: GetAvailablePokemonCount) - failed to count available pokemon: %w", err)
	}
	return count, nil
}

// removes all Pokemon from a league's draft pool (used when deleting a league)
func (r *leaguePokemonRepositoryImpl) DeleteAllLeaguePokemon(leagueID uuid.UUID) error {
	err := r.db.Where("league_id = ?", leagueID).Delete(&models.LeaguePokemon{}).Error
	if err != nil {
		return fmt.Errorf("(Error: DeleteAllLeaguePokemon) - failed to delete all league pokemon: %w", err)
	}
	return nil
}
