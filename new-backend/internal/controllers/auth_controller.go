package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/config"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var discordOauthConfig *oauth2.Config

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"` // deprecated in Discord but maybe it's needed
	Avatar        string `json:"avatar"`
}

type AuthController struct {
	userRepo   *repositories.UserRepository
	jwtService *services.JWTService
	cfg        *config.Config
}

func NewAuthController(userRepo *repositories.UserRepository, jwtService *services.JWTService,
	cfg *config.Config) *AuthController {

	discordOauthConfig = &oauth2.Config{
		ClientID:     cfg.DiscordClientID,
		ClientSecret: cfg.DiscordClientSecret,
		RedirectURL:  cfg.DiscordRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
		Scopes: []string{"identify"}, // Request User ID and Name
	}

	return &AuthController{
		userRepo:   userRepo,
		jwtService: jwtService,
		cfg:        cfg,
	}
}

// initiates the Discord OAuth2 login flow
func (aCtrl *AuthController) Login(ctx *gin.Context) {
	state := uuid.New().String() // to prevent CSRF attacks (apparently)
	// Store the state in a cookie or session for verification later
	ctx.SetCookie("oauthstate", state, 300, "/", aCtrl.cfg.AppBaseURL, false, true) // 5 minutes expiry

	url := discordOauthConfig.AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func (aCtrl *AuthController) DiscCallback(ctx *gin.Context) {
	storedState, err := ctx.Cookie("oauthstate")
	if err != nil || storedState == "" || storedState != ctx.Query("state") {
		log.Printf("(Error: DiscCallback) - OAuth state mismatch or missing. Stored=%s, query=%s, err=%v", storedState, ctx.Query("state"), err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Input"})
		return
	}

	ctx.SetCookie("oauthstate", "", -1, "/", aCtrl.cfg.AppBaseURL, false, true)

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	// Exchange the authorization code for an access token
	token, err := discordOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("(Error: DiscCallback) - Failed to exchange code for token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Discord"})
		return
	}

	discordUser, err := getDiscordUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("(Error: DiscCallback) - Failed to get Discord user info: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve Discord user information"})
		return
	}

	user, err := aCtrl.userRepo.GetUserByDiscordID(discordUser.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// User does not exist, create a new one
			log.Printf("Creating new user for Discord ID: %s", discordUser.ID)
			newUser := models.User{
				DiscordID:        discordUser.ID,
				DiscordUsername:  fmt.Sprintf("%s#%s", discordUser.Username, discordUser.Discriminator), // Combine if discriminator still exists for old users
				DiscordAvatarURL: getDiscordAvatarURL(discordUser.ID, discordUser.Avatar),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			createdUser, createErr := aCtrl.userRepo.CreateUser(&newUser)
			if createErr != nil {
				log.Printf("Failed to create new user: %v", createErr)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
			user = createdUser
		} else {
			log.Printf("Error checking user existence: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login"})
			return
		}
	} else {
		// User exists, update their username/avatar if needed
		updatedUsername := discordUser.Username
		if discordUser.Discriminator != "0" && discordUser.Discriminator != "" { // if has a discriminator
			updatedUsername = fmt.Sprintf("%s#%s", discordUser.Username, discordUser.Discriminator) // append the names
		}
		if user.DiscordUsername != updatedUsername || user.DiscordAvatarURL != getDiscordAvatarURL(discordUser.ID, discordUser.Avatar) {
			user.DiscordUsername = updatedUsername
			user.DiscordAvatarURL = getDiscordAvatarURL(discordUser.ID, discordUser.Avatar)
			user.UpdatedAt = time.Now()
			if _, updateErr := aCtrl.userRepo.UpdateUser(user); updateErr != nil {
				log.Printf("Warning: Failed to update user details for Discord ID %s: %v", discordUser.ID, updateErr)
			}
		}
	}

	// Generate JWT token
	jwtToken, err := aCtrl.jwtService.GenerateToken(user.Id)
	if err != nil {
		log.Printf("Failed to generate JWT for user %s: %v", user.Id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set JWT as an HTTP-only cookie
	const sessionTokenPeriod = int((time.Hour * 24 * 7 * 30 / time.Second)) // 30 days
	ctx.SetCookie("token", jwtToken, sessionTokenPeriod, "/", aCtrl.cfg.AppBaseURL, false, true)

	// Redirect to a dashboard
	ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/dashboard", aCtrl.cfg.AppBaseURL))
}

func getDiscordUserInfo(accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("(Error: getDiscordUserInfo) - Failed to create Discord user info request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("(Error: getDiscordUserInfo) - Failed to create Discord user info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("(Error: getDiscordUserInfo) - Discord API returned non-OK status: %d, body: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var discordUser DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return nil, fmt.Errorf("failed to decode Discord user info: %w", err)
	}
	return &discordUser, nil
}

// getDiscordAvatarURL constructs the Discord user's avatar URL.
// Discord avatars can be animated (gif) or static (png).
// If the avatar hash starts with 'a_', it's animated.
func getDiscordAvatarURL(userID, avatarHash string) string {
	if avatarHash == "" {
		// Default Discord avatar for users without a custom one
		// For now, returning an empty string or a placeholder. (could use own default)
		return ""
	}

	// Check if the avatar is animated (starts with 'a_')
	if strings.HasPrefix(avatarHash, "a_") {
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.gif", userID, avatarHash)
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", userID, avatarHash)
}
