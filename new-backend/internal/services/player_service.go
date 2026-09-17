package services

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerService interface {
	GetByID(playerID uuid.UUID) (*models.Player, error)
	GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error)
	GetByLeague(leagueID uuid.UUID) ([]models.Player, error)
	GetWithFullRoster(playerID uuid.UUID) (*models.Player, error)

	Create(currentUser *models.User, input *requests.PlayerCreateRequestDTO) (*models.Player, error)
	UpdateProfile(playerID uuid.UUID, inLeagueName, teamName *string) (*models.Player, error)
	UpdateOwnProfile(userID, leagueID uuid.UUID, inLeagueName, teamName *string) (*models.Player, error)
	UpdateDraftPoints(playerID uuid.UUID, draftPoints *int) (*models.Player, error)
	UpdateRecord(playerID uuid.UUID, wins, losses int) (*models.Player, error)
	UpdateDraftPosition(playerID uuid.UUID, draftPosition int) (*models.Player, error)
	UpdateRole(playerID uuid.UUID, newRole rbac.MemberRole) (*models.Player, error)
}

type playerServiceImpl struct {
	playerRepo    repositories.PlayerRepository
	leagueRepo    repositories.LeagueRepository
	userRepo      repositories.UserRepository
	spectatorRepo repositories.SpectatorRepository
	logger        *slog.Logger
}

func NewPlayerService(
	logger *slog.Logger,
	playerRepo repositories.PlayerRepository,
	leagueRepo repositories.LeagueRepository,
	userRepo repositories.UserRepository,
	spectatorRepo repositories.SpectatorRepository,
) PlayerService {
	return &playerServiceImpl{
		playerRepo:    playerRepo,
		leagueRepo:    leagueRepo,
		userRepo:      userRepo,
		spectatorRepo: spectatorRepo,
		logger:        utils.LoggerWithService(logger, "PlayerService"),
	}
}

func (s *playerServiceImpl) GetByID(playerID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		s.logger.Error("Failed to retrieve player", "player_id", playerID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", types.ErrInternalService)
	}
	return player, nil
}

func (s *playerServiceImpl) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		s.logger.Error("Failed to retrieve player", "user_id", userID, "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", types.ErrInternalService)
	}
	return player, nil
}

func (s *playerServiceImpl) GetByLeague(leagueID uuid.UUID) ([]models.Player, error) {
	players, err := s.playerRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("Failed to retrieve players", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve players data", types.ErrInternalService)
	}
	return players, nil
}

func (s *playerServiceImpl) GetWithFullRoster(playerID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetWithFullRoster(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		s.logger.Error("Failed to retrieve player with full roster", "player_id", playerID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", types.ErrInternalService)
	}
	return player, nil
}

func (s *playerServiceImpl) Create(currentUser *models.User, input *requests.PlayerCreateRequestDTO) (*models.Player, error) {
	league, err := s.leagueRepo.GetLeagueByID(input.LeagueID)
	if err != nil {
		s.logger.Error("Failed to fetch league", "league_id", input.LeagueID, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve league data", types.ErrInternalService)
	}

	if league.Status != enums.LeagueStatusSetup {
		s.logger.Warn("League is not in SETUP status", "league_id", input.LeagueID, "error", err)
		return nil, types.ErrInvalidState
	}

	if league.MaxPlayers > 0 && league.PlayerCount >= league.MaxPlayers {
		s.logger.Warn("League is full", "league_id", input.LeagueID, "player_count", league.PlayerCount, "max_players", league.MaxPlayers)
		return nil, types.ErrLeagueFull
	}

	user, err := s.userRepo.GetUserByID(input.UserID)
	if err != nil {
		s.logger.Error("Failed to fetch user", "user_id", input.UserID, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve user data", types.ErrInternalService)
	}

	if input.InLeagueName == nil {
		name := user.DiscordUsername
		input.InLeagueName = &name
	}
	if input.TeamName == nil {
		teamName := fmt.Sprintf("%s's Team", user.DiscordUsername)
		input.TeamName = &teamName
	}

	existingByUser, err := s.playerRepo.FindByUserAndLeague(input.UserID, input.LeagueID)
	if err != nil {
		s.logger.Error("Failed to check existing player by user ID", "user_id", input.UserID, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to check existing player data", types.ErrInternalService)
	}
	if existingByUser != nil {
		return nil, types.ErrUserAlreadyInLeague
	}

	// A spectator may not also become a player. If the user is already a
	// spectator, joining is denied — with a dedicated error if that spectator
	// is the league owner, since owner-as-player is handled separately.
	existingSpectator, spectatorErr := s.spectatorRepo.GetByUserAndLeague(input.UserID, input.LeagueID)
	if spectatorErr != nil && !errors.Is(spectatorErr, gorm.ErrRecordNotFound) {
		s.logger.Error("Failed to check existing spectator", "user_id", input.UserID, "league_id", input.LeagueID, "error", spectatorErr)
		return nil, fmt.Errorf("%w: failed to check existing spectator data", types.ErrInternalService)
	}
	if existingSpectator != nil {
		if existingSpectator.Role.IsOwner() {
			return nil, types.ErrOwnerSpectatorJoin
		}
		return nil, types.ErrUserAlreadySpectator
	}

	existingByName, err := s.playerRepo.FindByInLeagueName(*input.InLeagueName, input.LeagueID)
	if err != nil {
		s.logger.Error("Failed to check existing player by name", "name", *input.InLeagueName, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to check in-league name uniqueness", types.ErrInternalService)
	}
	if existingByName != nil {
		return nil, fmt.Errorf("%w: '%s'", types.ErrInLeagueNameTaken, *input.InLeagueName)
	}

	existingByTeam, err := s.playerRepo.FindByTeamName(*input.TeamName, input.LeagueID)
	if err != nil {
		s.logger.Error("Failed to check existing player by team name", "team_name", *input.TeamName, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to check team name uniqueness", types.ErrInternalService)
	}
	if existingByTeam != nil {
		return nil, fmt.Errorf("%w: '%s'", types.ErrTeamNameTaken, *input.TeamName)
	}

	inLeagueName := *input.InLeagueName
	teamName := *input.TeamName

	player := models.Player{
		UserID:        input.UserID,
		LeagueID:      input.LeagueID,
		InLeagueName:  &inLeagueName,
		TeamName:      &teamName,
		DraftPoints:   int(league.StartingDraftPoints),
		Wins:          0,
		Losses:        0,
		DraftPosition: 0,
		GroupNumber:   league.NewPlayerGroupNumber,
		SkipsLeft:     league.MaxPokemonPerPlayer - league.MinPokemonPerPlayer,
		Role:          rbac.MRoleRegular,
	}

	created, err := s.playerRepo.Create(&player)
	if err != nil {
		s.logger.Error("Failed to create player", "user_id", input.UserID, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to add player to league", types.ErrFailedToCreatePlayer)
	}

	league.PlayerCount++
	groups := max(league.Format.GroupCount, 1)
	league.NewPlayerGroupNumber = ((league.NewPlayerGroupNumber + 1) % groups) + 1
	if _, err = s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("Failed to update league", "league_id", league.ID, "user_id", input.UserID, "error", err)
		return nil, types.ErrInternalService
	}

	s.logger.Info("Player created", "player_id", created.ID, "user_id", input.UserID, "league_id", input.LeagueID)
	return created, nil
}

func (s *playerServiceImpl) UpdateProfile(playerID uuid.UUID, inLeagueName, teamName *string) (*models.Player, error) {
	existing, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve player for update", types.ErrInternalService)
	}

	return s.updateProfileDetails(existing, inLeagueName, teamName)
}

func (s *playerServiceImpl) UpdateOwnProfile(userID, leagueID uuid.UUID, inLeagueName, teamName *string) (*models.Player, error) {
	existing, err := s.playerRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve player for update", types.ErrInternalService)
	}

	return s.updateProfileDetails(existing, inLeagueName, teamName)
}

func (s *playerServiceImpl) updateProfileDetails(existing *models.Player, inLeagueName, teamName *string) (*models.Player, error) {
	updated := false
	if inLeagueName != nil {
		if existing.InLeagueName == nil || *inLeagueName != *existing.InLeagueName {
			if *inLeagueName != "" {
				existingByName, err := s.playerRepo.FindByInLeagueName(*inLeagueName, existing.LeagueID)
				if err != nil {
					return nil, fmt.Errorf("%w: failed to check in-league name uniqueness", types.ErrInternalService)
				}
				if existingByName != nil && existingByName.ID != existing.ID {
					return nil, fmt.Errorf("%w: '%s'", types.ErrInLeagueNameTaken, *inLeagueName)
				}
			}
			existing.InLeagueName = inLeagueName
			updated = true
		}
	}

	if teamName != nil {
		if existing.TeamName == nil || *teamName != *existing.TeamName {
			if *teamName != "" {
				existingByTeam, err := s.playerRepo.FindByTeamName(*teamName, existing.LeagueID)
				if err != nil {
					return nil, fmt.Errorf("%w: failed to check team name uniqueness", types.ErrInternalService)
				}
				if existingByTeam != nil && existingByTeam.ID != existing.ID {
					return nil, fmt.Errorf("%w: '%s'", types.ErrTeamNameTaken, *teamName)
				}
			}
			existing.TeamName = teamName
			updated = true
		}
	}

	if !updated {
		return existing, nil
	}

	updatedPlayer, err := s.playerRepo.Update(existing)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to save player profile updates", types.ErrInternalService)
	}

	return updatedPlayer, nil
}

func (s *playerServiceImpl) UpdateDraftPoints(playerID uuid.UUID, draftPoints *int) (*models.Player, error) {
	if draftPoints == nil {
		return nil, types.ErrInvalidInput
	}

	existing, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve player for draft points update", types.ErrInternalService)
	}

	err = s.playerRepo.UpdateDraftPoints(existing.ID, *draftPoints)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update player draft points", types.ErrInternalService)
	}

	updated, err := s.playerRepo.GetByID(existing.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to re-fetch updated player", types.ErrInternalService)
	}

	return updated, nil
}

func (s *playerServiceImpl) UpdateRecord(playerID uuid.UUID, wins, losses int) (*models.Player, error) {
	existing, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve player for record update", types.ErrInternalService)
	}

	err = s.playerRepo.UpdateRecord(existing.ID, wins, losses)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update player record", types.ErrInternalService)
	}

	updated, err := s.playerRepo.GetByID(existing.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to re-fetch updated player", types.ErrInternalService)
	}

	return updated, nil
}

func (s *playerServiceImpl) UpdateDraftPosition(playerID uuid.UUID, draftPosition int) (*models.Player, error) {
	existing, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve player for draft position update", types.ErrInternalService)
	}

	err = s.playerRepo.UpdateDraftPosition(existing.ID, draftPosition)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update player draft position", types.ErrInternalService)
	}

	updated, err := s.playerRepo.GetByID(existing.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to re-fetch updated player", types.ErrInternalService)
	}

	return updated, nil
}

func (s *playerServiceImpl) UpdateRole(playerID uuid.UUID, newRole rbac.MemberRole) (*models.Player, error) {
	// Upgrade/downgrade within the player roles; spectator roles live on the
	// spectators table and are handled by the spectator service.
	if !newRole.IsParticipant() {
		return nil, types.ErrInvalidInput
	}

	if _, err := s.playerRepo.GetByID(playerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, types.ErrInternalService
	}

	if err := s.playerRepo.UpdateRole(playerID, newRole); err != nil {
		return nil, types.ErrInternalService
	}

	updated, err := s.playerRepo.GetByID(playerID)
	if err != nil {
		return nil, types.ErrInternalService
	}

	return updated, nil
}
