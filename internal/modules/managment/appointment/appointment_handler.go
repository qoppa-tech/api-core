package appointment

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type AppointmentHandler struct {
	service *AppointmentService
}

func NewAppointmentHandler(db *sql.DB) *AppointmentHandler {
	return &AppointmentHandler{
		service: NewAppointmentService(db),
	}
}

type CreateAppointmentRequest struct {
	SalonID     uuid.UUID `json:"salon_id" binding:"required"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	ServiceID   uuid.UUID `json:"service_id" binding:"required"`
	ClientName  string    `json:"client_name" binding:"required,min=1"`
	ClientPhone string    `json:"client_phone" binding:"required"`
	ClientEmail *string   `json:"client_email" binding:"omitempty,email"`
	Date        string    `json:"date" binding:"required"`
	StartTime   string    `json:"start_time" binding:"required"`
	EndTime     string    `json:"end_time" binding:"required"`
	Status      *string   `json:"status" binding:"omitempty"`
	Notes       *string   `json:"notes" binding:"omitempty"`
}

type UpdateAppointmentRequest struct {
	Date      *string `json:"date" binding:"omitempty"`
	StartTime *string `json:"start_time" binding:"omitempty"`
	EndTime   *string `json:"end_time" binding:"omitempty"`
	Status    *string `json:"status" binding:"omitempty"`
	Notes     *string `json:"notes" binding:"omitempty"`
}

type UpdateAppointmentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type AppointmentResponse struct {
	ID          uuid.UUID `json:"id"`
	SalonID     uuid.UUID `json:"salon_id"`
	UserID      uuid.UUID `json:"user_id"`
	ServiceID   uuid.UUID `json:"service_id"`
	ClientName  string    `json:"client_name"`
	ClientPhone string    `json:"client_phone"`
	ClientEmail *string   `json:"client_email,omitempty"`
	Date        string    `json:"date"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	Status      string    `json:"status"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func toAppointmentResponse(a sqlc.Appointment) AppointmentResponse {
	var clientEmail *string
	if a.ClientEmail.Valid {
		clientEmail = &a.ClientEmail.String
	}

	var notes *string
	if a.Notes.Valid {
		notes = &a.Notes.String
	}

	return AppointmentResponse{
		ID:          a.ID,
		SalonID:     a.SalonID,
		UserID:      a.UserID,
		ServiceID:   a.ServiceID,
		ClientName:  a.ClientName,
		ClientPhone: a.ClientPhone,
		ClientEmail: clientEmail,
		Date:        a.Date.Format("2006-01-02"),
		StartTime:   a.StartTime.Format("15:04"),
		EndTime:     a.EndTime.Format("15:04"),
		Status:      string(a.Status),
		Notes:       notes,
		CreatedAt:   a.CreatedAt,
	}
}

func (h *AppointmentHandler) CreateAppointment(ctx *gin.Context) {
	var req CreateAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use HH:MM"})
		return
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, use HH:MM"})
		return
	}

	status := sqlc.AppointmentStatusPending
	if req.Status != nil {
		status = sqlc.AppointmentStatus(*req.Status)
	}

	params := sqlc.CreateAppointmentParams{
		SalonID:     req.SalonID,
		UserID:      req.UserID,
		ServiceID:   req.ServiceID,
		ClientName:  req.ClientName,
		ClientPhone: req.ClientPhone,
		Date:        date,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      status,
	}

	if req.ClientEmail != nil {
		params.ClientEmail = sql.NullString{String: *req.ClientEmail, Valid: true}
	}

	if req.Notes != nil {
		params.Notes = sql.NullString{String: *req.Notes, Valid: true}
	}

	appointment, err := h.service.CreateAppointment(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create appointment"})
		return
	}

	ctx.JSON(http.StatusCreated, toAppointmentResponse(appointment))
}

func (h *AppointmentHandler) GetAppointment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment id"})
		return
	}

	appointment, err := h.service.GetAppointmentByID(ctx.Request.Context(), id)
	if err != nil {
		switch err {
		case ErrAppointmentNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get appointment"})
		}
		return
	}

	ctx.JSON(http.StatusOK, toAppointmentResponse(appointment))
}

func (h *AppointmentHandler) ListAppointments(ctx *gin.Context) {
	salonIDStr := ctx.Query("salon_id")
	if salonIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "salon_id is required"})
		return
	}

	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon_id"})
		return
	}

	appointments, err := h.service.ListAppointmentsBySalonID(ctx.Request.Context(), salonID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list appointments"})
		return
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, a := range appointments {
		response[i] = toAppointmentResponse(a)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *AppointmentHandler) ListAppointmentsByDate(ctx *gin.Context) {
	salonIDStr := ctx.Query("salon_id")
	if salonIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "salon_id is required"})
		return
	}

	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon_id"})
		return
	}

	dateStr := ctx.Query("date")
	if dateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	appointments, err := h.service.ListAppointmentsByDate(ctx.Request.Context(), sqlc.ListAppointmentsByDateParams{
		SalonID: salonID,
		Date:    date,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list appointments"})
		return
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, a := range appointments {
		response[i] = toAppointmentResponse(a)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *AppointmentHandler) ListAppointmentsByDateRange(ctx *gin.Context) {
	salonIDStr := ctx.Query("salon_id")
	if salonIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "salon_id is required"})
		return
	}

	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon_id"})
		return
	}

	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
		return
	}

	appointments, err := h.service.ListAppointmentsByDateRange(ctx.Request.Context(), sqlc.ListAppointmentsByDateRangeParams{
		SalonID: salonID,
		Date:    startDate,
		Date_2:  endDate,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list appointments"})
		return
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, a := range appointments {
		response[i] = toAppointmentResponse(a)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *AppointmentHandler) ListAppointmentsByUser(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	dateStr := ctx.Query("date")
	if dateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	appointments, err := h.service.ListAppointmentsByUser(ctx.Request.Context(), sqlc.ListAppointmentsByUserParams{
		UserID: userID,
		Date:   date,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list appointments"})
		return
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, a := range appointments {
		response[i] = toAppointmentResponse(a)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *AppointmentHandler) ListAppointmentsByStatus(ctx *gin.Context) {
	salonIDStr := ctx.Query("salon_id")
	if salonIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "salon_id is required"})
		return
	}

	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon_id"})
		return
	}

	status := ctx.Query("status")
	if status == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	appointments, err := h.service.ListAppointmentsByStatus(ctx.Request.Context(), sqlc.ListAppointmentsByStatusParams{
		SalonID: salonID,
		Status:  sqlc.AppointmentStatus(status),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list appointments"})
		return
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, a := range appointments {
		response[i] = toAppointmentResponse(a)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *AppointmentHandler) ListAppointmentsByClientPhone(ctx *gin.Context) {
	salonIDStr := ctx.Query("salon_id")
	if salonIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "salon_id is required"})
		return
	}

	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon_id"})
		return
	}

	phone := ctx.Query("phone")
	if phone == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "phone is required"})
		return
	}

	appointments, err := h.service.ListAppointmentsByClientPhone(ctx.Request.Context(), sqlc.ListAppointmentsByClientPhoneParams{
		SalonID:     salonID,
		ClientPhone: phone,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list appointments"})
		return
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, a := range appointments {
		response[i] = toAppointmentResponse(a)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *AppointmentHandler) UpdateAppointment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment id"})
		return
	}

	var req UpdateAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Date == nil && req.StartTime == nil && req.EndTime == nil && req.Status == nil && req.Notes == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return
	}

	params := sqlc.UpdateAppointmentParams{
		ID: id,
	}

	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		params.Date = sql.NullTime{Time: date, Valid: true}
	}

	if req.StartTime != nil {
		startTime, err := time.Parse("15:04", *req.StartTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use HH:MM"})
			return
		}
		params.StartTime = sql.NullTime{Time: startTime, Valid: true}
	}

	if req.EndTime != nil {
		endTime, err := time.Parse("15:04", *req.EndTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, use HH:MM"})
			return
		}
		params.EndTime = sql.NullTime{Time: endTime, Valid: true}
	}

	if req.Status != nil {
		params.Status = sqlc.NullAppointmentStatus{
			AppointmentStatus: sqlc.AppointmentStatus(*req.Status),
			Valid:             true,
		}
	}

	if req.Notes != nil {
		params.Notes = sql.NullString{String: *req.Notes, Valid: true}
	}

	appointment, err := h.service.UpdateAppointment(ctx.Request.Context(), params)
	if err != nil {
		switch err {
		case ErrAppointmentNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update appointment"})
		}
		return
	}

	ctx.JSON(http.StatusOK, toAppointmentResponse(appointment))
}

func (h *AppointmentHandler) UpdateAppointmentStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment id"})
		return
	}

	var req UpdateAppointmentStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointment, err := h.service.UpdateAppointmentStatus(ctx.Request.Context(), sqlc.UpdateAppointmentStatusParams{
		ID:     id,
		Status: sqlc.AppointmentStatus(req.Status),
	})
	if err != nil {
		switch err {
		case ErrAppointmentNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update appointment status"})
		}
		return
	}

	ctx.JSON(http.StatusOK, toAppointmentResponse(appointment))
}

func (h *AppointmentHandler) DeleteAppointment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment id"})
		return
	}

	if err := h.service.DeleteAppointment(ctx.Request.Context(), id); err != nil {
		switch err {
		case ErrAppointmentNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete appointment"})
		}
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *AppointmentHandler) CountAppointmentsByUserAndDate(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	dateStr := ctx.Query("date")
	if dateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	count, err := h.service.CountAppointmentsByUserAndDate(ctx.Request.Context(), sqlc.CountAppointmentsByUserAndDateParams{
		UserID: userID,
		Date:   date,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count appointments"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}
