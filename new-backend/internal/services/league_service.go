package services

import (
	"fmt"
	"log"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
)

// defines the interface for league-related business logic.
type LeagueService interface {
	// handles the business logic for creating a new league.
	CreateLeague(userID uuid.UUID, req *common.LeagueCreateRequestDTO) (*models.League, error)
	// Get league entity using leagueID
	GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error)
	// gets all Leagues where userID is the commissioner
	GetLeaguesByCommissioner(userID uuid.UUID, currentUser *models.User) ([]models.League, error)
	// fetches all Leagues where the given userID is a player.
	GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error)
	ProcessWeeklyTick(leagueID uuid.UUID) error
	SetSchedulerService(schedulerService SchedulerService)
	SetGameService(gameService GameService)
	SetTransferService(transferService TransferService)
	StartRegularSeason(leagueID uuid.UUID) error
}

type leagueServiceImpl struct {
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	draftedPokemonRepo repositories.DraftedPokemonRepository
	draftRepo          repositories.DraftRepository
	gameRepo           repositories.GameRepository
	schedulerService   SchedulerService
	transferService    TransferService
	gameService        GameService // New dependency
}

func NewLeagueService(
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
	leaguePokemonRepo repositories.LeaguePokemonRepository,
	draftedPokemonRepo repositories.DraftedPokemonRepository,
	draftRepo repositories.DraftRepository,
	gameRepo repositories.GameRepository,
) LeagueService {
	return &leagueServiceImpl{
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
		draftedPokemonRepo: draftedPokemonRepo,
		draftRepo:          draftRepo,
		gameRepo:           gameRepo,
	}
}

func (s *leagueServiceImpl) SetSchedulerService(schedulerService SchedulerService) {
	s.schedulerService = schedulerService
}

func (s *leagueServiceImpl) SetGameService(gameService GameService) {
	s.gameService = gameService
}

func (s *leagueServiceImpl) SetTransferService(transferService TransferService) {
	s.transferService = transferService
}

// handles the business logic for creating a new league.
func (s *leagueServiceImpl) CreateLeague(userID uuid.UUID, input *common.LeagueCreateRequestDTO) (*models.League, error) {
	const maxLeaguesCommisionable = 2
	const maxGroupsAllowed = 2

	// check if user already has two owned leagues
	count, err := s.leagueRepo.GetLeaguesCountWhereOwner(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Could not get commissioner league count for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to check commissioner league count: %w", err)
	}

	if count >= maxLeaguesCommisionable {
		return nil, common.ErrMaxLeagueCreationLimitReached
	}

	if input.Format.GroupCount > maxGroupsAllowed {
		return nil, common.ErrExceedsMaxAllowableGroupCount
	}

	if input.Format.AllowTransfers && input.Format.TransferWindowFrequencyDays%7 != 0 {
		return nil, fmt.Errorf("%w: TransferWindowFrequencyDays must be a multiple of 7", common.ErrInvalidLeagueConfiguration)
	}

	newPlayerGroupNumber := 1
	if input.Format.GroupCount > 1 {
		// owner is first player and auto assigned 1. So next player will have to be group 2
		// will need to be changed if we decide to make use of Player.IsParticapating
		newPlayerGroupNumber = 2
	}

	league := &models.League{
		Name:                 input.Name,
		RulesetDescription:   input.RulesetDescription,
		MaxPokemonPerPlayer:  input.MaxPokemonPerPlayer,
		StartingDraftPoints:  input.StartingDraftPoints,
		StartDate:            input.StartDate,
		NewPlayerGroupNumber: newPlayerGroupNumber,
		Format: &models.LeagueFormat{
			SeasonType:                  input.Format.SeasonType,
			GroupCount:                  input.Format.GroupCount,
			PlayoffType:                 input.Format.PlayoffType,
			PlayoffParticipantCount:     input.Format.PlayoffParticipantCount,
			PlayoffByesCount:            input.Format.PlayoffByesCount,
			PlayoffSeedingType:          input.Format.PlayoffSeedingType,
			IsSnakeRoundDraft:           input.Format.IsSnakeRoundDraft,
			AllowTransfers:              input.Format.AllowTransfers,
			TransfersCostCredits:        input.Format.TransfersCostCredits,
			TransferCreditsPerWindow:    input.Format.TransferCreditsPerWindow,
			TransferCreditCap:           input.Format.TransferCreditCap,
			TransferWindowDuration:      input.Format.TransferWindowDuration,
			TransferWindowFrequencyDays: input.Format.TransferWindowFrequencyDays,
			DropCost:                    input.Format.DropCost,
			PickupCost:                  input.Format.PickupCost,
		},
	}

	createdLeague, err := s.leagueRepo.CreateLeague(league)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Failed to create league for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	// TODO: this should maybe not be done
	ownerPlayer := &models.Player{
		UserID:          userID,
		LeagueID:        createdLeague.ID,
		InLeagueName:    "League Owner",                       // Default, can be updated later
		TeamName:        fmt.Sprintf("%s's Team", input.Name), // Default, can be updated later
		IsParticipating: false,
		DraftPoints:     int(createdLeague.StartingDraftPoints),
		TransferCredits: 0,
		GroupNumber:     1, // first player for the league so assigned this
		Role:            rbac.PRoleOwner,
	}

	_, err = s.playerRepo.CreatePlayer(ownerPlayer)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Failed to create owner player for league %s: %v\n", createdLeague.ID, err)
		// TODO: Consider rolling back league creation if player creation fails
		return nil, fmt.Errorf("failed to create league owner player: %w", err)
	}

	return createdLeague, nil
}

// Get league entity using leagueID
func (s *leagueServiceImpl) GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error) {
	// User in league checks done at middleware

	// Retrieve the league
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeagueByIDForUser) - Could not get league %s for user %d: %v\n", leagueID, userID, err)
		return nil, fmt.Errorf("failed to retrieve league: %w", err)
	}

	return league, nil
}

// gets all Leagues where userID is the owner
func (s *leagueServiceImpl) GetLeaguesByCommissioner(
	userID uuid.UUID,
	currentUser *models.User,
) ([]models.League, error) {

	leagues, err := s.leagueRepo.GetLeaguesByOwner(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeaguesByCommissioner) - Failed to get commissioner leagues for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to retrieve commissioner leagues: %w", err)
	}

	return leagues, nil
}

// fetches all Leagues where the given userID is a player.
func (s *leagueServiceImpl) GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error) {

	leagues, err := s.leagueRepo.GetLeaguesByUser(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeaguesByUser) - Failed to get leagues for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to retrieve leagues: %w", err)
	}
	return leagues, nil
}

// StartRegularSeason orchestrates the beginning of a league's regular season.
// It generates all regular season games, updates the league status to REGULAR_SEASON,
// sets the initial current week number, schedules the very first weekly tick,
// and potentially triggers the first transfer window.
func (s *leagueServiceImpl) StartRegularSeason(leagueID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (LeagueService: StartRegularSeason) - Failed to fetch league %s: %v\n", leagueID, err)
		return common.ErrLeagueNotFound
	}

	// 1. Validate League Status
	if league.Status != enums.LeagueStatusPostDraft {
		log.Printf("ERROR: (LeagueService: StartRegularSeason) - League %s is not in POST_DRAFT status, cannot start regular season. Current status: %s\n", leagueID, league.Status)
		return common.ErrInvalidState
	}

	// 2. Generate Regular Season Games
	if err := s.gameService.GenerateRegularSeasonGames(leagueID); err != nil {
		log.Printf("ERROR: (LeagueService: StartRegularSeason) - Failed to generate regular season games for league %s: %v\n", leagueID, err)
		return fmt.Errorf("failed to generate regular season games: %w", err)
	}

	// 3. Update League Status, CurrentWeekNumber, and RegularSeasonStartDate
	now := time.Now()
	league.Status = enums.LeagueStatusRegularSeason
	league.CurrentWeekNumber = 1 // Season starts at Week 1
	league.RegularSeasonStartDate = &now
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (LeagueService: StartRegularSeason) - Failed to update league %s status, current week number, and regular season start date: %v\n", leagueID, err)
		return fmt.Errorf("failed to update league status: %w", err)
	}
	log.Printf("LOG: (LeagueService: StartRegularSeason) - League %s status updated to REGULAR_SEASON, CurrentWeekNumber set to %d, RegularSeasonStartDate set to %s.\n", leagueID, league.CurrentWeekNumber, league.RegularSeasonStartDate.String())

	// 4. Schedule the very first LeagueWeeklyTick
	// The first tick should occur 7 days from now to advance to Week 2.
	firstTickTime := now.Add(7 * 24 * time.Hour)
	league.NextWeeklyTick = &firstTickTime
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (LeagueService: StartRegularSeason) - Failed to update league %s with next weekly tick time: %v\n", leagueID, err)
		return fmt.Errorf("failed to update league with next weekly tick time: %w", err)
	}

	firstTickTask := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", utils.TaskTypeLeagueWeeklyTick, league.ID),
		ExecuteAt: firstTickTime,
		Type:      utils.TaskTypeLeagueWeeklyTick,
		Payload: utils.PayloadLeagueWeeklyTick{
			LeagueID: league.ID,
		},
	}
	s.schedulerService.RegisterTask(firstTickTask)
	log.Printf("LOG: (LeagueService: StartRegularSeason) - First weekly tick for league %s scheduled for %s.\n", leagueID, firstTickTime.String())

	// 5. Trigger first transfer window if applicable
	// This will call transferService.StartTransferPeriod, which will then schedule its own EndTransferPeriod.
	if league.Format.AllowTransfers && league.Format.TransferWindowFrequencyDays > 0 {
		weeksBetweenWindows := league.Format.TransferWindowFrequencyDays / 7
		if weeksBetweenWindows > 0 && (league.CurrentWeekNumber-1)%weeksBetweenWindows == 0 {
			log.Printf("LOG: (LeagueService: StartRegularSeason) - Triggering initial transfer window for league %s.\n", leagueID)
			if err := s.transferService.StartTransferPeriod(leagueID); err != nil {
				log.Printf("ERROR: (LeagueService: StartRegularSeason) - Failed to trigger initial transfer period for league %s: %v\n", leagueID, err)
				// Log but don't fail the whole season start, transfer window issues can be manually resolved.
			}
		}
	}

	return nil
}

// ProcessWeeklyTick handles the automatic progression of a league's week.
// It recalculates the current week number based on `RegularSeasonStartDate`
// to ensure consistency even after server restarts or missed ticks.
func (s *leagueServiceImpl) ProcessWeeklyTick(leagueID uuid.UUID) error {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("ERROR: (LeagueService: ProcessWeeklyTick) - Failed to fetch league %s: %v\n", leagueID, err)
		return common.ErrLeagueNotFound
	}

	if league.Status != enums.LeagueStatusRegularSeason {
		log.Printf("INFO: (LeagueService: ProcessWeeklyTick) - League %s is not in REGULAR_SEASON. Skipping weekly tick. Status: %s\n", leagueID, league.Status)
		return nil // Not an error, just not applicable
	}

	if league.RegularSeasonStartDate == nil {
		log.Printf("ERROR: (LeagueService: ProcessWeeklyTick) - League %s has no RegularSeasonStartDate. Cannot process tick.\n", leagueID)
		return fmt.Errorf("league %s is missing RegularSeasonStartDate", leagueID)
	}

	oldWeekNumber := league.CurrentWeekNumber
	now := time.Now()

	// Calculate the correct current week based on RegularSeasonStartDate
	durationSinceSeasonStart := now.Sub(*league.RegularSeasonStartDate)
	calculatedCurrentWeek := int(durationSinceSeasonStart.Hours()/(24*7)) + 1

	if calculatedCurrentWeek > oldWeekNumber {
		// Weeks were missed or it's a natural advancement. Update the CurrentWeekNumber.
		log.Printf("INFO: (LeagueService: ProcessWeeklyTick) - Advancing week for league %s from %d to %d.\n", leagueID, oldWeekNumber, calculatedCurrentWeek)
		league.CurrentWeekNumber = calculatedCurrentWeek
	} else {
		// System is already up-to-date. Log this and fall through to ensure next tick is scheduled correctly.
		log.Printf("INFO: (LeagueService: ProcessWeeklyTick) - League %s: Already at or beyond calculated week %d (current: %d). Re-scheduling next tick.\n", leagueID, calculatedCurrentWeek, oldWeekNumber)
	}

	// Calculate and schedule the next LeagueWeeklyTick based on the CURRENT correct week
	nextTickTime := league.RegularSeasonStartDate.Add(time.Duration(league.CurrentWeekNumber) * 7 * 24 * time.Hour)
	league.NextWeeklyTick = &nextTickTime

	// Save the updated league state BEFORE scheduling the task
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("ERROR: (LeagueService: ProcessWeeklyTick) - Failed to save updated league %s after weekly tick processing: %v\n", leagueID, err)
		return fmt.Errorf("failed to save league after weekly tick: %w", err)
	}

	// Register the next tick task with the scheduler
	nextTickTask := &utils.ScheduledTask{
		ID:        fmt.Sprintf("%d_%s", utils.TaskTypeLeagueWeeklyTick, league.ID),
		ExecuteAt: nextTickTime,
		Type:      utils.TaskTypeLeagueWeeklyTick,
		Payload:   utils.PayloadLeagueWeeklyTick{LeagueID: league.ID},
	}
	s.schedulerService.RegisterTask(nextTickTask)
	log.Printf("LOG: (LeagueService: ProcessWeeklyTick) - Next weekly tick for league %s scheduled for %s (Start of Week %d).\n", leagueID, nextTickTime.String(), league.CurrentWeekNumber+1)

	// Check for transfer windows ONLY on a natural single-week advancement.
	if calculatedCurrentWeek == oldWeekNumber+1 {
		if league.Format.AllowTransfers {
			weeksBetweenWindows := league.Format.TransferWindowFrequencyDays / 7
			if weeksBetweenWindows > 0 && (league.CurrentWeekNumber-1)%weeksBetweenWindows == 0 {
				log.Printf("LOG: (LeagueService: ProcessWeeklyTick) - Natural week advancement. Triggering transfer window for league %s for Week %d.\n", leagueID, league.CurrentWeekNumber)
				if err := s.transferService.StartTransferPeriod(leagueID); err != nil {
					log.Printf("ERROR: (LeagueService: ProcessWeeklyTick) - Failed to trigger transfer period for league %s: %v\n", leagueID, err)
				}
			}
		}
	} else if calculatedCurrentWeek > oldWeekNumber {
		// A multi-week jump occurred. Log it, but do not trigger any transfer windows.
		log.Printf("WARN: (LeagueService: ProcessWeeklyTick) - League %s jumped from week %d to %d. Transfer window checks are bypassed for this tick.\n", leagueID, oldWeekNumber, calculatedCurrentWeek)
	}

	return nil
}
