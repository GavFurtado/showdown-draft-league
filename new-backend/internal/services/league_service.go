package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
)

// defines the interface for league-related business logic.
type LeagueService interface {
	// handles the business logic for creating a new league.
	CreateLeague(userID uuid.UUID, req *common.LeagueRequest) (*models.League, error)
	// Get league entity using leagueID
	GetLeagueByIDForUser(userID, leagueID uuid.UUID) (*models.League, error)
	// gets all Leagues where userID is the commissioner
	GetLeaguesByCommissioner(userID uuid.UUID, currentUser *models.User) ([]models.League, error)
	// fetches all Leagues where the given userID is a player.
	GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error)
}

type leagueServiceImpl struct {
	leagueRepo         repositories.LeagueRepository
	playerRepo         repositories.PlayerRepository
	leaguePokemonRepo  repositories.LeaguePokemonRepository
	draftedPokemonRepo repositories.DraftedPokemonRepository
	draftRepo          repositories.DraftRepository
	gameRepo           repositories.GameRepository
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

// handles the business logic for creating a new league.
func (s *leagueServiceImpl) CreateLeague(userID uuid.UUID, input *common.LeagueRequest) (*models.League, error) {
	const maxLeaguesCommisionable = 2

	// check if user already has two owned leagues
	count, err := s.leagueRepo.GetLeaguesCountWhereOwner(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Could not get commissioner league count for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to check commissioner league count: %w", err)
	}

	if count >= maxLeaguesCommisionable {
		return nil, fmt.Errorf("max league creation limit reached: %d", maxLeaguesCommisionable)
	}

	league := &models.League{
		Name:                input.Name,
		RulesetDescription:  input.RulesetDescription,
		MaxPokemonPerPlayer: input.MaxPokemonPerPlayer,
		StartingDraftPoints: input.StartingDraftPoints,
		StartDate:           input.StartDate,
		EndDate:             input.EndDate,
		Format: models.LeagueFormat{
			SeasonType:              input.Format.SeasonType,
			GroupCount:              input.Format.GroupCount,
			GamesPerOpponent:        input.Format.GamesPerOpponent,
			PlayoffType:             input.Format.PlayoffType,
			PlayoffTeams:            input.Format.PlayoffTeams,
			PlayoffByes:             input.Format.PlayoffByes,
			PlayoffSeedingType:      input.Format.PlayoffSeedingType,
			IsSnakeRoundDraft:       input.Format.IsSnakeRoundDraft,
			AllowTrading:            input.Format.AllowTrading,
			AllowTransferCredits:    input.Format.AllowTransferCredits,
			TransferCreditsPerRound: input.Format.TransferCreditsPerRound,
		},
	}

	createdLeague, err := s.leagueRepo.CreateLeague(league)
	if err != nil {
		log.Printf("(Error: LeagueService.CreateLeague) - Failed to create league for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	ownerPlayer := &models.Player{
		UserID:          userID,
		LeagueID:        createdLeague.ID,
		InLeagueName:    "League Owner",                       // Default, can be updated later
		TeamName:        fmt.Sprintf("%s's Team", input.Name), // Default, can be updated later
		IsParticipating: false,
		DraftPoints:     int(createdLeague.StartingDraftPoints),
		Role:            rbac.Owner,
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

// gets all Leagues where userID is the commissioner
func (s *leagueServiceImpl) GetLeaguesByCommissioner(
	userID uuid.UUID,
	currentUser *models.User,
) ([]models.League, error) {
	// Authorization: Only admin or the user themselves can view their commissioner leagues
	if currentUser.Role != "admin" && currentUser.ID != userID {
		log.Printf("(Error: LeagueService.GetLeaguesByCommissioner) - Unauthorized access attempt by user %s to view commissioner leagues for user %s", currentUser.ID, userID)
		return nil, errors.New("not authorized to view these leagues")
	}

	leagues, err := s.leagueRepo.GetLeaguesByOwner(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeaguesByCommissioner) - Failed to get commissioner leagues for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to retrieve commissioner leagues: %w", err)
	}

	return leagues, nil
}

// fetches all Leagues where the given userID is a player.
func (s *leagueServiceImpl) GetLeaguesByUser(userID uuid.UUID, currentUser *models.User) ([]models.League, error) {
	// Authorization: Only admin or the user themselves can view their leagues
	if currentUser.Role != "admin" && currentUser.ID != userID {
		log.Printf("(Error: LeagueService.GetLeaguesByUser) - Unauthorized access attempt by user %s to view leagues for user %s", currentUser.ID, userID)
		return nil, errors.New("not authorized to view these leagues")
	}

	leagues, err := s.leagueRepo.GetLeaguesByUser(userID)
	if err != nil {
		log.Printf("(Error: LeagueService.GetLeaguesByUser) - Failed to get leagues for user %s: %v\n", userID, err)
		return nil, fmt.Errorf("failed to retrieve leagues: %w", err)
	}
	return leagues, nil
}
