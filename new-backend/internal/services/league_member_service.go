package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeagueMemberService interface {
	GetByID(memberID uuid.UUID) (*models.LeagueMember, error)
	GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error)
	GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error)
	GetByUser(userID uuid.UUID) ([]models.LeagueMember, error)
	GetWithFullRoster(memberID uuid.UUID) (*models.LeagueMember, error)
	GetRosterByWeek(memberID uuid.UUID, weekNumber int) ([]models.DraftedPokemon, error)

	Create(currentUser *models.User, input *requests.LeagueMemberCreateRequestDTO) (*models.LeagueMember, error)
	UpdateProfile(currentUser *models.User, memberID uuid.UUID, inLeagueName, teamName *string) (*models.LeagueMember, error)
	UpdateDraftPoints(currentUser *models.User, memberID uuid.UUID, draftPoints *int) (*models.LeagueMember, error)
	UpdateRecord(currentUser *models.User, memberID uuid.UUID, wins, losses int) (*models.LeagueMember, error)
	UpdateDraftPosition(currentUser *models.User, memberID uuid.UUID, draftPosition int) (*models.LeagueMember, error)
	UpdateRole(currentUserID, memberID uuid.UUID, newRole rbac.MemberRole) (*models.LeagueMember, error)
}

type leagueMemberServiceImpl struct {
	memberRepo         repositories.LeagueMemberRepository
	playerRepo         repositories.PlayerRepository
	leagueRepo         repositories.LeagueRepository
	userRepo           repositories.UserRepository
	draftedPokemonRepo repositories.DraftedPokemonRepository
}

func NewLeagueMemberService(
	memberRepo repositories.LeagueMemberRepository,
	playerRepo repositories.PlayerRepository,
	leagueRepo repositories.LeagueRepository,
	userRepo repositories.UserRepository,
	draftedPokemonRepo repositories.DraftedPokemonRepository,
) LeagueMemberService {
	return &leagueMemberServiceImpl{
		memberRepo:         memberRepo,
		playerRepo:         playerRepo,
		leagueRepo:         leagueRepo,
		userRepo:           userRepo,
		draftedPokemonRepo: draftedPokemonRepo,
	}
}

func (s *leagueMemberServiceImpl) GetByID(memberID uuid.UUID) (*models.LeagueMember, error) {
	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		log.Printf("Service: LeagueMemberService.GetByID - Failed to retrieve member %s: %v", memberID, err)
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
		log.Printf("Service: LeagueMemberService.GetByUserAndLeague - Failed to retrieve member (userID %s; leagueID %s): %v", userID, leagueID, err)
		return nil, fmt.Errorf("%w: failed to retrieve member data", types.ErrInternalService)
	}
	return member, nil
}

func (s *leagueMemberServiceImpl) GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error) {
	members, err := s.memberRepo.GetByLeague(leagueID)
	if err != nil {
		log.Printf("Service: LeagueMemberService.GetByLeague - Failed to retrieve members for league %s: %v", leagueID, err)
		return nil, fmt.Errorf("%w: failed to retrieve members data", types.ErrInternalService)
	}
	return members, nil
}

func (s *leagueMemberServiceImpl) GetByUser(userID uuid.UUID) ([]models.LeagueMember, error) {
	members, err := s.memberRepo.GetByUser(userID)
	if err != nil {
		log.Printf("Service: LeagueMemberService.GetByUser - Failed to retrieve members for user %s: %v", userID, err)
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
		log.Printf("Service: LeagueMemberService.GetWithFullRoster - Failed to retrieve member %s with full roster: %v", memberID, err)
		return nil, fmt.Errorf("%w: failed to retrieve member data", types.ErrInternalService)
	}
	return member, nil
}

func (s *leagueMemberServiceImpl) GetRosterByWeek(memberID uuid.UUID, weekNumber int) ([]models.DraftedPokemon, error) {
	_, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve member data: %v", types.ErrInternalService, err)
	}

	allDrafted, err := s.draftedPokemonRepo.GetAllDraftedPokemonByPlayer(memberID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve all drafted pokemon for member: %v", types.ErrInternalService, err)
	}

	var rosterForWeek []models.DraftedPokemon
	for _, pokemon := range allDrafted {
		isAcquired := pokemon.AcquiredWeek <= weekNumber
		isNotReleasedYet := !pokemon.IsReleased || (pokemon.ReleasedWeek != nil && *pokemon.ReleasedWeek > weekNumber)
		if isAcquired && isNotReleasedYet {
			rosterForWeek = append(rosterForWeek, pokemon)
		}
	}

	return rosterForWeek, nil
}

func (s *leagueMemberServiceImpl) Create(currentUser *models.User, input *requests.LeagueMemberCreateRequestDTO) (*models.LeagueMember, error) {
	league, err := s.leagueRepo.GetLeagueByID(input.LeagueID)
	if err != nil {
		log.Printf("Service: LeagueMemberService.Create - Failed to fetch league %s: %v", input.LeagueID, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrLeagueNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve league data", types.ErrInternalService)
	}

	if league.Status != enums.LeagueStatusSetup {
		log.Printf("Service: LeagueMemberService.Create - League %s is not in SETUP status: %v", input.LeagueID, err)
		return nil, types.ErrInvalidState
	}

	user, err := s.userRepo.GetUserByID(input.UserID)
	if err != nil {
		log.Printf("Service: LeagueMemberService.Create - Failed to fetch user %s: %v", input.UserID, err)
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
		log.Printf("Service: LeagueMemberService.Create - Failed to check existing member by user ID %s in league %s: %v", input.UserID, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to check existing member data", types.ErrInternalService)
	}
	if existingByUser != nil {
		return nil, types.ErrUserAlreadyInLeague
	}

	existingByName, err := s.memberRepo.FindByInLeagueName(*input.InLeagueName, input.LeagueID)
	if err != nil {
		log.Printf("Service: LeagueMemberService.Create - Failed to check existing member by name '%s' in league %s: %v", *input.InLeagueName, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to check in-league name uniqueness", types.ErrInternalService)
	}
	if existingByName != nil {
		return nil, fmt.Errorf("%w: '%s'", types.ErrInLeagueNameTaken, *input.InLeagueName)
	}

	existingByTeam, err := s.memberRepo.FindByTeamName(*input.TeamName, input.LeagueID)
	if err != nil {
		log.Printf("Service: LeagueMemberService.Create - Failed to check existing member by team name '%s' in league %s: %v", *input.TeamName, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to check team name uniqueness", types.ErrInternalService)
	}
	if existingByTeam != nil {
		return nil, fmt.Errorf("%w: '%s'", types.ErrTeamNameTaken, *input.TeamName)
	}

	inLeagueName := *input.InLeagueName
	teamName := *input.TeamName

	member := models.LeagueMember{
		UserID:       input.UserID,
		LeagueID:     input.LeagueID,
		InLeagueName: &inLeagueName,
		TeamName:     &teamName,
		DraftPoints:  int(league.StartingDraftPoints),
		Wins:         0,
		Losses:       0,
		DraftPosition: 0,
		GroupNumber:   league.NewPlayerGroupNumber,
		SkipsLeft:     league.MaxPokemonPerPlayer - league.MinPokemonPerPlayer,
		Role:          rbac.MRoleMember,
	}

	created, err := s.memberRepo.Create(&member)
	if err != nil {
		log.Printf("Service: LeagueMemberService.Create - Failed to create member for user %s in league %s: %v", input.UserID, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to add member to league", types.ErrFailedToCreatePlayer)
	}

	league.PlayerCount++
	league.NewPlayerGroupNumber = ((league.NewPlayerGroupNumber + 1) % league.Format.GroupCount) + 1
	if _, err = s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("Service: LeagueMemberService.Create - Failed to update league %s for user %s: %v", league.ID, input.UserID, err)
		return nil, types.ErrInternalService
	}

	log.Printf("Service: LeagueMemberService.Create - Member %s created for user %s in league %s.", created.ID, input.UserID, input.LeagueID)
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
