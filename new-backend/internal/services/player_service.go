package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerService interface {
	CreatePlayerHandler(input *common.PlayerCreateRequest) (*models.Player, error)

	GetPlayerByIDHandler(playerID uuid.UUID, currentUser *models.User) (*models.Player, error)
	GetPlayerByUserIDAndLeagueID(userID uuid.UUID, leagueID uuid.UUID) (*models.Player, error)
	GetPlayersByLeagueHandler(leagueID, userID uuid.UUID) ([]models.Player, error)
	GetPlayersByUserHandler(userID, currentUserID uuid.UUID) ([]models.Player, error)
	GetPlayerWithFullRosterHandler(playerID, currentUserID uuid.UUID) (*models.Player, error)

	UpdatePlayerProfile(currentUser *models.User, playerID uuid.UUID, inLeagueName *string, teamName *string) (*models.Player, error)
	UpdatePlayerDraftPoints(currentUser *models.User, playerID uuid.UUID, draftPoints *int) (*models.Player, error)
	UpdatePlayerRecord(currentUser *models.User, playerID uuid.UUID, wins int, losses int) (*models.Player, error)
	UpdatePlayerDraftPosition(currentUser *models.User, playerID uuid.UUID, draftPosition int) (*models.Player, error)
	UpdatePlayerRole(currentUserID, playerID uuid.UUID, newPlayerRole rbac.PlayerRole) (*models.Player, error)
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

	if league.Status != enums.LeagueStatusSetup {
		log.Printf("Service: CreatePlayerHandler - League %s is not in SETUP status to add players: %v", input.LeagueID, err)
		return nil, common.ErrInvalidState
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
		username := fmt.Sprintf("%s's Team", user.DiscordUsername)
		input.TeamName = &username
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
		Wins:          0,
		Losses:        0,
		DraftPosition: 0,
		GroupNumber:   league.NewPlayerGroupNumber,
		SkipsLeft:     league.MaxPokemonPerPlayer - league.MinPokemonPerPlayer,
		Role:          rbac.PRoleMember,
	}

	createdPlayer, err := s.playerRepo.CreatePlayer(&player)
	if err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to create player for user %s in league %s: %v", input.UserID, input.LeagueID, err)
		return nil, fmt.Errorf("%w: failed to add player to league", common.ErrFailedToCreatePlayer)
	}

	league.PlayerCount++
	league.NewPlayerGroupNumber = ((league.NewPlayerGroupNumber + 1) % league.Format.GroupCount) + 1 // +1 due to 1-based GroupNumbers
	if _, err = s.leagueRepo.UpdateLeague(league); err != nil {
		log.Printf("Service: CreatePlayerHandler - Failed to update league %s after creating player for %s: %v", league.ID, input.UserID, err)
		return nil, common.ErrInternalService
	}

	log.Printf("Service: CreatePlayerHandler - Player %s created for user %s in league %s.", createdPlayer.ID, input.UserID, input.LeagueID)
	return createdPlayer, nil
}

func (s *playerServiceImpl) GetPlayerByIDHandler(playerID uuid.UUID, currentUser *models.User) (*models.Player, error) {

	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: GetPlayerByIDHandler - Player %s not found: %v", playerID, err)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: GetPlayerByIDHandler - Failed to retrieve player %s: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	return player, nil
}

func (s *playerServiceImpl) GetPlayerByUserIDAndLeagueID(userID, leagueID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetPlayerByUserAndLeague(userID, leagueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("ERROR: (Service: GetPlayerByUserIDAndLeagueID) - Player (userID %s; leagueID %s) not found: %v", userID, leagueID, err)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("ERROR: (Service: GetPlayerByUserIDAndLeagueID) - Failed to retrieve player with userID %s and leagueID %s: %v", userID, leagueID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	return player, nil
}

func (s *playerServiceImpl) GetPlayersByLeagueHandler(leagueID, userID uuid.UUID) ([]models.Player, error) {
	players, err := s.playerRepo.GetPlayersByLeague(leagueID)
	if err != nil {
		log.Printf("Service: GetPlayersByLeagueHandler - Failed to retrieve players for league %s: %v", leagueID, err)
		return nil, fmt.Errorf("%w: failed to retrieve players data", common.ErrInternalService)
	}

	return players, nil
}

func (s *playerServiceImpl) GetPlayersByUserHandler(
	userID, currentUserID uuid.UUID,
) ([]models.Player, error) {

	players, err := s.playerRepo.GetPlayersByUser(userID)
	if err != nil {
		log.Printf("Service: GetPlayersByUserHandler - Failed to retrieve players for user %s: %v", userID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	return players, nil
}

func (s *playerServiceImpl) GetPlayerWithFullRosterHandler(playerID, currentUserID uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetPlayerWithFullRoster(playerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Service: GetPlayerWithFullRosterHandler - Player %s not found: %v", playerID, err)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: GetPlayerWithFullRosterHandler - Failed to retrieve player %s with full roster: %v", playerID, err)
		return nil, fmt.Errorf("%w: failed to retrieve player data", common.ErrInternalService)
	}

	return player, nil
}

// fails tests.. needs urgent rework
// TODO: fix this
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

	// TODO: with the new rbac middleware there's extra checks here that i'm not touching because
	// intended working: user can update their own player profile. Admins or League Moderators and Owners can update anyone's
	// other players/users cannot update another player's profile.
	// fix at some point

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
// Intended for manual updates (like an override). It sets the points.
// fails tests.. needs urgent rework
// TODO: fix this
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

	// TODO: not touching authorization here for now. please come back to this

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
// fails tests.. needs urgent rework
// TODO: fix this
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

	// TODO: not touching authorization for this rn. come back to this

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

	// TODO: not touching this rn. fix later

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

func (s *playerServiceImpl) UpdatePlayerRole(currentUserID, playerID uuid.UUID, newPlayerRole rbac.PlayerRole) (*models.Player, error) {
	err := s.playerRepo.UpdatePlayerRole(playerID, newPlayerRole)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Service: UpdatePlayerRole - player %s not found to update role (Requesting user: %s): %v\n", playerID, currentUserID, err)
			return nil, common.ErrPlayerNotFound
		}
		log.Printf("Service: UpdatePlayerRole - Failed to update role (Requesting user: %s) for player %s: %v\n", currentUserID, playerID, err)
		return nil, common.ErrInternalService
	}

	updatedPlayer, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		// should be impossible for this to be a record not found again
		log.Printf("Service: UpdatePlayerDraftPosition - Failed to re-fetch (Requesting user: %s) player %s after update: %v\n", currentUserID, playerID, err)
		return nil, common.ErrInternalService
	}

	return updatedPlayer, nil
}

// not implemented for initial use case
// TODO: do this at some point
// func (s *playerServiceImpl) LeaveLeague(playerID uuid.UUID) error
