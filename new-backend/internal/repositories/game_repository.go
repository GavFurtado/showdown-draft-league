package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepository interface {
	// creates a new game
	CreateGame(game *models.Game) (*models.Game, error)
	// gets game by ID with relationships
	GetGameByID(id uuid.UUID) (models.Game, error)
	// gets all games for a specific league
	GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error)
	// gets all games for a specific player
	GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error)
	// gets games by round number (regular season) in a league
	GetGamesByLeagueAndRound(leagueID uuid.UUID, roundNumber int) ([]models.Game, error)
	// marks a game as disputed
	DisputeGame(gameID uuid.UUID, reporterID uuid.UUID) error
	// gets head-to-head record between two players
	GetHeadToHeadRecord(player1ID, player2ID uuid.UUID) ([]models.Game, error)
	// gets player's win-loss record in a specific league
	GetPlayerRecordInLeague(playerID, leagueID uuid.UUID) (wins, losses int64, err error)
	// bulk creates games
	CreateGames(games []*models.Game) error
	// soft deletes a game
	DeleteGame(gameID uuid.UUID) error
	// gets games that need to be played by a specific player (scheduled games involving the player)
	GetScheduledGamesByPlayer(playerID uuid.UUID) ([]models.Game, error)
	// gets scheduled games for a league
	GetScheduledGamesByLeague(leagueID uuid.UUID) ([]models.Game, error)
	// gets completed games for a league
	GetCompletedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error)
	// gets disputed games for a league
	GetDisputedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error)
	// checks if games of a specific type exist for a given league.
	HasGames(leagueID uuid.UUID, gameType enums.GameType) (bool, error)

	// New/Refactored methods for game updates
	UpdateGameReport(gameID uuid.UUID, loserID uuid.UUID, dto *common.ReportGameDTO) error
	FinalizeGameAndUpdateStats(gameID uuid.UUID, loserID uuid.UUID, dto *common.FinalizeGameDTO) error
}

type gameRepositoryImpl struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepositoryImpl{
		db: db,
	}
}

// creates a new game
func (r *gameRepositoryImpl) CreateGame(game *models.Game) (*models.Game, error) {
	err := r.db.Create(game).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: CreateGame) - failed to create game: %w", err)
	}
	return game, nil
}

// gets game by ID with relationships
func (r *gameRepositoryImpl) GetGameByID(id uuid.UUID) (models.Game, error) {
	var game models.Game
	err := r.db.
		Preload("Player1").
		Preload("Player2").
		Preload("Winner").
		Preload("Loser").
		Preload("ReportingPlayer").
		Preload("Approver"). // Add Approver preload
		First(&game, "id = ?", id).Error

	if err != nil {
		return game, fmt.Errorf("(Error: GetGameByID) - failed to get game: %w", err)
	}
	return game, nil
}

// gets all games for a specific league
func (r *gameRepositoryImpl) GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.
		Preload("Player1").
		Preload("Player2").
		Preload("Winner").
		Where("league_id = ?", leagueID).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGamesByLeague) - failed to get games by league: %w", err)
	}
	return games, nil
}

// gets all games for a specific player
func (r *gameRepositoryImpl) GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.
		Preload("Player1").
		Preload("Player2").
		Preload("Winner").
		Where("player1_id = ? OR player2_id = ?", playerID, playerID).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGamesByPlayer) - failed to get games by player: %w", err)
	}
	return games, nil
}

// gets games by round number (regular season) in a league
func (r *gameRepositoryImpl) GetGamesByLeagueAndRound(leagueID uuid.UUID, roundNumber int) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player2").
		Preload("Winner").
		Where("league_id = ? AND game_type = ? AND round_number = ?", leagueID, enums.GameTypeRegularSeason, roundNumber).
		Order("created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetGamesByLeagueAndRound) - failed to get games by league and round: %w", err)
	}
	return games, nil
}

// gets scheduled games for a league
func (r *gameRepositoryImpl) GetScheduledGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player2").
		Where("league_id = ? AND status = ?", leagueID, enums.GameStatusScheduled).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPendingGamesByLeague) - failed to get scheduled games: %w", err)
	}
	return games, nil
}

// gets scheduled games for a player
func (r *gameRepositoryImpl) GetScheduledGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player2").
		Where("(player1_id = ? OR player2_id = ?) AND status = ?", playerID, playerID, enums.GameStatusScheduled).
		Order("round_number ASC, created_at ASC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetPendingGamesByPlayer) - failed to get scheduled games: %w", err)
	}
	return games, nil
}

// gets completed games for a league
func (r *gameRepositoryImpl) GetCompletedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player2").
		Preload("Winner").
		Preload("Loser").
		Where("league_id = ? AND status = ?", leagueID, enums.GameStatusCompleted).
		Order("round_number ASC, updated_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetCompletedGamesByLeague) - failed to get completed games: %w", err)
	}
	return games, nil
}

// gets disputed games for a league
func (r *gameRepositoryImpl) GetDisputedGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("Player1").
		Preload("Player2").
		Preload("ReportingPlayer").
		Where("league_id = ? AND status = ?", leagueID, enums.GameStatusDisputed).
		Order("updated_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetDisputedGamesByLeague) - failed to get disputed games: %w", err)
	}
	return games, nil
}

// HasGames checks if any games of a specific type exist for a given league.
func (r *gameRepositoryImpl) HasGames(leagueID uuid.UUID, gameType enums.GameType) (bool, error) {
	var count int64
	err := r.db.Model(&models.Game{}).
		Where("league_id = ? AND game_type = ?", leagueID, gameType).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("(Error: HasGames) - failed to check for existing games: %w", err)
	}
	return count > 0, nil
}


// marks a game as disputed
func (r *gameRepositoryImpl) DisputeGame(gameID uuid.UUID, reporterID uuid.UUID) error {
	updates := map[string]any{
		"status":              enums.GameStatusDisputed,
		"reporting_player_id": reporterID,
	}

	err := r.db.Model(&models.Game{}).
		Where("id = ?", gameID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("(Repository: DisputeGame) - failed to dispute game: %w", err)
	}
	return nil
}


// gets head-to-head record between two players
func (r *gameRepositoryImpl) GetHeadToHeadRecord(player1ID, player2ID uuid.UUID) ([]models.Game, error) {
	var games []models.Game
	err := r.db.Preload("League").
		Preload("Winner").
		Preload("Winner.User").
		Where("(player1_id = ? AND player2_id = ?) OR (player1_id = ? AND player2_id = ?)",
			player1ID, player2ID, player2ID, player1ID).
		Where("status = ?", enums.GameStatusCompleted).
		Order("updated_at DESC").
		Find(&games).Error

	if err != nil {
		return nil, fmt.Errorf("(Error: GetHeadToHeadRecord) - failed to get head-to-head record: %w", err)
	}
	return games, nil
}

// gets player's win-loss record in a specific league
func (r *gameRepositoryImpl) GetPlayerRecordInLeague(playerID, leagueID uuid.UUID) (wins, losses int64, err error) {
	// Count wins
	err = r.db.Model(&models.Game{}).
		Where("league_id = ? AND winner_id = ? AND status = ?", leagueID, playerID, enums.GameStatusCompleted).
		Count(&wins).Error
	if err != nil {
		return 0, 0, fmt.Errorf("(Error: GetPlayerRecordInLeague) - failed to count wins: %w", err)
	}

	// Count losses
	err = r.db.Model(&models.Game{}).
		Where("league_id = ? AND loser_id = ? AND status = ?", leagueID, playerID, enums.GameStatusCompleted).
		Count(&losses).Error
	if err != nil {
		return 0, 0, fmt.Errorf("(Error: GetPlayerRecordInLeague) - failed to count losses: %w", err)
	}

	return wins, losses, nil
}

// bulk creates games
func (r *gameRepositoryImpl) CreateGames(games []*models.Game) error {
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


// soft deletes a game
func (r *gameRepositoryImpl) DeleteGame(gameID uuid.UUID) error {
	err := r.db.Delete(&models.Game{}, "id = ?", gameID).Error
	if err != nil {
		return fmt.Errorf("(Error: DeleteGame) - failed to delete game: %w", err)
	}
	return nil
}

func (r *gameRepositoryImpl) UpdateGameReport(gameID uuid.UUID, loserID uuid.UUID, dto *common.ReportGameDTO) error {
	updates := map[string]interface{}{
		"winner_id":             dto.WinnerID,
		"loser_id":              loserID,
		"player1_wins":          dto.Player1Wins,
		"player2_wins":          dto.Player2Wins,
		"showdown_replay_links": dto.ReplayLinks,
		"reporting_player_id":   dto.ReporterID,
		"status":                enums.GameStatusApprovalPending,
		"approver_id":           nil, // Clear approver when a new report comes in
	}

	err := r.db.Model(&models.Game{}).Where("id = ?", gameID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("(Repository: UpdateGameReport) - failed to update game with report: %w", err)
	}
	return nil
}

// FinalizeGameAndUpdateStats handles the entire process of finalizing a game within a single transaction.
func (r *gameRepositoryImpl) FinalizeGameAndUpdateStats(gameID uuid.UUID, loserID uuid.UUID, dto *common.FinalizeGameDTO) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Repository: FinalizeGameAndUpdateStats) - failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		}
	}()

	// Fetch current game state for comparison
	var oldGame models.Game
	if err := tx.First(&oldGame, "id = ?", gameID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("FinalizeGameAndUpdateStats: failed to get game %s: %w", gameID, err)
	}

	// If game was already completed, revert old player stats
	if oldGame.Status == enums.GameStatusCompleted && oldGame.WinnerID != nil && oldGame.LoserID != nil {
		if err := r.decrementPlayerStats(tx, *oldGame.WinnerID, *oldGame.LoserID); err != nil {
			tx.Rollback()
			return fmt.Errorf("FinalizeGameAndUpdateStats: failed to decrement old player stats for game %s: %w", gameID, err)
		}
	}

	// Update the game record with the final results
	if err := r.finalizeGame(tx, gameID, loserID, dto); err != nil {
		tx.Rollback()
		return fmt.Errorf("FinalizeGameAndUpdateStats: failed to finalize game %s: %w", gameID, err)
	}

	// Apply new player stats
	if err := r.incrementPlayerStats(tx, dto.WinnerID, loserID); err != nil {
		tx.Rollback()
		return fmt.Errorf("FinalizeGameAndUpdateStats: failed to increment new player stats for game %s: %w", gameID, err)
	}

	return tx.Commit().Error
}

// finalizeGame is a private helper to update the game record within a transaction.
func (r *gameRepositoryImpl) finalizeGame(tx *gorm.DB, gameID uuid.UUID, loserID uuid.UUID, dto *common.FinalizeGameDTO) error {
	updates := map[string]interface{}{
		"winner_id":             dto.WinnerID,
		"loser_id":              loserID,
		"player1_wins":          dto.Player1Wins,
		"player2_wins":          dto.Player2Wins,
		"showdown_replay_links": dto.ReplayLinks,
		"approver_id":           dto.FinalizerID,
		"status":                enums.GameStatusCompleted,
	}

	err := tx.Model(&models.Game{}).Where("id = ?", gameID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("(Repository: finalizeGame) - failed to finalize game: %w", err)
	}
	return nil
}

// incrementPlayerStats is a private helper to atomically increment player stats within a transaction.
func (r *gameRepositoryImpl) incrementPlayerStats(tx *gorm.DB, winnerID, loserID uuid.UUID) error {
	if err := tx.Model(&models.Player{}).Where("id = ?", winnerID).Update("wins", gorm.Expr("wins + 1")).Error; err != nil {
		return fmt.Errorf("(Repository: incrementPlayerStats) - failed to increment winner's wins: %w", err)
	}
	if err := tx.Model(&models.Player{}).Where("id = ?", loserID).Update("losses", gorm.Expr("losses + 1")).Error; err != nil {
		return fmt.Errorf("(Repository: incrementPlayerStats) - failed to increment loser's losses: %w", err)
	}
	return nil
}

// decrementPlayerStats is a private helper to atomically decrement player stats within a transaction.
func (r *gameRepositoryImpl) decrementPlayerStats(tx *gorm.DB, winnerID, loserID uuid.UUID) error {
	if err := tx.Model(&models.Player{}).Where("id = ?", winnerID).Update("wins", gorm.Expr("wins - 1")).Error; err != nil {
		return fmt.Errorf("(Repository: decrementPlayerStats) - failed to decrement winner's wins: %w", err)
	}
	if err := tx.Model(&models.Player{}).Where("id = ?", loserID).Update("losses", gorm.Expr("losses - 1")).Error; err != nil {
		return fmt.Errorf("(Repository: decrementPlayerStats) - failed to decrement loser's losses: %w", err)
	}
	return nil
}
