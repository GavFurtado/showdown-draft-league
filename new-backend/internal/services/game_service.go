package services

import (
	"errors"
	"fmt"
	"log"
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

	if (league.Format.SeasonType == enums.LeagueSeasonTypeBracketOnly && league.Status != enums.LeagueStatusPostDraft) ||
		(league.Format.SeasonType == enums.LeagueSeasonTypeHybrid && league.Status != enums.LeagueStatusPostRegularSeason) {
		log.Printf("ERROR: (Service: GeneratePlayoffBracket) - League not in valid status to generate playoff bracket.\n")
		return err
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

	seededPlayers, err := getSeededPlayers(league, playersByGroup)

	return nil
}

// PRIVATE HELPERS
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
	for rank := 0; rank < numPlayersToQualifyPerGroup; rank++ {
		// Then iterate through each group to pick the player at the current rank
		for groupIdx := 0; groupIdx < league.Format.GroupCount; groupIdx++ {
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
	numActualPlayers := len(players)
	if numActualPlayers < 2 {
		// impossible since games cannot be scheduled in the first place
		// as draft cannot be started there's just one player
		// and games scheduling must happen after draft
		return nil, nil
	}

	playerIDsForSchedule := make([]uuid.UUID, numActualPlayers)
	for i, p := range players {
		playerIDsForSchedule[i] = p.ID
	}

	// if the group has an odd number of players, we create a dummy player
	// with uuid.Nil to indicate a bye
	if numActualPlayers%2 == 1 {
		playerIDsForSchedule = append(playerIDsForSchedule, uuid.Nil)
	}

	numPlayerInCircle := len(playerIDsForSchedule)
	numRounds := numPlayerInCircle - 1

	// assign fixed player and rotating players
	fixedPlayerID := playerIDsForSchedule[0]
	rotatingPlayers := make([]uuid.UUID, numPlayerInCircle-1)
	copy(rotatingPlayers, playerIDsForSchedule[1:]) // rest of the players

	var games []*models.Game // Game but with temporary group numbers that are later re-assigned
	for RoundIdx := 0; RoundIdx < numRounds; RoundIdx++ {
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
