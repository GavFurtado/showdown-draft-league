package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

// creates a new game
func (r *GameRepository) CreateGame(game *models.Game) (*models.Game, error) {
	err := r.db.Create(game).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreateGame) - failed to create game: %w", err)
	}
	return game, nil
}

// gets game by ID with relationships
func (r *GameRepository) GetGameByID(id uuid.UUID) (*models.Game, error) {
	var game models.Game
	err := r.db.Preload("League").
		Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Preload("Winner").
		Preload("Winner.User").
		Preload("Loser").
		Preload("Loser.User").
		Preload("Reporter").
		First(&game, "id = ?", id).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGameByID) - failed to get game: %w", err)
	}
	return &game, nil
}

// gets all games for a specific league
func (r *GameRepository) GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Preload("Winner").
		Preload("Winner.User").
		Where("league_id = ?", leagueID).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGamesByLeague) - failed to get games by league: %w", err)
	}
	return games, nil
}

// gets all games for a specific player
func (r *GameRepository) GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("League").
		Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Preload("Winner").
		Preload("Winner.User").
		Where("player_1_id = ? OR player_2_id = ?", playerID, playerID).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGamesByPlayer) - failed to get games by player: %w", err)
	}
	return games, nil
}

// gets games by round number in a league
func (r *GameRepository) GetGamesByLeagueAndRound(leagueID uuid.UUID, roundNumber int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Preload("Winner").
		Preload("Winner.User").
		Where("league_id = ? AND round_number = ?", leagueID, roundNumber).
		Order("created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGamesByLeagueAndRound) - failed to get games by league and round: %w", err)
	}
	return games, nil
}

// gets pending games for a league
func (r *GameRepository) GetPendingGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Where("league_id = ? AND status = ?", leagueID, models.GameStatusPending).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPendingGamesByLeague) - failed to get pending games: %w", err)
	}
	return games, nil
}

// gets completed games for a league
func (r *GameRepository) GetCompletedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Preload("Winner").
		Preload("Winner.User").
		Preload("Loser").
		Preload("Loser.User").
		Where("league_id = ? AND status = ?", leagueID, models.GameStatusCompleted).
		Order("round_number ASC, updated_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetCompletedGamesByLeague) - failed to get completed games: %w", err)
	}
	return games, nil
}

// gets disputed games for a league
func (r *GameRepository) GetDisputedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Preload("Reporter").
		Where("league_id = ? AND status = ?", leagueID, models.GameStatusDisputed).
		Order("updated_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetDisputedGamesByLeague) - failed to get disputed games: %w", err)
	}
	return games, nil
}

// updates game with score and potentially marks it as completed
func (r *GameRepository) UpdateGameScore(gameID uuid.UUID, player1Wins, player2Wins int) error {
	err := r.db.Model(&models.Game{}).
		Where("id = ?", gameID).
		Updates(map[string]interface{}{
			"player_1_wins": player1Wins,
			"player_2_wins": player2Wins,
		}).Error

	if err != nil {
		return fmt.Errorf("(Error: UpdateGameScore) - failed to update game score: %w", err)
	}
	return nil
}

// reports game result with winner/loser and replay links
func (r *GameRepository) ReportGameResult(
	gameID uuid.UUID,
	winnerID, loserID, reporterID uuid.UUID,
	player1Wins, player2Wins int,
	replayLinks []string,
) error {
	updates := map[string]interface{}{
		"winner_id":             winnerID,
		"loser_id":              loserID,
		"reported_by_user_id":   reporterID,
		"player_1_wins":         player1Wins,
		"player_2_wins":         player2Wins,
		"showdown_replay_links": replayLinks,
		"status":                models.GameStatusCompleted,
	}

	err := r.db.Model(&models.Game{}).
		Where("id = ?", gameID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("(Error: ReportGameResult) - failed to report game result: %w", err)
	}
	return nil
}

// marks a game as disputed
func (r *GameRepository) DisputeGame(gameID uuid.UUID, reporterID uuid.UUID) error {
	updates := map[string]interface{}{
		"status":              models.GameStatusDisputed,
		"reported_by_user_id": reporterID,
	}

	err := r.db.Model(&models.Game{}).
		Where("id = ?", gameID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("(Error: DisputeGame) - failed to dispute game: %w", err)
	}
	return nil
}

// resolves a disputed game (commissioner action)
func (r *GameRepository) ResolveDisputedGame(
	gameID uuid.UUID,
	winnerID, loserID uuid.UUID,
	player1Wins, player2Wins int,
	replayLinks []string,
) error {
	updates := map[string]interface{}{
		"winner_id":             winnerID,
		"loser_id":              loserID,
		"player_1_wins":         player1Wins,
		"player_2_wins":         player2Wins,
		"showdown_replay_links": replayLinks,
		"status":                models.GameStatusCompleted,
	}

	err := r.db.Model(&models.Game{}).
		Where("id = ?", gameID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("(Error: ResolveDisputedGame) - failed to resolve disputed game: %w", err)
	}
	return nil
}

// gets head-to-head record between two players
func (r *GameRepository) GetHeadToHeadRecord(player1ID, player2ID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("League").
		Preload("Winner").
		Preload("Winner.User").
		Where("(player_1_id = ? AND player_2_id = ?) OR (player_1_id = ? AND player_2_id = ?)",
			player1ID, player2ID, player2ID, player1ID).
		Where("status = ?", models.GameStatusCompleted).
		Order("updated_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetHeadToHeadRecord) - failed to get head-to-head record: %w", err)
	}
	return games, nil
}

// gets player's win-loss record in a specific league
func (r *GameRepository) GetPlayerRecordInLeague(playerID, leagueID uuid.UUID) (wins, losses int64, err error) {
	// Count wins
	err = r.db.Model(&models.Game{}).
		Where("league_id = ? AND winner_id = ? AND status = ?", leagueID, playerID, models.GameStatusCompleted).
		Count((*int64)(&wins)).Error
	if err != nil {
		return 0, 0, fmt.Errorf("(Error: GetPlayerRecordInLeague) - failed to count wins: %w", err)
	}

	// Count losses
	err = r.db.Model(&models.Game{}).
		Where("league_id = ? AND loser_id = ? AND status = ?", leagueID, playerID, models.GameStatusCompleted).
		Count((*int64)(&losses)).Error
	if err != nil {
		return 0, 0, fmt.Errorf("(Error: GetPlayerRecordInLeague) - failed to count losses: %w", err)
	}

	return wins, losses, nil
}

// gets current round number for a league (highest round with games)
func (r *GameRepository) GetCurrentRoundNumber(leagueID uuid.UUID) (int, error) {
	var maxRound int
	err := r.db.Model(&models.Game{}).
		Select("COALESCE(MAX(round_number), 0)").
		Where("league_id = ?", leagueID).
		Scan(&maxRound).Error

	if err != nil {
		return 0, fmt.Errorf("(Error: GetCurrentRoundNumber) - failed to get current round: %w", err)
	}
	return maxRound, nil
}

// bulk creates games for a round (useful for scheduling)
func (r *GameRepository) CreateGamesForRound(games []models.Game) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: CreateGamesForRound) - failed to start transaction: %w", tx.Error)
	}

	// if fails at any point due to panic, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, game := range games {
		if err := tx.Create(&game).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("(Error: CreateGamesForRound) - failed to create game: %w", err)
		}
	}

	return tx.Commit().Error
}

// updates player records after game completion (transaction)
func (r *GameRepository) UpdatePlayerRecordsAfterGame(
	winnerID, loserID uuid.UUID,
	winnerNewWins, winnerNewLosses, loserNewWins, loserNewLosses int,
) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: UpdatePlayerRecordsAfterGame) - failed to start transaction: %w", tx.Error)
	}

	// if fails at any point due to panic, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update winner's record
	if err := tx.Model(&models.Player{}).
		Where("id = ?", winnerID).
		Updates(map[string]interface{}{
			"wins":   winnerNewWins,
			"losses": winnerNewLosses,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: UpdatePlayerRecordsAfterGame) - failed to update winner record: %w", err)
	}

	// Update loser's record
	if err := tx.Model(&models.Player{}).
		Where("id = ?", loserID).
		Updates(map[string]interface{}{
			"wins":   loserNewWins,
			"losses": loserNewLosses,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: UpdatePlayerRecordsAfterGame) - failed to update loser record: %w", err)
	}

	return tx.Commit().Error
}

// soft deletes a game
func (r *GameRepository) DeleteGame(gameID uuid.UUID) error {
	err := r.db.Delete(&models.Game{}, "id = ?", gameID).Error
	if err != nil {
		return fmt.Errorf("(Error: DeleteGame) - failed to delete game: %w", err)
	}
	return nil
}

// gets games that need to be played by a specific player (pending games involving the player)
func (r *GameRepository) GetPendingGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("League").
		Preload("Player1").
		Preload("Player1.User").
		Preload("Player2").
		Preload("Player2.User").
		Where("(player_1_id = ? OR player_2_id = ?) AND status = ?", playerID, playerID, models.GameStatusPending).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPendingGamesByPlayer) - failed to get pending games by player: %w", err)
	}
	return games, nil
}
