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

// initiates the Discord OAuth2 login flow
func (aCtrl *AuthController) Login(ctx *gin.Context) {
	state := uuid.New().String()
	ctx.SetCookie("oauthstate", state, 300, "/", aCtrl.cfg.AppBaseURL, false, true)

	// Use the injected discordOauthConfig
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
	const sessionTokenPeriod = int((time.Hour * 24 * 7 * 30 / time.Second)) // 30 days
	ctx.SetCookie("token", jwtToken, sessionTokenPeriod, "/", aCtrl.cfg.AppBaseURL, false, true)

	// Redirect to dashboard
	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/dashboard", aCtrl.cfg.AppBaseURL))
}
