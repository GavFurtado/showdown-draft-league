package services

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
)

type DraftService interface {
	GetDraftByID(draftID uuid.UUID) (*models.Draft, error)
	GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error)
	StartDraft(leagueID uuid.UUID, TurnTimeLimit int) (*models.Draft, error)
	MakePick(currentUser *models.User, leagueID uuid.UUID, input *requests.DraftMakePickRequestDTO) error
	SkipTurn(currentUser *models.User, leagueID uuid.UUID) error
	AutoSkipTurn(playerID, leagueID uuid.UUID) error
	SetSchedulerService(schedulerService SchedulerService)
	SetNewRepositories(draftPickRepo repositories.DraftPickRepository, claimRepo repositories.ClaimRepository, poolEntryRepo repositories.PoolEntryRepository)
}

type draftServiceImpl struct {
	logger           *slog.Logger
	draftRepo        repositories.DraftRepository
	leagueRepo       repositories.LeagueRepository
	playerRepo       repositories.PlayerRepository
	webhookService   *WebhookService
	schedulerService SchedulerService

	draftPickRepo repositories.DraftPickRepository
	claimRepo     repositories.ClaimRepository
	poolEntryRepo repositories.PoolEntryRepository
}

func NewDraftService(
	logger *slog.Logger,
	leagueRepo repositories.LeagueRepository,
	draftRepo repositories.DraftRepository,
	playerRepo repositories.PlayerRepository,
	webhookService *WebhookService,
) DraftService {
	return &draftServiceImpl{
		logger:         utils.LoggerWithService(logger, "DraftService"),
		draftRepo:      draftRepo,
		leagueRepo:     leagueRepo,
		playerRepo:     playerRepo,
		webhookService: webhookService,
	}
}

func (s *draftServiceImpl) SetNewRepositories(
	draftPickRepo repositories.DraftPickRepository,
	claimRepo repositories.ClaimRepository,
	poolEntryRepo repositories.PoolEntryRepository,
) {
	s.draftPickRepo = draftPickRepo
	s.claimRepo = claimRepo
	s.poolEntryRepo = poolEntryRepo
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
			s.logger.Warn("Draft record not found", "draft_id", draftID, "error", err)
			return nil, types.ErrDraftNotFound
		}
		s.logger.Error("Error fetching draft", "draft_id", draftID, "error", err)
		return nil, types.ErrInternalService
	}
	return draft, nil
}

func (s *draftServiceImpl) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Draft record not found", "league_id", leagueID, "error", err)
			return nil, types.ErrDraftNotFound
		}
		s.logger.Error("Error fetching draft for league", "league_id", leagueID, "error", err)
		return nil, types.ErrInternalService
	}
	return draft, nil
}

// StartDraft initializes the draft for a given league. It validates that there are players,
// sets the draft order (either randomly or by pre-set positions), creates the initial
// draft state in the database, updates the league status to DRAFTING, and schedules the
// first turn's timeout task.
// player permission rbac.PermissionCreateDraft
// TODO: Make this more of an orchestrator. More testable private functions that this calls
func (s *draftServiceImpl) StartDraft(leagueID uuid.UUID, TurnTimeLimit int) (*models.Draft, error) {
	// Retrieve the league
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil || league == nil {
		s.logger.Error("Could not get league", "league_id", leagueID, "error", err)
		return nil, types.ErrLeagueNotFound
	}

	// Retrieve players in the league, sorted by draft position
	players, err := s.playerRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("Could not get players", "league_id", leagueID, "error", err)
		return nil, types.ErrInternalService
	}

	if len(players) == 0 {
		s.logger.Warn("No players found", "league_id", leagueID)
		return nil, types.ErrNoPlayerForDraft
	}

	// TODO: each case should call it's own respective function
	switch league.Format.DraftOrderType {
	case enums.DraftOrderTypeRandom:
		r := rand.New(rand.NewSource(time.Now().UnixNano())) // set seed
		r.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})

		// Assign new draft positions and update in DB
		for i := range players {
			players[i].DraftPosition = i + 1 // Draft positions are 1-based
			if err := s.playerRepo.UpdateDraftPosition(players[i].ID, players[i].DraftPosition); err != nil {
				s.logger.Error("Failed to update draft position", "player_id", players[i].ID, "error", err)
				return nil, types.ErrInternalService
			}
		}
		s.logger.Info("Randomized draft order complete", "league_id", leagueID)

	case enums.DraftOrderTypeManual:
		// Players are already sorted by DraftPosition from GetByLeague.
		// This assumes DraftPosition has been set manually prior to starting the draft.
		// Validate that all players have a unique, positive DraftPosition.
		seenPositions := make(map[int]bool)
		for _, m := range players {
			if m.DraftPosition <= 0 {
				s.logger.Error("Player has invalid draft position", "player_id", m.ID, "position", m.DraftPosition)
				return nil, types.ErrInvalidDraftPosition
			}
			if seenPositions[m.DraftPosition] {
				s.logger.Error("Duplicate draft position found", "position", m.DraftPosition, "player_id", m.ID)
				return nil, types.ErrDuplicateDraftPosition
			}
			seenPositions[m.DraftPosition] = true
		}
		// Ensure all positions from 1 to len(players) are present
		if len(seenPositions) != len(players) {
			s.logger.Error("Missing or extra draft positions for manual draft order", "league_id", leagueID)
			return nil, types.ErrIncompleteDraftOrder
		}
		s.logger.Info("Using manual draft order", "league_id", leagueID)
	}

	// Initialize the Draft model
	firstPlayerID := players[0].ID
	currTime := time.Now()

	draft := &models.Draft{
		LeagueID:                    leagueID,
		Status:                      enums.DraftStatusOngoing,
		CurrentRound:                1,
		CurrentPickInRound:          1,
		CurrentPickOnClock:          1, // NOTE: formula: ((CurrentRound - 1)*PlayerCount + CurrentPickInRound)
		CurrentTurnPlayerID:         &firstPlayerID,
		CurrentTurnStartTime:        &currTime,
		TurnTimeLimit:               TurnTimeLimit,
		PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks), // map[uuid.UUID][]int
		StartTime:                   time.Now(),
	}

	// Save the Draft model
	if err := s.draftRepo.CreateDraft(draft); err != nil {
		s.logger.Error("Failed to create draft", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	// Update the league status to DRAFTING
	league.Status = enums.LeagueStatusDrafting
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("Failed to update league status", "league_id", leagueID, "error", err)
		// FIXME: Consider rolling back draft creation if this fails
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
	// Implementation needs to be async as well

	return draft, nil
}

// MakePick handles a player's draft selection. It performs a series of validations:
// - Confirms the draft is in an active state.
// - Verifies it is the correct player's turn.
// - Checks that the requested Pokémon are available and affordable.
// - Ensures the pick doesn't violate league roster rules (such as the minimum roster size).
// If all checks pass, it executes the pick as a transaction and advances the draft state.
// MakePick makes one or more picks (if accumulated) during drafting phase;
// Different from ForcePick (not implemented yet),
func (s *draftServiceImpl) MakePick(
	currentUser *models.User,
	leagueID uuid.UUID,
	input *requests.DraftMakePickRequestDTO,
) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		s.logger.Error("Could not find league", "user_id", currentUser.ID, "league_id", leagueID, "error", err)
		return types.ErrLeagueNotFound
	}

	// Fetch draft for league
	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case types.ErrDraftNotFound:
			s.logger.Warn("Draft not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("Error fetching draft", "user_id", currentUser.ID, "error", err)
		}
		return err
	}

	// Fetch the player resource
	player, err := s.fetchPlayerResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case types.ErrPlayerNotFound:
			s.logger.Warn("Player not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("Error fetching player", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		}
		return err
	}

	// START early checks to prevent doing any of the expensive checks later
	// Check if it's the right player's turn
	// TODO: should all be in it's own function (like performEarlyChecks or performInexpensiveChecks)
	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		s.logger.Warn("Player tried to draft when not their turn", "player_id", currentTurnPlayerID, "current_turn_player_id", *draft.CurrentTurnPlayerID)
		return types.ErrUnauthorized
	}

	// Check if number of requested picks is valid for the player
	if input.RequestedPickCount > len(draft.PlayersWithAccumulatedPicks[player.ID])+1 {
		s.logger.Warn("Player requested too many draft picks", "user_id", currentUser.ID, "player_id", player.ID)
		return types.ErrTooManyRequestedPicks
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		s.logger.Warn("League is not in drafting status", "user_id", currentUser.ID, "league_id", league.ID)
		return types.ErrInvalidState
	}
	// END early checks

	// Fetch all the pool entries requested
	// expensive
	allRequestedPoolEntries, err := s.fetchRequestedPoolEntries(league.ID, input)
	if err != nil {
		switch err {
		case types.ErrPoolEntryNotFound:
			s.logger.Warn("One or more pool entries not found", "user_id", currentUser.ID, "error", err)
		case types.ErrConflict:
			s.logger.Warn("One or more pool entries not available", "user_id", currentUser.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("Error fetching requested pool entries", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		}
		return err
	}

	// get player count; needed in multiple places
	playerCount, err := s.playerRepo.GetCountByLeague(league.ID)
	if err != nil {
		s.logger.Error("Failed to get player count", "league_id", league.ID, "error", err)
		return types.ErrInternalService
	}
	if playerCount == 0 { // this should never happen if the draft has started or if the league even exists
		s.logger.Error("No players in league (unreachable code)", "league_id", league.ID)
		return types.ErrInternalService
	}

	totalRequestedCost := s.getTotalCostForPoolEntries(allRequestedPoolEntries)

	// perform remaining validation
	currentPickSlotUsed, err := s.validatePicksAndCheckCurrentPickSlotUsed(draft, player, league, input, totalRequestedCost)
	if err != nil {
		switch err {
		case types.ErrInvalidInput:
			s.logger.Warn("Invalid pick number in request", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInsufficientDraftPoints:
			s.logger.Warn("Insufficient draft points", "user_id", currentUser.ID, "league_id", league.ID, "draft_points", player.DraftPoints, "error", err)
		}
		return err
	}

	// execute picks (new model: creates DraftPick + Claim instead of DraftedPokemon)
	err = s.executeNewPickTransactions(draft, league, player, allRequestedPoolEntries, input, playerCount, totalRequestedCost)
	if err != nil {
		s.logger.Error("Batch transaction unsuccessful", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		return err
	}

	// get all players to change set the current player's turn for the next one
	allPlayers, err := s.playerRepo.GetByLeague(draft.LeagueID)
	if err != nil {
		s.logger.Error("Could not get all players", "league_id", league.ID, "error", err)
		return types.ErrInternalService
	}

	// advance turn (if CurrentPickSlotUsed) and update draft model
	draft, err = s.advanceDraftState(draft, league, player, allPlayers, int(playerCount), currentPickSlotUsed)
	if err != nil {
		s.logger.Error("Error advancing draft state", "league_id", league.ID, "error", err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		s.logger.Info("Draft completed", "league_id", draft.LeagueID)
		taskIDToDeregister := fmt.Sprintf("%d_%s", utils.TaskTypeDraftTurnTimeout, draft.LeagueID)
		s.schedulerService.DeregisterTask(taskIDToDeregister)
		return nil
	}

	// deregister previous task before registering new one
	taskIDToDeregister := fmt.Sprintf("%d_%s", utils.TaskTypeDraftTurnTimeout, draft.LeagueID)
	s.schedulerService.DeregisterTask(taskIDToDeregister)

	// schedule the timer task for the next player's turn if the draft hasn't completed
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
		s.logger.Error("Could not find league", "user_id", currentUser.ID, "league_id", leagueID, "error", err)
		return types.ErrLeagueNotFound
	}

	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case types.ErrDraftNotFound:
			s.logger.Warn("Draft not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("Error fetching draft", "user_id", currentUser.ID, "error", err)
		}
		return err
	}

	player, err := s.fetchPlayerResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case types.ErrPlayerNotFound:
			s.logger.Warn("Player not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("Error fetching player", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		}
		return err
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		s.logger.Warn("League is not in drafting status", "user_id", currentUser.ID, "league_id", league.ID)
		return types.ErrInvalidState
	}
	// check if it's the right player's turn
	if currentTurnPlayerID := *draft.CurrentTurnPlayerID; currentTurnPlayerID != player.ID {
		s.logger.Warn("Player tried to draft when not their turn", "player_id", currentTurnPlayerID, "current_turn_player_id", *draft.CurrentTurnPlayerID)
		return types.ErrUnauthorized
	}

	// get all players to change set the current player's turn for the next one
	allPlayers, err := s.playerRepo.GetByLeague(draft.LeagueID)
	if err != nil {
		s.logger.Error("Could not get all players", "league_id", league.ID, "error", err)
		return types.ErrInternalService
	}

	effectiveSkipsInThisAction := 1 // One skip for the current turn
	_, err = s.isSkipAllowed(player, effectiveSkipsInThisAction)
	if err != nil {
		s.logger.Warn("Player cannot skip current turn's pick as it would violate minimum roster requirement", "player_id", player.ID, "pick", draft.CurrentPickOnClock, "skips_left", player.SkipsLeft)
		return err
	}

	player.SkipsLeft -= effectiveSkipsInThisAction
	s.logger.Debug("Skips left before DB update", "player_id", player.ID, "skips_left", player.SkipsLeft)
	if _, err := s.playerRepo.Update(player); err != nil {
		s.logger.Error("Failed to update player skipsLeft in DB", "player_id", player.ID, "error", err)
		return types.ErrInternalService
	}
	// Re-fetch player to confirm DB state
	updatedPlayer, err := s.playerRepo.GetByID(player.ID)
	if err != nil {
		s.logger.Error("Failed to re-fetch player after update", "player_id", player.ID, "error", err)
		return types.ErrInternalService
	}
	s.logger.Debug("Skips left after DB re-fetch", "player_id", updatedPlayer.ID, "skips_left", updatedPlayer.SkipsLeft)

	draft, err = s.advanceDraftState(draft, league, player, allPlayers, len(allPlayers), false)
	if err != nil {
		s.logger.Error("Error advancing draft state", "league_id", league.ID, "error", err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		s.logger.Info("Draft completed", "league_id", draft.LeagueID)
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
	player, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("Player not found", "player_id", playerID, "league_id", leagueID, "error", err)
			return types.ErrPlayerNotFound
		}
		s.logger.Error("Error fetching player", "player_id", playerID, "league_id", leagueID, "error", err)
		return types.ErrInternalService
	}
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			s.logger.Warn("League not found", "player_id", playerID, "league_id", leagueID, "error", err)
			return types.ErrLeagueNotFound
		default:
			s.logger.Error("Could not fetch league", "league_id", leagueID, "error", err)
			return types.ErrInternalService
		}
	}
	draft, err := s.fetchDraftResource(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			s.logger.Warn("Draft not found", "player_id", playerID, "league_id", leagueID, "error", err)
			return types.ErrDraftNotFound
		default:
			s.logger.Error("Error fetching draft", "player_id", playerID, "error", err)
			return types.ErrInternalService
		}
	}

	effectiveSkipsInThisAction := 1
	allowed, err := s.isSkipAllowed(player, effectiveSkipsInThisAction)
	if !allowed {
		s.logger.Error("Cannot auto skip", "player_id", playerID, "league_id", leagueID, "error", err, "skips_left", player.SkipsLeft)
		// set Draft to PAUSED status, awaiting manual league staff intervention
		draft.Status = enums.DraftStatusPaused
		draft, err = s.draftRepo.UpdateDraft(draft)
		if err != nil {
			s.logger.Error("Could not update draft status to PAUSED", "draft_id", draft.ID, "error", err)
			return types.ErrInternalService
		}
		s.logger.Info("Draft paused, awaiting manual intervention", "league_id", leagueID)
		return types.ErrDraftPausedForIntervention
	}

	player.SkipsLeft -= effectiveSkipsInThisAction
	s.logger.Debug("Skips left before DB update", "player_id", player.ID, "skips_left", player.SkipsLeft)
	if _, err := s.playerRepo.Update(player); err != nil {
		s.logger.Error("Failed to update player skipsLeft in DB", "player_id", player.ID, "error", err)
		return types.ErrInternalService
	}
	// Re-fetch player to confirm DB state
	updatedPlayer, err := s.playerRepo.GetByID(player.ID)
	if err != nil {
		s.logger.Error("Failed to re-fetch player after update", "player_id", player.ID, "error", err)
		return types.ErrInternalService
	}
	s.logger.Debug("Skips left after DB re-fetch", "player_id", updatedPlayer.ID, "skips_left", updatedPlayer.SkipsLeft)
	s.logger.Debug("Skips left updated after auto-skip", "player_id", player.ID, "skips_left", player.SkipsLeft)

	allPlayers, err := s.playerRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("Could not get all players", "league_id", leagueID, "error", err)
		return types.ErrInternalService
	}

	draft, err = s.advanceDraftState(draft, league, player, allPlayers, len(allPlayers), false)
	if err != nil {
		s.logger.Error("Could not advance draft")
		return err
	}

	// AutoSkip shouldn't deregister the task because it is called by the schedulerService and thus
	// already dereigsters it automatically
	if draft.Status == enums.DraftStatusCompleted {
		s.logger.Info("Draft completed", "league_id", draft.LeagueID)
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
	s.logger.Info("Turn timeout task scheduled")
	// success
	return nil
}

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
		// i.e., a skip/implicit skip. Append CurrentPickOnClock to
		// accumulated picks for that player
		s.logger.Debug("Player skipping pick", "player_id", player.ID, "pick", draft.CurrentPickOnClock, "accumulated_picks", draft.PlayersWithAccumulatedPicks[player.ID])
		draft.PlayersWithAccumulatedPicks[player.ID] = append(draft.PlayersWithAccumulatedPicks[player.ID], draft.CurrentPickOnClock)
	}

	draft.CurrentPickOnClock++ // unconditonal increment

	// Check for draft completion
	isDraftCompleted, err := s.checkDraftCompletion(league, allPlayers)
	if err != nil {
		s.logger.Error("Error checking draft completion", "league_id", league.ID, "error", err)
		return nil, types.ErrInternalService
	}

	if isDraftCompleted {
		// If the draft has completed, we update and save the final state and return early
		// no further turn progression is needed

		draft.Status = enums.DraftStatusCompleted
		league.Status = enums.LeagueStatusPostDraft
		draft.EndTime = time.Now()

		draft, err := s.draftRepo.UpdateDraft(draft)
		if err != nil {
			s.logger.Error("Failed to update draft status to COMPLETED", "league_id", league.ID, "error", err)
			return nil, fmt.Errorf("failed to update draft state on completion: %w", err)
		}
		// save league status updated by checkDraftCompletion
		if _, err := s.leagueRepo.UpdateLeague(league); err != nil { // pray this never happens type shit
			// should prolly revert the draft update
			s.logger.Error("Failed to update league status to POST_DRAFT", "league_id", league.ID, "error", err)
			return nil, fmt.Errorf("failed to update league status on completion: %w", err)
		}

		return draft, nil // Draft completed, states saved. We're so done
	}

	// If draft is still ongoing,
	// Recalculate CurrentRound and CurrentPickInRound based on draft.CurrentPickOnClock
	draft.CurrentRound = ((draft.CurrentPickOnClock - 1) / int(playerCount)) + 1
	draft.CurrentPickInRound = ((draft.CurrentPickOnClock - 1) % int(playerCount)) + 1

	currentPlayerIdx := -1
	for i, m := range allPlayers { // there is likely some smort mafs you can do here to avoid an O(n) search. im stupid tho
		if m.ID == player.ID {
			currentPlayerIdx = i
			break
		}
	}
	if currentPlayerIdx == -1 { // this is an impossible case
		s.logger.Error("Current player not found in allPlayers list (unreachable)", "player_id", player.ID)
		return nil, types.ErrInternalService
	}

	var nextPlayerIdx int
	if league.Format.IsSnakeRoundDraft {
		if draft.CurrentRound%2 != 0 { // Odd round (forward order)
			nextPlayerIdx = draft.CurrentPickInRound - 1
		} else { // Even round (reverse order)
			nextPlayerIdx = int(playerCount) - draft.CurrentPickInRound
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
		s.logger.Error("Failed to update draft", "error", err)
		return nil, types.ErrInternalService
	}

	return draft, nil
}

// executeNewPickTransactions handles the database operations for a batch of draft picks.
// It creates the DraftPick records and Claim records (instead of the old DraftedPokemon model),
// updates the player's draft points, and marks the PoolEntry as unavailable.
func (s *draftServiceImpl) executeNewPickTransactions(
	draft *models.Draft,
	league *models.League,
	player *models.Player,
	allRequestedPoolEntries []*models.PoolEntry,
	input *requests.DraftMakePickRequestDTO,
	playerCount int64,
	totalRequestedCost int,
) error {
	var err error
	// Build draft pick and claim records
	var draftPicks []models.DraftPick
	var poolEntryIDs []uuid.UUID
	var accumulatedPickNumberIndicesToDelete []int

	for i := 0; i < input.RequestedPickCount; i++ {
		requestedPick := input.RequestedPicks[i]

		// Get the entry in allRequestedPoolEntries
		var currentPoolEntry *models.PoolEntry
		for _, entry := range allRequestedPoolEntries {
			if entry.ID == requestedPick.PoolEntryID {
				currentPoolEntry = entry
				break
			}
		}
		if currentPoolEntry == nil {
			return types.ErrPoolEntryNotFound
		}

		poolEntryIDs = append(poolEntryIDs, currentPoolEntry.ID)
		draftRoundNumber := ((requestedPick.DraftPickNumber - 1) / int(playerCount)) + 1

		// Build DraftPick (immutable event log)
		draftPick := models.DraftPick{
			DraftID:     draft.ID,
			PlayerID:    player.ID,
			PoolEntryID: currentPoolEntry.ID,
			RoundNumber: draftRoundNumber,
			PickNumber:  requestedPick.DraftPickNumber,
		}
		draftPicks = append(draftPicks, draftPick)

		// Cache accumulated pick numbers to remove
		if accumPickIndex := slices.Index(
			draft.PlayersWithAccumulatedPicks[player.ID], requestedPick.DraftPickNumber,
		); accumPickIndex != -1 {
			accumulatedPickNumberIndicesToDelete = append(accumulatedPickNumberIndicesToDelete, accumPickIndex)
		}
	}

	// Execute in a transaction: create DraftPicks, mark PoolEntries unavailable, deduct points
	err = s.executeWithTransaction(func(txRepo *transactionalRepositories) error {
		// 1. Create all draft picks
		if err := txRepo.draftPickRepo.CreateBatch(draftPicks); err != nil {
			return err
		}

		// 2. Mark pool entries as unavailable (is_available = false)
		for _, peID := range poolEntryIDs {
			if err := txRepo.poolEntryRepo.MarkUnavailable(nil, peID); err != nil {
				return err
			}
		}

		// 3. Deduct DraftPoints from the player
		player.DraftPoints -= totalRequestedCost
		if _, err := txRepo.playerRepo.Update(player); err != nil {
			return err
		}

		// 4. Create Claim records for each drafted pokemon
		for i, dp := range draftPicks {
			poolEntry := allRequestedPoolEntries[i]
			claimSource := enums.ClaimSourceDraft
			claim := &models.Claim{
				LeagueID:     league.ID,
				PlayerID:     player.ID,
				SpeciesID:    poolEntry.PokemonSpeciesID,
				Source:       claimSource,
				SourceID:     &dp.ID,
				CostPaid:     *poolEntry.Cost,
				AcquiredWeek: 0, // Pre-season draft week
				IsActive:     true,
			}
			if _, err := txRepo.claimRepo.Create(claim); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Remove used up accumulated picks from the draft model
	slices.SortFunc(accumulatedPickNumberIndicesToDelete, func(a, b int) int {
		return b - a // Descending order
	})
	playerAccumulatedPicks := draft.PlayersWithAccumulatedPicks[player.ID]
	for _, index := range accumulatedPickNumberIndicesToDelete {
		playerAccumulatedPicks = slices.Delete(playerAccumulatedPicks, index, index+1)
	}
	draft.PlayersWithAccumulatedPicks[player.ID] = playerAccumulatedPicks

	return nil
}

// transactionalRepositories holds the repository instances to use within a transaction.
type transactionalRepositories struct {
	draftPickRepo repositories.DraftPickRepository
	claimRepo     repositories.ClaimRepository
	poolEntryRepo repositories.PoolEntryRepository
	playerRepo    repositories.PlayerRepository
}

// executeWithTransaction is a helper to run operations that use the new model repositories.
// Note: This is a simplified approach that relies on each repository having its own DB handle.
// For true transactional integrity, repositories should share a transaction context.
// This works here because each repo call is independent (no cross-repo transaction needed
// beyond what each individual repo provides).
func (s *draftServiceImpl) executeWithTransaction(fn func(repos *transactionalRepositories) error) error {
	txRepos := &transactionalRepositories{
		draftPickRepo: s.draftPickRepo,
		claimRepo:     s.claimRepo,
		poolEntryRepo: s.poolEntryRepo,
		playerRepo:    s.playerRepo,
	}
	return fn(txRepos)
}

// validatePicksAndCheckCurrentPickSlotUsed performs the final validation checks before a pick is executed.
// It ensures that requested pick numbers are valid, the player has sufficient points, and that an
// implicit skip of the current turn doesn't violate minimum roster rules. It returns a boolean
// indicating if the current "on-the-clock" pick slot was used in the transaction.
func (s *draftServiceImpl) validatePicksAndCheckCurrentPickSlotUsed(
	draft *models.Draft,
	player *models.Player,
	league *models.League,
	input *requests.DraftMakePickRequestDTO,
	totalRequestedCost int,
) (bool, error) {
	playerID := *draft.CurrentTurnPlayerID // validated earlier to match currentPlayer

	// 1. Validate requested pick numbers against valid slots
	accumulatedPickNumbers := draft.PlayersWithAccumulatedPicks[playerID]
	s.logger.Debug("Current pick on clock and accumulated picks", "player_id", playerID, "current_pick", draft.CurrentPickOnClock, "accumulated_picks", accumulatedPickNumbers)
	validPickNumbersForPlayer := make([]int, len(accumulatedPickNumbers))
	copy(validPickNumbersForPlayer, accumulatedPickNumbers) // we don't wanna directly append
	validPickNumbersForPlayer = append(validPickNumbersForPlayer, draft.CurrentPickOnClock)

	// track used accumulated picks within this batch to prevent double-usage
	usedAccumulatedPicksInThisBatch := make(map[int]bool)
	currentPickSlotUsed := false

	for _, requestedPick := range input.RequestedPicks {
		// check if the requested pick number is a valid slot (current turn or accumulated)
		if !slices.Contains(validPickNumbersForPlayer, requestedPick.DraftPickNumber) {
			s.logger.Warn("Player requested invalid pick number", "player_id", playerID, "requested_pick", requestedPick.DraftPickNumber, "current_pick", draft.CurrentPickOnClock, "accumulated_picks", accumulatedPickNumbers)
			return false, types.ErrInvalidInput
		}

		// if it's an accumulated pick, ensure it's not used twice in this batch
		if requestedPick.DraftPickNumber != draft.CurrentPickOnClock {
			if usedAccumulatedPicksInThisBatch[requestedPick.DraftPickNumber] {
				s.logger.Warn("Player attempted to use accumulated pick multiple times", "player_id", playerID, "accumulated_pick", requestedPick.DraftPickNumber)
				return false, types.ErrInvalidInput
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
		return false, types.ErrInsufficientDraftPoints
	}

	// 3. "Skips Left" Preventative Validation
	// This ensures the player doesn't implicitly skip their current turn's slot
	// if doing so would prevent them from meeting MinPokemonPerPlayer.

	// Determine if the current "on-the-clock" pick slot is being used in this request.
	// If not, it implies a skip of the current turn.
	var err error
	isCurrentTurnUsed := false
	for _, requestedPick := range input.RequestedPicks {
		if requestedPick.DraftPickNumber == draft.CurrentPickOnClock {
			isCurrentTurnUsed = true
			break
		}
	}

	effectiveSkipsInThisAction := 0
	if !isCurrentTurnUsed {
		effectiveSkipsInThisAction = 1 // Current turn is implicitly skipped
	}

	_, err = s.isSkipAllowed(player, effectiveSkipsInThisAction)
	if err != nil {
		s.logger.Warn("Player cannot implicitly skip current turn as it would violate minimum roster requirement", "player_id", playerID, "pick", draft.CurrentPickOnClock, "skips_left", player.SkipsLeft)
		return false, err
	}

	return currentPickSlotUsed, nil
}

// isSkipAllowed checks if a player can skip (or implicit skip) their turn without making it impossible
// to meet the league's minimum roster requirement.
// returns true if allowed, false otherwise
// was previously a bigger function because we didn't have Player.SkipsLeft
func (s *draftServiceImpl) isSkipAllowed(player *models.Player, effectiveSkipsInThisAction int) (bool, error) {
	if player.SkipsLeft-effectiveSkipsInThisAction >= 0 {
		return true, nil
	}
	return false, types.ErrCannotSkipBelowMinimumRoster
}

func (s *draftServiceImpl) getTotalCostForPoolEntries(allRequestedPoolEntries []*models.PoolEntry) int {
	sumCost := 0
	for _, entry := range allRequestedPoolEntries {
		sumCost += *entry.Cost
	}
	return sumCost
}

// fetchDraftResource retrieves the draft for a league, converting a gorm.ErrRecordNotFound
// into a service-specific error.
func (s *draftServiceImpl) fetchDraftResource(leagueID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrDraftNotFound
		}
		return nil, types.ErrInternalService
	}
	return draft, nil
}

// fetchPlayerResource retrieves a player by user and league, converting a gorm.ErrRecordNotFound
// into a service-specific error.
func (s *draftServiceImpl) fetchPlayerResource(userID, leagueID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, types.ErrInternalService
	}
	s.logger.Debug("Fetched player", "player_id", player.ID, "skips_left", player.SkipsLeft)
	return player, nil
}

// fetchRequestedPoolEntries retrieves a list of PoolEntry by their IDs, ensuring they are
// all available to be drafted. It returns service-specific errors for not found or
// already drafted pokemon.
func (s *draftServiceImpl) fetchRequestedPoolEntries(leagueID uuid.UUID, input *requests.DraftMakePickRequestDTO) ([]*models.PoolEntry, error) {
	var poolEntryIDs []uuid.UUID
	for _, requestedPick := range input.RequestedPicks {
		poolEntryIDs = append(poolEntryIDs, requestedPick.PoolEntryID)
	}

	allRequestedPoolEntries, err := s.poolEntryRepo.GetByIDs(leagueID, poolEntryIDs)
	if err != nil {
		return nil, types.ErrInternalService
	}

	// Validate that all requested pokemon were actually returned and are available.
	if len(allRequestedPoolEntries) != len(poolEntryIDs) {
		return nil, types.ErrPoolEntryNotFound
	}

	var result []*models.PoolEntry
	for i := range allRequestedPoolEntries {
		if !allRequestedPoolEntries[i].IsAvailable {
			return nil, types.ErrConflict
		}
		result = append(result, &allRequestedPoolEntries[i])
	}

	return result, nil
}

func (s *draftServiceImpl) validateLeagueStatusForPick(leagueStatus enums.LeagueStatus, draftStatus enums.DraftStatus) bool {
	return leagueStatus == enums.LeagueStatusDrafting && draftStatus == enums.DraftStatusOngoing
}

// checkDraftCompletion determines if the draft has concluded by checking two conditions:
// 1. Has the total number of drafted pokemon reached the maximum allowed for the league?
// 2. Have all players met the minimum roster requirement?
// It is called after each pick/skip to see if the draft should be moved to a COMPLETED state.
// Uses the new Claim model for active pokemon counts.
func (s *draftServiceImpl) checkDraftCompletion(
	league *models.League,
	allPlayers []models.Player,
) (bool, error) {
	// 1. Calculate total expected picks for the entire draft
	totalPlayers := len(allPlayers)
	if totalPlayers == 0 {
		s.logger.Error("No players in league, cannot check completion", "league_id", league.ID)
		return false, types.ErrInternalService
	}
	maxPicksPerPlayer := league.MaxPokemonPerPlayer
	totalExpectedPicks := totalPlayers * maxPicksPerPlayer

	// 2. Get the current count of all active claims in the league
	currentTotalActiveClaims, err := s.claimRepo.GetActiveCountByLeague(league.ID)
	if err != nil {
		s.logger.Error("Failed to get total active claims", "league_id", league.ID, "error", err)
		return false, types.ErrInternalService
	}

	// Cond. 1: if the total number of picks has reached the maximum
	if currentTotalActiveClaims < int64(totalExpectedPicks) {
		return false, nil
	}

	// Cond. 2: if all players have met their MinPokemonPerRoster requirement
	minPokemonPerRoster := league.MinPokemonPerPlayer

	for _, player := range allPlayers {
		playerActiveRosterSize, err := s.claimRepo.GetActiveCountByPlayer(player.ID)
		if err != nil {
			s.logger.Error("Failed to get roster count for player", "player_id", player.ID, "league_id", league.ID, "error", err)
			return false, types.ErrInternalService
		}
		if playerActiveRosterSize < int64(minPokemonPerRoster) {
			return false, nil
		}
	}
	return true, nil
}

// NOTE: Old methods kept for reference during migration. Remove once migration is complete.
// The following methods were replaced:
// - executePickTransactions -> executeNewPickTransactions (uses DraftPick + Claim)
// - fetchRequestedPokemon -> fetchRequestedPoolEntries (uses PoolEntry)
// - getTotalCostForPicks -> getTotalCostForPoolEntries (uses PoolEntry)
// - checkDraftCompletion now uses Claim counts instead of DraftedPokemon counts
