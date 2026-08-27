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
	UpsertMembership(leagueID, userID uuid.UUID, roleName string) (*models.LeagueMember, error)
}

type devAuthServiceImpl struct {
	userRepo         repositories.UserRepository
	leagueRepo       repositories.LeagueRepository
	leagueMemberRepo repositories.LeagueMemberRepository
	jwtService       *JWTService
	logger           *slog.Logger
}

func NewDevAuthService(
	logger *slog.Logger,
	userRepo repositories.UserRepository,
	leagueRepo repositories.LeagueRepository,
	leagueMemberRepo repositories.LeagueMemberRepository,
	jwtService *JWTService,
) DevAuthService {
	return &devAuthServiceImpl{
		userRepo:         userRepo,
		leagueRepo:       leagueRepo,
		leagueMemberRepo: leagueMemberRepo,
		jwtService:       jwtService,
		logger:           utils.LoggerWithService(logger, "DevAuthService"),
	}
}

func (s *devAuthServiceImpl) ListUsers() ([]models.User, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		s.logger.Error("ListUsers - failed to list users", "error", err)
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
		s.logger.Error("CreateDevUser - failed to create user", "name", name, "error", err)
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
		s.logger.Error("Impersonate - failed to fetch user", "user_id", userID, "error", err)
		return nil, "", types.ErrInternalService
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		s.logger.Error("Impersonate - failed to mint token", "user_id", userID, "error", err)
		return nil, "", types.ErrInternalService
	}

	return user, token, nil
}

func (s *devAuthServiceImpl) UpsertMembership(leagueID, userID uuid.UUID, roleName string) (*models.LeagueMember, error) {
	role, ok := rbac.ParseMemberRole(roleName)
	if !ok {
		return nil, types.ErrInvalidInput
	}

	if _, err := s.leagueRepo.GetLeagueByID(leagueID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		s.logger.Error("UpsertMembership - failed to fetch league", "league_id", leagueID, "error", err)
		return nil, types.ErrInternalService
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		s.logger.Error("UpsertMembership - failed to fetch user", "user_id", userID, "error", err)
		return nil, types.ErrInternalService
	}

	existing, err := s.leagueMemberRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("UpsertMembership - failed to look up membership", "user_id", userID, "league_id", leagueID, "error", err)
			return nil, types.ErrInternalService
		}

		teamName := fmt.Sprintf("%s's Dev Team", user.DiscordUsername)
		member := &models.LeagueMember{
			UserID:          userID,
			LeagueID:        leagueID,
			TeamName:        &teamName,
			Role:            role,
			IsParticipating: true,
		}
		created, createErr := s.leagueMemberRepo.Create(member)
		if createErr != nil {
			s.logger.Error("UpsertMembership - failed to create membership", "user_id", userID, "league_id", leagueID, "error", createErr)
			return nil, types.ErrFailedToCreatePlayer
		}
		s.logger.Info("created dev membership", "user_id", userID, "league_id", leagueID, "role", role)
		return created, nil
	}

	if err := s.leagueMemberRepo.UpdateRole(existing.ID, role); err != nil {
		s.logger.Error("UpsertMembership - failed to update role", "member_id", existing.ID, "error", err)
		return nil, types.ErrInternalService
	}
	existing.Role = role
	s.logger.Info("updated dev membership role", "member_id", existing.ID, "role", role)
	return existing, nil
}
