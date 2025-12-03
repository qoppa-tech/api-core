package sso

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/logger"
	"github.com/parlorhub/api-core/internal/modules/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config

type SSOHandler struct {
	queries     *sqlc.Queries
	authService *auth.AuthService
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

func NewSSOHandler(db *sql.DB, authService *auth.AuthService) *SSOHandler {
	return &SSOHandler{
		queries:     sqlc.New(db),
		authService: authService,
	}
}

func (h *SSOHandler) GoogleLoginHandler(ctx *gin.Context) {
	state := uuid.New().String()
	url := getGoogleOAuthConfig().AuthCodeURL(state)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *SSOHandler) GoogleCallbackHandler(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "code not found"})
		return
	}

	token, err := getGoogleOAuthConfig().Exchange(context.Background(), code)
	if err != nil {
		logger.Error("Failed to exchange code for token", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		logger.Error("Failed to get user info from Google", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	ssoRecord, err := h.queries.GetSSOByProvider(ctx.Request.Context(), sqlc.GetSSOByProviderParams{
		Provider:       "google",
		ProviderUserID: userInfo.ID,
	})

	var user sqlc.User

	if err == sql.ErrNoRows {
		existingUser, err := h.queries.GetUserByEmail(ctx.Request.Context(), userInfo.Email)
		if err == sql.ErrNoRows {
			user, err = h.queries.CreateUser(ctx.Request.Context(), sqlc.CreateUserParams{
				Email:        userInfo.Email,
				PasswordHash: "",
				Name:         userInfo.Name,
				Phone:        "",
				Role:         sqlc.UserRoleCustomer,
				SalonID:      uuid.NullUUID{Valid: false},
			})
			if err != nil {
				logger.Error("Failed to create user from SSO", logger.Err(err), logger.F("email", userInfo.Email))
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
		} else if err != nil {
			logger.Error("Failed to check existing user", logger.Err(err), logger.F("email", userInfo.Email))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		} else {
			user = existingUser
		}

		_, err = h.queries.CreateSSO(ctx.Request.Context(), sqlc.CreateSSOParams{
			UserID:         user.ID,
			Provider:       "google",
			ProviderUserID: userInfo.ID,
		})
		if err != nil {
			logger.Error("Failed to create SSO record", logger.Err(err), logger.F("provider", "google"), logger.F("user_id", user.ID))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create SSO record"})
			return
		}
	} else if err != nil {
		logger.Error("Failed to check SSO record", logger.Err(err), logger.F("provider", "google"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	} else {
		user, err = h.queries.GetUserByID(ctx.Request.Context(), ssoRecord.UserID)
		if err != nil {
			logger.Error("Failed to get user", logger.Err(err), logger.F("user_id", ssoRecord.UserID))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}

	}

	oldToken := auth.GetAccessTokenFromRequest(ctx)
	tokens, err := h.authService.CreateSession(ctx.Request.Context(), user, oldToken)
	if err != nil {
		logger.Error("Failed to create session", logger.Err(err), logger.F("user_id", user.ID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	auth.SetAuthCookies(ctx, tokens)

	frontendURL := os.Getenv("FRONTEND_URL")
	redirectURL := frontendURL + "/auth/callback"

	params := "?onboardingCompleted=" + strconv.FormatBool(user.OnboardingCompletedAt.Valid)
	if user.CurrentOnboardingStepID.Valid {
		params += "&onboardingStep=" + strconv.Itoa(int(user.CurrentOnboardingStepID.Int32))
	}
	redirectURL += params

	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
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
