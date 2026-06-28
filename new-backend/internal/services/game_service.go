package services

import (
	"errors"
	"fmt"
	"log"
	"math/bits"
	"math/rand/v2"
	"sort"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameService interface {
	GetGameByID(ID uuid.UUID) (*models.Game, error)
	GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error)
	GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error)
	GenerateRegularSeasonGames(leagueID uuid.UUID) error
	GeneratePlayoffBracket(leagueID uuid.UUID) error

	ReportGameResult(gameID uuid.UUID, dto *requests.ReportGameRequestDTO) error
	FinalizeGameResult(gameID uuid.UUID, dto *requests.FinalizeGameRequestDTO) error
	SetLeagueService(leagueService LeagueService)
}

type gameServiceImpl struct {
	gameRepo      repositories.GameRepository
	leagueRepo    repositories.LeagueRepository
	playerRepo    repositories.PlayerRepository
	memberRepo    repositories.LeagueMemberRepository
	leagueService LeagueService
}

func NewGameService(
	gameRepo repositories.GameRepository,
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
	memberRepo repositories.LeagueMemberRepository,
) GameService {
	return &gameServiceImpl{
		gameRepo:   gameRepo,
		leagueRepo: leagueRepo,
		playerRepo: playerRepo,
		memberRepo: memberRepo,
	}
}

func (s *gameServiceImpl) SetLeagueService(leagueService LeagueService) {
	s.leagueService = leagueService
}

func (s *gameServiceImpl) GetGameByID(ID uuid.UUID) (*models.Game, error) {
	game, err := s.gameRepo.GetGameByID(ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", types.ErrGameNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", types.ErrInternalService, err)
	}
	return &game, nil
}

func (s *gameServiceImpl) GetGamesByLeague(leagueID uuid.UUID) ([]models.Game, error) {
	games, err := s.gameRepo.GetGamesByLeague(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", types.ErrGameNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", types.ErrInternalService, err)
	}
	return games, nil
}

func (s *gameServiceImpl) GetGamesByPlayer(playerID uuid.UUID) ([]models.Game, error) {
	games, err := s.gameRepo.GetGamesByPlayer(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", types.ErrGameNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", types.ErrInternalService, err)
	}

	return games, nil
}

// ReportGameResult allows a player to report the result of a game.
func (s *gameServiceImpl) ReportGameResult(gameID uuid.UUID, dto *requests.ReportGameRequestDTO) error {
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrGameNotFound
		}
		return fmt.Errorf("%w: %s", types.ErrInternalService, err.Error())
	}

	if game.Status != enums.GameStatusScheduled {
		return types.ErrConflict
	}

	// Determine loser ID
	var loserID uuid.UUID
	if dto.WinnerID == game.Player1ID {
		loserID = game.Player2ID
	} else if dto.WinnerID == game.Player2ID {
		loserID = game.Player1ID
	} else {
		return types.ErrInvalidInput // Winner must be one of the players in the game
	}

	// Ensure winner and loser are distinct
	if dto.WinnerID == loserID {
		return fmt.Errorf("%w: winner and loser cannot be the same", types.ErrInvalidInput)
	}

	// Ensure scores are not tied if a winner is provided
	if dto.Player1Wins == nil || dto.Player2Wins == nil {
		return fmt.Errorf("%w: player wins must be provided", types.ErrInvalidInput)
	}
	if *dto.Player1Wins == *dto.Player2Wins {
		return fmt.Errorf("%w: scores cannot be tied for a reported result", types.ErrInvalidInput)
	}

	if err := s.gameRepo.UpdateGameReport(gameID, loserID, dto); err != nil {
		return fmt.Errorf("ReportGameResult: failed to update game report %s: %w", gameID, err)
	}

	return nil
}

// FinalizeGameResult allows league staff to approve, submit, or retroactively edit a game result.
func (s *gameServiceImpl) FinalizeGameResult(gameID uuid.UUID, dto *requests.FinalizeGameRequestDTO) error {
	// Fetch game to determine loser ID
	game, err := s.gameRepo.GetGameByID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrGameNotFound
		}
		return fmt.Errorf("%w: %s", types.ErrInternalService, err.Error())
	}

	if !(game.Status == enums.GameStatusApprovalPending || game.Status == enums.GameStatusDisputed) {
		return types.ErrConflict
	}

	// Determine loser ID for the final result
	var loserID uuid.UUID
	if dto.WinnerID == game.Player1ID {
		loserID = game.Player2ID
	} else if dto.WinnerID == game.Player2ID {
		loserID = game.Player1ID
	} else {
		return types.ErrInvalidInput // Winner must be one of the players in the game
	}

	// Ensure scores are not tied
	if dto.Player1Wins == nil || dto.Player2Wins == nil {
		return fmt.Errorf("%w: player wins must be provided", types.ErrInvalidInput)
	}
	if *dto.Player1Wins == *dto.Player2Wins {
		return fmt.Errorf("%w: scores cannot be tied for a finalized result", types.ErrInvalidInput)
	}

	// RBAC Check is handled in controller, service layer proceeds with business logic

	err = s.gameRepo.FinalizeGameAndUpdateStats(&game, loserID, dto)
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

	// Check if games have already been generated for this league
	gamesExist, err := s.gameRepo.HasGames(leagueID, enums.GameTypeRegularSeason)
	if err != nil {
		log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - Failed to check for existing games for league %s: %v\n", leagueID, err)
		return types.ErrInternalService
	}
	if gamesExist {
		return types.ErrGamesAlreadyGenerated
	}

	// League needs to be in POST_DRAFT status and not a BRACKET_ONLY Season League
	if league.Status != enums.LeagueStatusPostDraft && league.Format.SeasonType == enums.LeagueSeasonTypeBracketOnly {
		log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - League %s not in valid state to generate season bracket: %v\n", leagueID, err)
		return types.ErrInvalidState
	}

	membersByGroupNumber := make([][]models.LeagueMember, league.Format.GroupCount)
	for i := 0; i < league.Format.GroupCount; i++ {
		members, err := s.memberRepo.GetByLeagueAndGroup(league.ID, i+1)
		if err != nil {
			log.Printf("ERROR: (Service: GenerateRegularSeasonGames) - Repository error fetching Members by League %s with Group Number %d: %v\n", league.ID, i+1, err)
			return types.ErrInternalService
		}
		membersByGroupNumber[i] = members
	}

	var allGeneratedGames []*models.Game
	for groupIndex, membersInGroup := range membersByGroupNumber {
		groupNumber := groupIndex + 1
		games, err := s.generateRoundRobinGamesForGroup(league.ID, membersInGroup, groupNumber)
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
			return types.ErrInternalService
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

	membersByGroup := make([][]models.LeagueMember, league.Format.GroupCount)
	for i := 0; i < league.Format.GroupCount; i++ {
		membersOfGroupX, err := s.memberRepo.GetByLeagueAndGroup(league.ID, i+1)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket): error fetching members of group %d for league %s: %v", i+1, league.ID, err)
			return fmt.Errorf("failed to fetch members for group %d: %w", i+1, err)
		}
		membersByGroup[i] = membersOfGroupX
	}

	seededMembers, err := s.getSeededPlayers(league, membersByGroup)
	if err != nil {
		log.Printf("ERROR: (Service: GeneratePlayoffBracket): error seeding members for playoffs for league %s: %v", league.ID, err)
		return err
	}

	var generatedGames []*models.Game
	if league.Format.PlayoffType == enums.LeaguePlayoffTypeSingleElim {
		if league.Format.PlayoffSeedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
			return fmt.Errorf("%w: %s and %s are incompatible playoff options",
				types.ErrInvalidLeagueConfiguration,
				enums.LeaguePlayoffTypeSingleElim,
				enums.LeaguePlayoffSeedingTypeFullySeeded)
		}
		generatedGames, err = s.generateSingleEliminationBracket(league, seededMembers)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Error generating single elimination bracket for league %s: %v\n", leagueID, err)
			return err
		}
	} else {
		generatedGames, err = s.generateDoubleEliminationBracket(league, seededMembers)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Error generating single elimination bracket for league %s: %v\n", leagueID, err)
			return err
		}
	}
	if len(generatedGames) > 0 {
		err = s.gameRepo.CreateGames(generatedGames)
		if err != nil {
			log.Printf("ERROR: (Service: GeneratePlayoffBracket) - Repository error creating games for league %s: %v\n", leagueID, err)
			return types.ErrInternalService
		}
	}

	return nil
}

// PRIVATE HELPERS

// generateSingleEliminationBracket generates the games for the single elimination bracket
// It takes into account changes introduced by various Format.PlayoffSeedingType
// returns a slice of all the generated Games and an error if generation failed
func (s *gameServiceImpl) generateSingleEliminationBracket(league *models.League, seededMembers []models.LeagueMember) ([]*models.Game, error) {
	var generatedGames []*models.Game
	numParticipants := len(seededMembers)
	if numParticipants == 0 {
		return nil, fmt.Errorf("no members provided for bracket generation")
	}

	// Each round must have a power of 2 # of participants
	// If this is not the case, bye "psuedo-players" (they always lose) can be added to make it a power of 2 (i.e., naturalByesNeeded)
	seedingType := league.Format.PlayoffSeedingType
	nextPowerOfTwo := s.getClosestPowerOfTwo(numParticipants)
	naturalByesNeeded := nextPowerOfTwo - numParticipants

	if league.Format.PlayoffSeedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
		return nil, fmt.Errorf("%w: FULLY_SEEDED brackets not supported for single elimination. Use BYES_ONLY", types.ErrInvalidLeagueConfiguration)
	}

	if naturalByesNeeded > 0 && seedingType == enums.LeaguePlayoffSeedingTypeStandard {
		return nil, fmt.Errorf("%w: Cannot construct single elimination bracket. Either add %d participants or enable 'byes' for %d players",
			types.ErrInvalidLeagueConfiguration,
			naturalByesNeeded, naturalByesNeeded)
	}
	if naturalByesNeeded != league.Format.PlayoffByesCount {
		return nil, fmt.Errorf("%w: Bye count is set to %d but valid is %d for %d participants", types.ErrInvalidLeagueConfiguration, league.Format.PlayoffByesCount, naturalByesNeeded, numParticipants)
	}

	var membersGettingByes []models.LeagueMember
	var membersInRound1 []models.LeagueMember
	if seedingType == enums.LeaguePlayoffSeedingTypeByesOnly {
		membersGettingByes = seededMembers[:league.Format.PlayoffByesCount]
		membersInRound1 = seededMembers[league.Format.PlayoffByesCount:]
	} else {
		membersInRound1 = seededMembers
	}

	var currentRoundGames []*models.Game
	roundNumber, gameNumberInRound := 1, 0
	for l, r := 0, len(membersInRound1)-1; l <= r; l, r = l+1, r-1 {
		member1 := membersInRound1[l]
		member2 := membersInRound1[r]
		var bracketPositionStr string

		gameNumberInRound++
		bracketPositionStr = fmt.Sprintf("Round %d: Game %d", roundNumber, gameNumberInRound)

		newGame := &models.Game{
			ID:              uuid.New(),
			LeagueID:        league.ID,
			Player1ID:       member1.ID,
			Player2ID:       member2.ID,
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
			if roundNumber == 2 && j < len(membersGettingByes) {
				newGamePlayer1ID = membersGettingByes[j].ID
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

func (s *gameServiceImpl) generateDoubleEliminationBracket(league *models.League, seededMembers []models.LeagueMember) ([]*models.Game, error) {
	var generatedGames []*models.Game
	numParticipants := len(seededMembers)
	if numParticipants == 0 {
		return nil, fmt.Errorf("no members provided for bracket generation")
	}

	seedingType := league.Format.PlayoffSeedingType
	byeCount := league.Format.PlayoffByesCount
	nextPowerOfTwo := s.getClosestPowerOfTwo(numParticipants)
	membersGettingByes := []models.LeagueMember{}
	membersStartingInUB1 := []models.LeagueMember{}
	membersStartingInLB1 := []models.LeagueMember{}
	remainingMembers := []models.LeagueMember{}

	if byeCount != 0 && seedingType == enums.LeaguePlayoffSeedingTypeStandard {
		return nil, fmt.Errorf("%w: Bye count must be 0 for Standard Seeded brackets", types.ErrInvalidLeagueConfiguration)
	}

	if byeCount > 0 {
		membersGettingByes = seededMembers[:byeCount]
		remainingMembers = seededMembers[byeCount:]
	} else {
		remainingMembers = seededMembers
	}

	if seedingType == enums.LeaguePlayoffSeedingTypeByesOnly {
		naturalByesNeeded := nextPowerOfTwo - numParticipants
		if naturalByesNeeded != byeCount {
			return nil, fmt.Errorf("%w: For BYES_ONLY Double Elimination with %d particpants, number of byes (Current: %d) allowed is %d",
				types.ErrInvalidLeagueConfiguration, numParticipants, byeCount, naturalByesNeeded)
		}
	} else if seedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
		nMembersGettingByes := len(membersGettingByes)
		nRemainingMembers := len(remainingMembers)

		if nRemainingMembers%3 != 0 {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED Double Elimination, number of members that don't get a bye (Current: %d) must be divisible by 3",
				types.ErrInvalidLeagueConfiguration, nRemainingMembers)
		}

		nMembers_UB1 := (2 * nRemainingMembers) / 3
		if nMembers_UB1 <= 0 || nMembers_UB1%2 != 0 {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED Double Elimination, number of members starting in Upper Bracket Round 1 (Current: %d) must be a positive even number",
				types.ErrInvalidLeagueConfiguration, nMembersGettingByes)
		}

		nEffectiveMembers_UB2 := byeCount + (nMembers_UB1 / 2)
		if nEffectiveMembers_UB2 <= 0 || !isPowerOfTwo(nEffectiveMembers_UB2) {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED Double Elimination, the total effective number of members in Upper Bracket 2 (Current: %d) must be a positive power of two",
				types.ErrInvalidLeagueConfiguration, nEffectiveMembers_UB2)
		}

		if nMembers_UB1 < 2*byeCount {
			return nil, fmt.Errorf("%w: For FULLY_SEEDED double elimination, the number of members receiving byes (Current: %d) cannot exceed the number of members starting in Upper Bracket Round 1 (%d) to maintain a balanced bracket structure",
				types.ErrInvalidLeagueConfiguration, byeCount, nMembers_UB1)
		}

		membersStartingInUB1 = remainingMembers[:nMembers_UB1]
		membersStartingInLB1 = remainingMembers[nMembers_UB1:]
	} else {
		membersStartingInUB1 = seededMembers
		if !isPowerOfTwo(numParticipants) {
			return nil, fmt.Errorf("%w: For STANDARD double elimination, number of participants must be a positive power of two", types.ErrInvalidLeagueConfiguration)
		}
	}

	var nEffectiveMembers_UB2 int
	if seedingType == enums.LeaguePlayoffSeedingTypeStandard {
		nEffectiveMembers_UB2 = numParticipants / 2
	} else {
		nEffectiveMembers_UB2 = len(membersGettingByes) + (len(membersStartingInUB1) / 2)
	}
	if nEffectiveMembers_UB2 <= 0 || !isPowerOfTwo(nEffectiveMembers_UB2) {
		return nil, fmt.Errorf("%w: Internal error: Calculated effective members for Upper Bracket Round 2 (%d) is not a positive power of two. This indicates a logic flaw in seeding validation",
			types.ErrInternalService, nEffectiveMembers_UB2)
	}

	var currentUBRoundGames []*models.Game
	var currentLBRoundGames []*models.Game

	ubRoundNumber, lbRoundNumber, gameNumberInRound := 1, 1, 0
	for l, r := 0, len(membersStartingInUB1)-1; l <= r; l, r = l+1, r-1 {
		member1 := membersStartingInUB1[l]
		member2 := membersStartingInUB1[r]

		gameNumberInRound++
		bracketPositionStr := fmt.Sprintf("Upper Round %d: Game %d", ubRoundNumber, gameNumberInRound)

		newGame := &models.Game{
			ID:              uuid.New(),
			LeagueID:        league.ID,
			Player1ID:       member1.ID,
			Player2ID:       member2.ID,
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

	gameNumberInRound = 0
	i := 0
	for i < len(currentUBRoundGames) {
		gameNumberInRound++
		bracketPositionStr := fmt.Sprintf("Lower Round %d: Game %d", lbRoundNumber, gameNumberInRound)
		newGameID := uuid.New()

		player1ID := uuid.Nil
		player2ID := uuid.Nil

		if seedingType == enums.LeaguePlayoffSeedingTypeFullySeeded {
			if i < len(membersStartingInLB1) {
				player1ID = membersStartingInLB1[i].ID
			}
			currentUBRoundGames[i].LoserToGameID = newGameID
			i++
		} else {
			currentUBRoundGames[i].LoserToGameID = newGameID
			if i+1 < len(currentUBRoundGames) {
				currentUBRoundGames[i+1].LoserToGameID = newGameID
			}
			i += 2
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

	totalUBRounds := getLog2(nextPowerOfTwo)
	totalLBRounds := 2 * (totalUBRounds - 1)
	j := 0
	for ubRoundNumber <= totalUBRounds {
		var nextUBRoundGames []*models.Game
		var nextLBRoundGames1 []*models.Game
		var nextLBRoundGames2 []*models.Game
		gameNumberInRound = 0

		i := 0
		for i < len(currentUBRoundGames) {
			gameNumberInRound++
			bracketPositionStr := fmt.Sprintf("Upper Round %d: Game %d", ubRoundNumber, gameNumberInRound)
			newGameID := uuid.New()
			player1ID := uuid.Nil

			byeGranted := false
			if ubRoundNumber == 2 && len(membersGettingByes) != 0 && j < len(membersGettingByes) {
				player1ID = membersGettingByes[j].ID
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
func (s *gameServiceImpl) getSeededPlayers(league *models.League, membersByGroup [][]models.LeagueMember) ([]models.LeagueMember, error) {
	var qualifyingMembers []models.LeagueMember

	numMembersToQualifyPerGroup := league.Format.PlayoffParticipantCount / league.Format.GroupCount

	for i := range membersByGroup {
		if len(membersByGroup[i]) == 0 {
			log.Printf("INFO: (Service: getSeededPlayers) - Encountered an empty member group %d for league %s. Skipping group.\n", i+1, league.ID)
			continue
		}
		sortMembers(membersByGroup[i])
	}

	for rank := range numMembersToQualifyPerGroup {
		for groupIdx := range league.Format.GroupCount {
			if rank < len(membersByGroup[groupIdx]) {
				qualifyingMembers = append(qualifyingMembers, membersByGroup[groupIdx][rank])
			}
		}
	}

	if len(qualifyingMembers) != league.Format.PlayoffParticipantCount {
		log.Printf("ERROR: (Service: getSeededPlayers) - Mismatch in qualified members count. Expected %d, got %d for league %s.\n",
			league.Format.PlayoffParticipantCount, len(qualifyingMembers), league.ID)
		return nil, types.ErrInsufficientPlayersForPlayoffs
	}

	return qualifyingMembers, nil
}

func sortMembers(members []models.LeagueMember) {
	sort.Slice(members, func(i, j int) bool {
		if members[i].Wins != members[j].Wins {
			return members[i].Wins > members[j].Wins
		}
		if members[i].Losses != members[j].Losses {
			return members[i].Losses < members[j].Losses
		}
		return members[i].ID.String() < members[j].ID.String()
	})
}

func (s *gameServiceImpl) generateRoundRobinGamesForGroup(leagueID uuid.UUID, members []models.LeagueMember, groupNumber int) ([]*models.Game, error) {
	nActualMembers := len(members)
	if nActualMembers < 2 {
		return nil, nil
	}

	memberIDsForSchedule := make([]uuid.UUID, nActualMembers)
	for i, m := range members {
		memberIDsForSchedule[i] = m.ID
	}

	if nActualMembers%2 == 1 {
		memberIDsForSchedule = append(memberIDsForSchedule, uuid.Nil)
	}

	numMembersInCircle := len(memberIDsForSchedule)
	numRounds := numMembersInCircle - 1

	fixedMemberID := memberIDsForSchedule[0]
	rotatingMembers := make([]uuid.UUID, numMembersInCircle-1)
	copy(rotatingMembers, memberIDsForSchedule[1:])

	var games []*models.Game
	for RoundIdx := range numRounds {
		memberOppositeID := rotatingMembers[len(rotatingMembers)/2]
		if fixedMemberID != uuid.Nil && memberOppositeID != uuid.Nil {
			games = append(games, &models.Game{
				LeagueID:    leagueID,
				Player1ID:   fixedMemberID,
				Player2ID:   memberOppositeID,
				Status:      enums.GameStatusScheduled,
				GameType:    enums.GameTypeRegularSeason,
				RoundNumber: RoundIdx,
				GroupNumber: &groupNumber,
			})
		} else {
			byeMemberID := fixedMemberID
			if byeMemberID == uuid.Nil {
				byeMemberID = memberOppositeID
			}
			fmt.Printf("\nINFO: (Service: generateRoundRobinGamesForGroup): Member %s (league %s) of group %d got a bye.\n", byeMemberID, leagueID, groupNumber)
		}

		for i := 0; i < len(rotatingMembers)/2; i++ {
			p1ID := rotatingMembers[i]
			p2ID := rotatingMembers[len(rotatingMembers)-1-i]
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
				byeMemberID := p1ID
				if byeMemberID == uuid.Nil {
					byeMemberID = p2ID
				}
				fmt.Printf("INFO: (Service: generateRoundRobinGamesForGroup) Member %s (league %s) of group %d got a bye.", byeMemberID, leagueID, groupNumber)
			}
		}

		if len(rotatingMembers) > 1 {
			lastMember := rotatingMembers[len(rotatingMembers)-1]
			copy(rotatingMembers[1:], rotatingMembers[:len(rotatingMembers)-1])
			rotatingMembers[0] = lastMember
		}
	}

	actualRoundNumbers := make([]int, numRounds)
	for i := range numRounds {
		actualRoundNumbers[i] = i + 1
	}
	rand.Shuffle(numRounds, func(i, j int) {
		actualRoundNumbers[i], actualRoundNumbers[j] = actualRoundNumbers[j], actualRoundNumbers[i]
	})

	for i := range games {
		conceptualRoundIdx := games[i].RoundNumber
		if conceptualRoundIdx >= 0 && conceptualRoundIdx < numRounds {
			games[i].RoundNumber = actualRoundNumbers[conceptualRoundIdx]
		} else {
			log.Printf("ERROR: (Service: generateRoundRobinGamesForGroup) - Invalid conceptual RoundIdx %d found in game for league %s, group %d", conceptualRoundIdx, leagueID, groupNumber)
			return nil, types.ErrInternalService
		}
	}

	return games, nil
}

func (s *gameServiceImpl) fetchLeagueResource(leagueID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		return nil, types.ErrInternalService
	}
	return league, nil
}
