package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/modules/auth/session"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	jwtSecret         []byte
	jwtIssuer         string
	JwtExpiration     time.Duration
	RefreshExpiration time.Duration
	CookieDomain      string
	CookieSecure      bool
	jwtOnce           sync.Once
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type AuthService struct {
	queries        *sqlc.Queries
	sessionService *session.SessionService
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	UserID  uuid.UUID  `json:"user_id"`
	Email   string     `json:"email"`
	Role    string     `json:"role"`
	SalonID *uuid.UUID `json:"salon_id,omitempty"`
	jwt.RegisteredClaims
}

func NewAuthService(db *sql.DB, sessionService *session.SessionService) *AuthService {
	return &AuthService{
		queries:        sqlc.New(db),
		sessionService: sessionService,
	}
}

func InitJWTConfig() {
	jwtOnce.Do(func() {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))

		jwtIssuer = os.Getenv("JWT_ISSUER")
		if jwtIssuer == "" {
			jwtIssuer = "parlor-api"
		}

		expHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
		if expHours == 0 {
			expHours = 24
		}
		JwtExpiration = time.Duration(expHours) * time.Hour

		refreshDays, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_DAYS"))
		if refreshDays <= 0 {
			refreshDays = 7
		}
		RefreshExpiration = time.Duration(refreshDays) * 24 * time.Hour

		CookieDomain = os.Getenv("COOKIE_DOMAIN")
		CookieSecure = os.Getenv("APP_ENV") == "production"
	})
}

// GenerateAccessToken creates a JWT access token for a user
func (s *AuthService) GenerateAccessToken(user sqlc.User) (string, error) {
	InitJWTConfig()

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

// GenerateRefreshToken creates a secure random refresh token
func (s *AuthService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ParseToken parses and validates a JWT token, returns claims even if expired
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	InitJWTConfig()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if token == nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Return claims even if token is expired (for refresh flow)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// Register creates a new user and returns tokens
func (s *AuthService) Register(ctx context.Context, email, password, name, phone, role string) (sqlc.User, error) {
	// Check if user exists
	existingUser, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil && existingUser.ID != uuid.Nil {
		return sqlc.User{}, ErrUserExists
	}
	if err != nil && err != sql.ErrNoRows {
		return sqlc.User{}, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return sqlc.User{}, err
	}

	userRole := sqlc.UserRoleCustomer
	switch role {
	case "admin":
		userRole = sqlc.UserRoleAdmin
	case "owner":
		userRole = sqlc.UserRoleOwner
	case "employee":
		userRole = sqlc.UserRoleEmployee
	case "customer":
		userRole = sqlc.UserRoleCustomer
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Phone:        phone,
		Role:         userRole,
		SalonID:      uuid.NullUUID{Valid: false},
	})
	if err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

// ValidateCredentials checks email/password and returns the user
func (s *AuthService) ValidateCredentials(ctx context.Context, email, password string) (sqlc.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc.User{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return sqlc.User{}, ErrInvalidCredentials
	}

	return user, nil
}

// CreateSession generates tokens, blacklists old ones, and stores new ones in Redis
func (s *AuthService) CreateSession(ctx context.Context, user sqlc.User, oldAccessToken string) (*TokenPair, error) {
	// Blacklist old access token if provided
	if oldAccessToken != "" {
		_ = s.sessionService.BlacklistToken(ctx, oldAccessToken, JwtExpiration)
	}

	// Blacklist any existing token for this user
	existingToken, err := s.sessionService.GetToken(ctx, user.ID.String())
	if err == nil && existingToken != "" {
		_ = s.sessionService.BlacklistToken(ctx, existingToken, JwtExpiration)
	}

	// Generate new tokens
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store tokens in Redis
	if err := s.sessionService.StoreToken(ctx, user.ID.String(), accessToken, JwtExpiration); err != nil {
		return nil, err
	}

	if err := s.sessionService.StoreRefreshToken(ctx, user.ID.String(), refreshToken, RefreshExpiration); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshSession validates refresh token and creates new session
func (s *AuthService) RefreshSession(ctx context.Context, oldAccessToken, refreshToken string) (sqlc.User, *TokenPair, error) {
	// Parse old access token to get user ID (even if expired)
	claims, err := s.ParseToken(oldAccessToken)
	if err != nil {
		return sqlc.User{}, nil, err
	}

	// Verify refresh token matches stored one
	storedRefreshToken, err := s.sessionService.GetRefreshToken(ctx, claims.UserID.String())
	if err != nil || storedRefreshToken != refreshToken {
		return sqlc.User{}, nil, ErrInvalidRefreshToken
	}

	// Get user from database
	user, err := s.queries.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return sqlc.User{}, nil, ErrUserNotFound
	}

	// Create new session (this will blacklist the old token)
	tokens, err := s.CreateSession(ctx, user, oldAccessToken)
	if err != nil {
		return sqlc.User{}, nil, err
	}

	return user, tokens, nil
}

// Logout blacklists token and deletes all user tokens from Redis
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, accessToken string) error {
	if accessToken != "" {
		_ = s.sessionService.BlacklistToken(ctx, accessToken, JwtExpiration)
	}

	return s.sessionService.DeleteAllUserTokens(ctx, userID.String())
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (sqlc.User, error) {
	return s.queries.GetUserByID(ctx, userID)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	return s.sessionService.IsTokenBlacklisted(ctx, token)
}
