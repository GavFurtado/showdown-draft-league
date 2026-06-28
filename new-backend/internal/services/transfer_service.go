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
	DropPokemon(currentUser *models.User, leagueID, claimID uuid.UUID) error
	PickupFreeAgent(currentUser *models.User, leagueID, poolEntryID uuid.UUID) error
	SetSchedulerService(schedulerService SchedulerService)
	SetNewRepositories(claimRepo repositories.ClaimRepository, poolEntryRepo repositories.PoolEntryRepository)
}

type transferServiceImpl struct {
	draftedPokemonRepo repositories.DraftedPokemonRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	schedulerService   SchedulerService

	// New redesign repositories
	claimRepo     repositories.ClaimRepository
	poolEntryRepo repositories.PoolEntryRepository
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

// SetNewRepositories injects the new redesign repositories.
func (s *transferServiceImpl) SetNewRepositories(claimRepo repositories.ClaimRepository, poolEntryRepo repositories.PoolEntryRepository) {
	s.claimRepo = claimRepo
	s.poolEntryRepo = poolEntryRepo
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

// DropPokemon allows a user to drop a claimed Pokemon, making it a free agent.
// This operation is only allowed during a transfer window and if the user owns the Pokemon.
// Uses the new Claim model for ownership tracking.
func (s *transferServiceImpl) DropPokemon(currentUser *models.User, leagueID, claimID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrLeagueNotFound
		}
		return types.ErrInternalService
	}

	claim, err := s.claimRepo.GetClaimByID(claimID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrDraftedPokemonNotFound
		}
		return types.ErrInternalService
	}
	if claim.LeagueID != leagueID {
		return types.ErrForbidden
	}

	// Authorize user is owner of claimed pokemon
	player, err := s.playerRepo.GetPlayerByID(claim.PlayerID)
	if err != nil {
		return types.ErrPlayerNotFound
	}
	if player.UserID != currentUser.ID {
		return types.ErrUnauthorized
	}

	inWindow, err := s.isLeagueInTransferWindow(claim.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return types.ErrInvalidState
	}

	if player.TransferCredits < league.Format.DropCost {
		return types.ErrInsufficientTransferCredits
	}

	if !claim.IsActive {
		return types.ErrPokemonAlreadyReleased
	}

	// Check if dropping this pokemon would put the player below the minimum
	currentPokemonCount, err := s.claimRepo.GetActiveClaimCountByPlayer(player.ID)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.DropPokemon) - could not get claim count for player %s: %v", player.ID, err)
		return types.ErrInternalService
	}
	if currentPokemonCount <= int64(league.MinPokemonPerPlayer) {
		return types.ErrBelowMinPokemon
	}

	// Find the pool entry to mark it as available again
	poolEntry, err := s.poolEntryRepo.GetPoolEntryByLeagueAndSpecies(claim.LeagueID, claim.SpeciesID)
	if err != nil {
		log.Printf("WARN: (TransferService.DropPokemon) - Could not find pool entry for species %d (dropping anyway): %v", claim.SpeciesID, err)
		// Continue even if we can't find the specific pool entry - the claim is still released
	}

	var poolEntryID uuid.UUID
	if poolEntry != nil {
		poolEntryID = poolEntry.ID
	}

	err = s.claimRepo.ReleaseClaimTransaction(claim, player, league.Format.DropCost, league.CurrentWeekNumber, poolEntryID)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.DropPokemon) - Failed to release claim with ID %s: %v", claimID, err)
		return types.ErrInternalService
	}
	return nil
}

// PickupFreeAgent allows a user to pick up a released Pokemon (free agent) using transfer credits.
// This operation is only allowed during a transfer window and if the player has enough credits.
// Uses the new Claim + PoolEntry models.
func (s *transferServiceImpl) PickupFreeAgent(currentUser *models.User, leagueID, poolEntryID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrLeagueNotFound
		}
		return types.ErrInternalService
	}

	poolEntry, err := s.poolEntryRepo.GetPoolEntryByID(poolEntryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrLeaguePokemonNotFound
		}
		return types.ErrInternalService
	}

	if poolEntry.LeagueID != leagueID {
		return types.ErrForbidden
	}

	inWindow, err := s.isLeagueInTransferWindow(poolEntry.LeagueID)
	if err != nil {
		return err
	}
	if !inWindow {
		return types.ErrInvalidState
	}

	player, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, poolEntry.LeagueID)
	if err != nil {
		return types.ErrPlayerNotFound
	}

	if player.TransferCredits < league.Format.PickupCost {
		return types.ErrInsufficientTransferCredits
	}

	if !poolEntry.IsAvailable {
		return types.ErrConflict // Pokemon not available
	}

	// Check if picking up this pokemon would put the player above the maximum
	currentPokemonCount, err := s.claimRepo.GetActiveClaimCountByPlayer(player.ID)
	if err != nil {
		log.Printf("LOG: (Error: TransferService.PickupFreeAgent) - could not get claim count for player %s: %v", player.ID, err)
		return types.ErrInternalService
	}
	if currentPokemonCount >= int64(league.MaxPokemonPerPlayer) {
		return types.ErrAboveMaxPokemon
	}

	newClaim := &models.Claim{
		LeagueID:     poolEntry.LeagueID,
		PlayerID:     player.ID,
		SpeciesID:    poolEntry.PokemonSpeciesID,
		Source:       enums.ClaimSourceFreeAgent,
		CostPaid:     *poolEntry.Cost,
		AcquiredWeek: league.CurrentWeekNumber,
		IsActive:     true,
	}

	if err := s.claimRepo.PickupFreeAgentTransaction(player, newClaim, poolEntry, league.Format.PickupCost); err != nil {
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
