package services

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DevAuthService backs the dev-only impersonation endpoints. It mints real
// JWTs for existing users and fabricates test users/memberships so flows and
// RBAC can be exercised in a browser without going through Discord OAuth.
//
// NEVER register routes backed by this service outside ENV=dev.
type DevAuthService interface {
	ListUsers() ([]models.User, error)
	CreateDevUser(name string, showdownUsername *string) (*models.User, error)
	Impersonate(userID uuid.UUID) (*models.User, string, error)
	UpsertMembership(leagueID, userID uuid.UUID, roleName string) (any, error)
}

type devAuthServiceImpl struct {
	userRepo      repositories.UserRepository
	leagueRepo    repositories.LeagueRepository
	playerRepo    repositories.PlayerRepository
	spectatorRepo repositories.SpectatorRepository
	jwtService    *JWTService
	logger        *slog.Logger
}

func NewDevAuthService(
	logger *slog.Logger,
	userRepo repositories.UserRepository,
	leagueRepo repositories.LeagueRepository,
	playerRepo repositories.PlayerRepository,
	spectatorRepo repositories.SpectatorRepository,
	jwtService *JWTService,
) DevAuthService {
	return &devAuthServiceImpl{
		userRepo:      userRepo,
		leagueRepo:    leagueRepo,
		playerRepo:    playerRepo,
		spectatorRepo: spectatorRepo,
		jwtService:    jwtService,
		logger:        utils.LoggerWithService(logger, "DevAuthService"),
	}
}

func (s *devAuthServiceImpl) ListUsers() ([]models.User, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		s.logger.Error("Failed to list users", "error", err)
		return nil, types.ErrInternalService
	}
	return users, nil
}

func (s *devAuthServiceImpl) CreateDevUser(name string, showdownUsername *string) (*models.User, error) {
	if name == "" {
		return nil, types.ErrInvalidInput
	}

	user := &models.User{
		// DiscordID is unique and required; fake one deterministically namespaced for dev users.
		DiscordID:        "dev-" + uuid.NewString(),
		DiscordUsername:  name,
		DiscordAvatarURL: "",
		ShowdownUsername: showdownUsername,
		Role:             "user",
	}

	created, err := s.userRepo.CreateUser(user)
	if err != nil {
		s.logger.Error("Failed to create user", "name", name, "error", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, types.ErrConflict
		}
		return nil, types.ErrInternalService
	}

	s.logger.Info("created dev user", "user_id", created.ID, "name", name)
	return created, nil
}

func (s *devAuthServiceImpl) Impersonate(userID uuid.UUID) (*models.User, string, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", types.ErrUserNotFound
		}
		s.logger.Error("Failed to fetch user", "user_id", userID, "error", err)
		return nil, "", types.ErrInternalService
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		s.logger.Error("Failed to mint token", "user_id", userID, "error", err)
		return nil, "", types.ErrInternalService
	}

	return user, token, nil
}

func (s *devAuthServiceImpl) UpsertMembership(leagueID, userID uuid.UUID, roleName string) (any, error) {
	role, ok := rbac.ParseMemberRole(roleName)
	if !ok {
		return nil, types.ErrInvalidInput
	}

	if _, err := s.leagueRepo.GetLeagueByID(leagueID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("Failed to fetch league", "league_id", leagueID, "error", err)
		return nil, types.ErrInternalService
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		s.logger.Error("Failed to fetch user", "user_id", userID, "error", err)
		return nil, types.ErrInternalService
	}

	if role.IsParticipant() {
		return s.upsertPlayer(leagueID, userID, role, user)
	}

	return s.upsertSpectator(leagueID, userID, role)
}

func (s *devAuthServiceImpl) upsertPlayer(leagueID, userID uuid.UUID, role rbac.MemberRole, user *models.User) (*models.Player, error) {
	existing, err := s.playerRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("Failed to look up membership",
				"user_id", userID, "league_id",
				leagueID,
				"error", err)
			return nil, types.ErrInternalService
		}

		teamName := fmt.Sprintf("%s's Dev Team", user.DiscordUsername)
		player := &models.Player{
			UserID:   userID,
			LeagueID: leagueID,
			TeamName: &teamName,
			Role:     role,
		}
		created, createErr := s.playerRepo.Create(player)
		if createErr != nil {
			s.logger.Error("Failed to create membership", "user_id", userID, "league_id", leagueID, "error", createErr)
			return nil, types.ErrFailedToCreatePlayer
		}
		s.logger.Info("created dev membership", "user_id", userID, "league_id", leagueID, "role", role)
		return created, nil
	}

	if err := s.playerRepo.UpdateRole(existing.ID, role); err != nil {
		s.logger.Error("Failed to update role", "player_id", existing.ID, "error", err)
		return nil, types.ErrInternalService
	}
	existing.Role = role
	s.logger.Info("updated dev membership role", "player_id", existing.ID, "role", role)
	return existing, nil
}

func (s *devAuthServiceImpl) upsertSpectator(leagueID, userID uuid.UUID, role rbac.MemberRole) (*models.Spectator, error) {
	existing, err := s.spectatorRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("Failed to look up spectator membership",
				"user_id", userID, "league_id", leagueID, "error", err)
			return nil, types.ErrInternalService
		}

		spectator := &models.Spectator{
			UserID:   userID,
			LeagueID: leagueID,
			Role:     role,
			IsViewer: false,
		}
		created, createErr := s.spectatorRepo.Create(spectator)
		if createErr != nil {
			s.logger.Error("Failed to create spectator membership", "user_id", userID, "league_id", leagueID, "error", createErr)
			return nil, types.ErrFailedToCreatePlayer
		}
		s.logger.Info("created dev spectator membership", "user_id", userID, "league_id", leagueID, "role", role)
		return created, nil
	}

	if err := s.spectatorRepo.UpdateRole(existing.ID, role); err != nil {
		s.logger.Error("Failed to update spectator role", "spectator_id", existing.ID, "error", err)
		return nil, types.ErrInternalService
	}
	existing.Role = role
	s.logger.Info("updated dev spectator membership role", "spectator_id", existing.ID, "role", role)
	return existing, nil
}
