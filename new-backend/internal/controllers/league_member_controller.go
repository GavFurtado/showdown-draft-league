package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeagueMemberController interface {
	GetByID(ctx *gin.Context)
	GetByUserAndLeague(ctx *gin.Context)
	GetByLeague(ctx *gin.Context)
	GetByUser(ctx *gin.Context)
	GetWithFullRoster(ctx *gin.Context)
	JoinLeague(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	UpdateDraftPoints(ctx *gin.Context)
	UpdateRecord(ctx *gin.Context)
	UpdateDraftPosition(ctx *gin.Context)
}

type leagueMemberControllerImpl struct {
	logger              *slog.Logger
	leagueMemberService services.LeagueMemberService
}

func NewLeagueMemberController(logger *slog.Logger, leagueMemberService services.LeagueMemberService) LeagueMemberController {
	return &leagueMemberControllerImpl{
		logger:              utils.LoggerWithService(logger, "LeagueMemberController"),
		leagueMemberService: leagueMemberService,
	}
}

// GetByID godoc
//
//	@Summary		Get a league member
//	@Description	Member details
//	@Tags			League Members
//	@Produce		json
//	@Param			id	path		string	true	"Member ID"
//	@Success		200	{object}	models.LeagueMember
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/members/{id} [get]
func (c *leagueMemberControllerImpl) GetByID(ctx *gin.Context) {
	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	member, err := c.leagueMemberService.GetByID(memberID)
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

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) GetByUserAndLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	userIDStr := ctx.DefaultQuery("userId", "")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "missing or invalid userId query")
		return
	}

	member, err := c.leagueMemberService.GetByUserAndLeague(userID, leagueID)
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

	ctx.JSON(http.StatusOK, member)
}

// GetByLeague godoc
//
//	@Summary		Get league members
//	@Description	All members of a league
//	@Tags			League Members
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.LeagueMember
//	@Failure		400			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/members [get]
func (c *leagueMemberControllerImpl) GetByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	members, err := c.leagueMemberService.GetByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByLeague")
		switch err {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, members)
}

// GetByUser godoc
//
//	@Summary		Get members by user
//	@Description	All league memberships for a user
//	@Tags			League Members
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{array}		models.LeagueMember
//	@Failure		400	{object}	map[string]interface{}
//	@Router			/api/users/{id}/members [get]
func (c *leagueMemberControllerImpl) GetByUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	members, err := c.leagueMemberService.GetByUser(userID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByUser")
		switch err {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, members)
}

// GetWithFullRoster godoc
//
//	@Summary		Get member with full roster
//	@Description	Member details including all drafted Pokemon
//	@Tags			League Members
//	@Produce		json
//	@Param			id	path		string	true	"Member ID"
//	@Success		200	{object}	models.LeagueMember
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/members/{id}/roster [get]
func (c *leagueMemberControllerImpl) GetWithFullRoster(ctx *gin.Context) {
	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	member, err := c.leagueMemberService.GetWithFullRoster(memberID)
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

	ctx.JSON(http.StatusOK, member)
}

// JoinLeague godoc
//
//	@Summary		Join a league
//	@Description	Add a member to a league
//	@Tags			League Members
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string									true	"League ID"
//	@Param			request		body		requests.LeagueMemberCreateRequestDTO	true	"Member details"
//	@Success		200			{object}	models.LeagueMember
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		401			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Failure		409			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId}/members/join [post]
func (c *leagueMemberControllerImpl) JoinLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	var req requests.LeagueMemberCreateRequestDTO
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

	member, err := c.leagueMemberService.Create(currentUser, &req)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "JoinLeague")
		switch err {
		case types.ErrUserNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrUserAlreadyInLeague, types.ErrInLeagueNameTaken, types.ErrTeamNameTaken:
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

	ctx.JSON(http.StatusOK, member)
}

// UpdateProfile godoc
//
//	@Summary		Update member profile
//	@Description	Update in-league name or team name
//	@Tags			League Members
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Member ID"
//	@Param			request	body		requests.UpdateLeagueMemberInfoRequestDTO	true	"Profile fields"
//	@Success		200		{object}	models.LeagueMember
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Router			/api/members/{id}/profile [put]
func (c *leagueMemberControllerImpl) UpdateProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	var req requests.UpdateLeagueMemberInfoRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if req.InLeagueName == nil && req.TeamName == nil {
		sendError(ctx, http.StatusBadRequest, "No fields to update")
		return
	}

	member, err := c.leagueMemberService.UpdateProfile(currentUser, memberID, req.InLeagueName, req.TeamName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateProfile")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInLeagueNameTaken, types.ErrTeamNameTaken:
			sendError(ctx, http.StatusConflict, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		case types.ErrUnauthorized:
			sendError(ctx, http.StatusUnauthorized, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

// UpdateDraftPoints godoc
//
//	@Summary		Update member's draft points
//	@Description	Set a member's draft points
//	@Tags			League Members
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Member ID"
//	@Success		200	{object}	models.LeagueMember
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/members/{id}/draft-points [put]
func (c *leagueMemberControllerImpl) UpdateDraftPoints(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
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

	member, err := c.leagueMemberService.UpdateDraftPoints(currentUser, memberID, req.DraftPoints)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateDraftPoints")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		case types.ErrUnauthorized:
			sendError(ctx, http.StatusUnauthorized, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

// UpdateRecord godoc
//
//	@Summary		Update member's record
//	@Description	Set wins and losses for a member
//	@Tags			League Members
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Member ID"
//	@Success		200	{object}	models.LeagueMember
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/members/{id}/record [put]
func (c *leagueMemberControllerImpl) UpdateRecord(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
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

	member, err := c.leagueMemberService.UpdateRecord(currentUser, memberID, *req.Wins, *req.Losses)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateRecord")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		case types.ErrUnauthorized:
			sendError(ctx, http.StatusUnauthorized, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

// UpdateDraftPosition godoc
//
//	@Summary		Update member's draft position
//	@Description	Set the draft order position for a member
//	@Tags			League Members
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Member ID"
//	@Success		200	{object}	models.LeagueMember
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/members/{id}/draft-position [put]
func (c *leagueMemberControllerImpl) UpdateDraftPosition(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
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

	member, err := c.leagueMemberService.UpdateDraftPosition(currentUser, memberID, *req.DraftPosition)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateDraftPosition")
		switch err {
		case types.ErrPlayerNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInternalService:
			sendError(ctx, http.StatusInternalServerError, err.Error())
		case types.ErrUnauthorized:
			sendError(ctx, http.StatusUnauthorized, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}
