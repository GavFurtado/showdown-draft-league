package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DraftPickController interface {
	GetByDraft(ctx *gin.Context)
	GetByPlayer(ctx *gin.Context)
	GetHistory(ctx *gin.Context)
	GetNextPickNumber(ctx *gin.Context)
}

type draftPickControllerImpl struct {
	logger           *slog.Logger
	draftPickService services.DraftPickService
	draftService     services.DraftService
}

func NewDraftPickController(
	logger *slog.Logger,
	draftPickService services.DraftPickService,
	draftService services.DraftService,
) DraftPickController {
	return &draftPickControllerImpl{
		logger:           utils.LoggerWithService(logger, "DraftPickController"),
		draftPickService: draftPickService,
		draftService:     draftService,
	}
}

// GetByDraft godoc
//
//	@Summary		Get draft picks
//	@Description	All picks made in a draft
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.DraftPick
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks [get]
func (c *draftPickControllerImpl) GetByDraft(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	draft, err := c.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		c.logger.Error("failed to get draft", "error", err, "method", "GetByDraft")
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			sendError(ctx, http.StatusNotFound, types.ErrDraftNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	picks, err := c.draftPickService.GetByDraft(draft.ID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByDraft")
		switch {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, picks)
}

// GetByPlayer godoc
//
//	@Summary		Get picks by player
//	@Description	Draft picks for a specific player
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{array}		models.DraftPick
//	@Failure		400			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks/player/{playerId} [get]
func (c *draftPickControllerImpl) GetByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	picks, err := c.draftPickService.GetByPlayer(playerID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByPlayer")
		switch {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, picks)
}

// GetHistory godoc
//
//	@Summary		Get draft pick history
//	@Description	Full history of all draft picks
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.DraftPick
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks/history [get]
func (c *draftPickControllerImpl) GetHistory(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	history, err := c.draftPickService.GetHistory(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetHistory")
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			sendError(ctx, http.StatusNotFound, types.ErrDraftNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, history)
}

// GetNextPickNumber godoc
//
//	@Summary		Get next pick number
//	@Description	The next pick number in the draft
//	@Tags			Draft Picks
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/draft-picks/next-pick-number [get]
func (c *draftPickControllerImpl) GetNextPickNumber(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	draft, err := c.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		c.logger.Error("failed to get draft", "error", err, "method", "GetNextPickNumber")
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			sendError(ctx, http.StatusNotFound, types.ErrDraftNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	nextPickNumber, err := c.draftPickService.GetNextPickNumber(draft.ID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetNextPickNumber")
		switch {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"next_pick_number": nextPickNumber})
}
