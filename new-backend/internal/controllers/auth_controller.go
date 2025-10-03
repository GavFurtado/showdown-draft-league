package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthController struct {
	authService        services.AuthService
	cfg                *config.Config
	discordOauthConfig *oauth2.Config
}

func NewAuthController(
	authService services.AuthService,
	cfg *config.Config,
	oauthConfig *oauth2.Config,
) *AuthController {
	return &AuthController{
		authService:        authService,
		cfg:                cfg,
		discordOauthConfig: oauthConfig,
	}
}

// NOTE: the skip to frontend dashboard could maybe cause problems in edge case scenarios i haven't thought of (probably)
func (aCtrl *AuthController) Login(ctx *gin.Context) {
	// 1. Check for existing JWT cookie
	token, err := ctx.Cookie("token")
	if err == nil {
		// 2. Try to validate it
		if userID, err := aCtrl.authService.VerifyToken(token); err == nil {
			// 3. Valid token -> go straight to frontend dashboard
			log.Printf("user already authenticated: %s\n", userID)
			ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/dashboard", aCtrl.cfg.AppBaseURL))
			return
		}
	}

	// 4. No valid token → begin Discord OAuth flow
	state := uuid.New().String()
	ctx.SetCookie("oauthstate", state, 300, "/", "localhost", false, true)

	url := aCtrl.discordOauthConfig.AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// handles the Discord OAuth2 callback
func (aCtrl *AuthController) DiscCallback(ctx *gin.Context) {
	storedState, err := ctx.Cookie("oauthstate")
	if err != nil || storedState == "" || storedState != ctx.Query("state") {
		log.Printf("(Error: DiscCallback) - OAuth state mismatch or missing. Stored=%s, query=%s, err=%v", storedState, ctx.Query("state"), err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Input"})
		return
	}

	ctx.SetCookie("oauthstate", "", -1, "/", aCtrl.cfg.AppBaseURL, false, true) // Clear the state cookie

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	_, jwtToken, err := aCtrl.authService.HandleDiscordCallback(ctx, code)
	if err != nil {
		log.Printf("(Error: DiscCallback) - AuthService failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Set JWT as an HTTP-only cookie
	const sessionTokenPeriod = int((time.Hour * 24 * 3 * 30 / time.Second)) // 90 days
	ctx.SetCookie("token", jwtToken, sessionTokenPeriod, "/", aCtrl.cfg.AppBaseURL, false, true)
	ctx.Header("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	// Redirect to dashboard
	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/dashboard", aCtrl.cfg.AppBaseURL))
}
