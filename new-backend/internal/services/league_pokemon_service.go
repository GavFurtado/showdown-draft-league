package services

import (
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaguePokemonService interface {
	CreatePokemonForLeague(currentUser *models.User, input *common.LeaguePokemonCreateRequest) (*models.LeaguePokemon, error)
}

type leaguePokemonServiceImpl struct {
	leaguePokemonRepo *repositories.LeaguePokemonRepository
	leagueRepo        *repositories.LeagueRepository
	userRepo          *repositories.UserRepository
}

func NewLeaguePokemonService(
	leaguePokemonRepo *repositories.LeaguePokemonRepository,
	leagueRepo *repositories.LeagueRepository,
	userRepo *repositories.UserRepository,
) LeaguePokemonService {
	return &leaguePokemonServiceImpl{
		leaguePokemonRepo: leaguePokemonRepo,
		leagueRepo:        leagueRepo,
		userRepo:          userRepo,
	}
}

// -- Private Helpers --
func (s *leaguePokemonServiceImpl) isUserCommissioner(userID, leagueID uuid.UUID) (bool, error) {
	isComm, err := s.leagueRepo.IsUserCommissioner(userID, leagueID)
	if err != nil {
		log.Printf("(Service: isUserCommissioner) - Failed to check commissioner status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check commissioner status: %w", err)
	}
	return isComm, nil
}

func (s *leaguePokemonServiceImpl) isUserPlayerInLeague(userID, leagueID uuid.UUID) (bool, error) {
	isPlayer, err := s.leagueRepo.IsUserPlayerInLeague(userID, leagueID)
	if err != nil {
		log.Printf("(Service: isUserPlayerInLeague) - Failed to check player status for user %s in league %s: %v", userID, leagueID, err)
		return false, fmt.Errorf("failed to check player status: %w", err)
	}
	return isPlayer, nil
}

func (s *leaguePokemonServiceImpl) getLeagueByID(leagueID, currentUserID uuid.UUID) (*models.League, error) {
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("(Service: getLeagueByID) - could not find league %d. (currentUser.ID: %d)\n", leagueID, currentUserID)
			return nil, common.ErrLeagueNotFound
		}
		// other error
		log.Printf("(Service: getLeagueByID) - could not retrieve league by leagueID %d (currentUser.ID: %d)\n", leagueID, currentUserID)
		return nil, common.ErrInternalService
	}
	return league, nil
}

func (s *leaguePokemonServiceImpl) CreatePokemonForLeague(currentUser *models.User, input *common.LeaguePokemonCreateRequest) (*models.LeaguePokemon, error) {
	// fetch league
	league, err := s.leagueRepo.GetLeagueByID(input.LeagueID)
	if err != nil {
		log.Printf("(Service: CreatePokemonForLeague) - Error:\n")
		return nil, err
	}

	// check if user has a player in league or if user is an admin
	if inLeague, err := s.isUserPlayerInLeague(currentUser.ID, league.ID); !inLeague && !currentUser.IsAdmin {
		if err != nil {
			log.Printf("(Service: CreatePokemonForLeague) - Error:\n")
			return nil, err
		}
		log.Printf("(Service: CreatePokemonForLeague) - user %d does not have a player in the league %d.\n", currentUser.ID, league.ID)
		return nil, common.ErrUnauthorized
	}

	// check if commisisoner or an admin
	if isComm, err := s.isUserCommissioner(currentUser.ID, league.ID); !isComm && !currentUser.IsAdmin {
		if err != nil {
			log.Printf("(Service: CreatePokemonForLeague) - Error:\n")
			return nil, err
		}
		log.Printf("(Service: CreatePokemonForLeague) - user %d does not have a commissioner player for the league %d:\n", currentUser.ID, league.ID)
		return nil, common.ErrUnauthorized
	}

	// user is a commissioner player in the league which exists
	// check league status
	// if league.Status != models.LeagueStatusSetup

	// if s.isUserCommissioner(currentUser.ID)
	return nil, nil
}
