package controllers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// AuthController does auth flow
// TODO: Refresh Token based auth flow
type AuthController interface {
	Login(ctx *gin.Context)
	DiscordCallback(ctx *gin.Context)
	Logout(ctx *gin.Context)
}

type authControllerImpl struct {
	logger             *slog.Logger
	authService        services.AuthService
	cfg                *config.Config
	discordOauthConfig *oauth2.Config
}

func NewAuthController(
	logger *slog.Logger,
	authService services.AuthService,
	cfg *config.Config,
	oauthConfig *oauth2.Config,
) AuthController {
	return &authControllerImpl{
		logger:             utils.LoggerWithService(logger, "AuthController"),
		authService:        authService,
		cfg:                cfg,
		discordOauthConfig: oauthConfig,
	}
}

// Login godoc
//
//	@Summary		Login with Discord
//	@Description	Redirects to Discord for authentication
//	@Tags			Auth
//	@Success		307
//	@Security		none
//	@Router			/auth/discord/login [get]
func (c *authControllerImpl) Login(ctx *gin.Context) {
	// 1. Check for existing JWT cookie
	token, err := ctx.Cookie("token")
	if err == nil {
		// 2. Try to validate it
		if userID, err := c.authService.VerifyToken(token); err == nil {
			// 3. Valid token -> go straight to frontend dashboard
			c.logger.Info("user already authenticated", "user_id", userID)
			ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/my-leagues", c.cfg.APP_BASE_URL))
			return
		}
	}

	// 4. No valid token -> begin Discord OAuth flow
	state := uuid.New().String()
	ctx.SetCookie("oauthstate", state, 300, "/", "", false, true)

	url := c.discordOauthConfig.AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// DiscordCallback godoc
//
//	@Summary		Discord OAuth callback
//	@Description	Handles the OAuth callback from Discord
//	@Tags			Auth
//	@Param			state	query	string	true	"OAuth state token"
//	@Param			code	query	string	true	"Authorization code from Discord"
//	@Success		200	{object}	responses.TokenResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Security		none
//	@Router			/auth/discord/callback [get]
func (c *authControllerImpl) DiscordCallback(ctx *gin.Context) {
	storedState, err := ctx.Cookie("oauthstate")
	if err != nil || storedState == "" || storedState != ctx.Query("state") {
		c.logger.Error("OAuth state mismatch or missing", "stored", storedState, "query", ctx.Query("state"), "error", err)
		sendError(ctx, http.StatusBadRequest, "Invalid Input")
		return
	}

	ctx.SetCookie("oauthstate", "", -1, "/", "", false, true)

	code := ctx.Query("code")
	if code == "" {
		sendError(ctx, http.StatusBadRequest, "Authorization code not provided")
		return
	}

	_, jwtToken, err := c.authService.HandleDiscordCallback(ctx, code)
	if err != nil {
		c.logger.Error("AuthService failed", "error", err)
		sendError(ctx, http.StatusInternalServerError, "Authentication failed")
		return
	}

	httpOnly := c.cfg.ENVIRONMENT != "dev" // set httpOnly to false if Environment is "dev"

	// Set JWT as an HTTP-only cookie
	const sessionTokenPeriod = int((time.Hour * 24 * 3 * 30 / time.Second)) // 90 days
	ctx.SetCookie("token", jwtToken, sessionTokenPeriod, "/", c.cfg.BACKEND_BASE_URL, false, httpOnly)

	// Return JWT as JSON
	ctx.JSON(http.StatusOK, responses.TokenResponse{Token: jwtToken})
}

// Logout godoc
//
//	@Summary		Logout
//	@Description	Clears the session cookie
//	@Tags			Auth
//	@Success		200	{object}	map[string]interface{}
//	@Security		none
//	@Router			/auth/logout [post]
func (c *authControllerImpl) Logout(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "/", c.cfg.BACKEND_BASE_URL, false, true) // clear the token cookie
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
