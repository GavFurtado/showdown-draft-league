package services

import (
	"errors"
	"fmt"
	"log"
	"math/bits"
	"math/rand/v2"
	"sort"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameService interface {
	GetGameByID(ID uuid.UUID) (*models.Game, error)
	GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error)
	GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error)
	GenerateRegularSeasonGames(leagueID uuid.UUID) error
	GeneratePlayoffBracket(leagueID uuid.UUID) error

	// New/Refactored methods
	ReportGameResult(gameID uuid.UUID, dto *common.ReportGameDTO) error
	FinalizeGameResult(gameID uuid.UUID, dto *common.FinalizeGameDTO) error
}

type gameServiceImpl struct {
	gameRepo   repositories.GameRepository
	leagueRepo repositories.LeagueRepository
	playerRepo repositories.PlayerRepository
}

func NewGameService(
	gameRepo repositories.GameRepository,
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
) GameService {
	return &gameServiceImpl{
		gameRepo:   gameRepo,
		leagueRepo: leagueRepo,
		playerRepo: playerRepo,
	}
}

func (s *gameServiceImpl) GetGameByID(ID uuid.UUID) (*models.Game, error) {
	game, err := s.gameRepo.GetGameByID(ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", common.ErrGameNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", common.ErrInternalService, err)
	}
	return &game, nil
}

func (s *gameServiceImpl) GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	games, err := s.gameRepo.GetGamesByLeague(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", common.ErrGameNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", common.ErrInternalService, err)
	}
	return games, nil
}

func (s *gameServiceImpl) GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	games, err := s.gameRepo.GetGamesByPlayer(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", common.ErrGameNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", common.ErrInternalService, err)
	}

	return games, nil
}

// ReportGameResult allows a player to report the result of a game.
func (s *gameServiceImpl) ReportGameResult(gameID uuid.UUID, dto *common.ReportGameDTO) error {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return fmt.Errorf("ReportGameResult: failed to get game %s: %w", gameID, err)
	}

	// Determine loser ID
	var loserID uuid.UUID
	if dto.WinnerID == game.Player1ID {
		loserID = game.Player2ID
	} else if dto.WinnerID == game.Player2ID {
		loserID = game.Player1ID
	} else {
		return common.ErrInvalidInput // Winner must be one of the players in the game
	}

	// Basic validation: reported player must be one of the game's participants
	if dto.ReporterID != game.Player1ID && dto.ReporterID != game.Player2ID {
		return fmt.Errorf("%w: reporter is not a participant in the game", common.ErrUnauthorized)
	}

	// Ensure winner and loser are distinct
	if dto.WinnerID == loserID {
		return fmt.Errorf("%w: winner and loser cannot be the same", common.ErrInvalidInput)
	}

	// Ensure scores are not tied if a winner is provided
	if dto.Player1Wins == dto.Player2Wins {
		return fmt.Errorf("%w: scores cannot be tied for a reported result", common.ErrInvalidInput)
	}

	if err := s.gameRepo.UpdateGameReport(gameID, loserID, dto); err != nil {
		return fmt.Errorf("ReportGameResult: failed to update game report %s: %w", gameID, err)
	}

	return nil
}

// FinalizeGameResult allows league staff to approve, submit, or retroactively edit a game result.
func (s *gameServiceImpl) FinalizeGameResult(gameID uuid.UUID, dto *common.FinalizeGameDTO) error {
	// Fetch game to determine loser ID
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		return fmt.Errorf("FinalizeGameResult: failed to get game %s: %w", gameID, err)
	}

	// Determine loser ID for the final result
	var loserID uuid.UUID
	if dto.WinnerID == game.Player1ID {
		loserID = game.Player2ID
	} else if dto.WinnerID == game.Player2ID {
		loserID = game.Player1ID
	} else {
		return common.ErrInvalidInput // Winner must be one of the players in the game
	}

	// RBAC Check is handled in controller, service layer proceeds with business logic

	err = s.gameRepo.FinalizeGameAndUpdateStats(gameID, loserID, dto)
	if err != nil {
		return fmt.Errorf("FinalizeGameResult: failed to finalize game and update stats for game %s: %w", gameID, err)
	}

	return nil
}

// GenerateRegularSeasonGames generates all the games of the regular season for every week assigning the correct RoundNumbers.
// For GroupCounts > 1 (only 1 or 2 is allowed), players are assigned opponents within their group.
func (s *gameServiceImpl) GenerateRegularSeasonGames(leagueID uuid.UUID) error {
	league, err := s.fetchLeagueResource(leagueID)
	if err != nil {
		log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - Couldn't fetch league %s: %v\n", leagueID, err)
		return err
	}

	// League needs to be in POST_DRAFT status and not a BRACKET_ONLY Season League
	if league.Status != enums.LeagueStatusPostDraft && league.Format.SeasonType == enums.LeagueSeasonTypeBracketOnly {
		log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - League %s not in valid state to generate season bracket: %v\n", leagueID, err)
		return common.ErrInvalidState
	}

	// GroupCount can only be 1 or 2
	// For GroupCount=1, all Players are auto assigned GroupNumber 1 on player creation
	playersByGroupNumber := make([][]models.Player, league.Format.GroupCount)
	for i := 0; i < league.Format.GroupCount; i++ {
		players, err := s.playerRepo.GetPlayersByLeagueAndGroupNumber(league.ID, i+1)
		if err != nil {
			log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - Repository error fetching Players by League %s with Group Number %d: %v\n", league.ID, i+1, err)
			return common.ErrInternalService
		}
		playersByGroupNumber[i] = players
	}

	var allGeneratedGames []*models.Game
	for groupIndex, playersInGroup := range playersByGroupNumber {
		groupNumber := groupIndex + 1
		games, err := s.generateRoundRobinGamesForGroup(league.ID, playersInGroup, groupNumber)
		if err != nil {
			log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - Error generating round-robin games for group %d in league %s: %v\n", groupNumber, leagueID, err)
			return err
		}
		allGeneratedGames = append(allGeneratedGames, games...)
	}

	if len(allGeneratedGames) > 0 {
		err = s.gameRepo.CreateGames(allGeneratedGames)
		if err != nil {
			log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - Repository error creating games for league %s: %v\n", leagueID, err)
			return common.ErrInternalService
		}
	}
	return nil
}

func (s *gameServiceImpl) GeneratePlayoffBracket(leagueID uuid.UUID) error {
	league, err := s.fetchLeagueResource(leagueID)
	if err != nil {
		log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Couldn't fetch league %s: %v\n", leagueID, err)
		return err
	}

	if league.Format.PlayoffType == enums.LeaguePlayoffTypeNone {
		return fmt.Errorf("playoffs are disabled for this league")
	}

	if (league.Format.SeasonType == enums.LeagueSeasonTypeBracketOnly && league.Status != enums.LeagueStatusPostDraft) ||
		(league.Format.SeasonType == enums.LeagueSeasonTypeHybrid && league.Status != enums.LeagueStatusPostRegularSeason) {
		log.Printf("ERROR: (Service: GeneratePlayoffBracket) - League not in valid status to generate playoff bracket.\n")
		return err
	}

	if league.Format.PlayoffParticipantCount%2 == 1 {
		return fmt.Errorf("cannot have odd number of participants for playoffs")
	}

	playersByGroup := make([][]models.Player, league.Format.GroupCount)
	for i := 0; i < league.Format.GroupCount; i++ {
		playersOfGroupX, err := s.playerRepo.GetPlayersByLeagueAndGroupNumber(league.ID, i+1)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket): error fetching players of group %d for league %s: %v", i+1, league.ID, err)
			return fmt.Errorf("failed to fetch players for group %d: %w", i+1, err)
		}
		playersByGroup[i] = playersOfGroupX
	}

	seededPlayers, err := s.getSeededPlayers(league, playersByGroup)
	if err != nil {
		log.Printf("ERROR: (Service: GeneratePlayoffBracket): error seeding players for playoffs for league %s: %v", league.ID, err)
		return err
	}

	var generatedGames []*models.Game
	if league.Format.PlayoffType == enums.LeaguePlayoffTypeSingleElim {
		// Single Elimination
		// Single Elim + Fully Seeded is an invalid league configuration
		// STANDARD or BYES_ONLY are allowed
		if league.Format.PlayoffSeedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
			return fmt.Errorf("%w: %s and %s are incompatible playoff options",
				common.ErrInvalidLeagueConfiguration,
				enums.LeaguePlayoffTypeSingleElim,
				enums.LeaguePlayoffSeedingTypeFullySeeded)
		}
		generatedGames, err = s.generateSingleEliminationBracket(league, seededPlayers)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Error generating single elimination bracket for league %s: %v\n", leagueID, err)
			return err
		}
	} else {
		// Double Elimination
		// compatible with all types of PlayoffSeedingType
		generatedGames, err = s.generateDoubleEliminationBracket(league, seededPlayers)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Error generating single elimination bracket for league %s: %v\n", leagueID, err)
			return err
		}
	}
	if len(generatedGames) > 0 {
		err = s.gameRepo.CreateGames(generatedGames)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Repository error creating games for league %s: %v\n", leagueID, err)
			return common.ErrInternalService
		}
	}

	return nil
}

// PRIVATE HELPERS

// generateSingleEliminationBracket generates the games for the single elimination bracket
// It takes into account changes introduced by various Format.PlayoffSeedingType
// returns a slice of all the generated Games and an error if generation failed
func (s *gameServiceImpl) generateSingleEliminationBracket(league *models.League, seededPlayers []models.Player) ([]*models.Game, error) {
	var generatedGames []*models.Game
	numParticipants := len(seededPlayers)
	if numParticipants == 0 { // should already be validated by this point
		return nil, fmt.Errorf("No players provided for bracket generation")
	}

	// Each round must have a power of 2 # of participants
	// If this is not the case, bye "psuedo-players" (they always lose) can be added to make it a power of 2 (i.e., naturalByesNeeded)
	seedingType := league.Format.PlayoffSeedingType
	nextPowerOfTwo := s.getClosestPowerOfTwo(numParticipants)
	naturalByesNeeded := nextPowerOfTwo - numParticipants

	if naturalByesNeeded > 0 && seedingType == enums.LeaguePlayoffSeedingTypeStandard {
		return nil, fmt.Errorf("%w: Cannot construct single elimination bracket. Either add %d participants or enable 'byes' for %d players",
			common.ErrInvalidLeagueConfiguration,
			naturalByesNeeded, naturalByesNeeded)
	}
	if naturalByesNeeded != league.Format.PlayoffByesCount {
		return nil, fmt.Errorf("%w: Bye count is set to %d but valid is %d for %d participants", common.ErrInvalidLeagueConfiguration, league.Format.PlayoffByesCount, naturalByesNeeded, numParticipants)
	}

	var playersGettingByes []models.Player
	var playersInRound1 []models.Player
	if seedingType == enums.LeaguePlayoffSeedingTypeByesOnly {
		// The first 'PlayoffByesCount' players get byes
		playersGettingByes = seededPlayers[:league.Format.PlayoffByesCount]
		// The remaining players play in Round 1
		playersInRound1 = seededPlayers[league.Format.PlayoffByesCount:]
	} else {
		// If no byes, all players are in Round 1
		playersInRound1 = seededPlayers
	}

	// --- First Round Game generation logic ---
	var currentRoundGames []*models.Game
	roundNumber, gameNumberInRound := 1, 0
	for l, r := 0, len(playersInRound1)-1; l <= r; l, r = l+1, r-1 {
		player1 := playersInRound1[l]
		player2 := playersInRound1[r]
		var bracketPositionStr string

		gameNumberInRound++
		bracketPositionStr = fmt.Sprintf("Round %d: Game %d", roundNumber, gameNumberInRound)

		// Create new game object
		// NOTE: Why uuid.New()? Everywhere else we let db/gorm do it.
		// because IDs are needed for other game objects.
		// It isn't really needed for round 1 but
		// left here for this explanation and for consistency
		newGame := &models.Game{
			ID:              uuid.New(),
			LeagueID:        league.ID,
			Player1ID:       player1.ID,
			Player2ID:       player2.ID,
			WinnerID:        nil,
			LoserID:         nil,
			Player1Wins:     0,
			Player2Wins:     0,
			RoundNumber:     roundNumber,
			GroupNumber:     nil,
			GameType:        enums.GameTypePlayoffSingleElim,
			Status:          enums.GameStatusScheduled,
			BracketPosition: &bracketPositionStr,
		}
		generatedGames = append(generatedGames, newGame)
		currentRoundGames = append(currentRoundGames, newGame)
	}

	// --- Subsequent Round Generation ---
	// For subsequent rounds, we generate placeholder game objects
	// with no player IDs filled in with exception to players
	// that were granted a bye to round 2
	totalRounds := getLog2(nextPowerOfTwo)
	j := 0
	for roundNumber := 2; roundNumber <= totalRounds; roundNumber++ {
		var nextRoundGames []*models.Game
		gameNumberInRound := 0

		// Each pair of games from the previous round feeds into one game in the new round
		i := 0
		for i < len(currentRoundGames) {
			gameNumberInRound++
			bracketPositionStr := fmt.Sprintf("Round %d: Game %d", roundNumber, gameNumberInRound)

			newGameID := uuid.New()
			newGamePlayer1ID := uuid.Nil

			byeGranted := false
			// for round 2 games, check the playersGettingByes slice
			// set the new round 2 game's player1ID to the valid player if it exists in the array
			if roundNumber == 2 && j < len(playersGettingByes) {
				// set uuid of newGamePlayer1ID
				newGamePlayer1ID = playersGettingByes[j].ID
				j++
				byeGranted = true
			}

			// link the previous round game(s) to this game
			// by setting winnerToGameID for those game(s)
			// 2 games when no bye is granted and 1 when bye is granted
			currentRoundGames[i].WinnerToGameID = newGameID
			if !byeGranted {
				currentRoundGames[i+1].WinnerToGameID = newGameID
			}

			newGame := &models.Game{
				ID:              newGameID,
				LeagueID:        league.ID,
				Player1ID:       newGamePlayer1ID, // playerID with bye or placeholder uuid.Nil
				Player2ID:       uuid.Nil,         // placeholder uuid.Nil
				WinnerID:        nil,
				LoserID:         nil,
				Player1Wins:     0,
				Player2Wins:     0,
				RoundNumber:     roundNumber,
				GroupNumber:     nil,
				GameType:        enums.GameTypePlayoffSingleElim,
				Status:          enums.GameStatusScheduled,
				BracketPosition: &bracketPositionStr,
			}
			nextRoundGames = append(nextRoundGames, newGame)
			generatedGames = append(generatedGames, newGame)

			if byeGranted {
				i += 1
			} else {
				i += 2
			}
		}
		currentRoundGames = nextRoundGames
	}

	// rename the last game to "Grand Final" and set the correct type
	if len(generatedGames) > 0 {
		bracketPositionStr := "Grand Final"
		generatedGames[len(generatedGames)-1].BracketPosition = &bracketPositionStr
		generatedGames[len(generatedGames)-1].GameType = enums.GameTypePlayoffGrandFinal
	}

	return generatedGames, nil
}

func (s *gameServiceImpl) generateDoubleEliminationBracket(league *models.League, seededPlayers []models.Player) ([]*models.Game, error) {
	// Upper Bracket (UB); Lower Bracket (LB)
	var generatedGames []*models.Game
	numParticipants := len(seededPlayers)
	if numParticipants == 0 { // should already be validated by this point
		return nil, fmt.Errorf("No players provided for bracket generation")
	}

	seedingType := league.Format.PlayoffSeedingType
	byeCount := league.Format.PlayoffByesCount
	nextPowerOfTwo := s.getClosestPowerOfTwo(numParticipants)
	playersGettingByes := []models.Player{}
	playersStartingInUB1 := []models.Player{}
	playersStartingInLB1 := []models.Player{} // only for fully seeded brackets
	remainingPlayers := []models.Player{}     // numParticipants - Format.PlayoffByeCount

	if byeCount != 0 && seedingType == enums.LeaguePlayoffSeedingTypeStandard {
		return nil, fmt.Errorf("%w: Bye count must be 0 for Standard Seeded brackets.", common.ErrInvalidLeagueConfiguration)
	}

	// initialize remainingPlayers and playersGettingByes
	if byeCount > 0 {
		playersGettingByes = seededPlayers[:byeCount]
		remainingPlayers = seededPlayers[byeCount:]
	} else {
		remainingPlayers = seededPlayers
	}

	// --- Initial validation of numParticipant ---
	if seedingType == enums.LeaguePlayoffSeedingTypeByesOnly {
		naturalByesNeeded := nextPowerOfTwo - numParticipants
		if naturalByesNeeded != byeCount {
			return nil, fmt.Errorf("%w: For BYES_ONLY Double Elimination with %d particpants, number of byes (Current: %d) allowed is %d.",
				common.ErrInvalidLeagueConfiguration, numParticipants, byeCount, naturalByesNeeded)
		}
	} else if seedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
		nPlayersGettingByes := len(playersGettingByes)
		nRemainingPlayers := len(remainingPlayers) // nPlayers_UB1 + nPlayers_LB1

		if nRemainingPlayers%3 != 0 {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED Double Elimination, number of players that don't get a bye (Current: %d) must be divisible by 3.",
				common.ErrInvalidLeagueConfiguration, nRemainingPlayers)
		}

		nPlayers_UB1 := (2 * nRemainingPlayers) / 3
		if nPlayers_UB1 <= 0 || nPlayers_UB1%2 != 0 { // Must be positive even number
			return nil, fmt.Errorf("%w: For FULLY_SEEDED Double Elimination, number of players starting in Upper Bracket Round 1 (Current: %d) must be a positive even number",
				common.ErrInvalidLeagueConfiguration, nPlayersGettingByes)
		}

		// Total effective player count in Upper Bracket Round 2
		nEffectivePlayers_UB2 := byeCount + (nPlayers_UB1 / 2)
		if nEffectivePlayers_UB2 <= 0 || !isPowerOfTwo(nEffectivePlayers_UB2) {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED Double Elimination, the total effective number of players in Upper Bracket 2 (Current: %d) must be a positive power of two",
				common.ErrInvalidLeagueConfiguration, nEffectivePlayers_UB2)
		}

		// For balanced bracket structure, # of players in UB Round 1 must be atleast twice the # of byes
		// Disallows brackets where more players start in UB Round 2 than UB Round 1 that are otherwise valid
		if nPlayers_UB1 < 2*byeCount {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED double elimination, the number of players receiving byes (Current: %d) cannot exceed the number of players starting in Upper Bracket Round 1 (%d) to maintain a balanced bracket structure.",
				common.ErrInvalidLeagueConfiguration, byeCount, nPlayers_UB1)
		}

		playersStartingInUB1 = remainingPlayers[:nPlayers_UB1]
		playersStartingInLB1 = remainingPlayers[nPlayers_UB1:] // nRemainingPlayers / 3
	} else { // Standard seeding
		playersStartingInUB1 = seededPlayers // All players start in UB Round 1
		if !isPowerOfTwo(numParticipants) {
			return nil, fmt.Errorf("%w: For STANDARD double elimination, number of participants must be a positive power of two", common.ErrInvalidLeagueConfiguration)
		}
	}

	// This may or may not be calculated depending on seeding type, so we're recalculating it
	var nEffectivePlayers_UB2 int
	if seedingType == enums.LeaguePlayoffSeedingTypeStandard {
		// For standard, all participants play in R1, so UB2 participants are half of total.
		// numParticipants is already validated to be a power of two.
		nEffectivePlayers_UB2 = numParticipants / 2
	} else {
		// For BYES_ONLY and FULLY_SEEDED, calculate based on byes and UB R1 players.
		nEffectivePlayers_UB2 = len(playersGettingByes) + (len(playersStartingInUB1) / 2)
	}
	if nEffectivePlayers_UB2 <= 0 || !isPowerOfTwo(nEffectivePlayers_UB2) {
		return nil, fmt.Errorf("%w: Internal error: Calculated effective players for Upper Bracket Round 2 (%d) is not a positive power of two. This indicates a logic flaw in seeding validation.",
			common.ErrInternalService, nEffectivePlayers_UB2)
	}

	var currentUBRoundGames []*models.Game
	var currentLBRoundGames []*models.Game

	// --- First UB Round Game generation ---
	ubRoundNumber, lbRoundNumber, gameNumberInRound := 1, 1, 0
	for l, r := 0, len(playersStartingInUB1)-1; l <= r; l, r = l+1, r-1 {
		player1 := playersStartingInUB1[l]
		player2 := playersStartingInUB1[r]

		gameNumberInRound++
		bracketPositionStr := fmt.Sprintf("Upper Round %d: Game %d", ubRoundNumber, gameNumberInRound)

		// Create new game object
		newGame := &models.Game{
			ID:              uuid.New(),
			LeagueID:        league.ID,
			Player1ID:       player1.ID,
			Player2ID:       player2.ID,
			WinnerID:        nil,
			LoserID:         nil,
			Player1Wins:     0,
			Player2Wins:     0,
			RoundNumber:     ubRoundNumber,
			GroupNumber:     nil,
			GameType:        enums.GameTypePlayoffUpper,
			Status:          enums.GameStatusScheduled,
			BracketPosition: &bracketPositionStr,
		}
		generatedGames = append(generatedGames, newGame)
		currentUBRoundGames = append(currentUBRoundGames, newGame)
	}
	ubRoundNumber++

	// ---  First LB Round Game Generation ---
	gameNumberInRound = 0
	i := 0
	for i < len(currentUBRoundGames) {
		gameNumberInRound++
		bracketPositionStr := fmt.Sprintf("Lower Round %d: Game %d", lbRoundNumber, gameNumberInRound)
		newGameID := uuid.New()

		player1ID := uuid.Nil // Default to nil, will be set for FULLY_SEEDED if applicable
		player2ID := uuid.Nil // Default to nil, as it's the feeder from UB

		if seedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
			if i < len(playersStartingInLB1) {
				player1ID = playersStartingInLB1[i].ID
			}
			// Link the one UB game whose loser feeds into this LB game
			currentUBRoundGames[i].LoserToGameID = newGameID

			i++ // Increment by 1 for fully seeded
		} else { // For STANDARD and BYES_ONLY, losers of UB R1 play each other.
			// Both player IDs are nil, to be filled by losers of currentUBRoundGames[i] and currentUBRoundGames[i+1]
			// Link the two UB games whose losers feed into this LB game
			currentUBRoundGames[i].LoserToGameID = newGameID
			if i+1 < len(currentUBRoundGames) {
				currentUBRoundGames[i+1].LoserToGameID = newGameID
			}
			i += 2 // Increment by 2 for standard/byes
		}

		newGame := &models.Game{
			ID:              newGameID,
			LeagueID:        league.ID,
			Player1ID:       player1ID,
			Player2ID:       player2ID,
			WinnerID:        nil,
			LoserID:         nil,
			Player1Wins:     0,
			Player2Wins:     0,
			RoundNumber:     lbRoundNumber,
			GroupNumber:     nil,
			GameType:        enums.GameTypePlayoffLower,
			Status:          enums.GameStatusScheduled,
			BracketPosition: &bracketPositionStr,
		}

		currentLBRoundGames = append(currentLBRoundGames, newGame)
		generatedGames = append(generatedGames, newGame)
	}
	lbRoundNumber++

	// --- Subsequent round generation for upper and lower bracket ---
	totalUBRounds := getLog2(nextPowerOfTwo)
	totalLBRounds := 2 * (totalUBRounds - 1)
	j := 0
	// Loop until only one game remains in UB round and one in LB round
	// For every Upper bracket game from round 2 onwards,
	// We do one "LB-Drop" round and then one "LB-Survival" round,
	// until the Lower Final ("LB-Drop") round which consists of a singular game
	// In "LB-Drop" rounds, Losers from the UB round enter the LB
	// In "LB-Survival" rounds, the winners from previous LB round play against each other
	// The UB Final is the exception. It corresponds to a single "LB-Drop" round
	// i.e., the LB final.
	for ubRoundNumber <= totalUBRounds {
		// upper and lower bracket round pairs
		var nextUBRoundGames []*models.Game
		var nextLBRoundGames1 []*models.Game // For "LB-Drop"
		var nextLBRoundGames2 []*models.Game // For "LB-Survival"
		gameNumberInRound = 0

		// Upper Bracket round
		i := 0
		for i < len(currentUBRoundGames) {
			gameNumberInRound++
			bracketPositionStr := fmt.Sprintf("Upper Round %d: Game %d", ubRoundNumber, gameNumberInRound)
			newGameID := uuid.New()
			player1ID := uuid.Nil

			// for round 2 games, check the playersGettingByes array
			// set the new round 2 game's player1ID to the valid player if it exists in the array
			byeGranted := false
			if ubRoundNumber == 2 && len(playersGettingByes) != 0 && j < len(playersGettingByes) {
				// set uuid of player1ID
				player1ID = playersGettingByes[j].ID
				j++
				byeGranted = true
			}

			// link the previous round game(s) to this game
			// by setting winnerToGameID for those game(s)
			// 2 games when no bye is granted and 1 when bye is granted
			currentUBRoundGames[i].WinnerToGameID = newGameID
			if !byeGranted {
				currentUBRoundGames[i+1].WinnerToGameID = newGameID
			}

			newGame := &models.Game{
				ID:              newGameID,
				LeagueID:        league.ID,
				Player1ID:       player1ID,
				Player2ID:       uuid.Nil, // placeholder uuid.Nil
				WinnerID:        nil,
				LoserID:         nil,
				Player1Wins:     0,
				Player2Wins:     0,
				RoundNumber:     ubRoundNumber,
				GroupNumber:     nil,
				GameType:        enums.GameTypePlayoffUpper,
				Status:          enums.GameStatusScheduled,
				BracketPosition: &bracketPositionStr,
			}
			nextUBRoundGames = append(nextUBRoundGames, newGame)
			generatedGames = append(generatedGames, newGame)

			// if a bye was granted, we fed one game of previous round to this new round game
			// so we check the next one (i+1)
			if byeGranted {
				i += 1
			} else {
				// if bye wasn't granted, we fed two games of the previous round to this game
				// so we check the next pair of games (i+2 and i+3)
				i += 2
			}
		}
		currentUBRoundGames = nextUBRoundGames

		nUBRoundGames := len(currentUBRoundGames)
		// If only one game was generated for the current UB round, it's the Upper Final
		if nUBRoundGames == 1 {
			bracketPositionStr := "Upper Final"
			currentUBRoundGames[0].BracketPosition = &bracketPositionStr
		}
		ubRoundNumber++

		// "LB-Drop" Round
		i, k := 0, 0
		gameNumberInRound = 0
		for lbRoundNumber <= totalLBRounds && i < len(currentLBRoundGames) {
			gameNumberInRound++
			bracketPositionStr := fmt.Sprintf("Lower Round %d: Game %d", lbRoundNumber, gameNumberInRound)

			newGameID := uuid.New()

			// Construct game object
			newGame := &models.Game{
				ID:              newGameID,
				LeagueID:        league.ID,
				Player1ID:       uuid.Nil,
				Player2ID:       uuid.Nil,
				WinnerID:        nil,
				LoserID:         nil,
				Player1Wins:     0,
				Player2Wins:     0,
				RoundNumber:     lbRoundNumber,
				GroupNumber:     nil,
				GameType:        enums.GameTypePlayoffLower,
				Status:          enums.GameStatusScheduled,
				BracketPosition: &bracketPositionStr,
			}

			// Link the games that feed into this game
			// Loser from one of the current UB round game
			// Winner from one of the previous LB round game
			if k < len(currentUBRoundGames) {
				currentUBRoundGames[k].LoserToGameID = newGameID
				k++
			}
			currentLBRoundGames[i].WinnerToGameID = newGameID
			i++

			nextLBRoundGames1 = append(nextLBRoundGames1, newGame)
			generatedGames = append(generatedGames, newGame)
		}
		currentLBRoundGames = nextLBRoundGames1
		lbRoundNumber++

		nLBRoundGames := len(currentLBRoundGames)
		// If only one game was generated for the current LB round, it's the Lower Final
		if nLBRoundGames == 1 {
			bracketPositionStr := "Lower Final"
			currentLBRoundGames[0].BracketPosition = &bracketPositionStr
			break // No more LB rounds after this
		}

		if len(currentLBRoundGames) == 0 {
			break
		}

		// "LB-Survival" Round
		gameNumberInRound = 0
		for i := 0; i < len(currentLBRoundGames); i += 2 {
			gameNumberInRound++
			bracketPositionStr := fmt.Sprintf("Lower Round %d: Game %d", lbRoundNumber, gameNumberInRound)

			newGameID := uuid.New()

			// Construct game object
			newGame := &models.Game{
				ID:              newGameID,
				LeagueID:        league.ID,
				Player1ID:       uuid.Nil,
				Player2ID:       uuid.Nil,
				WinnerID:        nil,
				LoserID:         nil,
				Player1Wins:     0,
				Player2Wins:     0,
				RoundNumber:     lbRoundNumber,
				GroupNumber:     nil,
				GameType:        enums.GameTypePlayoffLower,
				Status:          enums.GameStatusScheduled,
				BracketPosition: &bracketPositionStr,
			}

			// Link the games that feed into this game
			// Winner from ith previous LB round game
			// Winner from (i+1)th previous LB round game
			currentLBRoundGames[i].WinnerToGameID = newGameID
			currentLBRoundGames[i+1].WinnerToGameID = newGameID

			nextLBRoundGames2 = append(nextLBRoundGames2, newGame)
			generatedGames = append(generatedGames, newGame)
		}
		currentLBRoundGames = nextLBRoundGames2
		lbRoundNumber++
	}

	// Grand Final
	bracketPositionStr := "Grand Final"
	grandFinalGameID := uuid.New()
	grandFinalGame := &models.Game{
		ID:          grandFinalGameID,
		LeagueID:    league.ID,
		Player1ID:   uuid.Nil, // Winner of Upper Final
		Player2ID:   uuid.Nil, // Winner of Lower Final
		WinnerID:    nil,
		LoserID:     nil,
		Player1Wins: 0,
		Player2Wins: 0,
		// Grand Final is technically not part of the upper bracket, but issokay
		RoundNumber:     ubRoundNumber, // Should be after the last UB round
		GroupNumber:     nil,
		GameType:        enums.GameTypePlayoffGrandFinal,
		Status:          enums.GameStatusScheduled,
		BracketPosition: &bracketPositionStr,
	}
	generatedGames = append(generatedGames, grandFinalGame)

	// Link UB Final and LB Final to Grand Final
	if len(currentUBRoundGames) == 1 {
		currentUBRoundGames[0].WinnerToGameID = grandFinalGameID
	}
	if len(currentLBRoundGames) == 1 {
		currentLBRoundGames[0].WinnerToGameID = grandFinalGameID
	}

	return generatedGames, nil
}

// getNextPowerOfTwo returns the smallest power of two greater than or equal to n.
// Returns 1 if n is 0 or negative, as 1 is the smallest power of two relevant for bracket sizing.
// e.g., for: n = 8, p = 8; n = 7, p = 8; n = 12, p = 16
func (s *gameServiceImpl) getClosestPowerOfTwo(N int) int {
	if N <= 0 {
		return 1
	}
	p := 1
	for p < N {
		p = p << 1 // p = p*2
	}
	return p
}

// getLog2 returns the floored integer log to the base 2 of N
func getLog2(N int) int {
	return bits.Len(uint(N)) - 1
}

// isPowerOfTwo checks if a number is a power of two.
func isPowerOfTwo(N int) bool {
	return N > 0 && (N&(N-1) == 0)
}

// getSeededPlayers prepares a list of players for playoff bracket generation.
// It first sorts players within their respective groups. Then, it selects
// a qualifying number from each group and interleaves them to determine
// their overall seeding for the playoffs (e.g., 1st from Group A, 1st from Group B,
// 2nd from Group A, 2nd from Group B, etc.).
// It returns the final list of players who will participate in the bracket.
func (s *gameServiceImpl) getSeededPlayers(league *models.League, playersByGroup [][]models.Player) ([]models.Player, error) {
	var qualifyingPlayers []models.Player

	// Determine how many players should qualify from each group.
	// This assumes PlayoffParticipantCount is a multiple of GroupCount,
	// or that any remainder is handled by league rules (e.g., wildcards, or simply fewer players).
	// Validation for this divisibility should ideally happen at league creation/update.
	numPlayersToQualifyPerGroup := league.Format.PlayoffParticipantCount / league.Format.GroupCount

	// 1. Sort players within each group
	for i := range playersByGroup {
		// Ensure the group is not empty before attempting to sort
		if len(playersByGroup[i]) == 0 { // should never be true
			log.Printf("INFO: (Service: getSeededPlayers) - Encountered an empty player group %d for league %s. Skipping group.", i+1, league.ID)
			continue
		}
		sortPlayers(playersByGroup[i]) // Sorts in place
	}

	// 2. Interleave the top qualifiers from each group
	// Iterate through the 'rank' within each group
	for rank := range numPlayersToQualifyPerGroup {
		// Then iterate through each group to pick the player at the current rank
		for groupIdx := range league.Format.GroupCount {
			// Check if this group has a player at the current rank
			if rank < len(playersByGroup[groupIdx]) { // should never be false
				qualifyingPlayers = append(qualifyingPlayers, playersByGroup[groupIdx][rank])
			}
		}
	}

	// Ensure we have exactly PlayoffParticipantCount players.
	// This check is crucial if numPlayersToQualifyPerGroup * GroupCount < PlayoffParticipantCount
	// or if some groups had fewer players than expected.
	if len(qualifyingPlayers) != league.Format.PlayoffParticipantCount { // should never be true
		log.Printf("ERROR: (Service: getSeededPlayers) - Mismatch in qualified players count. Expected %d, got %d for league %s. This might indicate an issue with league configuration or player data.",
			league.Format.PlayoffParticipantCount, len(qualifyingPlayers), league.ID)
		return nil, common.ErrInsufficientPlayersForPlayoffs
	}

	return qualifyingPlayers, nil
}

// sortPlayers sorts (inplace) a slice of models.Player based on seeding criteria.
// Primary sort: Wins (descending)
// Secondary sort: Losses (ascending)
// Tertiary sort: Player ID (arbitrary but consistent tie-breaker)
func sortPlayers(players []models.Player) {
	sort.Slice(players, func(i, j int) bool {
		// Primary sort: More wins come first (descending)
		if players[i].Wins != players[j].Wins {
			return players[i].Wins > players[j].Wins
		}

		// Secondary sort: Fewer losses come first (ascending)
		if players[i].Losses != players[j].Losses {
			return players[i].Losses < players[j].Losses
		}

		// Tertiary sort (tie-breaker): Use player ID for consistent ordering
		// (UUID comparison directly might not be stable, converting to string is safer for consistent sort)
		return players[i].ID.String() < players[j].ID.String()
	})
}

// generateRoundRobinGamesForGroup returns all the []*models.Game with Game.GroupNumber groupNumber,
// and a shuffled Game.RoundNumber.
// Does not persist to the database.
// Uses the Circle Method algorithm. https://en.wikipedia.org/wiki/Round-robin_tournament#Circle_method
// For groups with odd player counts, every round a player gets a bye (no game for that week)
func (s *gameServiceImpl) generateRoundRobinGamesForGroup(leagueID uuid.UUID, players []models.Player, groupNumber int) ([]*models.Game, error) {
	nActualPlayers := len(players)
	if nActualPlayers < 2 {
		// impossible since games cannot be scheduled in the first place
		// as draft cannot be started there's just one player
		// and games scheduling must happen after draft
		return nil, nil
	}

	playerIDsForSchedule := make([]uuid.UUID, nActualPlayers)
	for i, p := range players {
		playerIDsForSchedule[i] = p.ID
	}

	// if the group has an odd number of players, we create a dummy player
	// with uuid.Nil to indicate a bye
	if nActualPlayers%2 == 1 {
		playerIDsForSchedule = append(playerIDsForSchedule, uuid.Nil)
	}

	numPlayerInCircle := len(playerIDsForSchedule)
	numRounds := numPlayerInCircle - 1

	// assign fixed player and rotating players
	fixedPlayerID := playerIDsForSchedule[0]
	rotatingPlayers := make([]uuid.UUID, numPlayerInCircle-1)
	copy(rotatingPlayers, playerIDsForSchedule[1:]) // rest of the players

	var games []*models.Game // Game but with temporary group numbers that are later re-assigned
	for RoundIdx := range numRounds {
		// Pairings for the current round

		// Pair 1: Fixed Player vs Player opposite in the circle
		playerOppositeID := rotatingPlayers[len(rotatingPlayers)/2]
		if fixedPlayerID != uuid.Nil && playerOppositeID != uuid.Nil {
			games = append(games, &models.Game{
				LeagueID:    leagueID,
				Player1ID:   fixedPlayerID,
				Player2ID:   playerOppositeID,
				Status:      enums.GameStatusScheduled,
				GameType:    enums.GameTypeRegularSeason,
				RoundNumber: RoundIdx,
				GroupNumber: &groupNumber,
			})
		} else {
			// One of the players has a bye for this game
			// For regular season games, we do not make a game record of type Bye
			// since the player doesn't get an advantage due to the bye
			// They simply don't have a game to play this week
			// Nothing has to be done. Absence of a game for the week indicates a bye
			// log for debugging purposes
			byePlayerID := fixedPlayerID
			if byePlayerID == uuid.Nil {
				byePlayerID = playerOppositeID
			}
			fmt.Printf("INFO: (Service: generateRoundRobinGamesForGroup) Player %s (league %s) of group %d got a bye.", byePlayerID, leagueID, groupNumber)
		}

		// Remaining Pairs: Pair rest of the players with their opposite
		// match first half with the other half and don't include the fixed player's opposite
		for i := 0; i < len(rotatingPlayers)/2; i++ {
			p1ID := rotatingPlayers[i]
			p2ID := rotatingPlayers[len(rotatingPlayers)-1-i] // opposite player of 'i'
			if p1ID != uuid.Nil && p2ID != uuid.Nil {
				games = append(games, &models.Game{
					LeagueID:    leagueID,
					Player1ID:   p1ID,
					Player2ID:   p2ID,
					Status:      enums.GameStatusScheduled,
					GameType:    enums.GameTypeRegularSeason,
					RoundNumber: RoundIdx,
					GroupNumber: &groupNumber,
				})
			} else {
				// One of the players has a bye
				byePlayerID := p1ID
				if byePlayerID == uuid.Nil {
					byePlayerID = p2ID
				}
				fmt.Printf("INFO: (Service: generateRoundRobinGamesForGroup) Player %s (league %s) of group %d got a bye.", byePlayerID, leagueID, groupNumber)
			}
		}

		// Rotate players for the next round
		if len(rotatingPlayers) > 1 {
			lastPlayer := rotatingPlayers[len(rotatingPlayers)-1]
			// Shift all elements to right by 1
			copy(rotatingPlayers[1:], rotatingPlayers[:len(rotatingPlayers)-1])
			rotatingPlayers[0] = lastPlayer // move last player to first position
		}
	}

	// Create a slice of actual RoundNumbers (1-indexed) and shuffle it.
	actualRoundNumbers := make([]int, numRounds)
	for i := range numRounds {
		actualRoundNumbers[i] = i + 1
	}
	rand.Shuffle(numRounds, func(i, j int) {
		actualRoundNumbers[i], actualRoundNumbers[j] = actualRoundNumbers[j], actualRoundNumbers[i]
	})

	// Assign the shuffled actual RoundNumbers to games based on their conceptual RoundIdx.
	for i := range games {
		conceptualRoundIdx := games[i].RoundNumber
		if conceptualRoundIdx >= 0 && conceptualRoundIdx < numRounds {
			games[i].RoundNumber = actualRoundNumbers[conceptualRoundIdx]
		} else {
			// should never happen
			log.Printf("ERROR: (Service: generateRoundRobinGamesForGroup) - Invalid conceptual RoundIdx %d found in game for league %s, group %d", conceptualRoundIdx, leagueID, groupNumber)
			return nil, common.ErrInternalService
		}
	}

	return games, nil
}

func (s *gameServiceImpl) fetchLeagueResource(leagueID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrLeagueNotFound
		}
		return nil, common.ErrInternalService
	}
	return league, nil
}
