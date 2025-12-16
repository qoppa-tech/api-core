package dashboard

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{service: NewService(db)}
}

func parseMonth(ctx *gin.Context) (time.Time, bool) {
	yearStr := ctx.Query("year")
	monthStr := ctx.Query("month")
	if yearStr == "" || monthStr == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), true
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return time.Time{}, false
	}

	monthInt, err := strconv.Atoi(monthStr)
	if err != nil || monthInt < 1 || monthInt > 12 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return time.Time{}, false
	}

	return time.Date(year, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC), true
}

// GetSummary aggregates monthly KPIs for the dashboard.
func (h *Handler) GetSummary(ctx *gin.Context) {
	var salonID uuid.UUID

	if salonValue, ok := ctx.Get("salon_id"); ok {
		id, ok := salonValue.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		salonID = id
	} else {
		salonIDStr := ctx.Query("salon_id")
		if salonIDStr == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "salon_id is required"})
			return
		}

		id, err := uuid.Parse(salonIDStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon_id"})
			return
		}
		salonID = id
	}

	userValue, ok := ctx.Get("user_id")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userValue.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	month, ok := parseMonth(ctx)
	if !ok {
		return
	}

	if err := h.service.ensureSalonOwnership(ctx.Request.Context(), salonID, userID); err != nil {
		switch err {
		case ErrSalonNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "salon not found"})
		case ErrSalonForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch salon"})
		}
		return
	}

	summary, err := h.service.GetSummary(ctx.Request.Context(), salonID, month)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch dashboard data"})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}
