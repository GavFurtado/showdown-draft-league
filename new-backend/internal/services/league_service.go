package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
)

// LeagueService defines the interface for league-related business logic.
type LeagueService interface {
	// handles the business logic for creating a new league.
	CreateLeague(userID uuid.UUID, req *requests.LeagueCreateRequestDTO) (*models.League, error)
	// Get league entity using leagueID
	GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error)
	// gets all Leagues where userID is the commissioner
	GetLeaguesByCommissioner(userID uuid.UUID, currentUser *models.User) ([]models.League, error)
	// fetches all Leagues where the given userID is a player.
	GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error)

	StartRegularSeason(leagueID uuid.UUID) error

	ProcessWeeklyTick(leagueID uuid.UUID) error

	SetSchedulerService(schedulerService SchedulerService)
	SetGameService(gameService GameService)
	SetTransferService(transferService TransferService)
}

type leagueServiceImpl struct {
	logger           *slog.Logger
	leagueRepo       repositories.LeagueRepository
	memberRepo       repositories.LeagueMemberRepository
	draftRepo        repositories.DraftRepository
	gameRepo         repositories.GameRepository
	schedulerService SchedulerService
	transferService  TransferService
	gameService      GameService
}

func NewLeagueService(
	logger *slog.Logger,
	leagueRepo repositories.LeagueRepository,
	memberRepo repositories.LeagueMemberRepository,
	draftRepo repositories.DraftRepository,
	gameRepo repositories.GameRepository,
) LeagueService {
	return &leagueServiceImpl{
		logger:     utils.LoggerWithService(logger, "LeagueService"),
		leagueRepo: leagueRepo,
		memberRepo: memberRepo,
		draftRepo:  draftRepo,
		gameRepo:   gameRepo,
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

// CreateLeague handles the business logic for creating a new league.
func (s *leagueServiceImpl) CreateLeague(userID uuid.UUID, input *requests.LeagueCreateRequestDTO) (*models.League, error) {
	const maxLeaguesCommisionable = 20
	const maxGroupsAllowed = 2

	// check if user already has two owned leagues
	count, err := s.leagueRepo.GetLeaguesCountWhereOwner(userID)
	if err != nil {
		s.logger.Error("CreateLeague - could not get commissioner league count", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to check commissioner league count: %w", err)
	}

	if count >= maxLeaguesCommisionable {
		return nil, types.ErrMaxLeagueCreationLimitReached
	}

	if input.Format.GroupCount > maxGroupsAllowed {
		return nil, types.ErrExceedsMaxAllowableGroupCount
	}

	if input.Format.GroupCount < 1 {
		return nil, fmt.Errorf("%w: GroupCount must be at least 1", types.ErrInvalidLeagueConfiguration)
	}

	if input.Format.AllowTransfers && input.Format.TransferWindowFrequencyDays%7 != 0 {
		return nil, fmt.Errorf("%w: TransferWindowFrequencyDays must be a multiple of 7", types.ErrInvalidLeagueConfiguration)
	}

	newPlayerGroupNumber := 1
	if input.Format.GroupCount > 1 {
		// Owner is the first player and auto assigned 1. So, next player will have to be group 2
		// will need to be changed if we decide to make use of Player.IsParticapating
		newPlayerGroupNumber = 2
	}

	// Create the league's ID here so we can defer league creation to after the ownerPlayer is created
	// whilst still having the league ID available to us, to set for the player
	leagueID := uuid.New()

	league := &models.League{
		ID:                   leagueID,
		Name:                 input.Name,
		OwnerUserID:          userID,
		RulesetDescription:   input.RulesetDescription,
		MaxPokemonPerPlayer:  input.MaxPokemonPerPlayer,
		MinPokemonPerPlayer:  input.MinPokemonPerPlayer,
		StartingDraftPoints:  input.StartingDraftPoints,
		NewPlayerGroupNumber: newPlayerGroupNumber,
		Format:               input.Format.ToLeagueFormatPtr(),
		Visibility:           input.Visibility,
		MaxPlayers:           input.MaxPlayers,
		PlayerCount:          0,
	}
	league.StartDate = time.Now()

	// Defer league creation, to avoid the extra DB update for PlayerCount update

	// Create ownerPlayer
	inLeagueName := "League Owner"
	teamName := fmt.Sprintf("%s's Team", input.Name)

	ownerPlayer := &models.LeagueMember{
		UserID:       userID,
		LeagueID:     leagueID,
		InLeagueName: &inLeagueName,
		TeamName:     &teamName,
		DraftPoints:  int(league.StartingDraftPoints),
		GroupNumber:  1,
		Role:         rbac.MRoleOwner,
	}

	_, err = s.memberRepo.Create(ownerPlayer)
	if err != nil {
		s.logger.Error("CreateLeague - failed to create owner", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("failed to create league owner: %w", err)
	}

	// Update PlayerCount now that the ownerPlayer was successful
	league.PlayerCount = 1

	// Persist league
	createdLeague, err := s.leagueRepo.CreateLeague(league)
	if err != nil {
		s.logger.Error("CreateLeague - failed to create league", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	return createdLeague, nil
}

// GetLeagueByIDForUser Get league entity using leagueID
func (s *leagueServiceImpl) GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error) {
	// User in league checks done at middleware

	// Retrieve the league
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		s.logger.Error("GetLeagueByIDForUser - could not get league", "league_id", leagueID, "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to retrieve league: %w", err)
	}

	return league, nil
}

// GetLeaguesByCommissioner gets all Leagues where userID is the owner
func (s *leagueServiceImpl) GetLeaguesByCommissioner(
	userID uuid.UUID,
	currentUser *models.User,
) ([]models.League, error) {
	leagues, err := s.leagueRepo.GetLeaguesByOwner(userID)
	if err != nil {
		s.logger.Error("GetLeaguesByCommissioner - failed to get commissioner leagues", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to retrieve commissioner leagues: %w", err)
	}

	return leagues, nil
}

// fetches all Leagues where the given userID is a player.
func (s *leagueServiceImpl) GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error) {
	leagues, err := s.leagueRepo.GetLeaguesByUser(userID)
	if err != nil {
		s.logger.Error("GetLeaguesByUser - failed to get leagues", "user_id", userID, "error", err)
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
		s.logger.Error("StartRegularSeason - failed to fetch league", "league_id", leagueID, "error", err)
		return types.ErrLeagueNotFound
	}

	// 1. Validate League Status
	if league.Status != enums.LeagueStatusPostDraft {
		s.logger.Error("StartRegularSeason - league is not in POST_DRAFT status", "league_id", leagueID, "status", league.Status)
		return types.ErrInvalidState
	}

	// 2. Generate Regular Season Games
	if err := s.gameService.GenerateRegularSeasonGames(leagueID); err != nil {
		s.logger.Error("StartRegularSeason - failed to generate regular season games", "league_id", leagueID, "error", err)
		return fmt.Errorf("failed to generate regular season games: %w", err)
	}

	// 3. Update League Status, CurrentWeekNumber, and RegularSeasonStartDate
	now := time.Now()
	league.Status = enums.LeagueStatusRegularSeason
	league.CurrentWeekNumber = 1 // Season starts at Week 1
	league.RegularSeasonStartDate = &now
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("StartRegularSeason - failed to update league status and week info", "league_id", leagueID, "error", err)
		return fmt.Errorf("failed to update league status: %w", err)
	}
	s.logger.Info("StartRegularSeason - league status updated to REGULAR_SEASON", "league_id", leagueID, "week", league.CurrentWeekNumber, "start_date", league.RegularSeasonStartDate.String())

	// 4. Schedule the very first LeagueWeeklyTick
	// The first tick should occur 7 days from now to advance to Week 2.
	firstTickTime := now.Add(7 * 24 * time.Hour)
	league.NextWeeklyTick = &firstTickTime
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("StartRegularSeason - failed to update league with next weekly tick time", "league_id", leagueID, "error", err)
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
	s.logger.Info("StartRegularSeason - first weekly tick scheduled", "league_id", leagueID, "tick_time", firstTickTime.String())

	// 5. Trigger first transfer window if applicable
	// This will call transferService.StartTransferPeriod, which will then schedule its own EndTransferPeriod.
	if league.Format.AllowTransfers && league.Format.TransferWindowFrequencyDays > 0 {
		weeksBetweenWindows := league.Format.TransferWindowFrequencyDays / 7
		if weeksBetweenWindows > 0 && (league.CurrentWeekNumber-1)%weeksBetweenWindows == 0 {
			s.logger.Info("StartRegularSeason - triggering initial transfer window", "league_id", leagueID)
			if err := s.transferService.StartTransferPeriod(leagueID); err != nil {
				s.logger.Error("StartRegularSeason - failed to trigger initial transfer period", "league_id", leagueID, "error", err)
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
		s.logger.Error("ProcessWeeklyTick - failed to fetch league", "league_id", leagueID, "error", err)
		return types.ErrLeagueNotFound
	}

	if league.Status != enums.LeagueStatusRegularSeason {
		s.logger.Info("ProcessWeeklyTick - league not in REGULAR_SEASON, skipping weekly tick", "league_id", leagueID, "status", league.Status)
		return nil // Not an error, just not applicable
	}

	if league.RegularSeasonStartDate == nil {
		s.logger.Error("ProcessWeeklyTick - league has no RegularSeasonStartDate", "league_id", leagueID)
		return fmt.Errorf("league %s is missing RegularSeasonStartDate", leagueID)
	}

	oldWeekNumber := league.CurrentWeekNumber
	now := time.Now()

	// Calculate the correct current week based on RegularSeasonStartDate
	durationSinceSeasonStart := now.Sub(*league.RegularSeasonStartDate)
	calculatedCurrentWeek := int(durationSinceSeasonStart.Hours()/(24*7)) + 1

	if calculatedCurrentWeek > oldWeekNumber {
		// Weeks were missed or it's a natural advancement. Update the CurrentWeekNumber.
		s.logger.Info("ProcessWeeklyTick - advancing week", "league_id", leagueID, "old_week", oldWeekNumber, "new_week", calculatedCurrentWeek)
		league.CurrentWeekNumber = calculatedCurrentWeek
	} else {
		// System is already up-to-date. Log this and fall through to ensure next tick is scheduled correctly.
		s.logger.Info("ProcessWeeklyTick - already at or beyond calculated week, re-scheduling next tick", "league_id", leagueID, "calculated_week", calculatedCurrentWeek, "current_week", oldWeekNumber)
	}

	// Calculate and schedule the next LeagueWeeklyTick based on the CURRENT correct week
	nextTickTime := league.RegularSeasonStartDate.Add(time.Duration(league.CurrentWeekNumber) * 7 * 24 * time.Hour)
	league.NextWeeklyTick = &nextTickTime

	// Save the updated league state BEFORE scheduling the task
	if _, err := s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("ProcessWeeklyTick - failed to save updated league after weekly tick processing", "league_id", leagueID, "error", err)
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
	s.logger.Info("ProcessWeeklyTick - next weekly tick scheduled", "league_id", leagueID, "tick_time", nextTickTime.String(), "next_week", league.CurrentWeekNumber+1)

	// Check for transfer windows ONLY on a natural single-week advancement.
	if calculatedCurrentWeek == oldWeekNumber+1 {
		if league.Format.AllowTransfers {
			weeksBetweenWindows := league.Format.TransferWindowFrequencyDays / 7
			if weeksBetweenWindows > 0 && (league.CurrentWeekNumber-1)%weeksBetweenWindows == 0 {
				s.logger.Info("ProcessWeeklyTick - natural week advancement, triggering transfer window", "league_id", leagueID, "week", league.CurrentWeekNumber)
				if err := s.transferService.StartTransferPeriod(leagueID); err != nil {
					s.logger.Error("ProcessWeeklyTick - failed to trigger transfer period", "league_id", leagueID, "error", err)
				}
			}
		}
	} else if calculatedCurrentWeek > oldWeekNumber {
		// A multi-week jump occurred. Log it, but do not trigger any transfer windows.
		s.logger.Warn("ProcessWeeklyTick - league jumped weeks, transfer window checks bypassed", "league_id", leagueID, "old_week", oldWeekNumber, "new_week", calculatedCurrentWeek)
	}

	return nil
}
