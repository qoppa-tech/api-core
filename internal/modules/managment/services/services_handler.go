package services

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type ServicesHandler struct {
	queries *sqlc.Queries
}

func NewServicesHandler(db *sql.DB) *ServicesHandler {
	return &ServicesHandler{
		queries: sqlc.New(db),
	}
}

type CreateServiceRequest struct {
	SalonID  uuid.UUID  `json:"salon_id" binding:"required"`
	UserID   *uuid.UUID `json:"user_id" binding:"omitempty"`
	Name     string     `json:"name" binding:"required,min=1"`
	Duration int32      `json:"duration" binding:"required,gt=0"`
	Price    int32      `json:"price" binding:"required,gte=0"`
	Active   *bool      `json:"active" binding:"omitempty"`
}

type UpdateServiceRequest struct {
	Name     *string    `json:"name" binding:"omitempty,min=1"`
	Duration *int32     `json:"duration" binding:"omitempty,gt=0"`
	Price    *int32     `json:"price" binding:"omitempty,gte=0"`
	Active   *bool      `json:"active" binding:"omitempty"`
	UserID   *uuid.UUID `json:"user_id" binding:"omitempty"`
}

type ServiceResponse struct {
	ID        uuid.UUID  `json:"id"`
	SalonID   uuid.UUID  `json:"salon_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Name      string     `json:"name"`
	Duration  int32      `json:"duration"`
	Price     int32      `json:"price"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
}

func toServiceResponse(s sqlc.Service) ServiceResponse {
	var userID *uuid.UUID
	if s.UserID.Valid {
		userID = &s.UserID.UUID
	}

	return ServiceResponse{
		ID:        s.ID,
		SalonID:   s.SalonID,
		UserID:    userID,
		Name:      s.Name,
		Duration:  s.Duration,
		Price:     s.Price,
		Active:    s.Active,
		CreatedAt: s.CreatedAt,
	}
}

// CreateService creates a new service
// @Summary Create a new service
// @Tags services
// @Accept json
// @Produce json
// @Param request body CreateServiceRequest true "Service data"
// @Success 201 {object} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services [post]
func (h *ServicesHandler) CreateService(ctx *gin.Context) {
	var req CreateServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.CreateServiceParams{
		SalonID:  req.SalonID,
		Name:     req.Name,
		Duration: req.Duration,
		Price:    req.Price,
		Active:   true,
	}

	if req.UserID != nil {
		params.UserID = uuid.NullUUID{UUID: *req.UserID, Valid: true}
	}

	if req.Active != nil {
		params.Active = *req.Active
	}

	service, err := h.queries.CreateService(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service"})
		return
	}

	ctx.JSON(http.StatusCreated, toServiceResponse(service))
}

// GetService gets a service by ID
// @Summary Get a service by ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/{id} [get]
func (h *ServicesHandler) GetService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	service, err := h.queries.GetServiceByID(ctx.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get service"})
		return
	}

	ctx.JSON(http.StatusOK, toServiceResponse(service))
}

// ListServices lists services by salon ID
// @Summary List services by salon ID
// @Tags services
// @Produce json
// @Param salon_id query string true "Salon ID"
// @Success 200 {array} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services [get]
func (h *ServicesHandler) ListServices(ctx *gin.Context) {
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

	services, err := h.queries.ListServicesBySalonID(ctx.Request.Context(), salonID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list services"})
		return
	}

	response := make([]ServiceResponse, len(services))
	for i, s := range services {
		response[i] = toServiceResponse(s)
	}

	ctx.JSON(http.StatusOK, response)
}

// ListActiveServices lists active services by salon ID
// @Summary List active services by salon ID
// @Tags services
// @Produce json
// @Param salon_id query string true "Salon ID"
// @Success 200 {array} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/active [get]
func (h *ServicesHandler) ListActiveServices(ctx *gin.Context) {
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

	services, err := h.queries.ListActiveServicesBySalonID(ctx.Request.Context(), salonID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list active services"})
		return
	}

	response := make([]ServiceResponse, len(services))
	for i, s := range services {
		response[i] = toServiceResponse(s)
	}

	ctx.JSON(http.StatusOK, response)
}

// ListServicesByUser lists services by user ID
// @Summary List services by user ID
// @Tags services
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {array} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/by-user [get]
func (h *ServicesHandler) ListServicesByUser(ctx *gin.Context) {
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

	services, err := h.queries.ListServicesByUserID(ctx.Request.Context(), uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list services"})
		return
	}

	response := make([]ServiceResponse, len(services))
	for i, s := range services {
		response[i] = toServiceResponse(s)
	}

	ctx.JSON(http.StatusOK, response)
}

// ListServicesBySalonAndUser lists services by salon and user ID
// @Summary List services by salon and user ID (includes services with no user assigned)
// @Tags services
// @Produce json
// @Param salon_id query string true "Salon ID"
// @Param user_id query string true "User ID"
// @Success 200 {array} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/by-salon-and-user [get]
func (h *ServicesHandler) ListServicesBySalonAndUser(ctx *gin.Context) {
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

	services, err := h.queries.ListServicesBySalonAndUser(ctx.Request.Context(), sqlc.ListServicesBySalonAndUserParams{
		SalonID: salonID,
		UserID:  uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list services"})
		return
	}

	response := make([]ServiceResponse, len(services))
	for i, s := range services {
		response[i] = toServiceResponse(s)
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateService updates a service
// @Summary Update a service
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param request body UpdateServiceRequest true "Service data to update"
// @Success 200 {object} ServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/{id} [patch]
func (h *ServicesHandler) UpdateService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	var req UpdateServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.UpdateServiceParams{
		ID: id,
	}

	if req.Name != nil {
		params.Name = sql.NullString{String: *req.Name, Valid: true}
	}

	if req.Duration != nil {
		params.Duration = sql.NullInt32{Int32: *req.Duration, Valid: true}
	}

	if req.Price != nil {
		params.Price = sql.NullInt32{Int32: *req.Price, Valid: true}
	}

	if req.Active != nil {
		params.Active = sql.NullBool{Bool: *req.Active, Valid: true}
	}

	if req.UserID != nil {
		params.UserID = uuid.NullUUID{UUID: *req.UserID, Valid: true}
	}

	service, err := h.queries.UpdateService(ctx.Request.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service"})
		return
	}

	ctx.JSON(http.StatusOK, toServiceResponse(service))
}

// DeleteService deletes a service
// @Summary Delete a service
// @Tags services
// @Param id path string true "Service ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/{id} [delete]
func (h *ServicesHandler) DeleteService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	err = h.queries.DeleteService(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete service"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
