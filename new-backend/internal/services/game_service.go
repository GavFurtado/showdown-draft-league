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
	GenerateRegularSeasonGames(leagueID uuid.UUID) error
	GeneratePlayoffBracket(leagueID uuid.UUID) error
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

	var allGeneratedGames []models.Game
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

	var generatedGames []models.Game
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
func (s *gameServiceImpl) generateSingleEliminationBracket(league *models.League, seededPlayers []models.Player) ([]models.Game, error) {
	var generatedGames []models.Game
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

	// First Round Game generation logic
	// (the first round is handled separately due to the possibility of byes going into the following round)
	var currentRoundGames []models.Game
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
		newGame := models.Game{
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

	// For subsequent rounds, we generate placeholder game objects
	// with no player IDs filled in with exception to players
	// that were granted a bye to round 2
	totalRounds := bits.Len(uint(nextPowerOfTwo)) - 1 // integer log 2 (floor)
	j := 0
	for roundNumber := 2; roundNumber <= totalRounds; roundNumber++ {
		var nextRoundGames []models.Game
		gameNumberInRound := 0

		// Each pair of games from the previous round feeds into one game in the new round
		i := 0
		for i < len(currentRoundGames) {
			gameNumberInRound++
			bracketPositionStr := fmt.Sprintf("Round %d: Game %d", roundNumber, gameNumberInRound)

			newGameID := uuid.New()
			newGamePlayer1ID := uuid.Nil

			byeGranted := false
			// for round 2 games, check the playersGettingByesArray
			// if yes, set the new round 2 game's player1ID to the valid player
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

			newGame := models.Game{
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

	// rename the last game to "Grand Final"
	if len(generatedGames) > 0 {
		bracketPositionStr := "Grand Final"
		generatedGames[len(generatedGames)-1].BracketPosition = &bracketPositionStr
	}

	return generatedGames, nil
}

func (s *gameServiceImpl) generateDoubleEliminationBracket(league *models.League, seededPlayers []models.Player) ([]models.Game, error) {
	// Upper Bracket (UB); Lower Bracket (LB)
	var generatedGames []models.Game
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

	// initial validation of numParticipant
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

	} else { // Standard seeding
		if !isPowerOfTwo(numParticipants) {
			return nil, fmt.Errorf("%w: For STANDARD double elimination, number of participants must be a positive power of two", common.ErrInvalidLeagueConfiguration)
		}
	}

	return generatedGames, nil
}

// getNextPowerOfTwo returns the smallest power of two greater than or equal to n.
// Returns 1 if n is 0 or negative, as 1 is the smallest power of two relevant for bracket sizing.
// e.g., for: n = 8, p = 8; n = 7, p = 8; n = 12, p = 16
func (s *gameServiceImpl) getClosestPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	p := 1
	for p < n {
		p = p << 1 // p = p*2
	}
	return p
}

// isPowerOfTwo checks if a number is a power of two.
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1) == 0)
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
func (s *gameServiceImpl) generateRoundRobinGamesForGroup(leagueID uuid.UUID, players []models.Player, groupNumber int) ([]models.Game, error) {
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

	var games []models.Game // Game but with temporary group numbers that are later re-assigned
	for RoundIdx := range numRounds {
		// Pairings for the current round

		// Pair 1: Fixed Player vs Player opposite in the circle
		playerOppositeID := rotatingPlayers[len(rotatingPlayers)/2]
		if fixedPlayerID != uuid.Nil && playerOppositeID != uuid.Nil {
			games = append(games, models.Game{
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
				games = append(games, models.Game{
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
