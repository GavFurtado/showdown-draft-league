package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameController interface {
	ReportGame(c *gin.Context)
	FinalizeGame(c *gin.Context)
	GetGameByID(c *gin.Context)
	GetGamesByLeague(c *gin.Context)
	GetGamesByPlayer(c *gin.Context)
	GenerateRegularSeasonGames(c *gin.Context)
	GeneratePlayoffBracket(c *gin.Context)
}
type gameControllerImpl struct {
	gameService services.GameService
	rbacService services.RBACService
}

func NewGameController(gameService services.GameService, rbacService services.RBACService) GameController {
	return &gameControllerImpl{
		gameService: gameService,
		rbacService: rbacService,
	}
}

// ReportGame handles a player reporting a game result.
func (ctrl *gameControllerImpl) ReportGame(c *gin.Context) {
	gameID, err := uuid.Parse(c.Param("gameId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
		return
	}

	var dto common.ReportGameDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request payload: %v", err)})
		return
	}

	// reporterID is the current player
	// arguably unecessary but idgaf
	reporterIDStr, exists := c.Get("playerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player ID not found in context"})
		return
	}
	dto.ReporterID = reporterIDStr.(uuid.UUID)

	if err := ctrl.gameService.ReportGameResult(gameID, &dto); err != nil {
		switch {
		case errors.Is(err, common.ErrInvalidInput), errors.Is(err, common.ErrUnauthorized):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrGameNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to report game result: %v", err)})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game result reported successfully for approval"})
}

// FinalizeGame handles league staff finalizing a game result (approve, submit, or retroactively edit).
func (ctrl *gameControllerImpl) FinalizeGame(c *gin.Context) {
	gameID, err := uuid.Parse(c.Param("gameId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
		return
	}

	var dto common.FinalizeGameDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request payload: %v", err)})
		return
	}

	finalizerIDStr, exists := c.Get("playerID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player ID not found in context"})
		return
	}
	finalizerID := finalizerIDStr.(uuid.UUID)
	dto.FinalizerID = finalizerID

	if err := ctrl.gameService.FinalizeGameResult(gameID, &dto); err != nil {
		switch {
		case errors.Is(err, common.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrGameNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to finalize game result: %v", err)})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game result finalized successfully"})
}
