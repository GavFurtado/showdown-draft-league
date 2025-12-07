package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DraftedPokemonRepository interface {
	// creates a new drafted Pokemon entry
	CreateDraftedPokemon(draftedPokemon *models.DraftedPokemon) (*models.DraftedPokemon, error)
	// gets drafted Pokemon by ID with relationships
	GetDraftedPokemonByID(id uuid.UUID) (*models.DraftedPokemon, error)
	// gets all Pokemon drafted by a specific player in a league
	GetDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all Pokemon drafted in a specific league
	GetDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all active (non-released) Pokemon drafted in a league
	GetActiveDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets all released Pokemon (free agents) in a league
	GetReleasedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// checks if a Pokemon species has been drafted in a league
	IsPokemonDrafted(leagueID uuid.UUID, pokemonSpeciesID int64) (bool, error)
	// gets the next draft pick number for a league
	GetNextDraftPickNumber(leagueID uuid.UUID) (int, error)
	// releases a Pokemon back to free agents
	ReleasePokemonTransaction(draftedPokemonID uuid.UUID, player *models.Player, dropCost int, releasedWeek int) error
	// gets count of active Pokemon drafted by a player
	GetDraftedPokemonCountByPlayer(playerID uuid.UUID) (int64, error)
	// gets all Pokemon drafted by a specific player (including released)
	GetAllDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error)
	// gets the actively drafted pokemon count by league
	GetActiveDraftedPokemonCountByLeague(leagueID uuid.UUID) (int64, error)
	// gets draft history for a league (all picks in order)
	GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// UNUSED trades a Pokemon from one player to another
	TradePokemon(draftedPokemonID, newPlayerID uuid.UUID) error
	// soft deletes a drafted Pokemon entry
	DeleteDraftedPokemon(draftedPokemonID uuid.UUID) error
	// performs a batch draft transaction (draft multiple Pokemon, update player points, and mark league Pokemon unavailable)
	DraftPokemonBatchTransaction(draftedPokemon []*models.DraftedPokemon, player *models.Player, leaguePokemonIDs []uuid.UUID, totalCost int) error
	// performs a transaction to pick up a free agent
	PickupFreeAgentTransaction(player *models.Player, newDraftedPokemon *models.DraftedPokemon, leaguePokemon *models.LeaguePokemon, pickupCost int) error
}

type draftedPokemonRepositoryImpl struct {
	db *gorm.DB
}

func NewDraftedPokemonRepository(db *gorm.DB) *draftedPokemonRepositoryImpl {
	return &draftedPokemonRepositoryImpl{db: db}
}

// creates a new drafted Pokemon entry
func (r *draftedPokemonRepositoryImpl) CreateDraftedPokemon(draftedPokemon *models.DraftedPokemon) (*models.DraftedPokemon, error) {
	err := r.db.Create(draftedPokemon).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreateDraftedPokemon) - failed to create drafted pokemon: %w", err)
	}
	return draftedPokemon, nil
}

// gets drafted Pokemon by ID with relationships
func (r *draftedPokemonRepositoryImpl) GetDraftedPokemonByID(id uuid.UUID) (*models.DraftedPokemon, error) {
	var draftedPokemon models.DraftedPokemon
	err := r.db.Preload("League").
		Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		First(&draftedPokemon, "id = ?", id).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftedPokemonByID) - failed to get drafted pokemon: %w", err)
	}
	return &draftedPokemon, nil
}

// gets all Pokemon drafted by a specific player in a league
func (r *draftedPokemonRepositoryImpl) GetDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.
		Preload("PokemonSpecies").
		Preload("LeaguePokemon").
		Where("player_id = ? AND is_released = ?", playerID, false).
		Order("draft_pick_number ASC").
		Find(&draftedPokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftedPokemonByPlayer) - failed to get drafted pokemon by player: %w", err)
	}
	return draftedPokemon, nil
}

// gets all Pokemon drafted in a specific league
func (r *draftedPokemonRepositoryImpl) GetDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ?", leagueID).
		Order("draft_pick_number ASC").
		Find(&draftedPokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftedPokemonByLeague) - failed to get drafted pokemon by league: %w", err)
	}
	return draftedPokemon, nil
}

// gets all active (non-released) Pokemon drafted in a league
func (r *draftedPokemonRepositoryImpl) GetActiveDraftedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ? AND is_released = ?", leagueID, false).
		Order("draft_pick_number ASC").
		Find(&draftedPokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetActiveDraftedPokemonByLeague) - failed to get active drafted pokemon: %w", err)
	}
	return draftedPokemon, nil
}

// gets all released Pokemon (free agents) in a league
func (r *draftedPokemonRepositoryImpl) GetReleasedPokemonByLeague(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ? AND is_released = ?", leagueID, true).
		Order("updated_at DESC"). // Most recently released first
		Find(&draftedPokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetReleasedPokemonByLeague) - failed to get released pokemon: %w", err)
	}
	return draftedPokemon, nil
}

// checks if a Pokemon species has been drafted in a league
func (r *draftedPokemonRepositoryImpl) IsPokemonDrafted(leagueID uuid.UUID, pokemonSpeciesID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.DraftedPokemon{}).
		Where("league_id = ? AND pokemon_species_id = ? AND is_released = ?", leagueID, pokemonSpeciesID, false).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("(Error: IsPokemonDrafted) - failed to check if pokemon is drafted: %w", err)
	}
	return count > 0, nil
}

// gets the next draft pick number for a league
func (r *draftedPokemonRepositoryImpl) GetNextDraftPickNumber(leagueID uuid.UUID) (int, error) {
	var maxPickNumber int
	err := r.db.Model(&models.DraftedPokemon{}).
		Select("COALESCE(MAX(draft_pick_number), 0)").
		Where("league_id = ?", leagueID).
		Scan(&maxPickNumber).Error

	if err != nil {
		return 0, fmt.Errorf("(Error: GetNextDraftPickNumber) - failed to get next draft pick number: %w", err)
	}
	return maxPickNumber + 1, nil
}

// ReleasePokemonTransaction releases a Pokemon back to free agents in a transaction
func (r *draftedPokemonRepositoryImpl) ReleasePokemonTransaction(draftedPokemonID uuid.UUID, player *models.Player, dropCost int, releasedWeek int) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: ReleasePokemonTransaction) - failed to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get the drafted pokemon to find the corresponding league_pokemon_id
	var draftedPokemon models.DraftedPokemon
	if err := tx.First(&draftedPokemon, "id = ?", draftedPokemonID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("(Error: ReleasePokemonTransaction) - drafted pokemon not found: %w", err)
		}
		return fmt.Errorf("(Error: ReleasePokemonTransaction) - failed to get drafted pokemon: %w", err)
	}

	// 2. Update the drafted pokemon to be released
	if err := tx.Model(&draftedPokemon).Updates(map[string]interface{}{"is_released": true, "released_week": releasedWeek}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: ReleasePokemonTransaction) - failed to release pokemon: %w", err)
	}

	// 3. Update the corresponding league pokemon to be available again
	if draftedPokemon.LeaguePokemonID != uuid.Nil {
		if err := tx.Model(&models.LeaguePokemon{}).Where("id = ?", draftedPokemon.LeaguePokemonID).Update("is_available", true).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("(Error: ReleasePokemonTransaction) - failed to update league pokemon availability: %w", err)
		}
	}

	// 4. Decrement player's TransferCredits
	player.TransferCredits -= dropCost
	if err := tx.Save(player).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: ReleasePokemonTransaction) - failed to update player transfer credits: %w", err)
	}

	return tx.Commit().Error
}

// GetDraftedPokemonCountByPlayer gets count of Pokemon drafted by a player
func (r *draftedPokemonRepositoryImpl) GetDraftedPokemonCountByPlayer(playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.DraftedPokemon{}).
		Where("player_id = ? AND is_released = ?", playerID, false).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("(Error: GetDraftedPokemonCountByPlayer) - failed to count drafted pokemon: %w", err)
	}
	return count, nil
}

// GetAllDraftedPokemonByPlayer gets all Pokemon drafted by a specific player (including released).
func (r *draftedPokemonRepositoryImpl) GetAllDraftedPokemonByPlayer(playerID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.
		Preload("PokemonSpecies").
		Preload("LeaguePokemon").
		Where("player_id = ?", playerID).
		Order("acquired_week ASC, released_week ASC").
		Find(&draftedPokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetAllDraftedPokemonByPlayer) - failed to get all drafted pokemon by player: %w", err)
	}
	return draftedPokemon, nil
}

// GetDraftHistory gets draft history for a league (all picks in order)
func (r *draftedPokemonRepositoryImpl) GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.Preload("Player").
		Preload("PokemonSpecies").
		Preload("LeaguePokemon").
		Where("league_id = ?", leagueID).
		Order("draft_pick_number ASC").
		Find(&draftedPokemon).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftHistory) - failed to get draft history: %w", err)
	}
	return draftedPokemon, nil
}

// TradePokemon trades a Pokemon from one player to another
func (r *draftedPokemonRepositoryImpl) TradePokemon(draftedPokemonID, newPlayerID uuid.UUID) error {
	err := r.db.Model(&models.DraftedPokemon{}).
		Where("id = ?", draftedPokemonID).
		Update("player_id", newPlayerID).Error

	if err != nil {
		return fmt.Errorf("(Error: TradePokemon) - failed to trade pokemon: %w", err)
	}
	return nil
}

// soft deletes a drafted Pokemon entry
func (r *draftedPokemonRepositoryImpl) DeleteDraftedPokemon(draftedPokemonID uuid.UUID) error {
	err := r.db.Delete(&models.DraftedPokemon{}, "id = ?", draftedPokemonID).Error
	if err != nil {
		return fmt.Errorf("(Error: DeleteDraftedPokemon) - failed to delete drafted pokemon: %w", err)
	}
	return nil
}

// DraftPokemonBatchTransaction performs a batch draft transaction (draft multiple Pokemon, update player points, and mark league Pokemon unavailable)
func (r *draftedPokemonRepositoryImpl) DraftPokemonBatchTransaction(draftedPokemon []*models.DraftedPokemon, player *models.Player, leaguePokemonIDs []uuid.UUID, totalCost int) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: DraftPokemonBatchTransaction) - failed to start transaction: %w", tx.Error)

	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Batch create the drafted Pokemon entries
	var draftedPokemonModels []models.DraftedPokemon
	for _, dp := range draftedPokemon {
		draftedPokemonModels = append(draftedPokemonModels, *dp)
	}

	if err := tx.CreateInBatches(draftedPokemonModels, 100).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonBatchTransaction) - failed to create drafted pokemon batch: %w", err)
	}

	// 2. Mark the Pokemon as unavailable in the league pool
	if err := tx.Model(&models.LeaguePokemon{}).
		Where("id IN (?) ", leaguePokemonIDs).
		Update("is_available", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonBatchTransaction) - failed to mark pokemon unavailable in batch: %w", err)
	}

	// 3. Deduct DraftPoints from the player
	player.DraftPoints -= totalCost
	if err := tx.Save(player).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonBatchTransaction) - failed to update player points: %w", err)
	}

	return tx.Commit().Error
}

// PickupFreeAgentTransaction performs a transaction to pick up a free agent
func (r *draftedPokemonRepositoryImpl) PickupFreeAgentTransaction(player *models.Player, newDraftedPokemon *models.DraftedPokemon, leaguePokemon *models.LeaguePokemon, pickupCost int) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Decrement player's TransferCredits
	player.TransferCredits -= pickupCost
	if err := tx.Save(player).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to update player transfer credits: %w", err)
	}

	// 2. Create new DraftedPokemon entry
	if err := tx.Create(newDraftedPokemon).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to create new drafted pokemon: %w", err)
	}

	// 3. Mark LeaguePokemon as unavailable
	leaguePokemon.IsAvailable = false
	if err := tx.Save(leaguePokemon).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to update league pokemon availability: %w", err)
	}

	return tx.Commit().Error
}

// GetActiveDraftedPokemonCountByLeague gets count of all active (non-released) Pokemon drafted in a league
func (r *draftedPokemonRepositoryImpl) GetActiveDraftedPokemonCountByLeague(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.DraftedPokemon{}).
		Where("league_id = ? AND is_released = ?", leagueID, false).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("(Error: GetActiveDraftedPokemonCountByLeague) - failed to count active drafted pokemon: %w", err)
	}
	return count, nil
}
