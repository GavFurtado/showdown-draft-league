package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DevAuthController backs the dev-only impersonation endpoints.
// Routes are only registered when ENV=dev — see routes.go.
type DevAuthController interface {
	ListUsers(ctx *gin.Context)
	CreateUser(ctx *gin.Context)
	Impersonate(ctx *gin.Context)
	UpsertMembership(ctx *gin.Context)
}

type devAuthControllerImpl struct {
	logger         *slog.Logger
	devAuthService services.DevAuthService
}

func NewDevAuthController(
	logger *slog.Logger,
	devAuthService services.DevAuthService,
) DevAuthController {
	return &devAuthControllerImpl{
		logger:         utils.LoggerWithService(logger, "DevAuthController"),
		devAuthService: devAuthService,
	}
}

// ListUsers godoc
//
//	@Summary		(DEV) List users
//	@Description	Dev-only. Lists all users for impersonation. Only registered when ENV=dev.
//	@Tags			Dev
//	@Success		200	{array}		models.User
//	@Failure		500	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/auth/dev/users [get]
func (c *devAuthControllerImpl) ListUsers(ctx *gin.Context) {
	users, err := c.devAuthService.ListUsers()
	if err != nil {
		c.logger.Error("Service failed", "error", err, "method", "ListUsers")
		sendError(ctx, http.StatusInternalServerError, "Failed to list users")
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// CreateUser godoc
//
//	@Summary		(DEV) Create a test user
//	@Description	Dev-only. Creates a fake user for impersonation. Only registered when ENV=dev.
//	@Tags			Dev
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.DevCreateUserRequestDTO	true	"New user"
//	@Success		201	{object}	models.User
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		409	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/auth/dev/users [post]
func (c *devAuthControllerImpl) CreateUser(ctx *gin.Context) {
	var req requests.DevCreateUserRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing field(s) in the payload")
		return
	}

	user, err := c.devAuthService.CreateDevUser(req.Name, req.ShowdownUsername)
	if err != nil {
		c.handleServiceError(ctx, "CreateUser", err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// Impersonate godoc
//
//	@Summary		(DEV) Log in as any user
//	@Description	Dev-only. Mints a real session JWT for the given user, bypassing Discord OAuth. Only registered when ENV=dev.
//	@Tags			Dev
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.DevImpersonateRequestDTO	true	"Target user"
//	@Success		200	{object}	responses.TokenResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/auth/dev/login [post]
func (c *devAuthControllerImpl) Impersonate(ctx *gin.Context) {
	var req requests.DevImpersonateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing field(s) in the payload")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid user ID")
		return
	}

	_, token, err := c.devAuthService.Impersonate(userID)
	if err != nil {
		c.handleServiceError(ctx, "Impersonate", err)
		return
	}

	ctx.JSON(http.StatusOK, responses.TokenResponse{Token: token})
}

// UpsertMembership godoc
//
//	@Summary		(DEV) Add/update a league membership role
//	@Description	Dev-only. Adds the user to the league with the given role (or updates their existing role). Only registered when ENV=dev.
//	@Tags			Dev
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.DevUpsertMembershipRequestDTO	true	"Membership"
//	@Success		200	{object}	models.LeagueMember
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/auth/dev/memberships [post]
func (c *devAuthControllerImpl) UpsertMembership(ctx *gin.Context) {
	var req requests.DevUpsertMembershipRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing field(s) in the payload")
		return
	}

	leagueID, err := uuid.Parse(req.LeagueID)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid league ID")
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid user ID")
		return
	}

	member, err := c.devAuthService.UpsertMembership(leagueID, userID, req.Role)
	if err != nil {
		c.handleServiceError(ctx, "UpsertMembership", err)
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *devAuthControllerImpl) handleServiceError(ctx *gin.Context, method string, err error) {
	switch err {
	case types.ErrInvalidInput:
		sendError(ctx, http.StatusBadRequest, err.Error())
	case types.ErrUserNotFound, types.ErrLeagueNotFound:
		sendError(ctx, http.StatusNotFound, err.Error())
	case types.ErrConflict, types.ErrFailedToCreatePlayer:
		sendError(ctx, http.StatusConflict, err.Error())
	default:
		c.logger.Error("Service failed", "error", err, "method", method)
		sendError(ctx, http.StatusInternalServerError, "Internal Server Error")
	}
}
