package auth

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/modules/auth/session"
	"golang.org/x/crypto/bcrypt"
)

// TODO: need to add oauth, token cache, rbac
const (
	bcryptCost = 12
)

var (
	jwtSecret     []byte
	jwtIssuer     string
	JwtExpiration time.Duration
	jwtOnce       sync.Once
)

func initJWTConfig() {
	jwtOnce.Do(func() {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))
		if len(jwtSecret) == 0 {
			jwtSecret = []byte("default-secret-change-in-production")
		}

		jwtIssuer = os.Getenv("JWT_ISSUER")
		if jwtIssuer == "" {
			jwtIssuer = "parlor-api"
		}

		expHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
		if expHours == 0 {
			expHours = 24
		}
		JwtExpiration = time.Duration(expHours) * time.Hour
	})
}

type AuthHandler struct {
	queries        *sqlc.Queries
	sessionService *session.SessionService
}

func NewAuthHandler(db *sql.DB, sessionService *session.SessionService) *AuthHandler {
	return &AuthHandler{
		queries:        sqlc.New(db),
		sessionService: sessionService,
	}
}

type UserResponse struct {
	ID      uuid.UUID  `json:"id"`
	Email   string     `json:"email"`
	Name    string     `json:"name"`
	Phone   string     `json:"phone"`
	Role    string     `json:"role"`
	SalonID *uuid.UUID `json:"salon_id,omitempty"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Role     string `json:"role" binding:"omitempty,oneof=admin staff solo"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type Claims struct {
	UserID  uuid.UUID  `json:"user_id"`
	Email   string     `json:"email"`
	Role    string     `json:"role"`
	SalonID *uuid.UUID `json:"salon_id,omitempty"`
	jwt.RegisteredClaims
}

func toUserResponse(user sqlc.User) UserResponse {
	var salonID *uuid.UUID
	if user.SalonID.Valid {
		salonID = &user.SalonID.UUID
	}
	return UserResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Phone:   user.Phone,
		Role:    string(user.Role),
		SalonID: salonID,
	}
}

func GenerateToken(user sqlc.User) (string, error) {
	initJWTConfig()

	var salonID *uuid.UUID
	if user.SalonID.Valid {
		salonID = &user.SalonID.UUID
	}

	claims := Claims{
		UserID:  user.ID,
		Email:   user.Email,
		Role:    string(user.Role),
		SalonID: salonID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (ah *AuthHandler) RegisterHandler(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingUser, err := ah.queries.GetUserByEmail(ctx.Request.Context(), req.Email)
	if err == nil && existingUser.ID != uuid.Nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking existing user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// TODO: IMPLEMENT SPECIFIC RBAC RULES TO WORKS IN HERE
	role := sqlc.UserRoleSolo
	if req.Role != "" {
		switch req.Role {
		case "admin":
			role = sqlc.UserRoleAdmin
		case "staff":
			role = sqlc.UserRoleStaff
		case "solo":
			role = sqlc.UserRoleSolo
		}
	}

	user, err := ah.queries.CreateUser(ctx.Request.Context(), sqlc.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Phone:        req.Phone,
		Role:         role,
		SalonID:      uuid.NullUUID{Valid: false},
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	token, err := GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	if err := ah.sessionService.StoreToken(ctx.Request.Context(), user.ID.String(), token, JwtExpiration); err != nil {
		log.Printf("Error storing token in Redis: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
		return
	}

	ctx.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	})
}

func (ah *AuthHandler) LoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ah.queries.GetUserByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := GenerateToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	if err := ah.sessionService.StoreToken(ctx.Request.Context(), user.ID.String(), token, JwtExpiration); err != nil {
		log.Printf("Error storing token in Redis: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
		return
	}

	_, _ = ah.queries.CreateSession(ctx.Request.Context(), sqlc.CreateSessionParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(JwtExpiration),
	})

	ctx.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	})
}

func (ah *AuthHandler) MeHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := ah.queries.GetUserByID(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}

func (ah *AuthHandler) LogoutHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
		_ = ah.sessionService.BlacklistToken(ctx.Request.Context(), token, JwtExpiration)
	}

	if err := ah.sessionService.DeleteAllUserTokens(ctx.Request.Context(), userID.(uuid.UUID).String()); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tokens"})
		return
	}

	err := ah.queries.DeleteUserSessions(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
