package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
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
	logger        *slog.Logger
	gameService   services.GameService
	leagueService services.LeagueService
}

func NewGameController(
	logger *slog.Logger,
	gameService services.GameService,
	leagueService services.LeagueService,
) GameController {
	return &gameControllerImpl{
		logger:        utils.LoggerWithService(logger, "GameController"),
		gameService:   gameService,
		leagueService: leagueService,
	}
}

// GetGameByID godoc
//
//	@Summary		Get a game
//	@Description	Game details
//	@Tags			Games
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			gameId		path		string	true	"Game ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games/{gameId} [get]
func (c *gameControllerImpl) GetGameByID(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("gameId"))
	if err != nil {
		c.logger.Error("Error parsing gameId param", "error", err)
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}
	leagueID, _ := uuid.Parse(ctx.Param("leagueId"))

	var game *models.Game
	if game, err = c.gameService.GetGameByID(gameID); err != nil {
		c.logger.Error("Error fetching game by ID", "league_id", leagueID, "game_id", gameID, "error", err)
		switch {
		case errors.Is(err, types.ErrGameNotFound):
			sendError(ctx, http.StatusNotFound, types.ErrGameNotFound.Error())
		case errors.Is(err, types.ErrInternalService):
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"game": game})
}

// GetGamesByLeague godoc
//
//	@Summary		Get league games
//	@Description	All games in a league
//	@Tags			Games
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games [get]
func (c *gameControllerImpl) GetGamesByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		c.logger.Error("Error parsing leagueId param", "error", err)
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var games []models.Game
	if games, err = c.gameService.GetGamesByLeague(leagueID); err != nil {
		c.logger.Error("Error fetching Games for League", "league_id", leagueID, "error", err)
		switch {
		case errors.Is(err, types.ErrGameNotFound):
			sendError(ctx, http.StatusBadRequest, "Games not found for league")
		case errors.Is(err, types.ErrInternalService):
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"games": games})
}

// GetGamesByPlayer godoc
//
//	@Summary		Get games by player
//	@Description	All games for a specific player
//	@Tags			Games
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			memberId	path		string	true	"Member ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games/members/{memberId} [get]
func (c *gameControllerImpl) GetGamesByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		c.logger.Error("Error parsing playerId param", "error", err)
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}
	leagueID, _ := uuid.Parse(ctx.Param("leagueId"))

	var games []models.Game
	if games, err = c.gameService.GetGamesByPlayer(playerID); err != nil {
		c.logger.Error("Error fetching Games for League", "league_id", leagueID, "error", err)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			sendError(ctx, http.StatusBadRequest, "Player/Games not found for league")
		case errors.Is(err, types.ErrInternalService):
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"games": games})
}

// ReportGame godoc
//
//	@Summary		Report a game result
//	@Description	Submit a game result for approval
//	@Tags			Games
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string							true	"League ID"
//	@Param			gameId		path		string							true	"Game ID"
//	@Param			request		body		requests.ReportGameRequestDTO	true	"Game result"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games/report/{gameId} [put]
func (c *gameControllerImpl) ReportGame(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("gameId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var dto requests.ReportGameRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		c.logger.Error("Error binding request", "error", err)
		sendError(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	reporterIDStr, exists := ctx.Get("playerID")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "Player ID not found in context")
		return
	}
	dto.ReporterID = reporterIDStr.(uuid.UUID)

	if err := c.gameService.ReportGameResult(gameID, &dto); err != nil {
		c.logger.Error("ReportGame error", "error", err)
		switch {
		case errors.Is(err, types.ErrConflict):
			sendError(ctx, http.StatusBadRequest, "This game is either already finalized or is pending approval")
		case errors.Is(err, types.ErrInvalidInput), errors.Is(err, types.ErrUnauthorized):
			sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.Is(err, types.ErrGameNotFound):
			sendError(ctx, http.StatusNotFound, "Game not found")
		default:
			sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to report game result: %v", err))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Game result reported successfully for approval"})
}

// FinalizeGame godoc
//
//	@Summary		Finalize a game result
//	@Description	Approve and finalize a reported game
//	@Tags			Games
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string							true	"League ID"
//	@Param			gameId		path		string							true	"Game ID"
//	@Param			request		body		requests.FinalizeGameRequestDTO	true	"Finalization data"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games/finalize/{gameId} [put]
func (c *gameControllerImpl) FinalizeGame(ctx *gin.Context) {
	gameID, err := uuid.Parse(ctx.Param("gameId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid game ID format")
		return
	}

	var dto requests.FinalizeGameRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		sendError(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	finalizerIDStr, exists := ctx.Get("playerID")
	if !exists {
		sendError(ctx, http.StatusUnauthorized, "Player ID not found in context")
		return
	}
	finalizerID := finalizerIDStr.(uuid.UUID)
	dto.FinalizerID = finalizerID

	if err := c.gameService.FinalizeGameResult(gameID, &dto); err != nil {
		c.logger.Error("FinalizeGameResult error", "error", err)
		switch {
		case errors.Is(err, types.ErrConflict):
			sendError(ctx, http.StatusBadRequest, "Game status not valid to Finalize")
		case errors.Is(err, types.ErrInvalidInput):
			sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.Is(err, types.ErrGameNotFound):
			sendError(ctx, http.StatusNotFound, "Game not found")
		default:
			sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to finalize game result: %v", err))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Game result finalized successfully"})
}

// StartRegularSeason godoc
//
//	@Summary		Start regular season
//	@Description	Generate the regular season schedule
//	@Tags			Games
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games/start-season [post]
func (c *gameControllerImpl) StartRegularSeason(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		c.logger.Error("Error parsing leagueId param", "error", err)
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err := c.leagueService.StartRegularSeason(leagueID); err != nil {
		c.logger.Error("Error starting regular season for League", "league_id", leagueID, "error", err)
		switch {
		case errors.Is(err, types.ErrUnauthorized):
			sendError(ctx, http.StatusUnauthorized, err.Error())
		case errors.Is(err, types.ErrLeagueNotFound):
			sendError(ctx, http.StatusNotFound, "League not found")
		case errors.Is(err, types.ErrInvalidState):
			sendError(ctx, http.StatusConflict, err.Error())
		case errors.Is(err, types.ErrInvalidInput):
			sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.Is(err, types.ErrGamesAlreadyGenerated):
			sendError(ctx, http.StatusConflict, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to start regular season: %v", err))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Regular season started successfully"})
}

// GeneratePlayoffBracket godoc
//
//	@Summary		Generate playoff bracket
//	@Description	Create the playoff bracket for a league
//	@Tags			Games
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/games/generate-playoffs [post]
func (c *gameControllerImpl) GeneratePlayoffBracket(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		c.logger.Error("Error parsing leagueId param", "error", err)
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err := c.gameService.GeneratePlayoffBracket(leagueID); err != nil {
		c.logger.Error("Error generating playoff bracket for League", "league_id", leagueID, "error", err)
		switch {
		case errors.Is(err, types.ErrUnauthorized):
			sendError(ctx, http.StatusUnauthorized, err.Error())
		case errors.Is(err, types.ErrLeagueNotFound):
			sendError(ctx, http.StatusNotFound, "League not found")
		case errors.Is(err, types.ErrInvalidInput):
			sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to generate playoff bracket: %v", err))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Playoff bracket generated successfully"})
}
