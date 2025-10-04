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
	ReleasePokemon(draftedPokemonID uuid.UUID) error
	// re-drafts a released Pokemon (from free agents)
	ReDraftPokemon(draftedPokemonID, newPlayerID uuid.UUID, newPickNumber int) error
	// gets count of Pokemon drafted by a player
	GetDraftedPokemonCountByPlayer(playerID uuid.UUID) (int64, error)
	// gets draft history for a league (all picks in order)
	GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error)
	// trades a Pokemon from one player to another
	TradePokemon(draftedPokemonID, newPlayerID uuid.UUID) error
	// performs a draft transaction (draft Pokemon and update league Pokemon availability)
	DraftPokemonTransaction(draftedPokemon *models.DraftedPokemon) error
	// soft deletes a drafted Pokemon entry
	DeleteDraftedPokemon(draftedPokemonID uuid.UUID) error
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
	err := r.db.Preload("PokemonSpecies").
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

// releases a Pokemon back to free agents
func (r *draftedPokemonRepositoryImpl) ReleasePokemon(draftedPokemonID uuid.UUID) error {
	err := r.db.Model(&models.DraftedPokemon{}).
		Where("id = ?", draftedPokemonID).
		Update("is_released", true).Error

	if err != nil {
		return fmt.Errorf("(Error: ReleasePokemon) - failed to release pokemon: %w", err)
	}
	return nil
}

// re-drafts a released Pokemon (from free agents)
func (r *draftedPokemonRepositoryImpl) ReDraftPokemon(draftedPokemonID, newPlayerID uuid.UUID, newPickNumber int) error {
	updates := map[string]interface{}{
		"player_id":         newPlayerID,
		"draft_pick_number": newPickNumber,
		"is_released":       false,
	}

	err := r.db.Model(&models.DraftedPokemon{}).
		Where("id = ?", draftedPokemonID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("(Error: ReDraftPokemon) - failed to re-draft pokemon: %w", err)
	}
	return nil
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

// GetDraftHistory gets draft history for a league (all picks in order)
func (r *draftedPokemonRepositoryImpl) GetDraftHistory(leagueID uuid.UUID) ([]models.DraftedPokemon, error) {
	var draftedPokemon []models.DraftedPokemon
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
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

// DraftPokemonTransaction performs a draft transaction (draft Pokemon and update league Pokemon availability)
func (r *draftedPokemonRepositoryImpl) DraftPokemonTransaction(draftedPokemon *models.DraftedPokemon) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: DraftPokemonTransaction) - failed to start transaction: %w", tx.Error)
	}

	// if fails at any point due to panic, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the drafted Pokemon entry
	if err := tx.Create(draftedPokemon).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonTransaction) - failed to create drafted pokemon: %w", err)
	}

	// Mark the Pokemon as unavailable in the league pool using LeaguePokemonID
	var leaguePokemon models.LeaguePokemon
	if err := tx.First(&leaguePokemon, "id = ?", draftedPokemon.LeaguePokemonID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonTransaction) - failed to get league pokemon: %w", err)
	}

	if err := tx.Model(&models.LeaguePokemon{}).
		Where("id = ?", draftedPokemon.LeaguePokemonID).
		Update("is_available", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonTransaction) - failed to mark pokemon unavailable: %w", err)
	}

	// Deduct DraftPoints from the player
	var player models.Player
	if err := tx.First(&player, "id = ?", draftedPokemon.PlayerID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonTransaction) - failed to get player: %w", err)
	}

	// Ensure player has enough points (this check should ideally also be done in service)
	if leaguePokemon.Cost == nil || player.DraftPoints < *leaguePokemon.Cost {
		tx.Rollback()
		// Return a specific error for insufficient points
		return fmt.Errorf("(Error: DraftPokemonTransaction) - insufficient draft points for player %s", player.ID)
	}

	player.DraftPoints -= *leaguePokemon.Cost
	if err := tx.Save(&player).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DraftPokemonTransaction) - failed to update player points: %w", err)
	}

	return tx.Commit().Error
}

// soft deletes a drafted Pokemon entry
func (r *draftedPokemonRepositoryImpl) DeleteDraftedPokemon(draftedPokemonID uuid.UUID) error {
	err := r.db.Delete(&models.DraftedPokemon{}, "id = ?", draftedPokemonID).Error
	if err != nil {
		return fmt.Errorf("(Error: DeleteDraftedPokemon) - failed to delete drafted pokemon: %w", err)
	}
	return nil
}
