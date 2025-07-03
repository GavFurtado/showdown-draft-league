package controllers

import (
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handles player-related HTTP requests.
type PlayerController interface {
	// POST to join a league (creates a new player entity); Expects a payload
	JoinLeague(ctx *gin.Context)
	// GET a player by their ID
	GetPlayerByID(ctx *gin.Context)
	// Get all players in a league with leagueID
	GetPlayersByLeague(ctx *gin.Context)
	// Get all players associated with a specific user
	GetPlayersByUser(ctx *gin.Context)
	// Get a player with their full roster
	GetPlayerWithFullRoster(ctx *gin.Context)
	// Update a player's profile; Expects a payload
	UpdatePlayerProfile(ctx *gin.Context)
}

type playerControllerImpl struct {
	playerService services.PlayerService
}

func NewPlayerController(playerService services.PlayerService) *playerControllerImpl {
	return &playerControllerImpl{
		playerService: playerService,
	}
}

// POST /api/league/:id/join (id being leagueID)
// Creates a player for the league :id, essentially joining the league
func (c *playerControllerImpl) JoinLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("PlayerController: JoinLeague - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	var req common.PlayerCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("PlayerController: JoinLeague - bad request\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: needs to be more complex for some kind of send JoinRequest thing in the future

	// service layer needs to do this checking
	if !currentUser.IsAdmin && req.UserID != currentUser.ID {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Cannot perform this transaction"})
	}

	if req.UserID == uuid.Nil || req.LeagueID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bad or Malformed Request"})
		return
	}

	log.Printf("PlayerController: JoinLeague - received join player request")
	player, err := c.playerService.CreatePlayerHandler(&req)
	if err != nil {
		log.Printf("PlayerController: JoinLeague - Service Method returned an error")
		switch err {
		case common.ErrUserNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrUserAlreadyInLeague, common.ErrInLeagueNameTaken, common.ErrTeamNameTaken:
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case common.ErrInternalService, common.ErrFailedToCreatePlayer:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		default:
			// fallback in case the error is unrecognized
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// POST /api/players/:id
func (c *playerControllerImpl) GetPlayerByID(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("PlayerController: GetPlayerByID - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	playerIDstr := ctx.Param("id")
	playerID, err := uuid.Parse(playerIDstr)
	if err != nil { // if the str was "" (which btw idk how that happens) it's still handled here
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	player, err := c.playerService.GetPlayerByIDHandler(playerID, currentUser)
	if err != nil {
		log.Printf("PlayerController: GetPlayerByID - Error occured in the Service Method")
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// GET /api/:id/players
// GET all players in a league :leagueId
func (c *playerControllerImpl) GetPlayersByLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("PlayerController: GetPlayersByLeague - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	// get param
	leagueIDStr := ctx.Param("id")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil { // if the str was "" (which btw idk how that happens) it's still handled here
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if leagueID == uuid.Nil || currentUser.ID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
	}

	players, err := c.playerService.GetPlayersByLeagueHandler(leagueID, currentUser.ID, currentUser.IsAdmin)
	if err != nil {
		log.Printf("PlayerController: GetPlayersByLeague - Error occured in the Service Method")
		switch err {
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, players)
}

// (GET /users/:id/players)
// Get all players associated with a specific user
func (c playerControllerImpl) GetPlayersByUser(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("PlayerController: GetPlayersByUser - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return

	}

	// get param
	userIDstr := ctx.Param("id")
	userID, err := uuid.Parse(userIDstr)
	if err != nil { // if the str was "" (which btw idk how that happens) it's still handled here
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	players, err := c.playerService.GetPlayersByUserHandler(userID, currentUser.ID, currentUser.IsAdmin)
	if err != nil {
		log.Printf("PlayerController: GetPlayersByUser - Error occured in the Service Method")
		switch err {
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return

	}

	ctx.JSON(http.StatusOK, players)
}

// GET /players/:id/roster
// Get a player with their full roster
func (c *playerControllerImpl) GetPlayerWithFullRoster(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("PlayerController: GetPlayersWithFullRoster - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return

	}

	// get param
	playerIDStr := ctx.Param("id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil { // if the str was "" (which btw idk how that happens) it's still handled here
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	player, err := c.playerService.GetPlayerWithFullRosterHandler(playerID, currentUser.ID, currentUser.IsAdmin)
	if err != nil {
		log.Printf("PlayerController: GetPlayersWithFullRoster - Error occured in the Service Method")
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return

	}

	ctx.JSON(http.StatusOK, player)
}

// PUT /players/:id/profile
// Update a player's profile; Expects a payload
func (c *playerControllerImpl) UpdatePlayerProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("PlayerController: UpdatePlayerProfile - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrNoUserInContext.Error()})
		return

	}

	var req common.UpdatePlayerInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("PlayerController: UpdatePlayerProfile - bad request\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playerIDstr := ctx.Param("id")
	playerID, err := uuid.Parse(playerIDstr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	var player *models.Player

	// validate fields in the req (four different service methods update the profile of a player)
	// use local helpers for selective updating
	if req.TeamName != nil || req.InLeagueName != nil {
		player = c.updatePlayerProfile(ctx, currentUser, playerID, req.InLeagueName, req.TeamName)
	}

	// TO POINTER OR NOT TO POINTER???????????????????????????????
	// i shall find out soon when i test

	if req.DraftPoints != nil {
		player = c.updatePlayerDraftPoints(ctx, currentUser, playerID, req.DraftPoints)
	}

	if req.DraftPosition != nil {
		player = c.updatePlayerDraftPosition(ctx, currentUser, playerID, req.DraftPosition)
	}

	if req.Wins != nil || req.Losses != nil {
		player = c.updatePlayerRecord(ctx, currentUser, playerID, *req.Wins, *req.Losses)
	}

	ctx.JSON(http.StatusOK, player)
}

// -- Update Player Profile helper functions --
// returns no errors because the errors are managed here (yes i am making exception)
func (c *playerControllerImpl) updatePlayerProfile(ctx *gin.Context, currentUser *models.User, playerID uuid.UUID, inLeagueName *string, teamName *string) *models.Player {
	player, err := c.playerService.UpdatePlayerProfile(currentUser, playerID, inLeagueName, teamName)
	if err != nil {
		// yes this is not the right method technically but it basically is
		log.Printf("PlayerController: UpdatePlayerProfile - error occured in the Service Method")
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrTeamNameTaken:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrTeamNameTaken:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return nil
	}
	return player
}

func (c *playerControllerImpl) updatePlayerDraftPoints(ctx *gin.Context, currentUser *models.User, playerID uuid.UUID, draftPoints *int) *models.Player {
	player, err := c.playerService.UpdatePlayerDraftPoints(currentUser, playerID, draftPoints)
	if err != nil {
		// yes this is not the right method technically but it basically is
		log.Printf("PlayerController: UpdatePlayerProfile - error occured in the Service Method")
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return nil
	}
	return player

}

func (c *playerControllerImpl) updatePlayerDraftPosition(ctx *gin.Context, currentUser *models.User, playerID uuid.UUID, draftPosition *int) *models.Player {

	player, err := c.playerService.UpdatePlayerDraftPosition(currentUser, playerID, draftPosition)
	if err != nil {
		// yes this is not the right method technically but it basically is
		log.Printf("PlayerController: UpdatePlayerProfile - error occured in the Service Method")
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return nil
	}
	return player
}

func (c *playerControllerImpl) updatePlayerRecord(ctx *gin.Context, currentUser *models.User, playerID uuid.UUID, wins int, losses int) *models.Player {
	player, err := c.playerService.UpdatePlayerRecord(currentUser, playerID, wins, losses)
	if err != nil {
		// yes this is not the right method technically but it basically is
		log.Printf("PlayerController: UpdatePlayerProfile - error occured in the Service Method")
		switch err {
		case common.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case common.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case common.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			// fallback
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unxpected error"})
		}
		return nil
	}
	return player
}
