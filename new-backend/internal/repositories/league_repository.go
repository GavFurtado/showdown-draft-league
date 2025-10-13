package repositories

import (
	"errors"
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac" // Import the new rbac package
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeagueRepository interface {
	// create a new league
	CreateLeague(league *models.League) (*models.League, error)
	// checks if a given user is a player in a specific league.
	IsUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error)
	// gets League by League ID with relationships preloaded
	GetLeagueByID(leagueID uuid.UUID) (*models.League, error)
	// gets a league (by id) with all its related data
	GetLeagueWithFullDetails(id uuid.UUID) (*models.League, error)
	// get all leagues where the user's player is owner
	GetLeaguesByOwner(userID uuid.UUID) ([]models.League, error)
	// gets total count of Leagues where userID's player is the owner
	GetLeaguesCountWhereOwner(userID uuid.UUID) (int64, error)
	// fetches all Leagues where the given userID is a player.
	GetLeaguesByUser(userID uuid.UUID) ([]models.League, error)
	// updates a league (name, start_date, ruleset_id, status, max_pokemon_per_player, free_agents)
	UpdateLeague(league *models.League) (*models.League, error)
	// soft deletes a league and all associated data
	DeleteLeague(leagueId uuid.UUID) error
	// Public helper to check if a user's player is the owner
	IsUserOwner(userID, leagueID uuid.UUID) (bool, error)
	// gets the current status of a league
	GetLeagueStatus(leagueID uuid.UUID) (enums.LeagueStatus, error)
	// retrieves all leagues with a specific status.
	GetAllLeaguesByStatus(status enums.LeagueStatus) ([]models.League, error)
	// retrieves all leagues that allow transfer credits.
	GetLeaguesThatAllowTransferCredits() ([]models.League, error)
}

type leagueRepositoryImpl struct {
	db *gorm.DB
}

func NewLeagueRepository(db *gorm.DB) LeagueRepository {
	return &leagueRepositoryImpl{
		db: db,
	}
}

// create a new league
func (r *leagueRepositoryImpl) CreateLeague(league *models.League) (*models.League, error) {
	err := r.db.Create(league).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreateLeague) - failed to create league: %v", err)
	}
	return league, nil
}

// checks if a given user is a player in a specific league.
func (r *leagueRepositoryImpl) IsUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {
	var player models.Player
	err := r.db.Where("user_id = ? AND league_id = ?", userID, leagueID).First(&player).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // User is not a player in this league
		}
		return false, fmt.Errorf("failed to check player membership: %w", err) // Other database error
	}
	return true, nil // User is a player in this league
}

// gets League by League ID with relationships preloaded
func (r *leagueRepositoryImpl) GetLeagueByID(leagueID uuid.UUID) (*models.League, error) {
	// Preload will load the associated relationships as opposed to lazy loading
	var league models.League

	// Removed Preload("CommissionerUser") as it no longer exists
	err := r.db.
		Preload("Players").
		Preload("Players.User").
		First(&league, "id = ?", leagueID).Error
	if err != nil {
		return nil, err
	}

	return &league, nil
}

// get all leagues where the user's player is owner
func (r *leagueRepositoryImpl) GetLeaguesByOwner(userID uuid.UUID) ([]models.League, error) {
	var leagues []models.League

	err := r.db.
		Joins("JOIN players ON players.league_id = leagues.id").                    // Join with the players table
		Where("players.user_id = ? AND players.role = ?", userID, rbac.PRoleOwner). // Filter by user_id and role
		Preload("Players").                                                         // Keep preloading players if needed
		Find(&leagues).Error
	if err != nil {
		return nil, err
	}

	return leagues, nil
}

// gets total count of Leagues where userID's player is the owner
func (r *leagueRepositoryImpl) GetLeaguesCountWhereOwner(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Player{}). // Change model to Player
						Where("user_id = ? AND role = ?", userID, rbac.PRoleOwner). // Filter by user_id and role
						Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// fetches all Leagues where the given userID is a player.
func (r *leagueRepositoryImpl) GetLeaguesByUser(userID uuid.UUID) ([]models.League, error) {
	var leagues []models.League

	err := r.db.
		// Joins with the Player table on the common LeagueID
		Joins("JOIN players ON players.league_id = leagues.id").
		// Filter the results where the player's user_id matches the provided userID
		Where("players.user_id = ? AND players.deleted_at IS NULL", userID). // Ensure only active players are considered
		Find(&leagues).Error                                                 // Finds the League records

	if err != nil {
		return nil, err
	}

	return leagues, nil
}

// updates a league (name, start_date, ruleset_id, status, max_pokemon_per_player, free_agents)
func (r *leagueRepositoryImpl) UpdateLeague(league *models.League) (*models.League, error) {
	err := r.db.Select(
		"name", "start_date", "end_date", "ruleset_id", "status",
		"max_pokemon_per_player", "allow_weekly_free_agents", "updated_at",
	).Updates(league).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: UpdateLeague) - failed to update league: %v", err)
	}

	return r.GetLeagueByID(league.ID)
}

// soft deletes a league and all associated data
func (r *leagueRepositoryImpl) DeleteLeague(leagueId uuid.UUID) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: DeleteLeague) - failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Soft delete all players in the league first
	if err := tx.Where("league_id = ?", leagueId).Delete(&models.Player{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DeleteLeague) - failed to delete league players: %w", err)
	}

	// Soft delete the league
	if err := tx.Delete(&models.League{}, "id = ?", leagueId).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: DeleteLeague) - failed to delete league: %w", err)
	}

	return tx.Commit().Error
}

// gets a league with all its related data
func (r *leagueRepositoryImpl) GetLeagueWithFullDetails(id uuid.UUID) (*models.League, error) {
	var league models.League

	// Removed Preload("CommissionerUser") as it no longer exists
	err := r.db.
		Preload("Players").
		Preload("Players.User").
		Preload("Players.Roster").
		Preload("Players.Roster.DraftedPokemon").
		Preload("Players.Roster.DraftedPokemon.PokemonSpecies").
		Preload("DefinedPokemon").
		Preload("DefinedPokemon.PokemonSpecies").
		First(&league, "id = ?", id).Error

	if err != nil {
		return nil, err
	}
	return &league, nil
}

// Public helper to check if a user's player is the owner
func (r *leagueRepositoryImpl) IsUserOwner(userID, leagueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Player{}). // Query the Player model
						Where("user_id = ? AND league_id = ? AND role = ?", userID, leagueID, rbac.PRoleOwner). // Check for owner role
						Count(&count).Error

	return count > 0, err
}

// gets the current status of a league
func (r *leagueRepositoryImpl) GetLeagueStatus(leagueID uuid.UUID) (enums.LeagueStatus, error) {
	var league models.League
	err := r.db.Select("status").First(&league, "id = ?", leagueID).Error
	if err != nil {
		return "", err
	}
	return league.Status, nil
}

// retrieves all leagues with a specific status.
func (r *leagueRepositoryImpl) GetAllLeaguesByStatus(status enums.LeagueStatus) ([]models.League, error) {
	var leagues []models.League
	if err := r.db.Where("status = ?", status).Find(&leagues).Error; err != nil {
		return nil, err
	}
	return leagues, nil
}

// retrieves all leagues that allow transfer credits.
func (r *leagueRepositoryImpl) GetLeaguesThatAllowTransferCredits() ([]models.League, error) {
	var leagues []models.League
	// Preload the LeagueFormat to access AllowTransferCredits
	if err := r.db.Where("format->>'allow_transfer_credits' = ?", "true").Find(&leagues).Error; err != nil {
		return nil, err
	}
	return leagues, nil
}
