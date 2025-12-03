package profile

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/logger"
)

type ProfileHandler struct {
	queries *sqlc.Queries
}

func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{
		queries: sqlc.New(db),
	}
}

type UpdateProfileRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=1"`
	Phone *string `json:"phone" binding:"omitempty"`
	Email *string `json:"email" binding:"omitempty,email"`
}

type ProfileResponse struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name"`
	Phone                 string     `json:"phone"`
	Role                  string     `json:"role"`
	SalonID               *uuid.UUID `json:"salon_id,omitempty"`
	OnboardingCompleted   bool       `json:"onboarding_completed"`
	CurrentOnboardingStep *int32     `json:"current_onboarding_step,omitempty"`
}

func toProfileResponse(user sqlc.User) ProfileResponse {
	var salonID *uuid.UUID
	if user.SalonID.Valid {
		salonID = &user.SalonID.UUID
	}

	var currentStep *int32
	if user.CurrentOnboardingStepID.Valid {
		currentStep = &user.CurrentOnboardingStepID.Int32
	}

	return ProfileResponse{
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

func (h *ProfileHandler) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.queries.GetUserByID(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, toProfileResponse(user))
}

func (h *ProfileHandler) UpdateProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == nil && req.Phone == nil && req.Email == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return
	}

	if req.Email != nil {
		existingUser, err := h.queries.GetUserByEmail(ctx.Request.Context(), *req.Email)
		if err == nil && existingUser.ID != userID.(uuid.UUID) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		if err != nil && err != sql.ErrNoRows {
			logger.Error("Failed to check email", logger.Err(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
			return
		}
	}

	params := sqlc.UpdateUserParams{
		ID: userID.(uuid.UUID),
	}

	if req.Name != nil {
		params.Name = sql.NullString{String: *req.Name, Valid: true}
	}
	if req.Phone != nil {
		params.Phone = sql.NullString{String: *req.Phone, Valid: true}
	}
	if req.Email != nil {
		params.Email = sql.NullString{String: *req.Email, Valid: true}
	}

	user, err := h.queries.UpdateUser(ctx.Request.Context(), params)
	if err != nil {
		logger.Error("Failed to update profile", logger.Err(err), logger.F("user_id", userID))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, toProfileResponse(user))
}
