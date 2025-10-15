package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
)

type DraftService interface {
	GetDraftByID(draftID uuid.UUID) (*models.Draft, error)
	GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error)
	StartDraft(leagueID uuid.UUID, TurnTimeLimit int) (*models.Draft, error)
	MakePick(currentUser *models.User, leagueID uuid.UUID, input *common.DraftMakePickDTO) error
	SkipTurn(currentUser *models.User, leagueID uuid.UUID) error
	AutoSkipTurn(playerID, leagueID uuid.UUID) error
	StartTransferPeriod(leagueID uuid.UUID) error
	EndTransferPeriod(leagueID uuid.UUID) error
	// DropFreeAgent(leagueID, draftedPokemonID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
	// PickupFreeAgent(leagueID, pokemonSpeciesID uuid.UUID, currentUser *models.User) (*models.DraftedPokemon, error)
	SetSchedulerService(schedulerService SchedulerService)
}

type draftServiceImpl struct {
	draftRepo          repositories.DraftRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	draftedPokemonRepo repositories.DraftedPokemonRepository
	webhookService     *WebhookService
	schedulerService   SchedulerService
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

// SetSchedulerService injects the SchedulerService dependency into the DraftService.
// This is called during application startup to break a circular dependency.
func (s *draftServiceImpl) SetSchedulerService(schedulerService SchedulerService) {
	s.schedulerService = schedulerService
}

func (s *draftServiceImpl) GetDraftByID(draftID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByID(draftID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("ERROR: (DraftService: GetDraftByID) - draft record for ID %s not found: %v", draftID, err)
			return nil, common.ErrDraftNotFound
		}
		log.Printf("ERROR: (DraftService: GetDraftByID) - Error fetching draft %s: %v", draftID, err)
		return nil, common.ErrInternalService
	}
	return draft, nil
}

func (s *draftServiceImpl) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("ERROR: (DraftService: GetDraftByID) - draft record for league ID %s not found: %v", leagueID, err)
			return nil, common.ErrDraftNotFound
		}
		log.Printf("ERROR: (DraftService: GetDraftByID) - Error fetching draft for league %s: %v", leagueID, err)
		return nil, common.ErrInternalService
	}
	return draft, nil
}

// StartDraft initializes the draft for a given league. It validates that there are players,
// sets the draft order (either randomly or by pre-set positions), creates the initial
// draft state in the database, updates the league status to DRAFTING, and schedules the
// first turn's timeout task.
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

	switch league.Format.DraftOrderType {
	case enums.DraftOrderTypeRandom:
		r := rand.New(rand.NewSource(time.Now().UnixNano())) // set seed
		r.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})

		// Assign new draft positions and update in DB
		for i := range players {
			players[i].DraftPosition = i + 1 // Draft positions are 1-based
			if err := s.playerRepo.UpdatePlayerDraftPosition(players[i].ID, players[i].DraftPosition); err != nil {
				log.Printf("LOG: (Error: DraftService.StartDraft) - Failed to update draft position for player %s: %v\n", players[i].ID, err)
				return nil, common.ErrInternalService
			}
		}
		log.Printf("LOG: (DraftService.StartDraft) - Randomized draft order for league %s complete.\n", leagueID)

	case enums.DraftOrderTypeManual:
		// Players are already sorted by DraftPosition from GetPlayersByLeague.
		// This assumes DraftPosition has been set manually prior to starting the draft.
		// Validate that all players have a unique, positive DraftPosition.
		seenPositions := make(map[int]bool)
		for _, p := range players {
			if p.DraftPosition <= 0 {
				log.Printf("ERROR: (DraftService: StartDraft) - Player %s has invalid draft position %d for manual draft order.\n", p.ID, p.DraftPosition)
				return nil, common.ErrInvalidDraftPosition
			}
			if seenPositions[p.DraftPosition] {
				log.Printf("ERROR: (DraftService: StartDraft) - Duplicate draft position %d found for player %s in manual draft order.\n", p.DraftPosition, p.ID)
				return nil, common.ErrDuplicateDraftPosition
			}
			seenPositions[p.DraftPosition] = true
		}
		// Ensure all positions from 1 to len(players) are present
		if len(seenPositions) != len(players) {
			log.Printf("ERROR: (DraftService: StartDraft) - Missing or extra draft positions for manual draft order in league %s.\n", leagueID)
			return nil, common.ErrIncompleteDraftOrder
		}
		log.Printf("LOG: (DraftService: StartDraft) - Using manual draft order for league %s.\n", leagueID)
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
		TurnTimeLimit:               TurnTimeLimit,
		PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks), // map[uuid.UUID][]int
		StartTime:                   time.Now(),
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

	taskType := utils.TaskTypeDraftTurnTimeout
	turnTimeLimit := draft.TurnTimeLimit
	turnStartTime := draft.CurrentTurnStartTime
	turnEndTime := turnStartTime.Add(time.Duration(turnTimeLimit) * time.Minute)

	task := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", taskType, draft.LeagueID),
		ExecuteAt: turnEndTime,
		Type:      taskType,
		Payload: utils.PayloadDraftTurnTimeout{
			LeagueID: draft.LeagueID,
			PlayerID: *draft.CurrentTurnPlayerID,
		},
	}

	s.schedulerService.RegisterTask(task)

	// Send an initial webhook notification
	// TODO: Implement webhook message creation logic
	// if err := (*s.webhookService).SendWebhookMessage(league.DiscordWebhookURL, "Draft has started!"); err != nil {
	// 	log.Printf("(Warning: DraftService.StartDraft) - Failed to send webhook for league %s: %v\n", leagueID, err)
	// 	// Continue execution, webhook failure shouldn't stop the draft
	// }

	return draft, nil
}

// MakePick handles a player's draft selection. It performs a series of validations:
// - Confirms the draft is in an active state.
// - Verifies it is the correct player's turn.
// - Checks that the requested Pokémon are available and affordable.
// - Ensures the pick doesn't violate league roster rules (e.g., minimum roster size).
// If all checks pass, it executes the pick as a transaction and advances the draft state.
// MakePick makes one or more picks (if accumulated) in a league's draft when league;
// Different from ForcePick, MakePick does all the required checks (there's a lot of checks) and validates the input
func (s *draftServiceImpl) MakePick(currentUser *models.User, leagueID uuid.UUID, input *common.DraftMakePickDTO) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftService: MakePick) - (user %s) could not find league %s: %v\n", currentUser.ID, leagueID, err)
		return common.ErrLeagueNotFound
	}

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
	// check if it's the right player's turn

	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		log.Printf("LOG: (DraftService: MakePick) - player %s tried to draft when it isn't their turn. Current Turn: Player %s\n", currentTurnPlayerID, *draft.CurrentTurnPlayerID)
		return common.ErrUnauthorized
	}

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

	totalRequestedCost := s.getTotalCostForPicks(allRequestedLeaguePokemon)

	// perform remaining validation
	currentPickSlotUsed, err := s.validatePicksAndCheckCurrentPickSlotUsed(draft, player, league, input, totalRequestedCost)
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
	draft, err = s.advanceDraftState(draft, league, player, allPlayers, int(playerCount), currentPickSlotUsed)
	if err != nil {
		log.Printf("LOG: (DraftService: MakePick) - Error occured when attempting to advance draft state for league %s: %v\n", league.ID, err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		fmt.Printf("INFO: (DraftService: advanceDraftState) - Draft Action (for league %s) was successful and Draft was detected to be COMPLETED. DraftStatus updated to COMPLETED.\n", draft.LeagueID)
		taskIDToDeregister := fmt.Sprintf("%d_%s", utils.TaskTypeDraftTurnTimeout, draft.LeagueID)
		s.schedulerService.DeregisterTask(taskIDToDeregister)
		return nil
	}

	// deregister prev task before registering new one
	taskIDToDeregister := fmt.Sprintf("%d_%s", utils.TaskTypeDraftTurnTimeout, draft.LeagueID)
	s.schedulerService.DeregisterTask(taskIDToDeregister)

	// schedule the timer task if the draft hasn't completed
	taskType := utils.TaskTypeDraftTurnTimeout
	turnTimeLimit := draft.TurnTimeLimit
	turnStartTime := draft.CurrentTurnStartTime
	turnEndTime := turnStartTime.Add(time.Duration(turnTimeLimit) * time.Minute)
	task := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", taskType, draft.LeagueID),
		ExecuteAt: turnEndTime,
		Type:      taskType,
		Payload: utils.PayloadDraftTurnTimeout{
			LeagueID: draft.LeagueID,
			PlayerID: *draft.CurrentTurnPlayerID,
		},
	}

	s.schedulerService.RegisterTask(task)

	// TODO: Trigger webhook notification for the pick that just happened as well as the turn change

	return nil // no errors
}

// SkipTurn allows a player to manually skip their current turn. It validates that the
// player is allowed to skip without violating minimum roster requirements and then
// advances the draft state, accumulating the skipped pick for the player.
func (s *draftServiceImpl) SkipTurn(currentUser *models.User, leagueID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("LOG: (DraftService: SkipTurn) - (user %s) could not find league %s: %v\n", currentUser.ID, leagueID, err)
		return common.ErrLeagueNotFound
	}

	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case common.ErrDraftNotFound:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Draft for leagueID %s not found: %v\n", currentUser.ID, league.ID, err)
		case common.ErrInternalService:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Error fetching draft: %v\n", currentUser.ID, err)
		}
		return err
	}

	player, err := s.fetchPlayerResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case common.ErrPlayerNotFound:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Player in league %s not found: %v\n", currentUser.ID, league.ID, err)
		case common.ErrInternalService:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Error fetching player in league %s: %v\n", currentUser.ID, league.ID, err)
		}
		return err
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		log.Printf("LOG: (DraftService: SkipTurn) - (user %s) league %s is not in drafting status: %v", currentUser.ID, league.ID, err)
		return common.ErrInvalidState
	}
	// check if it's the right player's turn
	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		log.Printf("LOG: (DraftService: SkipTurn) - player %s tried to draft when it isn't their turn. Current Turn: Player %s\n", currentTurnPlayerID, *draft.CurrentTurnPlayerID)
		return common.ErrUnauthorized
	}

	// get all players to change set the CurrentPlayer's turn for the next one
	allPlayers, err := s.playerRepo.GetPlayersByLeague(draft.LeagueID)
	if err != nil {
		log.Printf("DraftService: MakePick - Could not get all players in league %s: %v\n", league.ID, err)
		return common.ErrInternalService
	}

	playerRosterSize, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(player.ID)
	if err != nil {
		log.Printf("DraftService: MakePick - Failed to get player %d roster size\n", player.ID)
		return err
	}

	// validate if this skip action is allowed
	accumulatedPickCount := len(draft.PlayersWithAccumulatedPicks[player.ID])
	_, skipsLeft, err := s.isSkipAllowed(league, false, accumulatedPickCount, 0)
	if err != nil {
		log.Printf("LOG: (DraftService: SkipTurn) - Player %s cannot skip current turn's pick (%d) as it would violate minimum roster requirement.\nRoster size: %d, Min. required: %d, Skips left: %d.\n",
			player.ID, draft.CurrentPickOnClock, playerRosterSize, league.MinPokemonPerPlayer, skipsLeft)
		return err
	}

	draft, err = s.advanceDraftState(draft, league, player, allPlayers, len(allPlayers), false)
	if err != nil {
		log.Printf("LOG: (DraftService: MakePick) - Error occured when attempting to advance draft state for league %s: %v\n", league.ID, err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		fmt.Printf("INFO: (DraftService: advanceDraftState) - Draft Action (for league %s) was successful and Draft was detected to be COMPLETED. DraftStatus updated to COMPLETED.\n", draft.LeagueID)
		taskIDToDeregister := fmt.Sprintf("%d_%s", utils.TaskTypeDraftTurnTimeout, draft.LeagueID)
		s.schedulerService.DeregisterTask(taskIDToDeregister)
		return nil
	}

	taskIDToDeregister := fmt.Sprintf("%d_%s", utils.TaskTypeDraftTurnTimeout, draft.LeagueID)
	s.schedulerService.DeregisterTask(taskIDToDeregister)

	// schedule the timer task if the draft hasn't completed
	taskType := utils.TaskTypeDraftTurnTimeout
	turnTimeLimit := draft.TurnTimeLimit
	turnStartTime := draft.CurrentTurnStartTime
	turnEndTime := turnStartTime.Add(time.Duration(turnTimeLimit) * time.Minute)
	task := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", taskType, draft.LeagueID),
		ExecuteAt: turnEndTime,
		Type:      taskType,
		Payload: utils.PayloadDraftTurnTimeout{
			LeagueID: draft.LeagueID,
			PlayerID: *draft.CurrentTurnPlayerID,
		},
	}
	s.schedulerService.RegisterTask(task)

	// successful skip
	return nil
}

// AutoSkipTurn is called by the SchedulerService when a player's turn timer expires.
// It attempts to automatically skip the turn. If the skip is not allowed (e.g., it
// would violate minimum roster size), the draft is paused for manual intervention.
func (s *draftServiceImpl) AutoSkipTurn(playerID, leagueID uuid.UUID) error {
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		switch err {
		case common.ErrPlayerNotFound:
			log.Printf("ERROR: (DraftService: autoSkipTurn) - Player %s in league %s not found: %v\n", playerID, leagueID, err)
		case common.ErrInternalService:
			log.Printf("ERROR: (DraftService: autoSkipTurn) - Error fetching player %s in league %s: %v\n", playerID, leagueID, err)
		}
		return err
	}
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			log.Printf("LOG: (DraftService: autoSkipTurn) - (player %s) League %s not found: %v\n", playerID, leagueID, err)
			return common.ErrLeagueNotFound
		default:
			log.Printf("LOG: (DraftService: autoSkipTurn) - Could not fetch league %s: %v\n", leagueID, err)
			return common.ErrInternalService
		}
	}
	draft, err := s.fetchDraftResource(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			log.Printf("LOG: (DraftService: autoSkipTurn) - (player %s) Draft for leagueID %s not found: %v\n", playerID, leagueID, err)
			return common.ErrDraftNotFound
		default:
			log.Printf("LOG: (DraftService: autoSkipTurn) - (player %s) Error fetching draft: %v\n", playerID, err)
			return common.ErrInternalService
		}
	}

	accumulatedPicksForPlayer := draft.PlayersWithAccumulatedPicks[playerID]

	allowed, _, err := s.isSkipAllowed(league, false, len(accumulatedPicksForPlayer), 0)
	if !allowed {
		// common.ErrCannotSkipBelowMinimumRoster
		log.Printf("ERROR: (DraftService: autoSkipTurn) - Cannot auto skip for player %s, league %s: %v\n", playerID, leagueID, err)
		// set Draft to PAUSED status, awaiting manual league staff intervention
		draft.Status = enums.DraftStatusPaused
		draft, err = s.draftRepo.UpdateDraft(draft)
		if err != nil {
			log.Printf("ERROR: (DraftService: autoSkipTurn) - Could not update draft %d status to PAUSED: %v\n", draft.ID, err)
			return common.ErrInternalService
		}
		fmt.Printf("INFO: (DraftService: autoSkipTurn) - Draft for league %s paused. Awaiting Manual Intervention\n", leagueID)
		return common.ErrDraftPausedForIntervention
	}

	allPlayers, err := s.playerRepo.GetPlayersByLeague(leagueID)
	if err != nil {
		log.Printf("ERROR: (DraftService: autoSkipTurn) - Could not get all players in league %s: %v\n", leagueID, err)
		return common.ErrInternalService
	}

	draft, err = s.advanceDraftState(draft, league, player, allPlayers, len(allPlayers), false)
	if err != nil {
		log.Printf("ERROR: (DraftService: autoSkipTurn) - could not advance draft")
		return err
	}

	// AutoSkip shouldn't deregister the task because it is called by the schedulerService and thus
	// already dereigsters it automatically
	if draft.Status == enums.DraftStatusCompleted {
		fmt.Printf("INFO: (DraftService: advanceDraftState) - Draft Action (for league %s) was successful and Draft was detected to be COMPLETED. DraftStatus updated to COMPLETED.\n", draft.LeagueID)
		return nil
	}

	// schedule the timer task if the draft hasn't completed
	taskType := utils.TaskTypeDraftTurnTimeout
	turnTimeLimit := draft.TurnTimeLimit
	turnStartTime := draft.CurrentTurnStartTime
	turnEndTime := turnStartTime.Add(time.Duration(turnTimeLimit) * time.Minute)
	task := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", taskType, draft.LeagueID),
		ExecuteAt: turnEndTime,
		Type:      taskType,
		Payload: utils.PayloadDraftTurnTimeout{
			LeagueID: draft.LeagueID,
			PlayerID: *draft.CurrentTurnPlayerID,
		},
	}

	s.schedulerService.RegisterTask(task)
	fmt.Printf("INFO: (DraftService: AutoSkipTurn) - Success\n")
	// success
	return nil
}

// private helpers

// advanceDraftState moves the draft to the next turn or completes it.
// It increments the pick counter, checks if the draft's end conditions are met,
// determines the next player based on the draft order (linear or snake), and updates the draft model.
func (s *draftServiceImpl) advanceDraftState(
	draft *models.Draft,
	league *models.League,
	player *models.Player, // The player whose turn just ended/skipped
	allPlayers []models.Player, // All players in the league, for turn progression
	playerCount int,
	currentPickSlotUsed bool, // true if draft.CurrentPickOnClock was used in the request, false if skipped/implicitly skipped
) (*models.Draft, error) {
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
		return nil, common.ErrInternalService
	}

	if isDraftCompleted {
		// if the draft is completed, we update and save the final state and return early
		// no further turn progression is needed

		draft.Status = enums.DraftStatusCompleted
		league.Status = enums.LeagueStatusPostDraft
		draft.EndTime = time.Now()

		draft, err := s.draftRepo.UpdateDraft(draft)
		if err != nil {
			log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update draft status to COMPLETED for league %s:%v\n", league.ID, err)
			return nil, fmt.Errorf("failed to update draft state on completion: %w", err)
		}
		// save league status updated by checkDraftCompletion
		if _, err := s.leagueRepo.UpdateLeague(league); err != nil { // pray this never happens type shit
			// should prolly revert the draft update
			log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update league status to POST_DRAFT for league %s: %v\n", league.ID, err)
			return nil, fmt.Errorf("failed to update league status on completion: %w", err)
		}

		return draft, nil // Draft completed, states saved. we're so done
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
		return nil, common.ErrInternalService
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

	draft, err = s.draftRepo.UpdateDraft(draft)
	if err != nil {
		log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update draft: %v\n", err)
		return nil, common.ErrInternalService
	}

	return draft, nil
}

// executePickTransactions handles the database operations for a batch of draft picks.
// It creates the DraftedPokemon records, updates the player's draft points, and marks
// the LeaguePokemon as unavailable.
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

// validatePicksAndCheckCurrentPickSlotUsed performs the final validation checks before a pick is executed.
// It ensures that requested pick numbers are valid, the player has sufficient points, and that an
// implicit skip of the current turn doesn't violate minimum roster rules. It returns a boolean
// indicating if the current "on-the-clock" pick slot was used in the transaction.
func (s *draftServiceImpl) validatePicksAndCheckCurrentPickSlotUsed(
	draft *models.Draft,
	player *models.Player,
	league *models.League,
	input *common.DraftMakePickDTO,
	totalRequestedCost int,
) (bool, error) {
	playerID := *draft.CurrentTurnPlayerID // validated earlier to match currentPlayer

	// 1. Validate requested pick numbers against valid slots
	accumulatedPickNumbers := draft.PlayersWithAccumulatedPicks[playerID]
	validPickNumbersForPlayer := make([]int, len(accumulatedPickNumbers))
	copy(validPickNumbersForPlayer, accumulatedPickNumbers) // we don't wanna directly append
	validPickNumbersForPlayer = append(validPickNumbersForPlayer, draft.CurrentPickOnClock)

	// track used accumulated picks within this batch to prevent double-usage
	usedAccumulatedPicksInThisBatch := make(map[int]bool)
	currentPickSlotUsed := false
	accumulatedPickCount := len(accumulatedPickNumbers)

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
	_, skipsAllowed, err := s.isSkipAllowed(league, currentPickSlotUsed, accumulatedPickCount, input.RequestedPickCount)
	if err != nil {
		log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Player %s cannot implicitly skip current turn's pick (%d) as it would violate minimum roster requirement. Min required: %d, Skips left: %d.\n",
			playerID, draft.CurrentPickOnClock, league.MinPokemonPerPlayer, skipsAllowed)
		return false, err
	}

	return currentPickSlotUsed, nil
}

// isSkipAllowed checks if a player can skip (or implicit skip) their turn without making it impossible
// to meet the league's minimum roster requirement.
// returns true if allowed, false otherwise
// also returns number of skips allowed
func (s *draftServiceImpl) isSkipAllowed(
	league *models.League,
	currentPickSlotUsed bool, // true if the current "on-the-clock" pick slot was used in the request
	accumulatedPickCountForPlayer int, // Accumulated picks *before* this turn's action
	requestedPickCount int, // Number of picks the player is attempting to make in this request (0 for SkipTurn)
) (bool, int, error) {
	// Calculate the total "skip budget" for the entire draft.
	// This is the number of optional picks a player can forgo without violating the minimum roster size.
	totalSkipBudget := league.MaxPokemonPerPlayer - league.MinPokemonPerPlayer // if <0, invalid configuration of league
	totalSkipBudget = int(math.Max(0, float64(totalSkipBudget)))               // still doing the check though

	// Determine the total number of pick slots available to the player for this turn's action.
	// This includes any accumulated picks and the current "on-the-clock" pick if it's not being explicitly used.
	availablePickSlotsThisTurn := accumulatedPickCountForPlayer
	if !currentPickSlotUsed { // If current turn's pick is NOT used, it's an available slot
		availablePickSlotsThisTurn++
	}

	// Calculate the number of "effective skips" this action represents.
	// This is the number of available pick slots that are *not* being used.
	// 1 if regular turn skip (auto or otherwise),
	// `availablePickSlotsThisTurn` - `requestedPickCount` if implicit skip (MakePick)
	var effectiveSkipsInThisAction int
	if requestedPickCount > 0 { // This is a MakePick action
		effectiveSkipsInThisAction = availablePickSlotsThisTurn - requestedPickCount
	} else { // This is a SkipTurn or AutoSkipTurn action
		effectiveSkipsInThisAction = 1 // One skip for the current turn
	}
	// Ensure effectiveSkipsInThisAction is not negative
	// should be caught by other validation (and it is i'm pretty sure))
	effectiveSkipsInThisAction = int(math.Max(0, float64(effectiveSkipsInThisAction)))

	// The total number of skips the player will have made *after* this action.
	totalSkipsAfterAction := accumulatedPickCountForPlayer + effectiveSkipsInThisAction

	// Check if this action would exceed the total skip budget.
	if totalSkipsAfterAction > totalSkipBudget {
		return false, 0, common.ErrCannotSkipBelowMinimumRoster
	}

	// If we reach here, the action is allowed.
	// Calculate the remaining skips after this one.
	remainingSkips := totalSkipBudget - totalSkipsAfterAction
	remainingSkips = int(math.Max(0, float64(remainingSkips))) // Ensure >=0

	return true, remainingSkips, nil
}

func (s *draftServiceImpl) getTotalCostForPicks(allRequestedPokemon []*models.LeaguePokemon) int {
	sumCost := 0
	for _, pokemon := range allRequestedPokemon {
		sumCost += *pokemon.Cost
	}
	return sumCost
}

// fetchDraftResource retrieves the draft for a league, converting a gorm.ErrRecordNotFound
// into a service-specific error.
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

// fetchPlayerResource retrieves a player by user and league, converting a gorm.ErrRecordNotFound
// into a service-specific error.
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

// fetchRequestedPokemon retrieves a list of LeaguePokemon by their IDs, ensuring they are
// all available to be drafted. It returns service-specific errors for not found or
// already drafted pokemon.
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

// checkDraftCompletion determines if the draft has concluded by checking two conditions:
// 1. Has the total number of drafted pokemon reached the maximum allowed for the league?
// 2. Have all players met the minimum roster requirement?
// It is called after each pick/skip to see if the draft should be moved to a COMPLETED state.
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

// StartTransferPeriod begins the transfer window for a league. It updates the league status,
// allocates transfer credits to players if enabled, and schedules the end of the window.
func (s *draftServiceImpl) StartTransferPeriod(leagueID uuid.UUID) error {
	// 1. Fetch the League
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (DraftService: StartTransferPeriod) - Failed to fetch league %s: %v\n", leagueID, err)
		return common.ErrLeagueNotFound
	}

	// 2. Validate Status
	if league.Status != enums.LeagueStatusRegularSeason && league.Status != enums.LeagueStatusPostDraft {
		log.Printf("WARN: (DraftService: StartTransferPeriod) - League %s is not in a valid state to start a transfer window. Status: %s\n", leagueID, league.Status)
		return fmt.Errorf("invalid league status to start transfer window: %s", league.Status)
	}

	// 3. Update Player Credits (if applicable)
	if league.Format.AllowTransferCredits {
		players, err := s.playerRepo.GetPlayersByLeague(leagueID)
		if err != nil {
			log.Printf("ERROR: (DraftService: StartTransferPeriod) - Failed to get players for league %s: %v\n", leagueID, err)
			return common.ErrInternalService
		}

		for _, player := range players {
			player.TransferCredits += league.Format.TransferCreditsPerWindow
			if player.TransferCredits > league.Format.TransferCreditCap {
				player.TransferCredits = league.Format.TransferCreditCap
			}
			if _, err := s.playerRepo.UpdatePlayer(&player); err != nil {
				// Log the error but continue trying to update other players
				log.Printf("ERROR: (DraftService: StartTransferPeriod) - Failed to update transfer credits for player %s: %v\n", player.ID, err)
			}
		}
	}

	// 4. Update League Status
	league.Status = enums.LeagueStatusTransferWindow
	now := time.Now()
	league.Format.NextTransferWindowStart = &now // The window starts now

	// 5. Schedule EndTransferPeriod
	windowEndTime := now.Add(time.Duration(league.Format.TransferWindowDuration) * time.Minute)
	endTask := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", utils.TaskTypeTradingPeriodEnd, league.ID),
		ExecuteAt: windowEndTime,
		Type:      utils.TaskTypeTradingPeriodEnd,
		Payload: utils.PayloadTransferPeriodEnd{
			LeagueID: league.ID,
		},
	}
	s.schedulerService.RegisterTask(endTask)

	// 6. Save Changes
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (DraftService: StartTransferPeriod) - Failed to update league %s status: %v\n", leagueID, err)
		// Note: If this fails, the task is scheduled but the league status isn't updated.
		// This could be improved with a transactional approach if necessary.
		return common.ErrInternalService
	}

	log.Printf("LOG: (DraftService: StartTransferPeriod) - Transfer window started for league %s.\n", leagueID)
	return nil
}

// EndTransferPeriod concludes the transfer window for a league. It updates the league status
// and schedules the next transfer window to begin.
func (s *draftServiceImpl) EndTransferPeriod(leagueID uuid.UUID) error {
	// 1. Fetch the League
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (DraftService: EndTransferPeriod) - Failed to fetch league %s: %v\n", leagueID, err)
		return common.ErrLeagueNotFound
	}

	// 2. Validate Status
	if league.Status != enums.LeagueStatusTransferWindow {
		log.Printf("WARN: (DraftService: EndTransferPeriod) - League %s is not in a transfer window. Status: %s\n", leagueID, league.Status)
		return fmt.Errorf("invalid league status to end transfer window: %s", league.Status)
	}

	// 3. Update League Status
	league.Status = enums.LeagueStatusRegularSeason

	// 4. Schedule next StartTransferPeriod
	if league.Format.TransferWindowFrequencyDays > 0 {
		nextWindowStartTime := time.Now().AddDate(0, 0, league.Format.TransferWindowFrequencyDays)
		league.Format.NextTransferWindowStart = &nextWindowStartTime

		startTask := &utils.ScheduledTask{
			ID:        fmt.Sprintf("%d_%s", utils.TaskTypeTradingPeriodStart, league.ID),
			ExecuteAt: nextWindowStartTime,
			Type:      utils.TaskTypeTradingPeriodStart,
			Payload: utils.PayloadTransferPeriodStart{
				LeagueID: league.ID,
			},
		}
		s.schedulerService.RegisterTask(startTask)
	} else {
		// If frequency is 0 or less, don't schedule a next window.
		league.Format.NextTransferWindowStart = nil
	}

	// 5. Save Changes
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (DraftService: EndTransferPeriod) - Failed to update league %s: %v\n", leagueID, err)
		return common.ErrInternalService
	}

	log.Printf("LOG: (DraftService: EndTransferPeriod) - Transfer window ended for league %s.\n", leagueID)
	return nil
}
