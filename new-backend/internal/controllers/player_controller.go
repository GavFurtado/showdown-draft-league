package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PlayerController interface {
	GetByID(ctx *gin.Context)
	GetByUserAndLeague(ctx *gin.Context)
	GetByLeague(ctx *gin.Context)
	GetWithFullRoster(ctx *gin.Context)
	JoinLeague(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	UpdateOwnProfile(ctx *gin.Context)
	UpdateDraftPoints(ctx *gin.Context)
	UpdateRecord(ctx *gin.Context)
	UpdateDraftPosition(ctx *gin.Context)
	UpdateRole(ctx *gin.Context)
}

type playerControllerImpl struct {
	logger        *slog.Logger
	playerService services.PlayerService
}

func NewPlayerController(logger *slog.Logger, playerService services.PlayerService) PlayerController {
	return &playerControllerImpl{
		logger:        utils.LoggerWithService(logger, "PlayerController"),
		playerService: playerService,
	}
}

// GetByID godoc
//
//	@Summary		Get a player
//	@Description	Player details within a league
//	@Tags			Players
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId} [get]
func (c *playerControllerImpl) GetByID(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	player, err := c.playerService.GetByID(playerID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByID")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPlayerNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

func (c *playerControllerImpl) GetByUserAndLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "missing or invalid userId param")
		return
	}

	player, err := c.playerService.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByUserAndLeague")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPlayerNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// GetByLeague godoc
//
//	@Summary		Get league players
//	@Description	All players of a league
//	@Tags			Players
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players [get]
func (c *playerControllerImpl) GetByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	players, err := c.playerService.GetByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByLeague")
		switch err {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, players)
}

// GetWithFullRoster godoc
//
//	@Summary		Get player with full roster
//	@Description	Player details including all drafted Pokemon
//	@Tags			Players
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId}/roster [get]
func (c *playerControllerImpl) GetWithFullRoster(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	player, err := c.playerService.GetWithFullRoster(playerID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetWithFullRoster")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrPlayerNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// JoinLeague godoc
//
//	@Summary		Join a league
//	@Description	Add a player to a league
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string									true	"League ID"
//	@Param			request		body		requests.PlayerCreateRequestDTO	true	"Player details"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/join [post]
func (c *playerControllerImpl) JoinLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	var req requests.PlayerCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing required fields in payload")
		return
	}

	if currentUser.Role != "admin" && req.UserID != currentUser.ID {
		sendError(ctx, http.StatusUnauthorized, "Cannot perform this transaction")
		return
	}

	if req.UserID == uuid.Nil || req.LeagueID == uuid.Nil {
		sendError(ctx, http.StatusBadRequest, "Bad or Malformed Request")
		return
	}

	player, err := c.playerService.Create(currentUser, &req)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "JoinLeague")
		switch err {
		case types.ErrUserNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrUserAlreadyInLeague, types.ErrInLeagueNameTaken, types.ErrTeamNameTaken, types.ErrLeagueFull, types.ErrUserAlreadySpectator, types.ErrOwnerSpectatorJoin:
			sendError(ctx, http.StatusConflict, err.Error())
		case types.ErrInternalService, types.ErrFailedToCreatePlayer:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		case types.ErrInvalidState:
			sendError(ctx, http.StatusUnauthorized, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// UpdateProfile godoc
//
//	@Summary		Update a player's profile (staff)
//	@Description	Update in-league name or team name for any player. Requires update:player.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string										true	"League ID"
//	@Param			playerId	path		string										true	"Player ID"
//	@Param			request		body		requests.UpdatePlayerInfoRequestDTO	true	"Profile fields"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId}/profile [put]
func (c *playerControllerImpl) UpdateProfile(ctx *gin.Context) {
	var req requests.UpdatePlayerInfoRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if req.InLeagueName == nil && req.TeamName == nil {
		sendError(ctx, http.StatusBadRequest, "No fields to update")
		return
	}

	player, err := c.playerService.UpdateProfile(playerID, req.InLeagueName, req.TeamName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateProfile")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInLeagueNameTaken, types.ErrTeamNameTaken:
			sendError(ctx, http.StatusConflict, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// UpdateOwnProfile godoc
//
//	@Summary		Update your own player profile
//	@Description	Update your own in-league name or team name. Requires update:own_player.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string										true	"League ID"
//	@Param			request		body		requests.UpdatePlayerInfoRequestDTO	true	"Profile fields"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/me/profile [put]
func (c *playerControllerImpl) UpdateOwnProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req requests.UpdatePlayerInfoRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if req.InLeagueName == nil && req.TeamName == nil {
		sendError(ctx, http.StatusBadRequest, "No fields to update")
		return
	}

	player, err := c.playerService.UpdateOwnProfile(currentUser.ID, leagueID, req.InLeagueName, req.TeamName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateOwnProfile")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInLeagueNameTaken, types.ErrTeamNameTaken:
			sendError(ctx, http.StatusConflict, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// UpdateDraftPoints godoc
//
//	@Summary		Update player's draft points
//	@Description	Set a player's draft points. Requires update:player_score.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId}/draft-points [put]
func (c *playerControllerImpl) UpdateDraftPoints(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		DraftPoints *int `json:"DraftPoints" validate:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing or invalid DraftPoints")
		return
	}

	player, err := c.playerService.UpdateDraftPoints(playerID, req.DraftPoints)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateDraftPoints")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// UpdateRecord godoc
//
//	@Summary		Update player's record
//	@Description	Set wins and losses for a player. Requires update:player_score.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId}/record [put]
func (c *playerControllerImpl) UpdateRecord(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		Wins   *int `json:"Wins" validate:"required"`
		Losses *int `json:"Losses" validate:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing or invalid Wins/Losses")
		return
	}

	player, err := c.playerService.UpdateRecord(playerID, *req.Wins, *req.Losses)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateRecord")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// UpdateDraftPosition godoc
//
//	@Summary		Update player's draft position
//	@Description	Set the draft order position for a player. Requires update:player_score.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId}/draft-position [put]
func (c *playerControllerImpl) UpdateDraftPosition(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		DraftPosition *int `json:"DraftPosition" validate:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing or invalid DraftPosition")
		return
	}

	player, err := c.playerService.UpdateDraftPosition(playerID, *req.DraftPosition)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateDraftPosition")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}

// UpdateRole godoc
//
//	@Summary		Update a player's role
//	@Description	Change a player's role to another participant role. Requires update:role.
//	@Tags			Players
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string				true	"League ID"
//	@Param			playerId	path		string				true	"Player ID"
//	@Param			request		body		object				true	"Role"
//	@Success		200			{object}	models.Player
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/players/{playerId}/role [put]
func (c *playerControllerImpl) UpdateRole(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		Role string `json:"Role" validate:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing or invalid Role")
		return
	}

	newRole, ok := rbac.ParseMemberRole(req.Role)
	if !ok {
		sendError(ctx, http.StatusBadRequest, "Unrecognized role")
		return
	}

	player, err := c.playerService.UpdateRole(playerID, newRole)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateRole")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInvalidInput:
			sendError(ctx, http.StatusBadRequest, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, player)
}
