package repositories

import (
	"fmt"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

// creates a new player in a league
func (r *PlayerRepository) CreatePlayer(player *models.Player) (*models.Player, error) {
	err := r.db.Create(player).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreatePlayer) - failed to create player: %w", err)
	}
	return player, nil
}

// gets player by ID with preloaded relationships
func (r *PlayerRepository) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("User").
		Preload("League").
		Preload("PlayerRoster").
		Preload("PlayerRoster.DraftedPokemon").
		Preload("PlayerRoster.DraftedPokemon.PokemonSpecies").
		First(&player, "id = ?", id).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPlayerByID) - failed to get player: %w", err)
	}
	return &player, nil
}

// gets player by user ID and league ID
func (r *PlayerRepository) GetPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("User").
		Preload("League").
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&player).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPlayerByUserAndLeague) - failed to get player: %w", err)
	}
	return &player, nil
}

// gets all players in a specific league
func (r *PlayerRepository) GetPlayersByLeague(leagueID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := r.db.Preload("User").
		Where("league_id = ?", leagueID).
		Order("draft_position ASC").
		Find(&players).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPlayersByLeague) - failed to get players: %w", err)
	}
	return players, nil
}

// gets all players for a specific user across all leagues
func (r *PlayerRepository) GetPlayersByUser(userID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := r.db.Preload("League").
		Where("user_id = ?", userID).
		Find(&players).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPlayersByUser) - failed to get players: %w", err)
	}
	return players, nil
}

// updates player information
func (r *PlayerRepository) UpdatePlayer(player *models.Player) (*models.Player, error) {
	err := r.db.Select(
		"in_league_name", "team_name", "wins", "losses", "draft_points",
		"draft_position", "updated_at",
	).Updates(player).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: UpdatePlayer) - failed to update player: %w", err)
	}

	return r.GetPlayerByID(player.ID)
}

// updates player's draft points (used during drafting)
func (r *PlayerRepository) UpdatePlayerDraftPoints(playerID uuid.UUID, newPoints int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("draft_points", newPoints).Error

	if err != nil {
		return fmt.Errorf("(Error: UpdatePlayerDraftPoints) - failed to update draft points: %w", err)
	}
	return nil
}

// updates player's win/loss record
func (r *PlayerRepository) UpdatePlayerRecord(playerID uuid.UUID, wins, losses int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Updates(map[string]any{
			"wins":   wins,
			"losses": losses,
		}).Error

	if err != nil {
		return fmt.Errorf("(Error: UpdatePlayerRecord) - failed to update player record: %w", err)
	}
	return nil
}

// gets player count for a specific league
func (r *PlayerRepository) GetPlayerCountByLeague(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Player{}).
		Where("league_id = ?", leagueID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("(Error: GetPlayerCountByLeague) - failed to count players: %w", err)
	}
	return count, nil
}

// soft deletes a player from a league
func (r *PlayerRepository) DeletePlayer(playerID uuid.UUID) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: DeletePlayer) - failed to start transaction: %w", tx.Error)
	}

	// if fails at any point due to panic, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First soft delete all roster entries for this player
	if err := tx.Where("player_id = ?", playerID).Delete(&models.PlayerRoster{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DeletePlayer) - failed to delete player roster: %w", err)
	}

	// Then soft delete the player
	if err := tx.Delete(&models.Player{}, "id = ?", playerID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DeletePlayer) - failed to delete player: %w", err)
	}

	return tx.Commit().Error
}

// checks if a user is already a player in a specific league
func (r *PlayerRepository) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Player{}).
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("(Error: IsUserInLeague) - failed to check player membership: %w", err)
	}
	return count > 0, nil
}

// gets player with full roster details
func (r *PlayerRepository) GetPlayerWithFullRoster(playerID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("User").
		Preload("League").
		Preload("Roster").
		Preload("Roster.DraftedPokemon").
		Preload("Roster.DraftedPokemon.PokemonSpecies").
		First(&player, "id = ?", playerID).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPlayerWithFullRoster) - failed to get player with roster: %w", err)
	}
	return &player, nil
}
