package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerRepository interface {
	// Transactional methods
	Begin() *gorm.DB
	WithTx(tx *gorm.DB) PlayerRepository

	CreatePlayer(player *models.Player) (*models.Player, error)
	// gets player by ID with preloaded relationships
	GetPlayerByID(id uuid.UUID) (*models.Player, error)
	// gets player by user ID and league ID
	GetPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error)
	// gets all players in a specific league
	GetPlayersByLeague(leagueID uuid.UUID) ([]models.Player, error)
	// gets all players for a specific user across all leagues
	GetPlayersByUser(userID uuid.UUID) ([]models.Player, error)
	UpdatePlayer(player *models.Player) (*models.Player, error)
	// updates player's draft points (used during drafting)
	UpdatePlayerDraftPoints(playerID uuid.UUID, newPoints int) error
	// updates player's win/loss record
	UpdatePlayerRecord(playerID uuid.UUID, wins, losses int) error
	UpdatePlayerDraftPosition(playerID uuid.UUID, newPosition int) error
	UpdatePlayerRole(playerID uuid.UUID, playerRole rbac.PlayerRole) error
	GetPlayerCountByLeague(leagueID uuid.UUID) (int64, error)
	// soft deletes a player from a league
	DeletePlayer(playerID uuid.UUID) error
	// Helper: checks if a user is already a player in a specific league
	IsUserInLeague(userID, leagueID uuid.UUID) (bool, error)
	GetPlayerWithFullRoster(playerID uuid.UUID) (*models.Player, error)
	// finds a player by user ID and league ID.
	FindPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error)
	FindPlayerByInLeagueNameAndLeagueID(inLeagueName string, leagueID uuid.UUID) (*models.Player, error)
	FindPlayerByTeamNameAndLeagueID(teamName string, leagueID uuid.UUID) (*models.Player, error)
}

type playerRepositoryImpl struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepositoryImpl{
		db: db,
	}
}

// Begin starts a new transaction.
func (r *playerRepositoryImpl) Begin() *gorm.DB {
	return r.db.Begin()
}

// WithTx returns a new repository instance with the given transaction.
func (r *playerRepositoryImpl) WithTx(tx *gorm.DB) PlayerRepository {
	return &playerRepositoryImpl{db: tx}
}

// creates a new player in a league
func (r *playerRepositoryImpl) CreatePlayer(player *models.Player) (*models.Player, error) {
	err := r.db.Create(player).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreatePlayer) - failed to create player: %w", err)
	}
	return player, nil
}

// gets player by ID with preloaded relationships
func (r *playerRepositoryImpl) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("User").
		Preload("League").
		Preload("Roster").
		Preload("Roster.DraftedPokemon").
		Preload("Roster.DraftedPokemon.PokemonSpecies").
		First(&player, "id = ?", id).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPlayerByID) - failed to get player: %w", err)
	}
	return &player, nil
}

// gets player by user ID and league ID
func (r *playerRepositoryImpl) GetPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("User").
		Preload("League").
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&player).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return &player, nil
}

// gets all players in a specific league
func (r *playerRepositoryImpl) GetPlayersByLeague(leagueID uuid.UUID) ([]models.Player, error) {
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
func (r *playerRepositoryImpl) GetPlayersByUser(userID uuid.UUID) ([]models.Player, error) {
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
func (r *playerRepositoryImpl) UpdatePlayer(player *models.Player) (*models.Player, error) {
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
func (r *playerRepositoryImpl) UpdatePlayerDraftPoints(playerID uuid.UUID, newPoints int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("draft_points", newPoints).Error

	if err != nil {
		return fmt.Errorf("(Error: UpdatePlayerDraftPoints) - failed to update draft points: %w", err)
	}
	return nil
}

// updates player's win/loss record
func (r *playerRepositoryImpl) UpdatePlayerRecord(playerID uuid.UUID, wins, losses int) error {
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

func (r *playerRepositoryImpl) UpdatePlayerDraftPosition(playerID uuid.UUID, newPosition int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("draft_position", newPosition).Error

	if err != nil {
		return fmt.Errorf("(Error: UpdatePlayerDraftPosition) - failed to update draft position: %w", err)
	}
	return nil
}

func (r *playerRepositoryImpl) UpdatePlayerRole(playerID uuid.UUID, playerRole rbac.PlayerRole) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("role", playerRole).Error

	if err != nil {
		return fmt.Errorf("(Repository: UpdatePlayerRole) - failed to update player role: %w", err)
	}

	return nil
}

// gets player count for a specific league
func (r *playerRepositoryImpl) GetPlayerCountByLeague(leagueID uuid.UUID) (int64, error) {
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
func (r *playerRepositoryImpl) DeletePlayer(playerID uuid.UUID) error {
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
func (r *playerRepositoryImpl) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
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
func (r *playerRepositoryImpl) GetPlayerWithFullRoster(playerID uuid.UUID) (*models.Player, error) {

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

// finds a player by user ID and league ID.
// Returns (player, nil) if found, (nil, nil) if not found, (nil, error) for other DB errors.
func (r *playerRepositoryImpl) FindPlayerByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&player).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find player by user ID (%s) and league ID (%s): %w", userID, leagueID, err)
	}
	return &player, nil
}

// finds a player by their in-league name and league ID.
// Returns (player, nil) if found, (nil, nil) if not found, (nil, error) for other DB errors.
func (r *playerRepositoryImpl) FindPlayerByInLeagueNameAndLeagueID(inLeagueName string, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("in_league_name = ? AND league_id = ?", inLeagueName, leagueID).
		First(&player).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find player by in-league name (%s) and league ID (%s): %w", inLeagueName, leagueID, err)
	}
	return &player, nil
}

// finds a player by their team name and league ID.
// Returns (player, nil) if found, (nil, nil) if not found, (nil, error) for other DB errors.
func (r *playerRepositoryImpl) FindPlayerByTeamNameAndLeagueID(teamName string, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("team_name = ? AND league_id = ?", teamName, leagueID).
		First(&player).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find player by team name (%s) and league ID (%s): %w", teamName, leagueID, err)
	}
	return &player, nil
}
