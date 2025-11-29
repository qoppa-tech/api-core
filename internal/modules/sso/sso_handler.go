package sso

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/modules/auth"
	"github.com/parlorhub/api-core/internal/modules/auth/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

type SSOHandler struct {
	queries        *sqlc.Queries
	sessionService *session.SessionService
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func getGoogleOAuthConfig() *oauth2.Config {
	if googleOAuthConfig == nil {
		googleOAuthConfig = &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}
	return googleOAuthConfig
}

func NewSSOHandler(db *sql.DB, sessionService *session.SessionService) *SSOHandler {
	return &SSOHandler{
		queries:        sqlc.New(db),
		sessionService: sessionService,
	}
}

// GoogleLoginHandler redirects to Google OAuth consent page
func (h *SSOHandler) GoogleLoginHandler(ctx *gin.Context) {
	state := uuid.New().String() // Generate random state
	// TODO: Store state in Redis for validation
	url := getGoogleOAuthConfig().AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallbackHandler handles the OAuth callback from Google
func (h *SSOHandler) GoogleCallbackHandler(ctx *gin.Context) {
	// Get authorization code
	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "code not found"})
		return
	}

	// Exchange code for token
	token, err := getGoogleOAuthConfig().Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	// Get user info from Google
	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Check if SSO record exists
	ssoRecord, err := h.queries.GetSSOByProvider(ctx.Request.Context(), sqlc.GetSSOByProviderParams{
		Provider:       "google",
		ProviderUserID: userInfo.ID,
	})

	var user sqlc.User

	if err == sql.ErrNoRows {
		// Check if user with this email exists
		existingUser, err := h.queries.GetUserByEmail(ctx.Request.Context(), userInfo.Email)
		if err == sql.ErrNoRows {
			// Create new user
			user, err = h.queries.CreateUser(ctx.Request.Context(), sqlc.CreateUserParams{
				Email:        userInfo.Email,
				PasswordHash: "", // No password for SSO users
				Name:         userInfo.Name,
				Phone:        "",
				Role:         sqlc.UserRoleSolo,
				SalonID:      uuid.NullUUID{Valid: false},
			})
			if err != nil {
				log.Printf("Error creating user: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
		} else if err != nil {
			log.Printf("Error checking existing user: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		} else {
			user = existingUser
		}

		// Create SSO record
		_, err = h.queries.CreateSSO(ctx.Request.Context(), sqlc.CreateSSOParams{
			UserID:         user.ID,
			Provider:       "google",
			ProviderUserID: userInfo.ID,
			AccessToken:    sql.NullString{String: token.AccessToken, Valid: true},
			RefreshToken:   sql.NullString{String: token.RefreshToken, Valid: token.RefreshToken != ""},
			ExpiresAt:      sql.NullTime{Time: token.Expiry, Valid: !token.Expiry.IsZero()},
		})
		if err != nil {
			log.Printf("Error creating SSO record: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create SSO record"})
			return
		}
	} else if err != nil {
		log.Printf("Error checking SSO record: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	} else {
		// SSO record exists, get user
		user, err = h.queries.GetUserByID(ctx.Request.Context(), ssoRecord.UserID)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}

		// Update SSO tokens
		_, err = h.queries.UpdateSSOTokens(ctx.Request.Context(), sqlc.UpdateSSOTokensParams{
			ID:           ssoRecord.ID,
			AccessToken:  sql.NullString{String: token.AccessToken, Valid: true},
			RefreshToken: sql.NullString{String: token.RefreshToken, Valid: token.RefreshToken != ""},
			ExpiresAt:    sql.NullTime{Time: token.Expiry, Valid: !token.Expiry.IsZero()},
		})
		if err != nil {
			log.Printf("Error updating SSO tokens: %v", err)
		}
	}

	// Generate JWT token using auth package
	jwtToken, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Store token in Redis if session service is available
	if h.sessionService != nil {
		_ = h.sessionService.StoreToken(ctx.Request.Context(), user.ID.String(), jwtToken, auth.JwtExpiration)
	}

	// Redirect to frontend with token
	frontendURL := os.Getenv("FRONTEND_URL")
	ctx.Redirect(http.StatusTemporaryRedirect, frontendURL+"/auth/callback?token="+jwtToken)
}

func getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
