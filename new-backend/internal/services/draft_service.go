package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
)

type DraftService interface {
	StartDraft(leagueID uuid.UUID, TurnTimeLimit int) (*models.Draft, error)
	MakePick(currentUser *models.User, league *models.League, input *common.DraftMakePickDTO) error
	// SkipTurn(leagueID uuid.UUID, currentUser *models.User) (*models.Draft, error)
	// StartTradingPeriod(leagueID uuid.UUID, currentUser *models.User) (*models.League, error)
	// EndTradingPeriod(leagueID uuid.UUID, currentUser *models.User) (*models.League, error)

	// DropFreeAgent(leagueID, draftedPokemonID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
	// PickupFreeAgent(leagueID, pokemonSpeciesID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
}

type draftServiceImpl struct {
	draftRepo          repositories.DraftRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	draftedPokemonRepo repositories.DraftedPokemonRepository
	webhookService     *WebhookService
}

func NewDraftService(
	leagueRepo repositories.LeagueRepository,
	leaguePokemonRepo repositories.LeaguePokemonRepository,
	draftRepo repositories.DraftRepository,
	draftedPokemonRepo repositories.DraftedPokemonRepository,
	playerRepo repositories.PlayerRepository,
	webhookService *WebhookService,
) DraftService {
	return &draftServiceImpl{
		draftRepo:          draftRepo,
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
		draftedPokemonRepo: draftedPokemonRepo,
		webhookService:     webhookService,
	}
}

// player permission rbac.PermissionCreateDraft
func (s *draftServiceImpl) StartDraft(leagueID uuid.UUID, TurnTimeLimit int) (*models.Draft, error) {
	// Retrieve the league
	league, err := (s.leagueRepo).GetLeagueByID(leagueID)
	if err != nil || league == nil {
		log.Printf("LOG: (Error: DraftService.StartDraft) - Could not get league %s: %v\n", leagueID, err)
		return nil, common.ErrLeagueNotFound
	}

	// Retrieve players in the league, sorted by draft position
	players, err := s.playerRepo.GetPlayersByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (Error: DraftService.StartDraft) - Could not get players for league %s: %v\n", leagueID, err)
		return nil, common.ErrInternalService
	}

	if len(players) == 0 {
		log.Printf("LOG: (Error: DraftService.StartDraft) - No players found for league %s\n", leagueID)
		return nil, common.ErrNoPlayerForDraft
	}

	// Initialize the Draft model
	firstPlayerID := players[0].ID
	currTime := time.Now()

	draft := &models.Draft{
		LeagueID:                    leagueID,
		Status:                      enums.DraftStatusOngoing,
		CurrentRound:                1,
		CurrentPickInRound:          1,
		CurrentPickOnClock:          1, // formula: ((CurrentRound - 1)*PlayerCount + CurrentPickInRound)
		CurrentTurnPlayerID:         &firstPlayerID,
		CurrentTurnStartTime:        &currTime,
		PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks), // map[uuid.UUID][]int
		TurnTimeLimit:               TurnTimeLimit,
	}

	// Save the Draft model
	if err := s.draftRepo.CreateDraft(draft); err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Failed to create draft for league %s: %v\n", leagueID, err)
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	// Update the league status to DRAFTING
	league.Status = enums.LeagueStatusDrafting
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("(Error: DraftService.StartDraft) - Failed to update league status for league %s: %v\n", leagueID, err)
		// TODO: Consider rolling back draft creation if this fails
		return nil, fmt.Errorf("failed to update league status: %w", err)
	}

	// Send an initial webhook notification
	// TODO: Implement webhook message creation logic
	// if err := (*s.webhookService).SendWebhookMessage(league.DiscordWebhookURL, "Draft has started!"); err != nil {
	// 	log.Printf("(Warning: DraftService.StartDraft) - Failed to send webhook for league %s: %v\n", leagueID, err)
	// 	// Continue execution, webhook failure shouldn't stop the draft
	// }

	return draft, nil
}

// only used for the initial draft (not free agent transactions)

// NOTE: this stuff could possibly need a rework so it wasn't touched in the rbac refactor.
// current problem is that AccumulatedPicks isn't being used in the pick logic
// would require a change to the request. request would now send the requestedDraftPickNumber and that would need to be validated.

// MakePick makes one or more picks (if accumulated) in a league's draft when league;
// Different from ForcePick, MakePick does all the required checks (there's a lot of checks) and validates the input
func (s *draftServiceImpl) MakePick(currentUser *models.User, league *models.League, input *common.DraftMakePickDTO) error {
	// fetch draft for league
	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case common.ErrDraftNotFound:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) draft for leagueID %s not found: %v\n", currentUser.ID, league.ID, err)
		case common.ErrInternalService:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) error fetching draft: %v\n", currentUser.ID, err)
		}
		return err
	}

	player, err := s.fetchPlayerResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case common.ErrPlayerNotFound:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) Player in league %s not found: %v\n", currentUser.ID, league.ID, err)
		case common.ErrInternalService:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) Error fetching player in league %s: %v\n", currentUser.ID, league.ID, err)
		}
		return err
	}

	// early checks to prevent a potentially expensive check
	// check if number of requested picks is valid for the player
	if input.RequestedPickCount > len(draft.PlayersWithAccumulatedPicks[player.ID])+1 {
		log.Printf("LOG: (DraftService: MakePick) -  (user %s) Player %s requested too many draft picks\n", currentUser.ID, player.ID)
		return common.ErrTooManyRequestedPicks
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		log.Printf("LOG: (DraftService: MakePick) - (user %s) league %s is not in drafting status: %v", currentUser.ID, league.ID, err)
		return common.ErrInvalidState
	}

	// check if it's the right player's turn
	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		log.Printf("LOG: (DraftService: MakePick) - player %s tried to draft when it isn't their turn. Current Turn: Player %s\n", currentTurnPlayerID, *draft.CurrentTurnPlayerID)
		return common.ErrUnauthorized
	}
	// END early checks

	// fetch all the leaguePokemon requested
	// potentially expensive
	allRequestedLeaguePokemon, err := s.fetchRequestedPokemon(league.ID, input)
	if err != nil {
		switch err {
		case common.ErrLeaguePokemonNotFound:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) One or more League Pokemon were not found: %v\n", currentUser.ID, err)
		case common.ErrConflict:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) One or more League Pokemon are not available for drafting: %v\n", currentUser.ID, err)
		case common.ErrInternalService:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) error fetching requested league pokemon for league %s: %v\n", currentUser.ID, league.ID, err)
		}
		return err
	}

	// get PlayerCount; needed in multiple places
	playerCount, err := s.playerRepo.GetPlayerCountByLeague(league.ID)
	if err != nil {
		log.Printf("DraftService: MakePick - failed to get player count for league %s: %v\n", league.ID, err)
		return common.ErrInternalService
	}
	if playerCount == 0 { // this should never happen if the draft has started or if the league even exists
		log.Printf("DraftService: MakePick - no players in league %d. (Unreachable Code)\n", league.ID)
		return common.ErrInternalService
	}

	playerRosterSize, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(player.ID)
	if err != nil {
		log.Printf("DraftService: MakePick - Failed to get player %d roster size\n", player.ID)
		return err
	}
	totalRequestedCost := s.getTotalCostForPicks(allRequestedLeaguePokemon)

	// perform remaining validation
	currentPickSlotUsed, err := s.validatePicksAndCheckCurrentPickSlotUsed(draft, player, league,
		input, totalRequestedCost, int(playerRosterSize))
	if err != nil {
		switch err {
		case common.ErrInvalidInput:
			log.Printf("LOG: (DraftService: MakePick): (user %s; league %s) Invalid pick number in request: %v\n", currentUser.ID, league.ID, err)
		case common.ErrInsufficientDraftPoints:
			log.Printf("LOG: (DraftService: MakePick): (user %s; league %s) Insufficient draft points (%d) for transaction: %v\n", currentUser.ID, league.ID, player.DraftPoints, err)
		}
		return err
	}

	err = s.executePickTransactions(draft, league, player, allRequestedLeaguePokemon, input, playerCount, totalRequestedCost)
	if err != nil {
		log.Printf("LOG: (DraftService: MakePick): (user %s; league %s) Batch transaction unsucessful: %v\n", currentUser.ID, league.ID, err)
		return err
	}

	// get all players to change set the CurrentPlayer's turn for the next one
	allPlayers, err := s.playerRepo.GetPlayersByLeague(draft.LeagueID)
	if err != nil {
		log.Printf("DraftService: MakePick - Could not get all players in league %s: %v\n", league.ID, err)
		return common.ErrInternalService
	}

	// advance turn (if CurrentPickSlotUsed) and update draft model
	err = s.advanceDraftState(draft, league, player, allPlayers, playerCount, currentPickSlotUsed)
	if err != nil {
		log.Printf("DraftService: MakePick - Error occured when attempting to advance draft state for league %s: %v\n", league.ID, err)
		return err
	}

	// TODO: Trigger webhook notification for the pick that just happened as well as the turn change

	return nil // no errors
}

// private helpers
func (s *draftServiceImpl) advanceDraftState(
	draft *models.Draft,
	league *models.League,
	player *models.Player, // The player whose turn just ended/skipped
	allPlayers []models.Player, // All players in the league, for turn progression
	playerCount int64,
	currentPickSlotUsed bool, // true if draft.CurrentPickOnClock was used in the request, false if skipped/implicitly skipped
) error {
	if !currentPickSlotUsed {
		// i.e., a skip/implicit. append CurrentPickOnClock to accumulated picks for that player
		draft.PlayersWithAccumulatedPicks[player.ID] = append(draft.PlayersWithAccumulatedPicks[player.ID], draft.CurrentPickOnClock)
	}

	draft.CurrentPickOnClock++ // unconditonal increment

	// check for draft completion
	isDraftCompleted, err := s.checkDraftCompletion(league, allPlayers)
	if err != nil {
		log.Printf("LOG: (DraftService: advanceDraftState) - Error checking draft completion for league %s: %v\\n", league.ID,
			err)
		return common.ErrInternalService
	}

	if isDraftCompleted {
		// if the draft is completed, we update and save the final state and return early
		// no further turn progression is needed

		draft.Status = enums.DraftStatusCompleted
		league.Status = enums.LeagueStatusPostDraft

		if err := s.draftRepo.UpdateDraft(draft); err != nil {
			log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update draft status to COMPLETED for league %s:%v\n", league.ID, err)
			return fmt.Errorf("failed to update draft state on completion: %w", err)
		}
		// save league status updated by checkDraftCompletion
		if _, err := s.leagueRepo.UpdateLeague(league); err != nil { // pray this never happens type shit
			// should prolly revert the draft update
			log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update league status to POST_DRAFT for league %s: %v\n", league.ID, err)
			return fmt.Errorf("failed to update league status on completion: %w", err)
		}
		return nil // Draft completed, states saved. we're so done
	}

	// Recalculate CurrentRound and CurrentPickInRound based on draft.CurrentPickOnClock
	draft.CurrentRound = ((draft.CurrentPickOnClock - 1) / int(playerCount)) + 1
	draft.CurrentPickInRound = ((draft.CurrentPickOnClock - 1) % int(playerCount)) + 1

	currentPlayerIdx := -1
	for i, p := range allPlayers { // there is likely some smort mafs you can do here to avoid an O(n) search. im stupid tho
		if p.ID == player.ID {
			currentPlayerIdx = i
			break
		}
	}
	if currentPlayerIdx == -1 {
		log.Printf("LOG: (DraftService: advanceDraftState) - Current player %s not found in allPlayers list. This should not happen. (Unreachable Control Flow)\\n", player.ID)
		return common.ErrInternalService
	}

	var nextPlayerIdx int
	if league.Format.IsSnakeRoundDraft {
		if draft.CurrentRound%2 == 0 { // even rounds are reverse order
			nextPlayerIdx = currentPlayerIdx - 1
		} else { // odd rounds are forward order
			nextPlayerIdx = currentPlayerIdx + 1
		}
	} else { // linear draft
		nextPlayerIdx = currentPlayerIdx + 1
	}

	if nextPlayerIdx >= int(playerCount) || nextPlayerIdx < 0 {
		// The CurrentRound and CurrentPickInRound are already correctly set by recalculation.
		// We just need to adjust nextPlayerIdx for the start of the new round.
		if league.Format.IsSnakeRoundDraft && draft.CurrentRound%2 == 0 { // if snake round drafting and new round is even
			nextPlayerIdx = int(playerCount) - 1 // last player in reverse order
		} else {
			nextPlayerIdx = 0 // first player in forward order
		}
	}

	// finally set the next turn of player
	nextTurnPlayer := allPlayers[nextPlayerIdx]
	draft.CurrentTurnPlayerID = &nextTurnPlayer.ID
	draft.CurrentTurnStartTime = func() *time.Time { t := time.Now(); return &t }()

	if err := s.draftRepo.UpdateDraft(draft); err != nil {
		log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update draft: %v\n", err)
		return common.ErrInternalService
	}

	return nil
}

func (s *draftServiceImpl) executePickTransactions(
	draft *models.Draft,
	league *models.League,
	player *models.Player,
	allRequestedPokemon []*models.LeaguePokemon,
	input *common.DraftMakePickDTO,
	playerCount int64, // yes i regret not realising int64 and int are different in postgres but not really in Go
	totalRequestedCost int,
) error {
	// create draftedPokemon records
	var allCreatedDraftedPokemon []*models.DraftedPokemon
	var leaguePokemonIDs []uuid.UUID
	var accumulatedPickNumberIndicesToDelete []int
	for i := 0; i < input.RequestedPickCount; i++ { // restrict to only max requested pick count transactions if bad request
		requestedPick := input.RequestedPicks[i]
		// get the entry in allRequestedPokemon
		var currentLeaguePokemon *models.LeaguePokemon
		for _, requestedPokemon := range allRequestedPokemon {
			if requestedPokemon.ID == requestedPick.LeaguePokemonID {
				currentLeaguePokemon = requestedPokemon
				break
			}
		}

		leaguePokemonIDs = append(leaguePokemonIDs, currentLeaguePokemon.ID)
		draftRoundNumber := ((requestedPick.DraftPickNumber - 1) / int(playerCount)) + 1

		createdDraftedPokemon := &models.DraftedPokemon{
			LeagueID:         league.ID,
			PlayerID:         player.ID,
			PokemonSpeciesID: currentLeaguePokemon.PokemonSpeciesID,
			LeaguePokemonID:  currentLeaguePokemon.ID,
			DraftRoundNumber: draftRoundNumber,
			DraftPickNumber:  requestedPick.DraftPickNumber, // overall pick
			IsReleased:       false,
		}

		allCreatedDraftedPokemon = append(allCreatedDraftedPokemon, createdDraftedPokemon)
		// caching the pickNumbers to
		if accumPickIndex := slices.Index(
			draft.PlayersWithAccumulatedPicks[player.ID], requestedPick.DraftPickNumber,
		); accumPickIndex != -1 {
			accumulatedPickNumberIndicesToDelete = append(accumulatedPickNumberIndicesToDelete, accumPickIndex)

		}
	}

	err := s.draftedPokemonRepo.DraftPokemonBatchTransaction(allCreatedDraftedPokemon, player, leaguePokemonIDs, totalRequestedCost)
	if err != nil {
		return err
	}

	// remove used up accumulated picks and update the draft model
	// might look unoptimal. it is. but the slices aren't that big
	// sort indices in descending order to avoid invalidating later indices
	slices.SortFunc(accumulatedPickNumberIndicesToDelete, func(a, b int) int {
		return b - a // Descending order
	})
	playerAccumulatedPicks := draft.PlayersWithAccumulatedPicks[player.ID]
	// iterate through sorted indices and remove elements
	for _, index := range accumulatedPickNumberIndicesToDelete {
		playerAccumulatedPicks = slices.Delete(playerAccumulatedPicks, index, index+1)
	}
	// update the draft object's map entry with the modified slice
	draft.PlayersWithAccumulatedPicks[player.ID] = playerAccumulatedPicks

	return nil
}

// validatePicksAndCheckCurrentPickSlotUsed performs remaining validation checks for a batch of picks
// and determines if the current turn's pick slot was used.
func (s *draftServiceImpl) validatePicksAndCheckCurrentPickSlotUsed(
	draft *models.Draft,
	player *models.Player,
	league *models.League,
	input *common.DraftMakePickDTO,
	totalRequestedCost int,
	playerCurrentRosterSize int,
) (bool, error) {
	playerID := *draft.CurrentTurnPlayerID // validated earlier to match currentPlayer
	currentPickSlotUsed := false

	// 1. Validate requested pick numbers against valid slots
	accumulatedPickNumbers := draft.PlayersWithAccumulatedPicks[playerID]
	validPickNumbersForPlayer := make([]int, len(accumulatedPickNumbers))
	copy(validPickNumbersForPlayer, accumulatedPickNumbers) // we don't wanna directly append
	validPickNumbersForPlayer = append(validPickNumbersForPlayer, draft.CurrentPickOnClock)

	// track used accumulated picks within this batch to prevent double-usage
	usedAccumulatedPicksInThisBatch := make(map[int]bool)

	for _, requestedPick := range input.RequestedPicks {
		// check if the requested pick number is a valid slot (current turn or accumulated)
		if !slices.Contains(validPickNumbersForPlayer, requestedPick.DraftPickNumber) {
			log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Player %s requested invalid pick number %d. Not on clock (%d) and not in accumulated picks (%v).\n",
				playerID, requestedPick.DraftPickNumber, draft.CurrentPickOnClock, accumulatedPickNumbers)
			return false, common.ErrInvalidInput
		}

		// if it's an accumulated pick, ensure it's not used twice in this batch
		if requestedPick.DraftPickNumber != draft.CurrentPickOnClock {
			if usedAccumulatedPicksInThisBatch[requestedPick.DraftPickNumber] {
				log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Player %s attempted to use accumulated pick %d multiple times in one request.\n",
					playerID, requestedPick.DraftPickNumber)
				return false, common.ErrInvalidInput
			}
			usedAccumulatedPicksInThisBatch[requestedPick.DraftPickNumber] = true
		}

		// check if the current pick slot is being used in this request
		if requestedPick.DraftPickNumber == draft.CurrentPickOnClock {
			currentPickSlotUsed = true
		}
	}

	// 2. Check if player has enough draft points for the entire batch
	if player.DraftPoints < totalRequestedCost {
		return false, common.ErrInsufficientDraftPoints
	}

	// 3. "Skips Left" Preventative Validation
	// this ensures the player doesn't implicitly skip their current turn's slot
	// if doing so would prevent them from meeting MinPokemonPerPlayer.

	rosterSizeAfterThisRequest := playerCurrentRosterSize + len(input.RequestedPicks)
	// calc how many more picks are needed to meet the minimum roster size
	picksNeededToMeetMin := int(math.Max(0, float64(league.MinPokemonPerPlayer-rosterSizeAfterThisRequest)))
	totalAvailablePickSlots := len(accumulatedPickNumbers) + 1 // +1 for draft.CurrentPickOnClock
	// calc how many "skips" the player can still afford before violating MinPokemonPerRoster
	skipsAllowedBeforeMinViolation := totalAvailablePickSlots - picksNeededToMeetMin

	// if skipsAllowedBeforeMinViolation is 0 or negative, it means the player MUST use a pick
	// for their current turn's slot (or an accumulated pick if that's the only way to meet min).
	// if they are trying to implicitly skip their current turn's slot (currentPickSlotUsed is false)
	// when they have no skips left, then it's an invalid action.
	if skipsAllowedBeforeMinViolation <= 0 && !currentPickSlotUsed {
		log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Player %s cannot implicitly skip current turn's pick (%d) as it would violate minimum roster requirement. Roster after picks: %d, Min required: %d, Skips left: %d.\n",
			playerID, draft.CurrentPickOnClock, rosterSizeAfterThisRequest, league.MinPokemonPerPlayer, skipsAllowedBeforeMinViolation)
		return false, common.ErrCannotSkipBelowMinimumRoster
	}

	return currentPickSlotUsed, nil
}

func (s *draftServiceImpl) getTotalCostForPicks(allRequestedPokemon []*models.LeaguePokemon) int {
	sumCost := 0
	for _, pokemon := range allRequestedPokemon {
		sumCost += *pokemon.Cost
	}
	return sumCost
}

func (s *draftServiceImpl) fetchDraftResource(leagueID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrDraftNotFound
		}
		return nil, common.ErrInternalService
	}
	return draft, nil
}

func (s *draftServiceImpl) fetchPlayerResource(userID, leagueID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetPlayerByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrPlayerNotFound
		}
		return nil, common.ErrInternalService
	}
	return player, nil
}

func (s *draftServiceImpl) fetchRequestedPokemon(leagueID uuid.UUID, input *common.DraftMakePickDTO) ([]*models.LeaguePokemon, error) {
	var pokemonIDs []uuid.UUID
	for _, requestedPick := range input.RequestedPicks {
		pokemonIDs = append(pokemonIDs, requestedPick.LeaguePokemonID)
	}

	allRequestedLeaguePokemonStructs, err := s.leaguePokemonRepo.GetLeaguePokemonByIDs(leagueID, pokemonIDs)
	if err != nil {
		return nil, common.ErrInternalService
	}

	// Validate that all requested Pokémon were actually returned and are available.
	// This ensures no invalid IDs slipped through or were already drafted.
	if len(allRequestedLeaguePokemonStructs) != len(pokemonIDs) {
		// This means some requested Pokémon were not found or were filtered out by the repo.
		return nil, common.ErrLeaguePokemonNotFound
	}

	var allRequestedLeaguePokemon []*models.LeaguePokemon
	for _, lp := range allRequestedLeaguePokemonStructs {
		if !lp.IsAvailable {
			return nil, common.ErrConflict
		}
		allRequestedLeaguePokemon = append(allRequestedLeaguePokemon, &lp)
	}

	return allRequestedLeaguePokemon, nil
}

func (s *draftServiceImpl) validateLeagueStatusForPick(leagueStatus enums.LeagueStatus, draftStatus enums.DraftStatus) bool {
	return leagueStatus == enums.LeagueStatusDrafting && draftStatus == enums.DraftStatusOngoing
}

// checkDraftCompletion determines if the draft has concluded and updates statuses accordingly.
// It should be called after a successful pick or skip, and after draft state has been advanced.
func (s *draftServiceImpl) checkDraftCompletion(
	league *models.League,
	allPlayers []models.Player,
) (bool, error) { // Returns true if draft is completed, false otherwise, and an error if any.
	// 1. Calculate total expected picks for the entire draft
	//    This is based on the maximum roster size for each player.
	totalPlayers := len(allPlayers)
	if totalPlayers == 0 {
		// should be impossible to reach here
		log.Printf("LOG: (DraftService: checkDraftCompletion) - No players in league %s. Cannot check for draft completion.\\n", league.ID)
		return false, common.ErrInternalService
	}
	maxPicksPerPlayer := league.MaxPokemonPerPlayer
	totalExpectedPicks := totalPlayers * maxPicksPerPlayer

	// 2. Get the current count of all *active* drafted Pokémon in the league
	currentTotalDraftedPokemon, err := s.draftedPokemonRepo.GetActiveDraftedPokemonCountByLeague(league.ID)
	if err != nil {
		log.Printf("LOG: (DraftService: checkDraftCompletion) - Failed to get total drafted pokemon count for league %s: %v\\n", league.ID, err)
		return false, common.ErrInternalService
	}

	// Cond. 1: if the total number of picks has reached the maximum
	if currentTotalDraftedPokemon < int64(totalExpectedPicks) {
		// not all maximum picks have been made yet, so the draft is not complete.
		return false, nil
	}

	// Cond. 2: if all players have met their MinPokemonPerRoster requirement
	minPokemonPerRoster := league.MinPokemonPerPlayer

	for _, player := range allPlayers {
		playerActiveRosterSize, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(player.ID) // Reusing existing method
		if err != nil {
			log.Printf("LOG: (DraftService: checkDraftCompletion) - Failed to get roster count for player %s in league %s: %v\\n", player.ID, league.ID, err)
			return false, common.ErrInternalService
		}
		if playerActiveRosterSize < int64(minPokemonPerRoster) {
			// at least one player has not met their minimum roster size, so the draft is not complete.
			return false, nil
		}
	}
	return true, nil
}
