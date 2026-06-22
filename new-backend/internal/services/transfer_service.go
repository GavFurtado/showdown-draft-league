package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransferService interface {
	StartTransferPeriod(leagueID uuid.UUID) error
	EndTransferPeriod(leagueID uuid.UUID) error
	DropPokemon(currentUser *models.User, leagueID, draftedPokemonID uuid.UUID) error
	PickupFreeAgent(currentUser *models.User, leagueID, leaguePokemonID uuid.UUID) error
	SetSchedulerService(schedulerService SchedulerService)
}

type transferServiceImpl struct {
	draftedPokemonRepo repositories.DraftedPokemonRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	schedulerService   SchedulerService
}

func NewTransferService(
	draftedPokemonRepo repositories.DraftedPokemonRepository,
	leaguePokemonRepo repositories.LeaguePokemonRepository,
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
) TransferService {
	return &transferServiceImpl{
		draftedPokemonRepo: draftedPokemonRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
	}
}

func (s *transferServiceImpl) SetSchedulerService(schedulerService SchedulerService) {
	s.schedulerService = schedulerService
}

// StartTransferPeriod begins the transfer window for a league. It updates the league status,
// allocates transfer credits to players if enabled, and schedules the end of the window.
func (s *transferServiceImpl) StartTransferPeriod(leagueID uuid.UUID) error {
	// 1. Fetch the League
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to fetch league %s: %v\n", leagueID, err)
		return types.ErrLeagueNotFound
	}

	// 2. Validate Status
	if !league.Format.AllowTransfers {
		return fmt.Errorf("%w: Transfers are disabled for this league", types.ErrUnauthorized)
	}

	if league.Status != enums.LeagueStatusRegularSeason && league.Status != enums.LeagueStatusPostDraft {
		log.Printf("WARN: (TransferService: StartTransferPeriod) - League %s is not in a valid state to start a transfer window. Status: %s\n", leagueID, league.Status)
		return fmt.Errorf("invalid league status to start transfer window: %s", league.Status)
	}

	// 3. Update Player Credits (if applicable)
	didAllPlayersAccrueCredits := true
	if league.Format.TransfersCostCredits {
		players, err := s.playerRepo.GetPlayersByLeague(leagueID)
		if err != nil {
			log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to get players for league %s: %v\n", leagueID, err)
			return types.ErrInternalService
		}

		for _, player := range players {
			player.TransferCredits += league.Format.TransferCreditsPerWindow
			if player.TransferCredits > league.Format.TransferCreditCap {
				player.TransferCredits = league.Format.TransferCreditCap
			}
			if _, err := s.playerRepo.UpdatePlayer(&player); err != nil {
				// Log the error but continue trying to update other players
				log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to update transfer credits for player %s (%s): %v\n", player.InLeagueName, player.ID, err)
				didAllPlayersAccrueCredits = false
			}
		}
	}

	// 4. Update League Status
	league.Status = enums.LeagueStatusTransferWindow
	now := time.Now()
	league.Format.NextTransferWindowStart = &now // The window starts now

	// 5. Schedule EndTransferPeriod
	windowEndTime := now.Add(time.Duration(league.Format.TransferWindowDuration) * time.Hour)
	taskID := fmt.Sprintf("%d_%s", utils.TaskTypeTransferPeriodEnd, league.ID)
	endTask := &utils.ScheduledTask{
		ID:        taskID,
		ExecuteAt: windowEndTime,
		Type:      utils.TaskTypeTransferPeriodEnd,
		Payload: utils.PayloadTransferPeriodEnd{
			LeagueID: league.ID,
		},
	}
	s.schedulerService.RegisterTask(endTask)

	// 6. Save Changes
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to update league %s status: %v\n", leagueID, err)
		s.schedulerService.DeregisterTask(taskID)
		return types.ErrInternalService
	}

	if !league.Format.TransfersCostCredits {
		log.Printf("LOG: (TransferService: StartTransferPeriod) - Transfer window started for league %s.\n", leagueID)
	} else {
		if !didAllPlayersAccrueCredits {
			log.Printf("LOG: (TransferService: StartTransferPeriod) - Transfer window started for league %s but there was an error in credit accrual for one or more players.\n", leagueID)
		} else {
			log.Printf("LOG: (TransferService: StartTransferPeriod) - Transfer window started for league %s and all players accrued credits.\n", leagueID)
		}
	}
	return nil
}

// EndTransferPeriod concludes the transfer window for a league. It updates the league status
// and schedules the next transfer window to begin.
func (s *transferServiceImpl) EndTransferPeriod(leagueID uuid.UUID) error {
	// 1. Fetch the League
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (TransferService: EndTransferPeriod) - Failed to fetch league %s: %v\n", leagueID, err)
		return types.ErrLeagueNotFound
	}

	// 2. Validate Status
	if league.Status != enums.LeagueStatusTransferWindow {
		log.Printf("WARN: (TransferService: EndTransferPeriod) - League %s is not in a transfer window. Status: %s\n", leagueID, league.Status)
		return fmt.Errorf("invalid league status to end transfer window: %s", league.Status)
	}

	// 3. Update League Status
	league.Status = enums.LeagueStatusRegularSeason

	// 4. Schedule next StartTransferPeriod
	taskID := fmt.Sprintf("%d_%s", utils.TaskTypeTransferPeriodStart, league.ID)
	if league.Format.TransferWindowFrequencyDays > 0 {
		nextWindowStartTime := time.Now().AddDate(0, 0, league.Format.TransferWindowFrequencyDays)
		league.Format.NextTransferWindowStart = &nextWindowStartTime

		startTask := &utils.ScheduledTask{
			ID:        taskID,
			Type:      utils.TaskTypeTransferPeriodStart,
			ExecuteAt: nextWindowStartTime,
			Payload: utils.PayloadTransferPeriodStart{
				LeagueID: league.ID,
			},
		}
		s.schedulerService.RegisterTask(startTask)
	} else {
		// If frequency is 0 or less, don't schedule a next window.
		league.Format.NextTransferWindowStart = nil
		league.Format.AllowTransfers = false
		league.Format.TransfersCostCredits = false
	}

	// 5. Save Changes
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (TransferService: EndTransferPeriod) - Failed to update league %s: %v\n", leagueID, err)
		s.schedulerService.DeregisterTask(taskID)
		return types.ErrInternalService
	}

	log.Printf("LOG: (TransferService: EndTransferPeriod) - Transfer window ended for league %s.\n", leagueID)
	return nil
}

// DropPokemon allows a user to drop a drafted Pokemon, making it a free agent.
// This operation is only allowed during a transfer window and if the user owns the Pokemon.
func (s *transferServiceImpl) DropPokemon(currentUser *models.User, leagueID, draftedPokemonID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrLeagueNotFound
		}
		return types.ErrInternalService
	}

	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrDraftedPokemonNotFound
		}
		return types.ErrInternalService
	}
	if draftedPokemon.LeagueID != leagueID {
		return types.ErrForbidden
	}

	// Authorize user is owner of pokemon
	player, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
		return types.ErrPlayerNotFound
	}
	if player.UserID != currentUser.ID {
		return types.ErrUnauthorized
	}

	inWindow, err := s.isLeagueInTransferWindow(draftedPokemon.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return types.ErrInvalidState
	}

	if player.TransferCredits < league.Format.DropCost {
		return types.ErrInsufficientTransferCredits
	}

	if draftedPokemon.IsReleased {
		return types.ErrPokemonAlreadyReleased
	}

	// Check if dropping this pokemon would put the player below the minimum
	currentPokemonCount, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(player.ID)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.DropPokemon) - could not get pokemon count for player %s: %v", player.ID, err)
		return types.ErrInternalService
	}
	if currentPokemonCount <= int64(league.MinPokemonPerPlayer) {
		return types.ErrBelowMinPokemon
	}

	err = s.draftedPokemonRepo.ReleasePokemonTransaction(draftedPokemon, player, league.Format.DropCost, league.CurrentWeekNumber)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.DropPokemon) - Failed to release pokemon with ID %s: %v", draftedPokemonID, err)
		return types.ErrInternalService
	}
	return nil
}

// PickupFreeAgent allows a user to pick up a released Pokemon (free agent) using transfer credits.
// This operation is only allowed during a transfer window and if the player has enough credits.
func (s *transferServiceImpl) PickupFreeAgent(currentUser *models.User, leagueID, leaguePokemonID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrLeagueNotFound
		}
		return types.ErrInternalService
	}

	leaguePokemon, err := s.leaguePokemonRepo.GetLeaguePokemonByID(leaguePokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrLeaguePokemonNotFound
		}
		return types.ErrInternalService
	}

	if leaguePokemon.LeagueID != leagueID {
		return types.ErrForbidden
	}

	inWindow, err := s.isLeagueInTransferWindow(leaguePokemon.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return types.ErrInvalidState // Or a more specific "not in transfer window" error
	}

	player, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, leaguePokemon.LeagueID)
	if err != nil {
		return types.ErrPlayerNotFound
	}

	if player.TransferCredits < league.Format.PickupCost {
		return types.ErrInsufficientTransferCredits
	}

	if !leaguePokemon.IsAvailable {
		return types.ErrConflict // Pokemon not available
	}

	// Check if picking up this pokemon would put the player above the maximum
	currentPokemonCount, err := s.draftedPokemonRepo.GetDraftedPokemonCountByPlayer(player.ID)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.PickupFreeAgent) - could not get pokemon count for player %s: %v", player.ID, err)
		return types.ErrInternalService
	}
	if currentPokemonCount >= int64(league.MaxPokemonPerPlayer) {
		return types.ErrAboveMaxPokemon
	}

	newDraftedPokemon := &models.DraftedPokemon{
		LeagueID:         leaguePokemon.LeagueID,
		PlayerID:         player.ID,
		PokemonSpeciesID: leaguePokemon.PokemonSpeciesID,
		LeaguePokemonID:  leaguePokemon.ID,
		IsReleased:       false,
		AcquiredWeek:     league.CurrentWeekNumber,
	}

	// Call the new repository transaction method
	if err := s.draftedPokemonRepo.PickupFreeAgentTransaction(player, newDraftedPokemon, leaguePokemon, league.Format.PickupCost); err != nil {
		log.Printf("LOG: (Error: TransferService.PickupFreeAgent) - Failed to complete pickup free agent transaction: %v", err)
		return types.ErrInternalService
	}

	return nil
}

func (s *transferServiceImpl) isLeagueInTransferWindow(leagueID uuid.UUID) (bool, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		return false, types.ErrLeagueNotFound
	}
	return league.Status == enums.LeagueStatusTransferWindow, nil
}
