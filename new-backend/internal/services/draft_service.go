package services

import (
	"errors"
	"fmt"
	"log"
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
	draftRepo        repositories.DraftRepository
	leagueRepo       repositories.LeagueRepository
	memberRepo       repositories.LeagueMemberRepository
	webhookService   *WebhookService
	schedulerService SchedulerService

	draftPickRepo repositories.DraftPickRepository
	claimRepo     repositories.ClaimRepository
	poolEntryRepo repositories.PoolEntryRepository
}

func NewDraftService(
	leagueRepo repositories.LeagueRepository,
	draftRepo repositories.DraftRepository,
	memberRepo repositories.LeagueMemberRepository,
	webhookService *WebhookService,
) DraftService {
	return &draftServiceImpl{
		draftRepo:      draftRepo,
		leagueRepo:     leagueRepo,
		memberRepo:     memberRepo,
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
			log.Printf("ERROR: (DraftService: GetDraftByID) - draft record for ID %s not found: %v", draftID, err)
			return nil, types.ErrDraftNotFound
		}
		log.Printf("ERROR: (DraftService: GetDraftByID) - Error fetching draft %s: %v", draftID, err)
		return nil, types.ErrInternalService
	}
	return draft, nil
}

func (s *draftServiceImpl) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	draft, err := s.draftRepo.GetDraftByLeagueID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("ERROR: (DraftService: GetDraftByID) - draft record for league ID %s not found: %v", leagueID, err)
			return nil, types.ErrDraftNotFound
		}
		log.Printf("ERROR: (DraftService: GetDraftByID) - Error fetching draft for league %s: %v", leagueID, err)
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
		log.Printf("LOG: (Error: DraftService.StartDraft) - Could not get league %s: %v\n", leagueID, err)
		return nil, types.ErrLeagueNotFound
	}

	// Retrieve members in the league, sorted by draft position
	members, err := s.memberRepo.GetByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (Error: DraftService.StartDraft) - Could not get members for league %s: %v\n", leagueID, err)
		return nil, types.ErrInternalService
	}

	if len(members) == 0 {
		log.Printf("LOG: (Error: DraftService.StartDraft) - No members found for league %s\n", leagueID)
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
				log.Printf("LOG: (Error: DraftService.StartDraft) - Failed to update draft position for member %s: %v\n", members[i].ID, err)
				return nil, types.ErrInternalService
			}
		}
		log.Printf("LOG: (DraftService.StartDraft) - Randomized draft order for league %s complete.\n", leagueID)

	case enums.DraftOrderTypeManual:
		// Members are already sorted by DraftPosition from GetByLeague.
		// This assumes DraftPosition has been set manually prior to starting the draft.
		// Validate that all members have a unique, positive DraftPosition.
		seenPositions := make(map[int]bool)
		for _, m := range members {
			if m.DraftPosition <= 0 {
				log.Printf("ERROR: (DraftService: StartDraft) - Member %s has invalid draft position %d for manual draft order.\n", m.ID, m.DraftPosition)
				return nil, types.ErrInvalidDraftPosition
			}
			if seenPositions[m.DraftPosition] {
				log.Printf("ERROR: (DraftService: StartDraft) - Duplicate draft position %d found for member %s in manual draft order.\n", m.DraftPosition, m.ID)
				return nil, types.ErrDuplicateDraftPosition
			}
			seenPositions[m.DraftPosition] = true
		}
		// Ensure all positions from 1 to len(members) are present
		if len(seenPositions) != len(members) {
			log.Printf("ERROR: (DraftService: StartDraft) - Missing or extra draft positions for manual draft order in league %s.\n", leagueID)
			return nil, types.ErrIncompleteDraftOrder
		}
		log.Printf("LOG: (DraftService: StartDraft) - Using manual draft order for league %s.\n", leagueID)
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
		log.Printf("LOG: (DraftService: MakePick) - (user %s) could not find league %s: %v\n", currentUser.ID, leagueID, err)
		return types.ErrLeagueNotFound
	}

	// fetch draft for league
	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case types.ErrDraftNotFound:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) draft for leagueID %s not found: %v\n", currentUser.ID, league.ID, err)
		case types.ErrInternalService:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) error fetching draft: %v\n", currentUser.ID, err)
		}
		return err
	}

	member, err := s.fetchMemberResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case types.ErrPlayerNotFound:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) Member in league %s not found: %v\n", currentUser.ID, league.ID, err)
		case types.ErrInternalService:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) Error fetching member in league %s: %v\n", currentUser.ID, league.ID, err)
		}
		return err
	}

	// START early checks to prevent a expensive checks later
	// check if it's the right member's turn
	if currentTurnMemberID := *draft.CurrentTurnMemberID; currentTurnMemberID != member.ID {
		log.Printf("LOG: (DraftService: MakePick) - member %s tried to draft when it isn't their turn. Current Turn: Member %s\n", currentTurnMemberID, *draft.CurrentTurnMemberID)
		return types.ErrUnauthorized
	}

	// check if number of requested picks is valid for the member
	if input.RequestedPickCount > len(draft.PlayersWithAccumulatedPicks[member.ID])+1 {
		log.Printf("LOG: (DraftService: MakePick) -  (user %s) Member %s requested too many draft picks\n", currentUser.ID, member.ID)
		return types.ErrTooManyRequestedPicks
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		log.Printf("LOG: (DraftService: MakePick) - (user %s) league %s is not in drafting status: %v", currentUser.ID, league.ID, err)
		return types.ErrInvalidState
	}
	// END early checks

	// fetch all the pool entries requested
	// expensive
	allRequestedPoolEntries, err := s.fetchRequestedPoolEntries(league.ID, input)
	if err != nil {
		switch err {
		case types.ErrLeaguePokemonNotFound:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) One or more pool entries were not found: %v\n", currentUser.ID, err)
		case types.ErrConflict:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) One or more pool entries are not available for drafting: %v\n", currentUser.ID, err)
		case types.ErrInternalService:
			log.Printf("LOG: (DraftService: MakePick) - (user %s) error fetching requested pool entries for league %s: %v\n", currentUser.ID, league.ID, err)
		}
		return err
	}

	// get member count; needed in multiple places
	memberCount, err := s.memberRepo.GetCountByLeague(league.ID)
	if err != nil {
		log.Printf("DraftService: MakePick - failed to get member count for league %s: %v\n", league.ID, err)
		return types.ErrInternalService
	}
	if memberCount == 0 { // this should never happen if the draft has started or if the league even exists
		log.Printf("DraftService: MakePick - no members in league %d. (Unreachable Code)\n", league.ID)
		return types.ErrInternalService
	}

	totalRequestedCost := s.getTotalCostForPoolEntries(allRequestedPoolEntries)

	// perform remaining validation
	currentPickSlotUsed, err := s.validatePicksAndCheckCurrentPickSlotUsed(draft, member, league, input, totalRequestedCost)
	if err != nil {
		switch err {
		case types.ErrInvalidInput:
			log.Printf("LOG: (DraftService: MakePick): (user %s; league %s) Invalid pick number in request: %v\n", currentUser.ID, league.ID, err)
		case types.ErrInsufficientDraftPoints:
			log.Printf("LOG: (DraftService: MakePick): (user %s; league %s) Insufficient draft points (%d) for transaction: %v\n", currentUser.ID, league.ID, member.DraftPoints, err)
		}
		return err
	}

	// execute picks (new model: creates DraftPick + Claim instead of DraftedPokemon)
	err = s.executeNewPickTransactions(draft, league, member, allRequestedPoolEntries, input, memberCount, totalRequestedCost)
	if err != nil {
		log.Printf("LOG: (DraftService: MakePick): (user %s; league %s) Batch transaction unsucessful: %v\n", currentUser.ID, league.ID, err)
		return err
	}

	// get all members to change set the current member's turn for the next one
	allMembers, err := s.memberRepo.GetByLeague(draft.LeagueID)
	if err != nil {
		log.Printf("DraftService: MakePick - Could not get all members in league %s: %v\n", league.ID, err)
		return types.ErrInternalService
	}

	// advance turn (if CurrentPickSlotUsed) and update draft model
	draft, err = s.advanceDraftState(draft, league, member, allMembers, int(memberCount), currentPickSlotUsed)
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
		log.Printf("LOG: (DraftService: SkipTurn) - (user %s) could not find league %s: %v\n", currentUser.ID, leagueID, err)
		return types.ErrLeagueNotFound
	}

	draft, err := s.fetchDraftResource(league.ID)
	if err != nil {
		switch err {
		case types.ErrDraftNotFound:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Draft for leagueID %s not found: %v\n", currentUser.ID, league.ID, err)
		case types.ErrInternalService:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Error fetching draft: %v\n", currentUser.ID, err)
		}
		return err
	}

	member, err := s.fetchMemberResource(currentUser.ID, league.ID)
	if err != nil {
		switch err {
		case types.ErrPlayerNotFound:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Member in league %s not found: %v\n", currentUser.ID, league.ID, err)
		case types.ErrInternalService:
			log.Printf("LOG: (DraftService: SkipTurn) - (user %s) Error fetching member in league %s: %v\n", currentUser.ID, league.ID, err)
		}
		return err
	}

	// check league status
	if isValidStatus := s.validateLeagueStatusForPick(league.Status, draft.Status); !isValidStatus {
		log.Printf("LOG: (DraftService: SkipTurn) - (user %s) league %s is not in drafting status: %v", currentUser.ID, league.ID, err)
		return types.ErrInvalidState
	}
	// check if it's the right member's turn
	if currentTurnMemberID := *draft.CurrentTurnMemberID; currentTurnMemberID != member.ID {
		log.Printf("LOG: (DraftService: SkipTurn) - member %s tried to draft when it isn't their turn. Current Turn: Member %s\n", currentTurnMemberID, *draft.CurrentTurnMemberID)
		return types.ErrUnauthorized
	}

	// get all members to change set the current member's turn for the next one
	allMembers, err := s.memberRepo.GetByLeague(draft.LeagueID)
	if err != nil {
		log.Printf("LOG: (DraftService: SkipTurn) - Could not get all members in league %s: %v\n", league.ID, err)
		return types.ErrInternalService
	}

	effectiveSkipsInThisAction := 1 // One skip for the current turn
	_, err = s.isSkipAllowed(member, effectiveSkipsInThisAction)
	if err != nil {
		log.Printf("LOG: (DraftService: SkipTurn) - Member %s cannot skip current turn's pick (%d) as it would violate minimum roster requirement. Skips left: %d.\n",
			member.ID, draft.CurrentPickOnClock, member.SkipsLeft)
		return err
	}

	member.SkipsLeft -= effectiveSkipsInThisAction
	log.Printf("DEBUG: (DraftService: SkipTurn) - Member %s SkipsLeft BEFORE DB update: %d\n", member.ID, member.SkipsLeft)
	if _, err := s.memberRepo.Update(member); err != nil {
		log.Printf("CRITICAL ERROR: (DraftService: SkipTurn) - Failed to update member %s skipsLeft in DB: %v\n", member.ID, err)
		return types.ErrInternalService
	}
	// Re-fetch member to confirm DB state
	updatedMember, err := s.memberRepo.GetByID(member.ID)
	if err != nil {
		log.Printf("CRITICAL ERROR: (DraftService: SkipTurn) - Failed to re-fetch member %s after update: %v\n", member.ID, err)
		return types.ErrInternalService
	}
	log.Printf("DEBUG: (DraftService: SkipTurn) - Member %s SkipsLeft AFTER DB re-fetch: %d\n", updatedMember.ID, updatedMember.SkipsLeft)

	draft, err = s.advanceDraftState(draft, league, member, allMembers, len(allMembers), false)
	if err != nil {
		log.Printf("LOG: (DraftService: SkipTurn) - Error occured when attempting to advance draft state for league %s: %v\n", league.ID, err)
		return err
	}

	if draft.Status == enums.DraftStatusCompleted {
		fmt.Printf("INFO: (DraftService: SkipTurn) - Draft Action (for league %s) was successful and Draft was detected to be COMPLETED. DraftStatus updated to COMPLETED.\n", draft.LeagueID)
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
			log.Printf("ERROR: (DraftService: AutoSkipTurn) - Member %s in league %s not found: %v\n", memberID, leagueID, err)
			return types.ErrPlayerNotFound
		}
		log.Printf("ERROR: (DraftService: AutoSkipTurn) - Error fetching member %s in league %s: %v\n", memberID, leagueID, err)
		return types.ErrInternalService
	}
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			log.Printf("LOG: (DraftService: AutoSkipTurn) - (member %s) League %s not found: %v\n", memberID, leagueID, err)
			return types.ErrLeagueNotFound
		default:
			log.Printf("LOG: (DraftService: AutoSkipTurn) - Could not fetch league %s: %v\n", leagueID, err)
			return types.ErrInternalService
		}
	}
	draft, err := s.fetchDraftResource(leagueID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			log.Printf("LOG: (DraftService: AutoSkipTurn) - (member %s) Draft for leagueID %s not found: %v\n", memberID, leagueID, err)
			return types.ErrDraftNotFound
		default:
			log.Printf("LOG: (DraftService: AutoSkipTurn) - (member %s) Error fetching draft: %v\n", memberID, err)
			return types.ErrInternalService
		}
	}

	effectiveSkipsInThisAction := 1
	allowed, err := s.isSkipAllowed(member, effectiveSkipsInThisAction)
	if !allowed {
		log.Printf("ERROR: (DraftService: AutoSkipTurn) - Cannot auto skip for member %s, league %s: %v. Skips left: %d\n", memberID, leagueID, err, member.SkipsLeft)
		// set Draft to PAUSED status, awaiting manual league staff intervention
		draft.Status = enums.DraftStatusPaused
		draft, err = s.draftRepo.UpdateDraft(draft)
		if err != nil {
			log.Printf("ERROR: (DraftService: AutoSkipTurn) - Could not update draft %d status to PAUSED: %v\n", draft.ID, err)
			return types.ErrInternalService
		}
		fmt.Printf("INFO: (DraftService: AutoSkipTurn) - Draft for league %s paused. Awaiting Manual Intervention\n", leagueID)
		return types.ErrDraftPausedForIntervention
	}

	member.SkipsLeft -= effectiveSkipsInThisAction
	log.Printf("DEBUG: (DraftService: AutoSkipTurn) - Member %s SkipsLeft BEFORE DB update: %d\n", member.ID, member.SkipsLeft)
	if _, err := s.memberRepo.Update(member); err != nil {
		log.Printf("CRITICAL ERROR: (DraftService: AutoSkipTurn) - Failed to update member %s skipsLeft in DB: %v\n", member.ID, err)
		return types.ErrInternalService
	}
	// Re-fetch member to confirm DB state
	updatedMember, err := s.memberRepo.GetByID(member.ID)
	if err != nil {
		log.Printf("CRITICAL ERROR: (DraftService: AutoSkipTurn) - Failed to re-fetch member %s after update: %v\n", member.ID, err)
		return types.ErrInternalService
	}
	log.Printf("DEBUG: (DraftService: AutoSkipTurn) - Member %s SkipsLeft AFTER DB re-fetch: %d\n", updatedMember.ID, updatedMember.SkipsLeft)
	log.Printf("DEBUG: (DraftService: AutoSkipTurn) - Member %s SkipsLeft updated to %d after auto-skip.\n", member.ID, member.SkipsLeft)

	allMembers, err := s.memberRepo.GetByLeague(leagueID)
	if err != nil {
		log.Printf("ERROR: (DraftService: AutoSkipTurn) - Could not get all members in league %s: %v\n", leagueID, err)
		return types.ErrInternalService
	}

	draft, err = s.advanceDraftState(draft, league, member, allMembers, len(allMembers), false)
	if err != nil {
		log.Printf("ERROR: (DraftService: AutoSkipTurn) - could not advance draft")
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
			PlayerID: *draft.CurrentTurnMemberID,
		},
	}

	s.schedulerService.RegisterTask(task)
	fmt.Printf("INFO: (DraftService: AutoSkipTurn) - Success\n")
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
		log.Printf("DEBUG: (DraftService: advanceDraftState) - Member %s skipping pick %d. Current accumulated picks: %v\n", member.ID, draft.CurrentPickOnClock, draft.PlayersWithAccumulatedPicks[member.ID])
		draft.PlayersWithAccumulatedPicks[member.ID] = append(draft.PlayersWithAccumulatedPicks[member.ID], draft.CurrentPickOnClock)
	}

	draft.CurrentPickOnClock++ // unconditonal increment

	// Check for draft completion
	isDraftCompleted, err := s.checkDraftCompletion(league, allMembers)
	if err != nil {
		log.Printf("LOG: (DraftService: advanceDraftState) - Error checking draft completion for league %s: %v\\n", league.ID,
			err)
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
			log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update draft status to COMPLETED for league %s:%v\n", league.ID, err)
			return nil, fmt.Errorf("failed to update draft state on completion: %w", err)
		}
		// save league status updated by checkDraftCompletion
		if _, err := s.leagueRepo.UpdateLeague(league); err != nil { // pray this never happens type shit
			// should prolly revert the draft update
			log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update league status to POST_DRAFT for league %s: %v\n", league.ID, err)
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
		log.Printf("LOG: (DraftService: advanceDraftState) - Current member %s not found in allMembers list. This should not happen. (Unreachable Control Flow)\\n", member.ID)
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
		log.Printf("LOG: (DraftService: advanceDraftState) - Failed to update draft: %v\n", err)
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
			if entry.ID == requestedPick.LeaguePokemonID {
				currentPoolEntry = entry
				break
			}
		}
		if currentPoolEntry == nil {
			return types.ErrLeaguePokemonNotFound
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
	log.Printf("DEBUG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Member %s. Current Pick On Clock: %d. Accumulated Picks: %v\n", memberID, draft.CurrentPickOnClock, accumulatedPickNumbers)
	validPickNumbersForMember := make([]int, len(accumulatedPickNumbers))
	copy(validPickNumbersForMember, accumulatedPickNumbers) // we don't wanna directly append
	validPickNumbersForMember = append(validPickNumbersForMember, draft.CurrentPickOnClock)

	// track used accumulated picks within this batch to prevent double-usage
	usedAccumulatedPicksInThisBatch := make(map[int]bool)
	currentPickSlotUsed := false

	for _, requestedPick := range input.RequestedPicks {
		// check if the requested pick number is a valid slot (current turn or accumulated)
		if !slices.Contains(validPickNumbersForMember, requestedPick.DraftPickNumber) {
			log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Member %s requested invalid pick number %d. Not on clock (%d) and not in accumulated picks (%v).\n",
				memberID, requestedPick.DraftPickNumber, draft.CurrentPickOnClock, accumulatedPickNumbers)
			return false, types.ErrInvalidInput
		}

		// if it's an accumulated pick, ensure it's not used twice in this batch
		if requestedPick.DraftPickNumber != draft.CurrentPickOnClock {
			if usedAccumulatedPicksInThisBatch[requestedPick.DraftPickNumber] {
				log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Member %s attempted to use accumulated pick %d multiple times in one request.\n",
					memberID, requestedPick.DraftPickNumber)
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
		log.Printf("LOG: (DraftService: validatePicksAndCheckCurrentPickSlotUsed) - Member %s cannot implicitly skip current turn's pick (%d) as it would violate minimum roster requirement. Skips left: %d.\n",
			memberID, draft.CurrentPickOnClock, member.SkipsLeft)
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
	log.Printf("DEBUG: (DraftService: fetchMemberResource) - Fetched member %s. SkipsLeft: %d\n", member.ID, member.SkipsLeft)
	return member, nil
}

// fetchRequestedPoolEntries retrieves a list of PoolEntry by their IDs, ensuring they are
// all available to be drafted. It returns service-specific errors for not found or
// already drafted pokemon.
func (s *draftServiceImpl) fetchRequestedPoolEntries(leagueID uuid.UUID, input *requests.DraftMakePickRequestDTO) ([]*models.PoolEntry, error) {
	var poolEntryIDs []uuid.UUID
	for _, requestedPick := range input.RequestedPicks {
		poolEntryIDs = append(poolEntryIDs, requestedPick.LeaguePokemonID)
	}

	allRequestedPoolEntries, err := s.poolEntryRepo.GetByIDs(leagueID, poolEntryIDs)
	if err != nil {
		return nil, types.ErrInternalService
	}

	// Validate that all requested pokemon were actually returned and are available.
	if len(allRequestedPoolEntries) != len(poolEntryIDs) {
		return nil, types.ErrLeaguePokemonNotFound
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
		log.Printf("LOG: (DraftService: checkDraftCompletion) - No members in league %s. Cannot check for draft completion.\\n", league.ID)
		return false, types.ErrInternalService
	}
	maxPicksPerMember := league.MaxPokemonPerPlayer
	totalExpectedPicks := totalMembers * maxPicksPerMember

	// 2. Get the current count of all active claims in the league
	currentTotalActiveClaims, err := s.claimRepo.GetActiveCountByLeague(league.ID)
	if err != nil {
		log.Printf("LOG: (DraftService: checkDraftCompletion) - Failed to get total active claims for league %s: %v\\n", league.ID, err)
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
			log.Printf("LOG: (DraftService: checkDraftCompletion) - Failed to get roster count for member %s in league %s: %v\\n", member.ID, league.ID, err)
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
