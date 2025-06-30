package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
)

// defines the interface for league-related business logic.
type LeagueService interface {
	CreateLeague(userID uuid.UUID, req common.LeagueRequest) (*models.League, error)
	GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error)
}

type leagueServiceImpl struct {
	leagueRepo         *repositories.LeagueRepository
	playerRepo         *repositories.PlayerRepository
	leaguePokemonRepo  *repositories.LeaguePokemonRepository
	draftedPokemonRepo *repositories.DraftedPokemonRepository
	gameRepo           *repositories.GameRepository
}

func NewLeagueService(
	leagueRepo *repositories.LeagueRepository,
	playerRepo *repositories.PlayerRepository,
	leaguePokemonRepo *repositories.LeaguePokemonRepository,
	draftedPokemonRepo *repositories.DraftedPokemonRepository,
	gameRepo *repositories.GameRepository,
) LeagueService {
	return &leagueServiceImpl{
		leagueRepo:         leagueRepo,
		playerRepo:         playerRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
		draftedPokemonRepo: draftedPokemonRepo,
		gameRepo:           gameRepo,
	}
}

// handles the business logic for creating a new league.
func (s *leagueServiceImpl) CreateLeague(userID uuid.UUID, input common.LeagueRequest) (*models.League, error) {
	const maxLeaguesCommisionable = 2

	count, err := s.leagueRepo.GetLeaguesCountByCommissioner(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Could not get commissioner league count for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to check commissioner league count: %w", err)
	}

	if count >= maxLeaguesCommisionable {
		return nil, fmt.Errorf("max league creation limit reached: %d", maxLeaguesCommisionable)
	}

	league := &models.League{
		Name:                  input.Name,
		CommissionerUserID:    userID,
		RulesetID:             input.RulesetID,
		MaxPokemonPerPlayer:   input.MaxPokemonPerPlayer,
		StartingDraftPoints:   int(input.StartingDraftPoints),
		AllowWeeklyFreeAgents: input.AllowWeeklyFreeAgents,
		StartDate:             input.StartDate,
		EndDate:               input.EndDate,
	}

	createdLeague, err := s.leagueRepo.CreateLeague(league)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Failed to create league for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	return createdLeague, nil
}

func (s *leagueServiceImpl) GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error) {
	// Check if user is a player in the league (or commissioner)
	isPlayer, err := s.leagueRepo.IsUserPlayerInLeague(userID, leagueID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeagueByIDForUser) - User in league check failed for user %s, league %s: %v\n", userID, leagueID, err)
		return nil, fmt.Errorf("failed to verify user's league membership: %w", err)
	}

	if !isPlayer {
		return nil, errors.New("not authorized to view this league")
	}

	// Retrieve the league
	league, err := s.leagueRepo.GetLeagueByID(leagueID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeagueByIDForUser) - Could not get league %s: %v\n", leagueID, err)
		return nil, fmt.Errorf("failed to retrieve league: %w", err)
	}

	return league, nil
}
