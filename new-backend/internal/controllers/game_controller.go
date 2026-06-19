package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameController interface {
	ReportGame(ctx *gin.Context)
	FinalizeGame(ctx *gin.Context)
	GetGameByID(ctx *gin.Context)
	GetGamesByLeague(ctx *gin.Context)
	GetGamesByPlayer(ctx *gin.Context)
	StartRegularSeason(ctx *gin.Context)
	GeneratePlayoffBracket(ctx *gin.Context)
}

type gameControllerImpl struct {
	gameService   services.GameService
	leagueService services.LeagueService
}

func NewGameController(
	gameService services.GameService,
	leagueService services.LeagueService,
) GameController {
	return &gameControllerImpl{
		gameService:   gameService,
		leagueService: leagueService,
	}

}

func (c *gameControllerImpl) GetGameByID(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("gameId"))
	if err != nil {
		log.Printf("ERROR: (Controller: GetGameByID) - Error parsing gameId param: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
	}
	leagueID, _ := uuid.Parse(ctx.Param("leagueId"))

	var game *models.Game
	if game, err = c.gameService.GetGameByID(gameID); err != nil {
		log.Printf("ERROR: (Controller: GetGameByID) - Error fetching game (League %s) by ID %s: %v", leagueID, gameID, err)
		switch {
		case errors.Is(err, common.ErrGameNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": common.ErrGameNotFound.Error()})
		case errors.Is(err, common.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"game": game})
}

func (c *gameControllerImpl) GetGamesByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		log.Printf("ERROR: (Controller: GetGamesByLeague) - Error parsing leagueId param: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
	}

	var games []models.Game
	if games, err = c.gameService.GetGamesByLeague(leagueID); err != nil {
		log.Printf("ERROR: (Controller: GetGamesByLeague) - Error fetching Games for League %s : %v", leagueID, err)
		switch {
		case errors.Is(err, common.ErrGameNotFound):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Games not found for league"})
		case errors.Is(err, common.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"games": games})
}

func (c *gameControllerImpl) GetGamesByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		log.Printf("ERROR: (Controller: GetGamesByPlayer) - Error parsing playerId param: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
	}
	leagueID, _ := uuid.Parse(ctx.Param("leagueId"))

	var games []models.Game
	if games, err = c.gameService.GetGamesByPlayer(playerID); err != nil {
		log.Printf("ERROR: (Controller: GetGameByID) - Error fetching Games for League %s : %v", leagueID, err)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Player/Games not found for league"})
		case errors.Is(err, common.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"games": games})
}

// ReportGame handles a player reporting a game result.
func (c *gameControllerImpl) ReportGame(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("gameId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	var dto common.ReportGameDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		log.Printf("ERROR: (Controller: ReportGame): Error binding request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request payload: %v", err)})
		return
	}

	// reporterID is the current player
	// arguably unecessary but idgaf
	reporterIDStr, exists := ctx.Get("playerID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Player ID not found in context"})
		return
	}
	dto.ReporterID = reporterIDStr.(uuid.UUID)

	if err := c.gameService.ReportGameResult(gameID, &dto); err != nil {
		log.Printf("ERROR: (Controller: ReportGame) - %s\n", err.Error())
		switch {
		case errors.Is(err, common.ErrConflict):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "This game is either already finalized or is pending approval"})
		case errors.Is(err, common.ErrInvalidInput), errors.Is(err, common.ErrUnauthorized):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrGameNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to report game result: %v", err)})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Game result reported successfully for approval"})
}

// FinalizeGame handles league staff finalizing a game result (approve, submit, or retroactively edit).
func (c *gameControllerImpl) FinalizeGame(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("gameId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID format"})
		return
	}

	var dto common.FinalizeGameDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request payload: %v", err)})
		return
	}

	finalizerIDStr, exists := ctx.Get("playerID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Player ID not found in context"})
		return
	}
	finalizerID := finalizerIDStr.(uuid.UUID)
	dto.FinalizerID = finalizerID

	if err := c.gameService.FinalizeGameResult(gameID, &dto); err != nil {
		log.Printf("ERROR: (Controller: FinalizeGameResult) - %s\n", err.Error())
		switch {
		case errors.Is(err, common.ErrConflict):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Game status not valid to Finalize"})
		case errors.Is(err, common.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrGameNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to finalize game result: %v", err)})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Game result finalized successfully"})
}

func (c *gameControllerImpl) StartRegularSeason(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		log.Printf("ERROR: (Controller: StartRegularSeason) - Error parsing leagueId param: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if err := c.leagueService.StartRegularSeason(leagueID); err != nil {
		log.Printf("ERROR: (Controller: StartRegularSeason) - Error starting regular season for League %s : %v", leagueID, err)
		switch {
		case errors.Is(err, common.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrLeagueNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
		case errors.Is(err, common.ErrInvalidState):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrGamesAlreadyGenerated):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start regular season: %v", err)})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Regular season started successfully"})
}

func (c *gameControllerImpl) GeneratePlayoffBracket(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		log.Printf("ERROR: (Controller: GeneratePlayoffBracket) - Error parsing leagueId param: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	if err := c.gameService.GeneratePlayoffBracket(leagueID); err != nil {
		log.Printf("ERROR: (Controller: GeneratePlayoffBracket) - Error generating playoff bracket for League %s : %v", leagueID, err)
		switch {
		case errors.Is(err, common.ErrUnauthorized):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, common.ErrLeagueNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
		case errors.Is(err, common.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate playoff bracket: %v", err)})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Playoff bracket generated successfully"})
}
