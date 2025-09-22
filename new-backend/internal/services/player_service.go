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
	"gorm.io/gorm"
)

type PlayerService interface {
	CreatePlayerHandler(input *common.PlayerCreateRequest) (*models.Player, error)

	GetPlayerByIDHandler(playerID uuid.UUID, currentUser *models.User) (*models.Player, error)
	GetPlayersByLeagueHandler(leagueID, userID uuid.UUID, isUserAnAdmin bool) ([]models.Player, error)
	GetPlayersByUserHandler(userID, currentUserID uuid.UUID, isCurrentUserAnAdmin bool) ([]models.Player, error)
	GetPlayerWithFullRosterHandler(playerID, currentUserID uuid.UUID, isCurrentUserAnAdmin bool) (*models.Player, error)

	UpdatePlayerProfile(currentUser *models.User, playerID uuid.UUID, inLeagueName *string, teamName *string) (*models.Player, error)
	UpdatePlayerDraftPoints(currentUser *models.User, playerID uuid.UUID, draftPoints *int) (*models.Player, error)
	UpdatePlayerRecord(currentUser *models.User, playerID uuid.UUID, wins int, losses int) (*models.Player, error)
	UpdatePlayerDraftPosition(currentUser *models.User, playerID uuid.UUID, draftPosition int) (*models.Player, error)

	// can't be implemented currently, see "definition" of this function for more information
	// SetCommissionerStatus(playerID uuid.UUID, isCommissioner bool, currentUser *models.User) error
	// (s *playerServiceImpl) LeaveLeague(playerID uuid.UUID) error
}

type playerServiceImpl struct {
	playerRepo repositories.PlayerRepository
	leagueRepo repositories.LeagueRepository
	userRepo   repositories.UserRepository
}

func NewPlayerService(
	playerRepo repositories.PlayerRepository,
	leagueRepo repositories.LeagueRepository,
	userRepo repositories.UserRepository,
) PlayerService {
	return &playerServiceImpl{
		playerRepo: playerRepo,
		leagueRepo: leagueRepo,
		userRepo:   userRepo,
	}
}

func (s *playerServiceImpl) CreatePlayerHandler(input *common.PlayerCreateRequest) (*models.Player, error) {
	// fetch League and User details
	league, err := s.leagueRepo.GetLeagueByID(input.LeagueID)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to fetch league %s: %v", input.LeagueID, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrLeagueNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve league data", common.ErrInternalService)
	}

	user, err := s.userRepo.GetUserByID(input.UserID)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to fetch user %s: %v", input.UserID, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: failed to retrieve user data", common.ErrInternalService)
	}

	// Fallback for InLeagueName and TeamName (always happens if empty)
	if input.InLeagueName == nil {
		input.InLeagueName = &user.DiscordUsername
	}
	if input.TeamName == nil {
		input.TeamName = &user.DiscordUsername
	}

	// --- UNIQUENESS CHECKS ---

	// a. Check for Existing Player (User can only be a player once per league)
	existingPlayerByUser, err := s.playerRepo.FindPlayerByUserAndLeague(input.UserID, input.LeagueID)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to check for existing player by user ID %s in league %s: %v", input.UserID, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to check existing player data", common.ErrInternalService)
	}
	if existingPlayerByUser != nil {
		return nil, common.ErrUserAlreadyInLeague
	}

	// b. Check for Unique InLeagueName within the League
	existingPlayerByName, err := s.playerRepo.FindPlayerByInLeagueNameAndLeagueID(*input.InLeagueName, input.LeagueID)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to check for existing player by in-league name '%s' in league %s: %v", *input.InLeagueName, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to check in-league name uniqueness", common.ErrInternalService)
	}
	if existingPlayerByName != nil {
		return nil, fmt.Errorf("%w: '%s'", common.ErrInLeagueNameTaken, *input.InLeagueName)
	}

	// c. Check for Unique TeamName within the League
	existingPlayerByTeamName, err := s.playerRepo.FindPlayerByTeamNameAndLeagueID(*input.TeamName, input.LeagueID)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to check for existing player by team name '%s' in league %s: %v", *input.TeamName, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to check team name uniqueness", common.ErrInternalService)
	}
	if existingPlayerByTeamName != nil {
		return nil, fmt.Errorf("%w: '%s'", common.ErrTeamNameTaken, *input.TeamName)
	}
	// --- END UNIQUENESS CHECKS ---

	// initialize player model
	player := models.Player{
		UserID:       input.UserID,
		LeagueID:     input.LeagueID,
		InLeagueName: *input.InLeagueName,
		TeamName:     *input.TeamName,
		DraftPoints:  int(league.StartingDraftPoints),

		// derived values
		Wins:   0,
		Losses: 0,
		Role:   rbac.Member, // Default role for new players
	}

	createdPlayer, err := s.playerRepo.CreatePlayer(&player)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to create player for user %s in league %s: %v", input.UserID, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to add player to league", common.ErrFailedToCreatePlayer)
	}

	log.Printf("Service: CreatePlayerHandler - Player %s created for user %s in league %s.", createdPlayer.ID, input.UserID, input.LeagueID)
	return createdPlayer, nil
}

func (s *playerServiceImpl) GetPlayerByIDHandler(playerID uuid.UUID, currentUser *models.User) (*models.Player, error) {
	// controller needs to ensure currentUser is not nil

	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: GetPlayerByIDHandler - Player %s not found: %v", playerID, err)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: GetPlayerByIDHandler - Failed to retrieve player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	// Authorization checks
	if currentUser.Role == "admin" {
		log.Printf("Service: GetPlayerByIDHandler - Skipped authorization checks for admin user %s.", currentUser.ID)
		return player, nil
	}

	if player.UserID == currentUser.ID { // User is viewing their own profile.
		log.Printf("Service: GetPlayerByIDHandler - User %s viewing their own player profile %s.", currentUser.ID, playerID)
		return player, nil
	}

	isCurrentUserInLeague, err := s.leagueRepo.IsUserPlayerInLeague(currentUser.ID, player.LeagueID)
	if err != nil {
		log.Printf("Service: GetPlayerByIDHandler - Error checking current user %s's league membership in league %s: %v", currentUser.ID, player.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to perform league membership authorization check", common.ErrInternalService)
	}

	if isCurrentUserInLeague { // if currentUser is a player in the same league as the player being viewed, allow access.
		log.Printf("Service: GetPlayerByIDHandler - User %s is a player in the same league as player %s. Access granted.", currentUser.ID, playerID)
		return player, nil
	}

	log.Printf("Service: GetPlayerByIDHandler - Unauthorized access attempt by user %s to player %s.", currentUser.ID, playerID)
	return nil, common.ErrUnauthorized
}

func (s *playerServiceImpl) GetPlayersByLeagueHandler(leagueID, userID uuid.UUID, isUserAnAdmin bool) ([]models.Player, error) {
	players, err := s.playerRepo.GetPlayersByLeague(leagueID)
	if err != nil {
		log.Printf("Service: GetPlayersByLeagueHandler - Failed to retrieve players for league %s: %v", leagueID, err)
		return nil, fmt.Errorf("%w: failed to retrieve players data", common.ErrInternalService)
	}

	if isUserAnAdmin {
		log.Printf("Service: GetPlayersByLeagueHandler - Skipped authorization checks for admin user %s.", userID)
		return players, nil
	}

	// check if the currentUser (not admin) is a player in that league
	isCurrentUserInLeague, err := s.leagueRepo.IsUserPlayerInLeague(userID, leagueID)
	if err != nil {
		log.Printf("Service: GetPlayersByLeagueHandler - Error checking current user %s's league membership in league %s: %v", userID, leagueID, err)
		return nil, fmt.Errorf("%w: failed to perform league membership authorization check", common.ErrInternalService)
	}

	if isCurrentUserInLeague {
		log.Printf("Service: GetPlayersByLeagueHandler - User %s is a player in league %s. Access granted.", userID, leagueID)
		return players, nil
	}

	log.Printf("Service: GetPlayersByLeagueHandler - Unauthorized access attempt by user %s to league %s players.", userID, leagueID)
	return nil, common.ErrUnauthorized
}

func (s *playerServiceImpl) GetPlayersByUserHandler(
	userID, currentUserID uuid.UUID,
	isCurrentUserAnAdmin bool,
) ([]models.Player, error) {

	players, err := s.playerRepo.GetPlayersByUser(userID)
	if err != nil {
		log.Printf("Service: GetPlayersByUserHandler - Failed to retrieve players for user %s: %v", userID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	// is currentUser an admin or is requesting their own player
	if isCurrentUserAnAdmin || currentUserID == userID {
		log.Printf("Service: GetPlayersByUserHandler - User %s (admin or self) accessing players for user %s.", currentUserID, userID)
		return players, nil
	}

	log.Printf("Service: GetPlayersByUserHandler - Unauthorized access attempt by user %s to view players for user %s.", currentUserID, userID)
	return nil, common.ErrUnauthorized
}

func (s *playerServiceImpl) GetPlayerWithFullRosterHandler(playerID, currentUserID uuid.UUID, isCurrentUserAnAdmin bool) (*models.Player, error) {

	player, err := s.playerRepo.GetPlayerWithFullRoster(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: GetPlayerWithFullRosterHandler - Player %s not found: %v", playerID, err)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: GetPlayerWithFullRosterHandler - Failed to retrieve player %s with full roster: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	if isCurrentUserAnAdmin {
		log.Printf("Service: GetPlayerWithFullRosterHandler - Skipped authorization for admin user %s.", currentUserID)
		return player, nil
	}
	if player.UserID == currentUserID {
		log.Printf("Service: GetPlayerWithFullRosterHandler - User %s viewing their own player roster for player %s.", currentUserID, playerID)
		return player, nil // User is viewing their own player profile.
	}

	isCurrentUserInLeague, err := s.leagueRepo.IsUserPlayerInLeague(currentUserID, player.LeagueID)
	if err != nil {
		log.Printf("Service: GetPlayerWithFullRosterHandler - Error checking current user %s's league membership in league %s: %v", currentUserID, player.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to perform league membership authorization check", common.ErrInternalService)
	}

	if isCurrentUserInLeague { // if currentUser is a player in the same league as the player being viewed, allow access.
		log.Printf("Service: GetPlayerWithFullRosterHandler - User %s is a player in the same league as player %s. Access granted to roster.", currentUserID, playerID)
		return player, nil
	}

	log.Printf("Service: GetPlayerWithFullRosterHandler - Unauthorized access attempt by user %s to player %s's roster.", currentUserID, playerID)
	return nil, common.ErrUnauthorized
}

func (s *playerServiceImpl) UpdatePlayerProfile(
	currentUser *models.User,
	playerID uuid.UUID,
	inLeagueName *string,
	teamName *string,
) (*models.Player, error) {
	// fetch existing user
	existingPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: UpdatePlayerProfile - Player %s not found.", playerID)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: UpdatePlayerProfile - Failed to fetch player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player for update", common.ErrInternalService)
	}

	// Authorization: Admin, or the player themselves, or a LeagueOwner/Moderator
	if currentUser.Role != "admin" && currentUser.ID != existingPlayer.UserID {
		// If not admin and not updating self, check if they are a LeagueOwner or Moderator
		requesterPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, existingPlayer.LeagueID)
		if err != nil {
			log.Printf("Service: UpdatePlayerProfile - Failed to get requester player for auth: %v", err)
			return nil, common.ErrInternalService
		}
		if requesterPlayer == nil || (!requesterPlayer.IsLeagueOwner() && !requesterPlayer.IsLeagueModerator()) {
			log.Printf("Service: UpdatePlayerProfile - Unauthorized access attempt by user %s to update player %s's profile.", currentUser.ID, playerID)
			return nil, common.ErrUnauthorized
		}
	}

	// Apply updates selectively and perform business validation
	updated := false
	if inLeagueName != nil && *inLeagueName != existingPlayer.InLeagueName {
		if *inLeagueName != "" {
			existing, err := s.playerRepo.FindPlayerByInLeagueNameAndLeagueID(*inLeagueName, existingPlayer.LeagueID)
			if err != nil {
				log.Printf("Service: UpdatePlayerProfile - DB error checking in-league name uniqueness: %v", err)
				return nil, fmt.Errorf("%w: failed to check in-league name uniqueness", common.ErrInternalService)
			}
			if existing != nil && existing.ID != existingPlayer.ID {
				return nil, fmt.Errorf("%w: '%s'", common.ErrInLeagueNameTaken, *inLeagueName)
			}
		}
		existingPlayer.InLeagueName = *inLeagueName
		updated = true
	}

	if teamName != nil && *teamName != existingPlayer.TeamName {
		if *teamName != "" {
			existing, err := s.playerRepo.FindPlayerByTeamNameAndLeagueID(*teamName, existingPlayer.LeagueID)
			if err != nil {
				log.Printf("Service: UpdatePlayerProfile - DB error checking team name uniqueness: %v", err)
				return nil, fmt.Errorf("%w: failed to check team name uniqueness", common.ErrInternalService)
			}
			if existing != nil && existing.ID != existingPlayer.ID {
				return nil, fmt.Errorf("%w: '%s'", common.ErrTeamNameTaken, *teamName)
			}
		}
		existingPlayer.TeamName = *teamName
		updated = true
	}

	// Only call update if changes were made
	if !updated {
		return existingPlayer, nil // No changes, return existing player without DB call
	}

	updatedPlayer, err := s.playerRepo.UpdatePlayer(existingPlayer)
	if err != nil {
		log.Printf("Service: UpdatePlayerProfile - Failed to save updated player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to save player profile updates", common.ErrInternalService)
	}

	log.Printf("Service: UpdatePlayerProfile - Player %s profile updated by user %s.", playerID, currentUser.ID)
	return updatedPlayer, nil
}

// UpdatePlayerDraftPoints allows a LeagueOwner/Moderator to update a player's draft points.
func (s *playerServiceImpl) UpdatePlayerDraftPoints(
	currentUser *models.User,
	playerID uuid.UUID,
	draftPoints *int,
) (*models.Player, error) {

	// Fetch the player to verify league context and authorization
	existingPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: UpdatePlayerDraftPoints - Player %s not found.", playerID)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: UpdatePlayerDraftPoints - Failed to fetch player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player for draft points update", common.ErrInternalService)
	}

	// Authorization: Only Admin or LeagueOwner/Moderator can update draft points
	if currentUser.Role != "admin" {
		requesterPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, existingPlayer.LeagueID)
		if err != nil {
			log.Printf("Service: UpdatePlayerDraftPoints - Failed to get requester player for auth: %v", err)
			return nil, common.ErrInternalService
		}
		if requesterPlayer == nil || (!requesterPlayer.IsLeagueOwner() && !requesterPlayer.IsLeagueModerator()) {
			log.Printf("Service: UpdatePlayerDraftPoints - Unauthorized attempt by user %s to update player %s's draft points.", currentUser.ID, playerID)
			return nil, common.ErrUnauthorized
		}
	}

	if draftPoints == nil {
		log.Printf("Service: UpdatePlayerDraftPoints - request draft points is somehow nil (should be impossible in the service layer)")
		return nil, common.ErrInternalService
	}

	// Perform update using the specific repository method
	err = s.playerRepo.UpdatePlayerDraftPoints(playerID, *draftPoints)
	if err != nil {
		log.Printf("Service: UpdatePlayerDraftPoints - Failed to update player %s draft points: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to update player draft points", common.ErrInternalService)
	}

	// Fetch the updated player to return
	updatedPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		log.Printf("Service: UpdatePlayerDraftPoints - Failed to re-fetch player %s after update: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to re-fetch updated player", common.ErrInternalService)
	}

	log.Printf("Service: UpdatePlayerDraftPoints - Player %s draft points updated to %d by user %s.", playerID, *draftPoints, currentUser.ID)
	return updatedPlayer, nil
}

// UpdatePlayerRecord allows a LeagueOwner/Moderator to update a player's win/loss record.
func (s *playerServiceImpl) UpdatePlayerRecord(
	currentUser *models.User,
	playerID uuid.UUID,
	wins int,
	losses int,
) (*models.Player, error) {
	// Fetch the player to verify league context and authorization
	existingPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: UpdatePlayerRecord - Player %s not found.", playerID)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: UpdatePlayerRecord - Failed to fetch player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player for record update", common.ErrInternalService)
	}

	// Authorization: Only Admin or LeagueOwner/Moderator can update win/loss record
	if currentUser.Role != "admin" {
		requesterPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, existingPlayer.LeagueID)
		if err != nil {
			log.Printf("Service: UpdatePlayerRecord - Failed to get requester player for auth: %v", err)
			return nil, common.ErrInternalService
		}
		if requesterPlayer == nil || (!requesterPlayer.IsLeagueOwner() && !requesterPlayer.IsLeagueModerator()) {
			log.Printf("Service: UpdatePlayerRecord - Unauthorized attempt by user %s to update player %s's record.", currentUser.ID, playerID)
			return nil, common.ErrUnauthorized
		}
	}

	// Perform update using the specific repository method
	err = s.playerRepo.UpdatePlayerRecord(playerID, wins, losses)
	if err != nil {
		log.Printf("Service: UpdatePlayerRecord - Failed to update player %s record: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to update player record", common.ErrInternalService)
	}

	// Fetch the updated player to return
	updatedPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		log.Printf("Service: UpdatePlayerRecord - Failed to re-fetch player %s after update: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to re-fetch updated player", common.ErrInternalService)
	}

	log.Printf("Service: UpdatePlayerRecord - Player %s record updated to W%d-L%d by user %s.", playerID, wins, losses, currentUser.ID)
	return updatedPlayer, nil
}

// UpdatePlayerDraftPosition allows a LeagueOwner/Moderator to update a player's draft position.
func (s *playerServiceImpl) UpdatePlayerDraftPosition(
	currentUser *models.User,
	playerID uuid.UUID,
	draftPosition int,
) (*models.Player, error) {
	// Fetch the player to verify league context and authorization
	existingPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: UpdatePlayerDraftPosition - Player %s not found.", playerID)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: UpdatePlayerDraftPosition - Failed to fetch player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player for draft position update", common.ErrInternalService)
	}

	// Authorization: Only Admin or LeagueOwner/Moderator can update draft position
	if currentUser.Role != "admin" {
		requesterPlayer, err := s.playerRepo.GetPlayerByUserAndLeague(currentUser.ID, existingPlayer.LeagueID)
		if err != nil {
			log.Printf("Service: UpdatePlayerDraftPosition - Failed to get requester player for auth: %v", err)
			return nil, common.ErrInternalService
		}
		if requesterPlayer == nil || (!requesterPlayer.IsLeagueOwner() && !requesterPlayer.IsLeagueModerator()) {
			log.Printf("Service: UpdatePlayerDraftPosition - Unauthorized attempt by user %s to update player %s's draft position.", currentUser.ID, playerID)
			return nil, common.ErrUnauthorized
		}
	}

	// Perform update using the specific repository method
	err = s.playerRepo.UpdatePlayerDraftPosition(playerID, draftPosition)
	if err != nil {
		log.Printf("Service: UpdatePlayerDraftPosition - Failed to update player %s draft position: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to update player draft position", common.ErrInternalService)
	}

	// Fetch the updated player to return
	updatedPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		log.Printf("Service: UpdatePlayerDraftPosition - Failed to re-fetch player %s after update: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to re-fetch updated player", common.ErrInternalService)
	}

	log.Printf("Service: UpdatePlayerDraftPosition - Player %s draft position updated to %d by user %s.", playerID, draftPosition, currentUser.ID)
	return updatedPlayer, nil
}

// NOTE: cannot implement this at this moment due to limitations in our league model. cba so saving for a future refactor
// future work is to have better RBAC inside of a league, which would allow this to be properly implemented
// TODO: define this function this when RBAC has been properly implemented.
// func (s *playerServiceImpl) SetCommissionerStatus(playerID uuid.UUID, isCommissioner bool, currentUser *models.User) error

// not implemented for initial use case
// TODO: do this at some point
// func (s *playerServiceImpl) LeaveLeague(playerID uuid.UUID) error
