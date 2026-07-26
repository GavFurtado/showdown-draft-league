package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DraftController interface {
	GetDraftByID(ctx *gin.Context)
	GetDraftByLeagueID(ctx *gin.Context)
	StartDraft(ctx *gin.Context)
	MakePick(ctx *gin.Context)
	SkipPick(ctx *gin.Context)
}

type draftControllerImpl struct {
	logger       *slog.Logger
	draftService services.DraftService
}

func NewDraftController(logger *slog.Logger, draftService services.DraftService) DraftController {
	return &draftControllerImpl{
		logger:       utils.LoggerWithService(logger, "DraftController"),
		draftService: draftService,
	}
}

// GetDraftByID godoc
//
//	@Summary		Get a draft
//	@Description	Draft details by ID
//	@Tags			Drafts
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			draftId		path		string	true	"Draft ID"
//	@Success		200			{object}	models.Draft
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/draft/{draftId} [get]
func (dc *draftControllerImpl) GetDraftByID(ctx *gin.Context) {
	draftIDStr := ctx.Param("draftId")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	draft, err := dc.draftService.GetDraftByID(draftID)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			sendError(ctx, http.StatusNotFound, types.ErrDraftNotFound.Error())
		case errors.Is(err, types.ErrInternalService):
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error()+" - "+err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// GetDraftByLeagueID godoc
//
//	@Summary		Get draft by league
//	@Description	Current draft for a league
//	@Tags			Drafts
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	models.Draft
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/draft [get]
func (dc *draftControllerImpl) GetDraftByLeagueID(ctx *gin.Context) {
	leagueIDStr := ctx.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	draft, err := dc.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			sendError(ctx, http.StatusNotFound, types.ErrDraftNotFound.Error())
		case errors.Is(err, types.ErrInternalService):
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// StartDraft godoc
//
//	@Summary		Start a draft
//	@Description	Start the draft for a league
//	@Tags			Drafts
//	@Produce		json
//	@Param			leagueId		path		int	true	"League ID"
//	@Param			turnTimeLimit	query		int	false	"Turn time limit in minutes"	default(120)
//	@Success		200				{object}	models.Draft
//	@Failure		400				{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/draft/start [post]
func (dc *draftControllerImpl) StartDraft(ctx *gin.Context) {
	leagueIDStr := ctx.Param("leagueId")

	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid league ID")
		return
	}

	turnTimeLimitStr := ctx.DefaultQuery("turnTimeLimit", "120")
	turnTimeLimit, err := strconv.Atoi(turnTimeLimitStr)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid turnTimeLimit value")
		return
	}

	draft, err := dc.draftService.StartDraft(leagueID, turnTimeLimit)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrLeagueNotFound):
			sendError(ctx, http.StatusBadRequest, types.ErrLeagueNotFound.Error())
		case errors.Is(err, types.ErrNoPlayerForDraft):
			sendError(ctx, http.StatusInternalServerError, types.ErrNoPlayerForDraft.Error())
		case errors.Is(err, types.ErrInvalidDraftPosition):
			sendError(ctx, http.StatusInternalServerError, types.ErrInvalidDraftPosition.Error())
		case errors.Is(err, types.ErrDuplicateDraftPosition):
			sendError(ctx, http.StatusInternalServerError, types.ErrDuplicateDraftPosition.Error())
		case errors.Is(err, types.ErrIncompleteDraftOrder):
			sendError(ctx, http.StatusInternalServerError, types.ErrIncompleteDraftOrder.Error())
		case errors.Is(err, types.ErrInternalService):
			sendError(ctx, http.StatusInternalServerError, types.ErrIncompleteDraftOrder.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "Failed to start draft")
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

// MakePick godoc
//
//	@Summary		Make a draft pick
//	@Description	Select a Pokemon in the draft
//	@Tags			Drafts
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string								true	"League ID"
//	@Param			request		body		requests.DraftMakePickRequestDTO	true	"Pick details"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/draft/pick [post]
func (dc *draftControllerImpl) MakePick(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var input requests.DraftMakePickRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, types.ErrInvalidInput.Error())
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		sendError(c, http.StatusUnauthorized, types.ErrNoUserInContext.Error())
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		sendError(c, http.StatusInternalServerError, "Failed to process user information")
		return
	}

	if err := dc.draftService.MakePick(user, leagueID, &input); err != nil {
		switch {
		case errors.Is(err, types.ErrUnauthorized):
			sendError(c, http.StatusUnauthorized, "Not your turn to pick")
		case errors.Is(err, types.ErrInvalidState):
			sendError(c, http.StatusConflict, "Draft is not in a valid state for picking")
		case errors.Is(err, types.ErrTooManyRequestedPicks):
			sendError(c, http.StatusBadRequest, "Requested too many picks")
		case errors.Is(err, types.ErrInsufficientDraftPoints):
			sendError(c, http.StatusForbidden, "Insufficient draft points")
		default:
			sendError(c, http.StatusInternalServerError, "Failed to make pick")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pick successful"})
}

// SkipPick godoc
//
//	@Summary		Skip your turn
//	@Description	Pass on your pick in the draft
//	@Tags			Drafts
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/draft/skip [post]
func (dc *draftControllerImpl) SkipPick(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid league ID")
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		sendError(c, http.StatusUnauthorized, "User not found in context")
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		sendError(c, http.StatusInternalServerError, "Failed to process user information")
		return
	}

	if err := dc.draftService.SkipTurn(user, leagueID); err != nil {
		switch {
		case errors.Is(err, types.ErrUnauthorized):
			sendError(c, http.StatusUnauthorized, "Not your turn to skip")
		case errors.Is(err, types.ErrInvalidState):
			sendError(c, http.StatusConflict, "Draft is not in a valid state for skipping")
		case errors.Is(err, types.ErrCannotSkipBelowMinimumRoster):
			sendError(c, http.StatusForbidden, "Cannot skip, minimum roster requirement not met")
		default:
			sendError(c, http.StatusInternalServerError, "Failed to skip turn")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Turn skipped successfully"})
}
