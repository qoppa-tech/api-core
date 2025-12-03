package auth

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/logger"
	"github.com/parlorhub/api-core/internal/modules/auth/session"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(db *sql.DB, sessionService *session.SessionService) *AuthHandler {
	return &AuthHandler{
		service: NewAuthService(db, sessionService),
	}
}

type UserResponse struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name"`
	Phone                 string     `json:"phone"`
	Role                  string     `json:"role"`
	SalonID               *uuid.UUID `json:"salon_id,omitempty"`
	OnboardingCompleted   bool       `json:"onboarding_completed"`
	CurrentOnboardingStep *int32     `json:"current_onboarding_step,omitempty"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Role     string `json:"role" binding:"omitempty,oneof=admin owner employee customer"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}

func SetAuthCookies(ctx *gin.Context, tokens *TokenPair) {
	InitJWTConfig()

	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(
		AccessTokenCookieName,
		tokens.AccessToken,
		int(JwtExpiration.Seconds()),
		"/",
		CookieDomain,
		CookieSecure,
		true,
	)

	ctx.SetCookie(
		RefreshTokenCookieName,
		tokens.RefreshToken,
		int(RefreshExpiration.Seconds()),
		"/auth/refresh",
		CookieDomain,
		CookieSecure,
		true,
	)
}

func ClearAuthCookies(ctx *gin.Context) {
	InitJWTConfig()

	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(AccessTokenCookieName, "", -1, "/", CookieDomain, CookieSecure, true)
	ctx.SetCookie(RefreshTokenCookieName, "", -1, "/auth/refresh", CookieDomain, CookieSecure, true)
}

func GetAccessTokenFromRequest(ctx *gin.Context) string {
	if token, err := ctx.Cookie(AccessTokenCookieName); err == nil && token != "" {
		return token
	}

	authHeader := ctx.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}

	return ""
}

func toUserResponse(user sqlc.User) UserResponse {
	var salonID *uuid.UUID
	if user.SalonID.Valid {
		salonID = &user.SalonID.UUID
	}

	var currentStep *int32
	if user.CurrentOnboardingStepID.Valid {
		currentStep = &user.CurrentOnboardingStepID.Int32
	}

	return UserResponse{
		ID:                    user.ID,
		Email:                 user.Email,
		Name:                  user.Name,
		Phone:                 user.Phone,
		Role:                  string(user.Role),
		SalonID:               salonID,
		OnboardingCompleted:   user.OnboardingCompletedAt.Valid,
		CurrentOnboardingStep: currentStep,
	}
}

func (h *AuthHandler) RegisterHandler(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Register(ctx.Request.Context(), req.Email, req.Password, req.Name, req.Phone, req.Role)
	if err == ErrUserExists {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}
	if err != nil {
		logger.Error("Failed to register user", logger.Err(err), logger.F("email", req.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	h.loginUser(ctx, user, http.StatusCreated)
}

func (h *AuthHandler) LoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.ValidateCredentials(ctx.Request.Context(), req.Email, req.Password)
	if err == ErrInvalidCredentials {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err != nil {
		logger.Error("Failed to validate credentials", logger.Err(err), logger.F("email", req.Email))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	h.loginUser(ctx, user, http.StatusOK)
}

func (h *AuthHandler) loginUser(ctx *gin.Context, user sqlc.User, statusCode int) {
	oldToken := GetAccessTokenFromRequest(ctx)

	tokens, err := h.service.CreateSession(ctx.Request.Context(), user, oldToken)
	if err != nil {
		logger.Error("Failed to create session", logger.Err(err), logger.F("user_id", user.ID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	SetAuthCookies(ctx, tokens)

	ctx.JSON(statusCode, AuthResponse{
		User: toUserResponse(user),
	})
}

func (h *AuthHandler) MeHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.service.GetUserByID(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) LogoutHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	token := GetAccessTokenFromRequest(ctx)

	if err := h.service.Logout(ctx.Request.Context(), userID.(uuid.UUID), token); err != nil {
		logger.Warn("Error during logout", logger.Err(err), logger.F("user_id", userID))
	}

	ClearAuthCookies(ctx)

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) RefreshHandler(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie(RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	oldAccessToken := GetAccessTokenFromRequest(ctx)
	if oldAccessToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}

	user, tokens, err := h.service.RefreshSession(ctx.Request.Context(), oldAccessToken, refreshToken)
	if err == ErrInvalidToken || err == ErrInvalidRefreshToken {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	if err == ErrUserNotFound {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		logger.Error("Failed to refresh session", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh session"})
		return
	}

	SetAuthCookies(ctx, tokens)

	ctx.JSON(http.StatusOK, AuthResponse{
		User: toUserResponse(user),
	})
}

func (h *AuthHandler) GetService() *AuthService {
	return h.service
}
