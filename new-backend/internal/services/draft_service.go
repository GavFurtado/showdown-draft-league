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
	logger         *slog.Logger
	draftRepo      repositories.DraftRepository
	leagueRepo     repositories.LeagueRepository
	memberRepo     repositories.LeagueMemberRepository
	webhookService *WebhookService
	schedulerService SchedulerService

	draftPickRepo repositories.DraftPickRepository
	claimRepo     repositories.ClaimRepository
	poolEntryRepo repositories.PoolEntryRepository
}

func NewDraftService(
	logger *slog.Logger,
	leagueRepo repositories.LeagueRepository,
	draftRepo repositories.DraftRepository,
	memberRepo repositories.LeagueMemberRepository,
	webhookService *WebhookService,
) DraftService {
	return &draftServiceImpl{
		logger:        utils.LoggerWithService(logger, "DraftService"),
		draftRepo:     draftRepo,
		leagueRepo:    leagueRepo,
		memberRepo:    memberRepo,
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
			s.logger.Warn("GetDraftByID - draft record not found", "draft_id", draftID, "error", err)
			return nil, types.ErrDraftNotFound
		}
		s.logger.Error("GetDraftByID - error fetching draft", "draft_id", draftID, "error", err)
		return nil, types.ErrInternalService
	}
	return draft, nil
}

func (s *draftServiceImpl) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("GetDraftByLeagueID - draft record not found", "league_id", leagueID, "error", err)
			return nil, types.ErrDraftNotFound
		}
		s.logger.Error("GetDraftByLeagueID - error fetching draft for league", "league_id", leagueID, "error", err)
		return nil, types.ErrInternalService
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
		s.logger.Error("StartDraft - could not get league", "league_id", leagueID, "error", err)
		return nil, types.ErrLeagueNotFound
	}

	// Retrieve members in the league, sorted by draft position
	members, err := s.memberRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("StartDraft - could not get members", "league_id", leagueID, "error", err)
		return nil, types.ErrInternalService
	}

	if len(members) == 0 {
		s.logger.Warn("StartDraft - no members found", "league_id", leagueID)
		return nil, types.ErrNoPlayerForDraft
	}

	switch league.Format.DraftOrderType {
	case enums.DraftOrderTypeRandom:
		r := rand.New(rand.NewSource(time.Now().UnixNano())) // set seed
		r.Shuffle(len(members), func(i, j int) {
			members[i], members[j] = members[j], members[i]
		})

		// Assign new draft positions and update in DB
		for i := range members {
			members[i].DraftPosition = i + 1 // Draft positions are 1-based
			if err := s.memberRepo.UpdateDraftPosition(members[i].ID, members[i].DraftPosition); err != nil {
				s.logger.Error("StartDraft - failed to update draft position", "member_id", members[i].ID, "error", err)
				return nil, types.ErrInternalService
			}
		}
		s.logger.Info("StartDraft - randomized draft order complete", "league_id", leagueID)

	case enums.DraftOrderTypeManual:
		// Members are already sorted by DraftPosition from GetByLeague.
		// This assumes DraftPosition has been set manually prior to starting the draft.
		// Validate that all members have a unique, positive DraftPosition.
		seenPositions := make(map[int]bool)
		for _, m := range members {
			if m.DraftPosition <= 0 {
				s.logger.Error("StartDraft - member has invalid draft position", "member_id", m.ID, "position", m.DraftPosition)
				return nil, types.ErrInvalidDraftPosition
			}
			if seenPositions[m.DraftPosition] {
				s.logger.Error("StartDraft - duplicate draft position found", "position", m.DraftPosition, "member_id", m.ID)
				return nil, types.ErrDuplicateDraftPosition
			}
			seenPositions[m.DraftPosition] = true
		}
		// Ensure all positions from 1 to len(members) are present
		if len(seenPositions) != len(members) {
			s.logger.Error("StartDraft - missing or extra draft positions for manual draft order", "league_id", leagueID)
			return nil, types.ErrIncompleteDraftOrder
		}
		s.logger.Info("StartDraft - using manual draft order", "league_id", leagueID)
	}

	// Initialize the Draft model
	firstMemberID := members[0].ID
	currTime := time.Now()

	draft := &models.Draft{
		LeagueID:                    leagueID,
		Status:                      enums.DraftStatusOngoing,
		CurrentRound:                1,
		CurrentPickInRound:          1,
		CurrentPickOnClock:          1, // formula: ((CurrentRound - 1)*PlayerCount + CurrentPickInRound)
		CurrentTurnMemberID:         &firstMemberID,
		CurrentTurnStartTime:        &currTime,
		TurnTimeLimit:               TurnTimeLimit,
		PlayersWithAccumulatedPicks: make(models.PlayerAccumulatedPicks), // map[uuid.UUID][]int
		StartTime:                   time.Now(),
	}

	// Save the Draft model
	if err := s.draftRepo.CreateDraft(draft); err != nil {
		s.logger.Error("StartDraft - failed to create draft", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	// Update the league status to DRAFTING
	league.Status = enums.LeagueStatusDrafting
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("StartDraft - failed to update league status", "league_id", leagueID, "error", err)
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
			PlayerID: *draft.CurrentTurnMemberID,
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
// MakePick makes one or more picks (if accumulated) during drafting phase;
// Different from ForcePick (not implemented yet),
// MakePick does all the required checks (there's a lot of checks) and validates the input
//
// NOTE: This method has been migrated to write DraftPick + Claim records instead of DraftedPokemon.
func (s *draftServiceImpl) MakePick(
	currentUser *models.User,
	leagueID uuid.UUID,
	input *requests.DraftMakePickRequestDTO,
) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		s.logger.Error("MakePick - could not find league", "user_id", currentUser.ID, "league_id", leagueID, "error", err)
		return types.ErrLeagueNotFound
	}

	// fetch draft for league
	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case types.ErrDraftNotFound:
			s.logger.Warn("MakePick - draft not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("MakePick - error fetching draft", "user_id", currentUser.ID, "error", err)
		}
		return err
	}

	member, err := s.fetchMemberResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case types.ErrPlayerNotFound:
			s.logger.Warn("MakePick - member not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("MakePick - error fetching member", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		}
		return err
	}

	// START early checks to prevent a expensive checks later
	// check if it's the right member's turn
	if currentTurnMemberID := *draft.CurrentTurnMemberID; currentTurnMemberID != member.ID {
		s.logger.Warn("MakePick - member tried to draft when not their turn", "member_id", currentTurnMemberID, "current_turn_member_id", *draft.CurrentTurnMemberID)
		return types.ErrUnauthorized
	}

	// check if number of requested picks is valid for the member
	if input.RequestedPickCount > len(draft.PlayersWithAccumulatedPicks[member.ID])+1 {
		s.logger.Warn("MakePick - member requested too many draft picks", "user_id", currentUser.ID, "member_id", member.ID)
		return types.ErrTooManyRequestedPicks
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		s.logger.Warn("MakePick - league is not in drafting status", "user_id", currentUser.ID, "league_id", league.ID)
		return types.ErrInvalidState
	}
	// END early checks

	// fetch all the pool entries requested
	// expensive
	allRequestedPoolEntries, err := s.fetchRequestedPoolEntries(league.ID, input)
	if err != nil {
		switch err {
		case types.ErrPoolEntryNotFound:
			s.logger.Warn("MakePick - one or more pool entries not found", "user_id", currentUser.ID, "error", err)
		case types.ErrConflict:
			s.logger.Warn("MakePick - one or more pool entries not available", "user_id", currentUser.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("MakePick - error fetching requested pool entries", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		}
		return err
	}

	// get member count; needed in multiple places
	memberCount, err := s.memberRepo.GetCountByLeague(league.ID)
	if err != nil {
		s.logger.Error("MakePick - failed to get member count", "league_id", league.ID, "error", err)
		return types.ErrInternalService
	}
	if memberCount == 0 { // this should never happen if the draft has started or if the league even exists
		s.logger.Error("MakePick - no members in league (unreachable code)", "league_id", league.ID)
		return types.ErrInternalService
	}

	totalRequestedCost := s.getTotalCostForPoolEntries(allRequestedPoolEntries)

	// perform remaining validation
	currentPickSlotUsed, err := s.validatePicksAndCheckCurrentPickSlotUsed(draft, member, league, input, totalRequestedCost)
	if err != nil {
		switch err {
		case types.ErrInvalidInput:
			s.logger.Warn("MakePick - invalid pick number in request", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInsufficientDraftPoints:
			s.logger.Warn("MakePick - insufficient draft points", "user_id", currentUser.ID, "league_id", league.ID, "draft_points", member.DraftPoints, "error", err)
		}
		return err
	}

	// execute picks (new model: creates DraftPick + Claim instead of DraftedPokemon)
	err = s.executeNewPickTransactions(draft, league, member, allRequestedPoolEntries, input, memberCount, totalRequestedCost)
	if err != nil {
		s.logger.Error("MakePick - batch transaction unsuccessful", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		return err
	}

	// get all members to change set the current member's turn for the next one
	allMembers, err := s.memberRepo.GetByLeague(draft.LeagueID)
	if err != nil {
		s.logger.Error("MakePick - could not get all members", "league_id", league.ID, "error", err)
		return types.ErrInternalService
	}

	// advance turn (if CurrentPickSlotUsed) and update draft model
	draft, err = s.advanceDraftState(draft, league, member, allMembers, int(memberCount), currentPickSlotUsed)
	if err != nil {
		s.logger.Error("MakePick - error advancing draft state", "league_id", league.ID, "error", err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		s.logger.Info("MakePick - draft completed", "league_id", draft.LeagueID)
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
			PlayerID: *draft.CurrentTurnMemberID,
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
		s.logger.Error("SkipTurn - could not find league", "user_id", currentUser.ID, "league_id", leagueID, "error", err)
		return types.ErrLeagueNotFound
	}

	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case types.ErrDraftNotFound:
			s.logger.Warn("SkipTurn - draft not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("SkipTurn - error fetching draft", "user_id", currentUser.ID, "error", err)
		}
		return err
	}

	member, err := s.fetchMemberResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case types.ErrPlayerNotFound:
			s.logger.Warn("SkipTurn - member not found", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		case types.ErrInternalService:
			s.logger.Error("SkipTurn - error fetching member", "user_id", currentUser.ID, "league_id", league.ID, "error", err)
		}
		return err
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		s.logger.Warn("SkipTurn - league is not in drafting status", "user_id", currentUser.ID, "league_id", league.ID)
		return types.ErrInvalidState
	}
	// check if it's the right member's turn
	if currentTurnMemberID := *draft.CurrentTurnMemberID; currentTurnMemberID != member.ID {
		s.logger.Warn("SkipTurn - member tried to draft when not their turn", "member_id", currentTurnMemberID, "current_turn_member_id", *draft.CurrentTurnMemberID)
		return types.ErrUnauthorized
	}

	// get all members to change set the current member's turn for the next one
	allMembers, err := s.memberRepo.GetByLeague(draft.LeagueID)
	if err != nil {
		s.logger.Error("SkipTurn - could not get all members", "league_id", league.ID, "error", err)
		return types.ErrInternalService
	}

	effectiveSkipsInThisAction := 1 // One skip for the current turn
	_, err = s.isSkipAllowed(member, effectiveSkipsInThisAction)
	if err != nil {
		s.logger.Warn("SkipTurn - member cannot skip current turn's pick as it would violate minimum roster requirement", "member_id", member.ID, "pick", draft.CurrentPickOnClock, "skips_left", member.SkipsLeft)
		return err
	}

	member.SkipsLeft -= effectiveSkipsInThisAction
	s.logger.Debug("SkipTurn - skips left before DB update", "member_id", member.ID, "skips_left", member.SkipsLeft)
	if _, err := s.memberRepo.Update(member); err != nil {
		s.logger.Error("SkipTurn - failed to update member skipsLeft in DB", "member_id", member.ID, "error", err)
		return types.ErrInternalService
	}
	// Re-fetch member to confirm DB state
	updatedMember, err := s.memberRepo.GetByID(member.ID)
	if err != nil {
		s.logger.Error("SkipTurn - failed to re-fetch member after update", "member_id", member.ID, "error", err)
		return types.ErrInternalService
	}
	s.logger.Debug("SkipTurn - skips left after DB re-fetch", "member_id", updatedMember.ID, "skips_left", updatedMember.SkipsLeft)

	draft, err = s.advanceDraftState(draft, league, member, allMembers, len(allMembers), false)
	if err != nil {
		s.logger.Error("SkipTurn - error advancing draft state", "league_id", league.ID, "error", err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		s.logger.Info("SkipTurn - draft completed", "league_id", draft.LeagueID)
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
			PlayerID: *draft.CurrentTurnMemberID,
		},
	}
	s.schedulerService.RegisterTask(task)

	// successful skip
	return nil
}

// AutoSkipTurn is called by the SchedulerService when a player's turn timer expires.
// It attempts to automatically skip the turn. If the skip is not allowed (e.g., it
// would violate minimum roster size), the draft is paused for manual intervention.
func (s *draftServiceImpl) AutoSkipTurn(memberID, leagueID uuid.UUID) error {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("AutoSkipTurn - member not found", "member_id", memberID, "league_id", leagueID, "error", err)
			return types.ErrPlayerNotFound
		}
		s.logger.Error("AutoSkipTurn - error fetching member", "member_id", memberID, "league_id", leagueID, "error", err)
		return types.ErrInternalService
	}
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			s.logger.Warn("AutoSkipTurn - league not found", "member_id", memberID, "league_id", leagueID, "error", err)
			return types.ErrLeagueNotFound
		default:
			s.logger.Error("AutoSkipTurn - could not fetch league", "league_id", leagueID, "error", err)
			return types.ErrInternalService
		}
	}
	draft, err := s.fetchDraftResource(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			s.logger.Warn("AutoSkipTurn - draft not found", "member_id", memberID, "league_id", leagueID, "error", err)
			return types.ErrDraftNotFound
		default:
			s.logger.Error("AutoSkipTurn - error fetching draft", "member_id", memberID, "error", err)
			return types.ErrInternalService
		}
	}

	effectiveSkipsInThisAction := 1
	allowed, err := s.isSkipAllowed(member, effectiveSkipsInThisAction)
	if !allowed {
		s.logger.Error("AutoSkipTurn - cannot auto skip", "member_id", memberID, "league_id", leagueID, "error", err, "skips_left", member.SkipsLeft)
		// set Draft to PAUSED status, awaiting manual league staff intervention
		draft.Status = enums.DraftStatusPaused
		draft, err = s.draftRepo.UpdateDraft(draft)
		if err != nil {
			s.logger.Error("AutoSkipTurn - could not update draft status to PAUSED", "draft_id", draft.ID, "error", err)
			return types.ErrInternalService
		}
		s.logger.Info("AutoSkipTurn - draft paused, awaiting manual intervention", "league_id", leagueID)
		return types.ErrDraftPausedForIntervention
	}

	member.SkipsLeft -= effectiveSkipsInThisAction
	s.logger.Debug("AutoSkipTurn - skips left before DB update", "member_id", member.ID, "skips_left", member.SkipsLeft)
	if _, err := s.memberRepo.Update(member); err != nil {
		s.logger.Error("AutoSkipTurn - failed to update member skipsLeft in DB", "member_id", member.ID, "error", err)
		return types.ErrInternalService
	}
	// Re-fetch member to confirm DB state
	updatedMember, err := s.memberRepo.GetByID(member.ID)
	if err != nil {
		s.logger.Error("AutoSkipTurn - failed to re-fetch member after update", "member_id", member.ID, "error", err)
		return types.ErrInternalService
	}
	s.logger.Debug("AutoSkipTurn - skips left after DB re-fetch", "member_id", updatedMember.ID, "skips_left", updatedMember.SkipsLeft)
	s.logger.Debug("AutoSkipTurn - skips left updated after auto-skip", "member_id", member.ID, "skips_left", member.SkipsLeft)

	allMembers, err := s.memberRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("AutoSkipTurn - could not get all members", "league_id", leagueID, "error", err)
		return types.ErrInternalService
	}

	draft, err = s.advanceDraftState(draft, league, member, allMembers, len(allMembers), false)
	if err != nil {
		s.logger.Error("AutoSkipTurn - could not advance draft")
		return err
	}

	// AutoSkip shouldn't deregister the task because it is called by the schedulerService and thus
	// already dereigsters it automatically
	if draft.Status == enums.DraftStatusCompleted {
		s.logger.Info("AutoSkipTurn - draft completed", "league_id", draft.LeagueID)
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
			PlayerID: *draft.CurrentTurnMemberID,
		},
	}

	s.schedulerService.RegisterTask(task)
	s.logger.Info("AutoSkipTurn - success")
	// success
	return nil
}

// advanceDraftState moves the draft to the next turn or completes it.
// It increments the pick counter, checks if the draft's end conditions are met,
// determines the next player based on the draft order (linear or snake), and updates the draft model.
func (s *draftServiceImpl) advanceDraftState(
	draft *models.Draft,
	league *models.League,
	member *models.LeagueMember, // The member whose turn just ended/skipped
	allMembers []models.LeagueMember, // All members in the league, for turn progression
	memberCount int,
	currentPickSlotUsed bool, // true if draft.CurrentPickOnClock was used in the request, false if skipped/implicitly skipped
) (*models.Draft, error) {
	if !currentPickSlotUsed {
		// i.e., a skip/implicit skip. Append CurrentPickOnClock to
		// accumulated picks for that member
		s.logger.Debug("advanceDraftState - member skipping pick", "member_id", member.ID, "pick", draft.CurrentPickOnClock, "accumulated_picks", draft.PlayersWithAccumulatedPicks[member.ID])
		draft.PlayersWithAccumulatedPicks[member.ID] = append(draft.PlayersWithAccumulatedPicks[member.ID], draft.CurrentPickOnClock)
	}

	draft.CurrentPickOnClock++ // unconditonal increment

	// Check for draft completion
	isDraftCompleted, err := s.checkDraftCompletion(league, allMembers)
	if err != nil {
		s.logger.Error("advanceDraftState - error checking draft completion", "league_id", league.ID, "error", err)
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
			s.logger.Error("advanceDraftState - failed to update draft status to COMPLETED", "league_id", league.ID, "error", err)
			return nil, fmt.Errorf("failed to update draft state on completion: %w", err)
		}
		// save league status updated by checkDraftCompletion
		if _, err := s.leagueRepo.UpdateLeague(league); err != nil { // pray this never happens type shit
			// should prolly revert the draft update
			s.logger.Error("advanceDraftState - failed to update league status to POST_DRAFT", "league_id", league.ID, "error", err)
			return nil, fmt.Errorf("failed to update league status on completion: %w", err)
		}

		return draft, nil // Draft completed, states saved. We're so done
	}

	// If draft is still ongoing,
	// Recalculate CurrentRound and CurrentPickInRound based on draft.CurrentPickOnClock
	draft.CurrentRound = ((draft.CurrentPickOnClock - 1) / int(memberCount)) + 1
	draft.CurrentPickInRound = ((draft.CurrentPickOnClock - 1) % int(memberCount)) + 1

	currentMemberIdx := -1
	for i, m := range allMembers { // there is likely some smort mafs you can do here to avoid an O(n) search. im stupid tho
		if m.ID == member.ID {
			currentMemberIdx = i
			break
		}
	}
	if currentMemberIdx == -1 { // this is an impossible case
		s.logger.Error("advanceDraftState - current member not found in allMembers list (unreachable)", "member_id", member.ID)
		return nil, types.ErrInternalService
	}

	var nextMemberIdx int
	if league.Format.IsSnakeRoundDraft {
		if draft.CurrentRound%2 != 0 { // Odd round (forward order)
			nextMemberIdx = draft.CurrentPickInRound - 1
		} else { // Even round (reverse order)
			nextMemberIdx = int(memberCount) - draft.CurrentPickInRound
		}
	} else { // linear draft
		nextMemberIdx = currentMemberIdx + 1
	}

	if nextMemberIdx >= int(memberCount) || nextMemberIdx < 0 {
		// The CurrentRound and CurrentPickInRound are already correctly set by recalculation.
		// We just need to adjust nextMemberIdx for the start of the new round.
		if league.Format.IsSnakeRoundDraft && draft.CurrentRound%2 == 0 { // if snake round drafting and new round is even
			nextMemberIdx = int(memberCount) - 1 // last member in reverse order
		} else {
			nextMemberIdx = 0 // first member in forward order
		}
	}

	// finally set the next turn of member
	nextTurnMember := allMembers[nextMemberIdx]
	draft.CurrentTurnMemberID = &nextTurnMember.ID
	draft.CurrentTurnStartTime = func() *time.Time { t := time.Now(); return &t }()

	draft, err = s.draftRepo.UpdateDraft(draft)
	if err != nil {
		s.logger.Error("advanceDraftState - failed to update draft", "error", err)
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
	member *models.LeagueMember,
	allRequestedPoolEntries []*models.PoolEntry,
	input *requests.DraftMakePickRequestDTO,
	memberCount int64,
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
		draftRoundNumber := ((requestedPick.DraftPickNumber - 1) / int(memberCount)) + 1

		// Build DraftPick (immutable event log)
		draftPick := models.DraftPick{
			DraftID:     draft.ID,
			PlayerID:    member.ID,
			PoolEntryID: currentPoolEntry.ID,
			RoundNumber: draftRoundNumber,
			PickNumber:  requestedPick.DraftPickNumber,
		}
		draftPicks = append(draftPicks, draftPick)

		// Cache accumulated pick numbers to remove
		if accumPickIndex := slices.Index(
			draft.PlayersWithAccumulatedPicks[member.ID], requestedPick.DraftPickNumber,
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

		// 3. Deduct DraftPoints from the member
		member.DraftPoints -= totalRequestedCost
		if _, err := txRepo.memberRepo.Update(member); err != nil {
			return err
		}

		// 4. Create Claim records for each drafted pokemon
		for i, dp := range draftPicks {
			poolEntry := allRequestedPoolEntries[i]
			claimSource := enums.ClaimSourceDraft
			claim := &models.Claim{
				LeagueID:     league.ID,
				PlayerID:     member.ID,
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
	memberAccumulatedPicks := draft.PlayersWithAccumulatedPicks[member.ID]
	for _, index := range accumulatedPickNumberIndicesToDelete {
		memberAccumulatedPicks = slices.Delete(memberAccumulatedPicks, index, index+1)
	}
	draft.PlayersWithAccumulatedPicks[member.ID] = memberAccumulatedPicks

	return nil
}

// transactionalRepositories holds the repository instances to use within a transaction.
type transactionalRepositories struct {
	draftPickRepo repositories.DraftPickRepository
	claimRepo     repositories.ClaimRepository
	poolEntryRepo repositories.PoolEntryRepository
	memberRepo    repositories.LeagueMemberRepository
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
		memberRepo:    s.memberRepo,
	}
	return fn(txRepos)
}

// validatePicksAndCheckCurrentPickSlotUsed performs the final validation checks before a pick is executed.
// It ensures that requested pick numbers are valid, the player has sufficient points, and that an
// implicit skip of the current turn doesn't violate minimum roster rules. It returns a boolean
// indicating if the current "on-the-clock" pick slot was used in the transaction.
func (s *draftServiceImpl) validatePicksAndCheckCurrentPickSlotUsed(
	draft *models.Draft,
	member *models.LeagueMember,
	league *models.League,
	input *requests.DraftMakePickRequestDTO,
	totalRequestedCost int,
) (bool, error) {
	memberID := *draft.CurrentTurnMemberID // validated earlier to match currentMember

	// 1. Validate requested pick numbers against valid slots
	accumulatedPickNumbers := draft.PlayersWithAccumulatedPicks[memberID]
	s.logger.Debug("validatePicksAndCheckCurrentPickSlotUsed - current pick on clock and accumulated picks", "member_id", memberID, "current_pick", draft.CurrentPickOnClock, "accumulated_picks", accumulatedPickNumbers)
	validPickNumbersForMember := make([]int, len(accumulatedPickNumbers))
	copy(validPickNumbersForMember, accumulatedPickNumbers) // we don't wanna directly append
	validPickNumbersForMember = append(validPickNumbersForMember, draft.CurrentPickOnClock)

	// track used accumulated picks within this batch to prevent double-usage
	usedAccumulatedPicksInThisBatch := make(map[int]bool)
	currentPickSlotUsed := false

	for _, requestedPick := range input.RequestedPicks {
		// check if the requested pick number is a valid slot (current turn or accumulated)
		if !slices.Contains(validPickNumbersForMember, requestedPick.DraftPickNumber) {
			s.logger.Warn("validatePicksAndCheckCurrentPickSlotUsed - member requested invalid pick number", "member_id", memberID, "requested_pick", requestedPick.DraftPickNumber, "current_pick", draft.CurrentPickOnClock, "accumulated_picks", accumulatedPickNumbers)
			return false, types.ErrInvalidInput
		}

		// if it's an accumulated pick, ensure it's not used twice in this batch
		if requestedPick.DraftPickNumber != draft.CurrentPickOnClock {
			if usedAccumulatedPicksInThisBatch[requestedPick.DraftPickNumber] {
				s.logger.Warn("validatePicksAndCheckCurrentPickSlotUsed - member attempted to use accumulated pick multiple times", "member_id", memberID, "accumulated_pick", requestedPick.DraftPickNumber)
				return false, types.ErrInvalidInput
			}
			usedAccumulatedPicksInThisBatch[requestedPick.DraftPickNumber] = true
		}

		// check if the current pick slot is being used in this request
		if requestedPick.DraftPickNumber == draft.CurrentPickOnClock {
			currentPickSlotUsed = true
		}
	}

	// 2. Check if member has enough draft points for the entire batch
	if member.DraftPoints < totalRequestedCost {
		return false, types.ErrInsufficientDraftPoints
	}

	// 3. "Skips Left" Preventative Validation
	// This ensures the member doesn't implicitly skip their current turn's slot
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

	_, err = s.isSkipAllowed(member, effectiveSkipsInThisAction)
	if err != nil {
			s.logger.Warn("validatePicksAndCheckCurrentPickSlotUsed - member cannot implicitly skip current turn as it would violate minimum roster requirement", "member_id", memberID, "pick", draft.CurrentPickOnClock, "skips_left", member.SkipsLeft)
		return false, err
	}

	return currentPickSlotUsed, nil
}

// isSkipAllowed checks if a player can skip (or implicit skip) their turn without making it impossible
// to meet the league's minimum roster requirement.
// returns true if allowed, false otherwise
// was previously a bigger function because we didn't have Player.SkipsLeft
func (s *draftServiceImpl) isSkipAllowed(member *models.LeagueMember, effectiveSkipsInThisAction int) (bool, error) {
	if member.SkipsLeft-effectiveSkipsInThisAction >= 0 {
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

// fetchMemberResource retrieves a member by user and league, converting a gorm.ErrRecordNotFound
// into a service-specific error.
func (s *draftServiceImpl) fetchMemberResource(userID, leagueID uuid.UUID) (*models.LeagueMember, error) {
	member, err := s.memberRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, types.ErrInternalService
	}
	s.logger.Debug("fetchMemberResource - fetched member", "member_id", member.ID, "skips_left", member.SkipsLeft)
	return member, nil
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
	allMembers []models.LeagueMember,
) (bool, error) {
	// 1. Calculate total expected picks for the entire draft
	totalMembers := len(allMembers)
	if totalMembers == 0 {
		s.logger.Error("checkDraftCompletion - no members in league, cannot check completion", "league_id", league.ID)
		return false, types.ErrInternalService
	}
	maxPicksPerMember := league.MaxPokemonPerPlayer
	totalExpectedPicks := totalMembers * maxPicksPerMember

	// 2. Get the current count of all active claims in the league
	currentTotalActiveClaims, err := s.claimRepo.GetActiveCountByLeague(league.ID)
	if err != nil {
		s.logger.Error("checkDraftCompletion - failed to get total active claims", "league_id", league.ID, "error", err)
		return false, types.ErrInternalService
	}

	// Cond. 1: if the total number of picks has reached the maximum
	if currentTotalActiveClaims < int64(totalExpectedPicks) {
		return false, nil
	}

	// Cond. 2: if all members have met their MinPokemonPerRoster requirement
	minPokemonPerRoster := league.MinPokemonPerPlayer

	for _, member := range allMembers {
		memberActiveRosterSize, err := s.claimRepo.GetActiveCountByPlayer(member.ID)
		if err != nil {
			s.logger.Error("checkDraftCompletion - failed to get roster count for member", "member_id", member.ID, "league_id", league.ID, "error", err)
			return false, types.ErrInternalService
		}
		if memberActiveRosterSize < int64(minPokemonPerRoster) {
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
