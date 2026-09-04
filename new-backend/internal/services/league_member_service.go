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

type LeagueMemberService interface {
	GetByID(memberID uuid.UUID) (*models.LeagueMember, error)
	GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error)
	GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error)
	GetByUser(userID uuid.UUID) ([]models.LeagueMember, error)
	GetWithFullRoster(memberID uuid.UUID) (*models.LeagueMember, error)

	Create(currentUser *models.User, input *requests.LeagueMemberCreateRequestDTO) (*models.LeagueMember, error)
	UpdateProfile(currentUser *models.User, memberID uuid.UUID, inLeagueName, teamName *string) (*models.LeagueMember, error)
	UpdateDraftPoints(currentUser *models.User, memberID uuid.UUID, draftPoints *int) (*models.LeagueMember, error)
	UpdateRecord(currentUser *models.User, memberID uuid.UUID, wins, losses int) (*models.LeagueMember, error)
	UpdateDraftPosition(currentUser *models.User, memberID uuid.UUID, draftPosition int) (*models.LeagueMember, error)
	UpdateRole(currentUserID, memberID uuid.UUID, newRole rbac.MemberRole) (*models.LeagueMember, error)
}

type leagueMemberServiceImpl struct {
	memberRepo repositories.LeagueMemberRepository
	leagueRepo repositories.LeagueRepository
	userRepo   repositories.UserRepository
	logger     *slog.Logger
}

func NewLeagueMemberService(
	logger *slog.Logger,
	memberRepo repositories.LeagueMemberRepository,
	leagueRepo repositories.LeagueRepository,
	userRepo repositories.UserRepository,
) LeagueMemberService {
	return &leagueMemberServiceImpl{
		memberRepo: memberRepo,
		leagueRepo: leagueRepo,
		userRepo:   userRepo,
		logger:     utils.LoggerWithService(logger, "LeagueMemberService"),
	}
}

func (s *leagueMemberServiceImpl) GetByID(memberID uuid.UUID) (*models.LeagueMember, error) {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		s.logger.Error("GetByID - failed to retrieve member", "member_id", memberID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve member data", types.ErrInternalService)
	}
	return member, nil
}

func (s *leagueMemberServiceImpl) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error) {
	member, err := s.memberRepo.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		s.logger.Error("GetByUserAndLeague - failed to retrieve member", "user_id", userID, "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve member data", types.ErrInternalService)
	}
	return member, nil
}

func (s *leagueMemberServiceImpl) GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error) {
	members, err := s.memberRepo.GetByLeague(leagueID)
	if err != nil {
		s.logger.Error("GetByLeague - failed to retrieve members", "league_id", leagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve members data", types.ErrInternalService)
	}
	return members, nil
}

func (s *leagueMemberServiceImpl) GetByUser(userID uuid.UUID) ([]models.LeagueMember, error) {
	members, err := s.memberRepo.GetByUser(userID)
	if err != nil {
		s.logger.Error("GetByUser - failed to retrieve members", "user_id", userID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve member data", types.ErrInternalService)
	}
	return members, nil
}

func (s *leagueMemberServiceImpl) GetWithFullRoster(memberID uuid.UUID) (*models.LeagueMember, error) {
	member, err := s.memberRepo.GetWithFullRoster(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		s.logger.Error("GetWithFullRoster - failed to retrieve member with full roster", "member_id", memberID, "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve member data", types.ErrInternalService)
	}
	return member, nil
}

func (s *leagueMemberServiceImpl) Create(currentUser *models.User, input *requests.LeagueMemberCreateRequestDTO) (*models.LeagueMember, error) {
	league, err := s.leagueRepo.GetLeagueByID(input.LeagueID)
	if err != nil {
		s.logger.Error("Create - failed to fetch league", "league_id", input.LeagueID, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve league data", types.ErrInternalService)
	}

	if league.Status != enums.LeagueStatusSetup {
		s.logger.Warn("Create - league is not in SETUP status", "league_id", input.LeagueID, "error", err)
		return nil, types.ErrInvalidState
	}

	if league.MaxPlayers > 0 && league.PlayerCount >= league.MaxPlayers {
		s.logger.Warn("Create - league is full", "league_id", input.LeagueID, "player_count", league.PlayerCount, "max_players", league.MaxPlayers)
		return nil, types.ErrLeagueFull
	}

	user, err := s.userRepo.GetUserByID(input.UserID)
	if err != nil {
		s.logger.Error("Create - failed to fetch user", "user_id", input.UserID, "error", err)
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

	existingByUser, err := s.memberRepo.FindByUserAndLeague(input.UserID, input.LeagueID)
	if err != nil {
		s.logger.Error("Create - failed to check existing member by user ID", "user_id", input.UserID, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to check existing member data", types.ErrInternalService)
	}
	if existingByUser != nil {
		return nil, types.ErrUserAlreadyInLeague
	}

	existingByName, err := s.memberRepo.FindByInLeagueName(*input.InLeagueName, input.LeagueID)
	if err != nil {
		s.logger.Error("Create - failed to check existing member by name", "name", *input.InLeagueName, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to check in-league name uniqueness", types.ErrInternalService)
	}
	if existingByName != nil {
		return nil, fmt.Errorf("%w: '%s'", types.ErrInLeagueNameTaken, *input.InLeagueName)
	}

	existingByTeam, err := s.memberRepo.FindByTeamName(*input.TeamName, input.LeagueID)
	if err != nil {
		s.logger.Error("Create - failed to check existing member by team name", "team_name", *input.TeamName, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to check team name uniqueness", types.ErrInternalService)
	}
	if existingByTeam != nil {
		return nil, fmt.Errorf("%w: '%s'", types.ErrTeamNameTaken, *input.TeamName)
	}

	inLeagueName := *input.InLeagueName
	teamName := *input.TeamName

	member := models.LeagueMember{
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
		Role:          rbac.MRoleMember,
	}

	created, err := s.memberRepo.Create(&member)
	if err != nil {
		s.logger.Error("Create - failed to create member", "user_id", input.UserID, "league_id", input.LeagueID, "error", err)
		return nil, fmt.Errorf("%w: failed to add member to league", types.ErrFailedToCreatePlayer)
	}

	league.PlayerCount++
	groups := max(league.Format.GroupCount, 1)
	league.NewPlayerGroupNumber = ((league.NewPlayerGroupNumber + 1) % groups) + 1
	if _, err = s.leagueRepo.UpdateLeague(league); err != nil {
		s.logger.Error("Create - failed to update league", "league_id", league.ID, "user_id", input.UserID, "error", err)
		return nil, types.ErrInternalService
	}

	s.logger.Info("Create - member created", "member_id", created.ID, "user_id", input.UserID, "league_id", input.LeagueID)
	return created, nil
}

func (s *leagueMemberServiceImpl) UpdateProfile(currentUser *models.User, memberID uuid.UUID, inLeagueName, teamName *string) (*models.LeagueMember, error) {
	existing, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve member for update", types.ErrInternalService)
	}

	if currentUser.Role != "admin" && currentUser.ID != existing.UserID {
		requester, err := s.memberRepo.GetByUserAndLeague(currentUser.ID, existing.LeagueID)
		if err != nil {
			return nil, types.ErrInternalService
		}
		if requester == nil || (!requester.IsLeagueOwner() && !requester.IsLeagueModerator()) {
			return nil, types.ErrUnauthorized
		}
	}

	updated := false
	if inLeagueName != nil {
		if existing.InLeagueName == nil || *inLeagueName != *existing.InLeagueName {
			if *inLeagueName != "" {
				existingByName, err := s.memberRepo.FindByInLeagueName(*inLeagueName, existing.LeagueID)
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
				existingByTeam, err := s.memberRepo.FindByTeamName(*teamName, existing.LeagueID)
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

	updatedMember, err := s.memberRepo.Update(existing)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to save member profile updates", types.ErrInternalService)
	}

	return updatedMember, nil
}

func (s *leagueMemberServiceImpl) UpdateDraftPoints(currentUser *models.User, memberID uuid.UUID, draftPoints *int) (*models.LeagueMember, error) {
	existing, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve member for draft points update", types.ErrInternalService)
	}

	if currentUser.Role != "admin" {
		requester, err := s.memberRepo.GetByUserAndLeague(currentUser.ID, existing.LeagueID)
		if err != nil {
			return nil, types.ErrInternalService
		}
		if requester == nil || (!requester.IsLeagueOwner() && !requester.IsLeagueModerator()) {
			return nil, types.ErrUnauthorized
		}
	}

	if draftPoints == nil {
		return nil, types.ErrInternalService
	}

	err = s.memberRepo.UpdateDraftPoints(memberID, *draftPoints)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update member draft points", types.ErrInternalService)
	}

	updated, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to re-fetch updated member", types.ErrInternalService)
	}

	return updated, nil
}

func (s *leagueMemberServiceImpl) UpdateRecord(currentUser *models.User, memberID uuid.UUID, wins, losses int) (*models.LeagueMember, error) {
	existing, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve member for record update", types.ErrInternalService)
	}

	if currentUser.Role != "admin" {
		requester, err := s.memberRepo.GetByUserAndLeague(currentUser.ID, existing.LeagueID)
		if err != nil {
			return nil, types.ErrInternalService
		}
		if requester == nil || (!requester.IsLeagueOwner() && !requester.IsLeagueModerator()) {
			return nil, types.ErrUnauthorized
		}
	}

	err = s.memberRepo.UpdateRecord(memberID, wins, losses)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update member record", types.ErrInternalService)
	}

	updated, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to re-fetch updated member", types.ErrInternalService)
	}

	return updated, nil
}

func (s *leagueMemberServiceImpl) UpdateDraftPosition(currentUser *models.User, memberID uuid.UUID, draftPosition int) (*models.LeagueMember, error) {
	existing, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve member for draft position update", types.ErrInternalService)
	}

	if currentUser.Role != "admin" {
		requester, err := s.memberRepo.GetByUserAndLeague(currentUser.ID, existing.LeagueID)
		if err != nil {
			return nil, types.ErrInternalService
		}
		if requester == nil || (!requester.IsLeagueOwner() && !requester.IsLeagueModerator()) {
			return nil, types.ErrUnauthorized
		}
	}

	err = s.memberRepo.UpdateDraftPosition(memberID, draftPosition)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update member draft position", types.ErrInternalService)
	}

	updated, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to re-fetch updated member", types.ErrInternalService)
	}

	return updated, nil
}

func (s *leagueMemberServiceImpl) UpdateRole(currentUserID, memberID uuid.UUID, newRole rbac.MemberRole) (*models.LeagueMember, error) {
	err := s.memberRepo.UpdateRole(memberID, newRole)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrPlayerNotFound
		}
		return nil, types.ErrInternalService
	}

	updated, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, types.ErrInternalService
	}

	return updated, nil
}
