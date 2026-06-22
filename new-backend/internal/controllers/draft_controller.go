package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
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
	draftService services.DraftService
}

func NewDraftController(draftService services.DraftService) DraftController {
	return &draftControllerImpl{
		draftService: draftService,
	}
}

func (dc *draftControllerImpl) GetDraftByID(ctx *gin.Context) {
	draftIDStr := ctx.Param("draftId")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	draft, err := dc.draftService.GetDraftByID(draftID)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrDraftNotFound.Error()})
		case errors.Is(err, types.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error(), "details": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

func (dc *draftControllerImpl) GetDraftByLeagueID(ctx *gin.Context) {
	leagueIDStr := ctx.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	draft, err := dc.draftService.GetDraftByLeagueID(leagueID)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrDraftNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrDraftNotFound.Error()})
		case errors.Is(err, types.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

func (dc *draftControllerImpl) StartDraft(ctx *gin.Context) {
	leagueIDStr := ctx.Param("leagueId")

	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	turnTimeLimitStr := ctx.DefaultQuery("turnTimeLimit", "120") // Default to 120 minutes
	turnTimeLimit, err := strconv.Atoi(turnTimeLimitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid turnTimeLimit value"})
		return
	}

	draft, err := dc.draftService.StartDraft(leagueID, turnTimeLimit)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrLeagueNotFound):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrLeagueNotFound.Error()})
		case errors.Is(err, types.ErrNoPlayerForDraft):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrNoPlayerForDraft.Error()})
		case errors.Is(err, types.ErrInvalidDraftPosition):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInvalidDraftPosition.Error()})
		case errors.Is(err, types.ErrDuplicateDraftPosition):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrDuplicateDraftPosition.Error()})
		case errors.Is(err, types.ErrIncompleteDraftOrder):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrIncompleteDraftOrder.Error()})
		case errors.Is(err, types.ErrInternalService):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrIncompleteDraftOrder.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start draft", "details": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, draft)
}

func (dc *draftControllerImpl) MakePick(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	var input requests.DraftMakePickRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": types.ErrInvalidInput.Error()})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user information"})
		return
	}

	if err := dc.draftService.MakePick(user, leagueID, &input); err != nil {
		// Handle specific errors from the service layer
		switch {
		case errors.Is(err, types.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not your turn to pick"})
		case errors.Is(err, types.ErrInvalidState):
			c.JSON(http.StatusConflict, gin.H{"error": "Draft is not in a valid state for picking"})
		case errors.Is(err, types.ErrTooManyRequestedPicks):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Requested too many picks"})
		case errors.Is(err, types.ErrInsufficientDraftPoints):
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient draft points"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make pick", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pick successful"})
}

func (dc *draftControllerImpl) SkipPick(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user information"})
		return
	}

	if err := dc.draftService.SkipTurn(user, leagueID); err != nil {
		switch {
		case errors.Is(err, types.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not your turn to skip"})
		case errors.Is(err, types.ErrInvalidState):
			c.JSON(http.StatusConflict, gin.H{"error": "Draft is not in a valid state for skipping"})
		case errors.Is(err, types.ErrCannotSkipBelowMinimumRoster):
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot skip, minimum roster requirement not met"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to skip turn", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Turn skipped successfully"})
}
