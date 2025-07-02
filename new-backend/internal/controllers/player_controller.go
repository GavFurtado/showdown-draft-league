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
	"golang.org/x/text/cases"
)

// handles player-related HTTP requests.
type PlayerController interface {
	// POST to join a league (creates a new player entity)
	JoinLeague(ctx *gin.Context)
	// GET a player by their ID
	GetPlayerByID(ctx *gin.Context)
}

type playerControllerImpl struct {
	playerService services.PlayerService
}

func NewPlayerController(playerService services.PlayerService) *playerControllerImpl {
	return &playerControllerImpl{
		playerService: playerService,
	}
}

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
