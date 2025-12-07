package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
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
		return common.ErrLeagueNotFound
	}

	// 2. Validate Status
	if !league.Format.AllowTransfers {
		return fmt.Errorf("%w: Transfers are disabled for this league", common.ErrUnauthorized)
	}

	if league.Status != enums.LeagueStatusRegularSeason && league.Status != enums.LeagueStatusPostDraft {
		log.Printf("WARN: (TransferService: StartTransferPeriod) - League %s is not in a valid state to start a transfer window. Status: %s\n", leagueID, league.Status)
		return fmt.Errorf("invalid league status to start transfer window: %s", league.Status)
	}

	// 3. Update Player Credits (if applicable)
	if league.Format.TransfersCostCredits {
		players, err := s.playerRepo.GetPlayersByLeague(leagueID)
		if err != nil {
			log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to get players for league %s: %v\n", leagueID, err)
			return common.ErrInternalService
		}

		for _, player := range players {
			player.TransferCredits += league.Format.TransferCreditsPerWindow
			if player.TransferCredits > league.Format.TransferCreditCap {
				player.TransferCredits = league.Format.TransferCreditCap
			}
			if _, err := s.playerRepo.UpdatePlayer(&player); err != nil {
				// Log the error but continue trying to update other players
				log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to update transfer credits for player %s: %v\n", player.ID, err)
			}
		}
	}

	// 4. Update League Status
	league.Status = enums.LeagueStatusTransferWindow
	now := time.Now()
	league.Format.NextTransferWindowStart = &now // The window starts now

	// 5. Schedule EndTransferPeriod
	windowEndTime := now.Add(time.Duration(league.Format.TransferWindowDuration) * time.Minute)
	taskID := fmt.Sprintf("%d_%s", utils.TaskTypeTradingPeriodEnd, league.ID)
	endTask := &utils.ScheduledTask{
		ID:        taskID,
		ExecuteAt: windowEndTime,
		Type:      utils.TaskTypeTradingPeriodEnd,
		Payload: utils.PayloadTransferPeriodEnd{
			LeagueID: league.ID,
		},
	}
	s.schedulerService.RegisterTask(endTask)

	// 6. Save Changes
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (TransferService: StartTransferPeriod) - Failed to update league %s status: %v\n", leagueID, err)
		s.schedulerService.DeregisterTask(taskID)
		return common.ErrInternalService
	}

	log.Printf("LOG: (TransferService: StartTransferPeriod) - Transfer window started for league %s.\n", leagueID)
	return nil
}

// EndTransferPeriod concludes the transfer window for a league. It updates the league status
// and schedules the next transfer window to begin.
func (s *transferServiceImpl) EndTransferPeriod(leagueID uuid.UUID) error {
	// 1. Fetch the League
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (TransferService: EndTransferPeriod) - Failed to fetch league %s: %v\n", leagueID, err)
		return common.ErrLeagueNotFound
	}

	// 2. Validate Status
	if league.Status != enums.LeagueStatusTransferWindow {
		log.Printf("WARN: (TransferService: EndTransferPeriod) - League %s is not in a transfer window. Status: %s\n", leagueID, league.Status)
		return fmt.Errorf("invalid league status to end transfer window: %s", league.Status)
	}

	// 3. Update League Status
	league.Status = enums.LeagueStatusRegularSeason

	// 4. Schedule next StartTransferPeriod
	taskID := fmt.Sprintf("%d_%s", utils.TaskTypeTradingPeriodStart, league.ID)
	if league.Format.TransferWindowFrequencyDays > 0 {
		nextWindowStartTime := time.Now().AddDate(0, 0, league.Format.TransferWindowFrequencyDays)
		league.Format.NextTransferWindowStart = &nextWindowStartTime

		startTask := &utils.ScheduledTask{
			ID:        taskID,
			Type:      utils.TaskTypeTradingPeriodStart,
			ExecuteAt: nextWindowStartTime,
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
		log.Printf("ERROR: (TransferService: EndTransferPeriod) - Failed to update league %s: %v\n", leagueID, err)
		s.schedulerService.DeregisterTask(taskID)
		return common.ErrInternalService
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
			return common.ErrLeagueNotFound
		}
		return common.ErrInternalService
	}

	draftedPokemon, err := s.draftedPokemonRepo.GetDraftedPokemonByID(draftedPokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrDraftedPokemonNotFound
		}
		return common.ErrInternalService
	}
	if draftedPokemon.LeagueID != leagueID {
		return common.ErrForbidden
	}

	// Authorize user is owner of pokemon
	player, err := s.playerRepo.GetPlayerByID(draftedPokemon.PlayerID)
	if err != nil {
		return common.ErrPlayerNotFound
	}
	if player.UserID != currentUser.ID {
		return common.ErrUnauthorized
	}

	inWindow, err := s.isLeagueInTransferWindow(draftedPokemon.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return common.ErrInvalidState
	}

	if player.TransferCredits < league.Format.DropCost {
		return common.ErrInsufficientTransferCredits
	}

	if draftedPokemon.IsReleased {
		return common.ErrPokemonAlreadyReleased
	}

	err = s.draftedPokemonRepo.ReleasePokemonTransaction(draftedPokemonID, player, league.Format.DropCost)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.DropPokemon) - Failed to release pokemon with ID %s: %v", draftedPokemonID, err)
		return common.ErrInternalService
	}
	return nil
}

// PickupFreeAgent allows a user to pick up a released Pokemon (free agent) using transfer credits.
// This operation is only allowed during a transfer window and if the player has enough credits.
func (s *transferServiceImpl) PickupFreeAgent(currentUser *models.User, leagueID, leaguePokemonID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrLeagueNotFound
		}
		return common.ErrInternalService
	}

	leaguePokemon, err := s.leaguePokemonRepo.GetLeaguePokemonByID(leaguePokemonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrLeaguePokemonNotFound
		}
		return common.ErrInternalService
	}

	if leaguePokemon.LeagueID != leagueID {
		return common.ErrForbidden
	}

	inWindow, err := s.isLeagueInTransferWindow(leaguePokemon.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return common.ErrInvalidState // Or a more specific "not in transfer window" error
	}

	player, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, leaguePokemon.LeagueID)
	if err != nil {
		return common.ErrPlayerNotFound
	}

	if player.TransferCredits < league.Format.PickupCost {
		return common.ErrInsufficientTransferCredits
	}

	if !leaguePokemon.IsAvailable {
		return common.ErrConflict // Pokemon not available
	}

	newDraftedPokemon := &models.DraftedPokemon{
		LeagueID:         leaguePokemon.LeagueID,
		PlayerID:         player.ID,
		PokemonSpeciesID: leaguePokemon.PokemonSpeciesID,
		LeaguePokemonID:  leaguePokemon.ID,
		IsReleased:       false,
	}

	// Call the new repository transaction method
	if err := s.draftedPokemonRepo.PickupFreeAgentTransaction(player, newDraftedPokemon, leaguePokemon, league.Format.PickupCost); err != nil {
		log.Printf("LOG: (Error: TransferService.PickupFreeAgent) - Failed to complete pickup free agent transaction: %v", err)
		return common.ErrInternalService
	}

	return nil
}

func (s *transferServiceImpl) isLeagueInTransferWindow(leagueID uuid.UUID) (bool, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		return false, common.ErrLeagueNotFound
	}
	return league.Status == enums.LeagueStatusTransferWindow, nil
}
