package onboarding

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/parlorhub/api-core/internal/logger"
)

type OnboardingHandler struct {
	queries *sqlc.Queries
}

func NewOnboardingHandler(db *sql.DB) *OnboardingHandler {
	return &OnboardingHandler{
		queries: sqlc.New(db),
	}
}

type OnboardingOption struct {
	ID          int32 `json:"id"`
	OptionIndex int32 `json:"option_index"`
}

type OnboardingStepResponse struct {
	ID      int32              `json:"id"`
	Options []OnboardingOption `json:"options"`
}

type UserSelection struct {
	StepID        int32   `json:"step_id"`
	OptionIndices []int32 `json:"option_indices"`
}

type OnboardingProgressResponse struct {
	Steps          []OnboardingStepResponse `json:"steps"`
	UserSelections []UserSelection          `json:"user_selections"`
	CurrentStep    *int32                   `json:"current_step"`
	Completed      bool                     `json:"completed"`
}

type OnboardingSelectionRequest struct {
	StepID        int32   `json:"step_id" binding:"required"`
	OptionIndices []int32 `json:"option_indices" binding:"required,min=1"`
}

type SaveOnboardingRequest struct {
	Selections []OnboardingSelectionRequest `json:"selections" binding:"required,min=1"`
	Completed  bool                         `json:"completed"`
}

func (h *OnboardingHandler) GetOnboarding(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	steps, err := h.queries.GetOnboardingSteps(ctx.Request.Context())
	if err != nil {
		logger.Error("Failed to get onboarding steps", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get onboarding steps"})
		return
	}

	allOptions, err := h.queries.GetAllOnboardingOptions(ctx.Request.Context())
	if err != nil {
		logger.Error("Failed to get onboarding options", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get onboarding options"})
		return
	}

	userSelections, err := h.queries.GetUserOnboardingSelections(ctx.Request.Context(), uuid.NullUUID{UUID: userID.(uuid.UUID), Valid: true})
	if err != nil && err != sql.ErrNoRows {
		logger.Error("Failed to get user onboarding selections", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user selections"})
		return
	}

	user, err := h.queries.GetUserByID(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		logger.Error("Failed to get user", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	response := buildOnboardingResponse(steps, allOptions, userSelections, user)

	ctx.JSON(http.StatusOK, response)
}

func (h *OnboardingHandler) SaveOnboarding(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SaveOnboardingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByID(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		logger.Error("Failed to get user", logger.Err(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	salonID := user.ID
	if user.SalonID.Valid {
		salonID = user.SalonID.UUID
	}

	for _, selection := range req.Selections {
		err := h.queries.DeleteUserOnboardingSelections(ctx.Request.Context(),
			sqlc.DeleteUserOnboardingSelectionsParams{
				UserID: uuid.NullUUID{UUID: userID.(uuid.UUID), Valid: true},
				StepID: selection.StepID,
			})
		if err != nil {
			logger.Error("Failed to delete old selections", logger.Err(err),
				logger.F("step_id", selection.StepID))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save selections"})
			return
		}

		for _, optionIndex := range selection.OptionIndices {
			err := h.queries.SaveOnboardingSelection(ctx.Request.Context(),
				sqlc.SaveOnboardingSelectionParams{
					UserID:      uuid.NullUUID{UUID: userID.(uuid.UUID), Valid: true},
					SalonID:     salonID,
					StepID:      selection.StepID,
					OptionIndex: optionIndex,
				})
			if err != nil {
				logger.Error("Failed to save onboarding selection", logger.Err(err),
					logger.F("step_id", selection.StepID),
					logger.F("option_index", optionIndex))
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save selections"})
				return
			}
		}

		err = h.queries.UpdateUserOnboardingStep(ctx.Request.Context(),
			sqlc.UpdateUserOnboardingStepParams{
				ID:                      userID.(uuid.UUID),
				CurrentOnboardingStepID: sql.NullInt32{Int32: selection.StepID, Valid: true},
			})
		if err != nil {
			logger.Warn("Failed to update user onboarding step", logger.Err(err))
		}
	}

	if req.Completed {
		maxStepID := int32(0)
		for _, selection := range req.Selections {
			if selection.StepID > maxStepID {
				maxStepID = selection.StepID
			}
		}

		err := h.queries.CompleteUserOnboarding(ctx.Request.Context(),
			sqlc.CompleteUserOnboardingParams{
				ID:                      userID.(uuid.UUID),
				CurrentOnboardingStepID: sql.NullInt32{Int32: maxStepID, Valid: true},
			})
		if err != nil {
			logger.Error("Failed to complete user onboarding", logger.Err(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete onboarding"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "onboarding saved successfully"})
}

func buildOnboardingResponse(
	steps []int32,
	allOptions []sqlc.OnboardingOption,
	userSelections []sqlc.SalonOnboardingSelection,
	user sqlc.User,
) OnboardingProgressResponse {
	optionsByStep := make(map[int32][]OnboardingOption)
	for _, opt := range allOptions {
		optionsByStep[opt.StepID] = append(optionsByStep[opt.StepID], OnboardingOption{
			ID:          opt.ID,
			OptionIndex: opt.OptionIndex,
		})
	}

	stepResponses := make([]OnboardingStepResponse, 0, len(steps))
	for _, stepID := range steps {
		stepResponses = append(stepResponses, OnboardingStepResponse{
			ID:      stepID,
			Options: optionsByStep[stepID],
		})
	}

	selectionsByStep := make(map[int32][]int32)
	for _, sel := range userSelections {
		selectionsByStep[sel.StepID] = append(selectionsByStep[sel.StepID], sel.OptionIndex)
	}

	userSelectionResponses := make([]UserSelection, 0)
	for stepID, optionIndices := range selectionsByStep {
		userSelectionResponses = append(userSelectionResponses, UserSelection{
			StepID:        stepID,
			OptionIndices: optionIndices,
		})
	}

	var currentStep *int32
	if user.CurrentOnboardingStepID.Valid {
		currentStep = &user.CurrentOnboardingStepID.Int32
	}

	completed := user.OnboardingCompletedAt.Valid

	return OnboardingProgressResponse{
		Steps:          stepResponses,
		UserSelections: userSelectionResponses,
		CurrentStep:    currentStep,
		Completed:      completed,
	}
}
