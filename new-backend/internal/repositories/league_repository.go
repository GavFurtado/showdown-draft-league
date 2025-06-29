package repositories

import (
	"errors"
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeagueRepository struct {
	db *gorm.DB
}

func NewLeagueRepository(db *gorm.DB) *LeagueRepository {
	return &LeagueRepository{db: db}
}

func (r *LeagueRepository) CreateLeague(league *models.League) (*models.League, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("(Error: CreateLeague) - failed to start transaction: %w", tx.Error)
	}

	// if fails at any point due to panic, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := tx.Create(league).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("(Error: CreateLeague) - failed to create transaction: %v", err)
	}

	// League was created
	// Add the commissioner (creator of league) as a player
	commisionerPlayer := &models.Player{
		UserID:         league.CommissionerUserID,
		LeagueID:       league.ID,
		InLeagueName:   "",                  // can be set by player later
		TeamName:       "Commissioner Team", // can be changed later (optional)
		DraftPoints:    int(league.StartingDraftPoints),
		IsCommissioner: true,
		DraftPosition:  1, // TODO: look into this. Should be random, no?
	}
	if err := tx.Create(commisionerPlayer).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("(Error: CreateLeague) - failed to create commissioner player: %w", err)
	}

	// Commissioner into league and League created successfully
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("(Error: CreateLeague) - failed to create transaction: %w", err)
	}

	return league, nil
}

// checks if a given user is a player in a specific league.
func (r *LeagueRepository) IsUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {
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
func (r *LeagueRepository) GetLeagueByID(id uuid.UUID) (*models.League, error) {
	// Preload will load the associated relationships as opposed to lazy loading
	var league models.League

	err := r.db.Preload("CommisionerUser").
		Preload("Players").
		Preload("Players.User").
		First(&league, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &league, nil
}

// gets all Leagues where userID is the commisioner
func (r *LeagueRepository) GetLeaguesByCommissioner(userID uuid.UUID) ([]models.League, error) {
	var leagues []models.League

	err := r.db.Where("commissioner_user_id = ?", userID).
		Preload("Players").
		Find(&leagues).Error
	if err != nil {
		return nil, err
	}

	return leagues, nil
}

// gets total count of Leagues where userID is the Commisioner
func (r *LeagueRepository) GetLeaguesCountByCommissioner(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.League{}).
		Where("commissioner_user_id = ?", userID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// fetches all Leagues where the given userID is a player.
func (r *LeagueRepository) GetLeaguesByUser(userID uuid.UUID) ([]models.League, error) {
	var leagues []models.League

	err := r.db.
		// Joins with the Player table on the common LeagueID
		Joins("JOIN players ON players.league_id = leagues.id").
		// Filter the results where the player's user_id matches the provided userID
		Where("players.user_id = ?", userID).
		Find(&leagues).Error // Finds the League records

	if err != nil {
		return nil, err
	}

	return leagues, nil
}

func (r *LeagueRepository) UpdateLeague(league *models.League) (*models.League, error) {
	err := r.db.Select(
		"name", "start_date", "end_date", "ruleset_id", "status",
		"max_pokemon_per_player", "allow_weekly_free_agents", "updated_at",
	).Updates(league).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: UpdateLeague) - failed to update league: %w", err)
	}

	return r.GetLeagueByID(league.ID)
}

// soft deletes a league and all associated data
func (r *LeagueRepository) DeleteLeague(leagueId uuid.UUID) error {
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
func (r *LeagueRepository) GetLeagueWithFullDetails(id uuid.UUID) (*models.League, error) {
	var league models.League

	err := r.db.Preload("CommisionerUser").
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

// Helper to check if a user is the commissioner of a league
func (r *LeagueRepository) IsUserCommissioner(userID, leagueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.League{}).
		Where("id = ? AND commissioner_user_id = ?", leagueID, userID).
		Count(&count).Error

	return count > 0, err
}
