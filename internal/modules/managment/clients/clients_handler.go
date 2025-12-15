package clients

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type ClientsHandler struct {
	service *ClientsService
}

func NewClientsHandler(db *sql.DB) *ClientsHandler {
	return &ClientsHandler{
		service: NewClientsService(db),
	}
}

type CreateClientRequest struct {
	SalonID  uuid.UUID  `json:"salon_id" binding:"required"`
	Name     string     `json:"name" binding:"required,min=1"`
	Phone    string     `json:"phone" binding:"required"`
	Email    *string    `json:"email" binding:"omitempty,email"`
	Birthday *time.Time `json:"birthday" binding:"omitempty"`
}

type UpdateClientRequest struct {
	Name     *string    `json:"name" binding:"omitempty,min=1"`
	Phone    *string    `json:"phone" binding:"omitempty"`
	Email    *string    `json:"email" binding:"omitempty,email"`
	Birthday *time.Time `json:"birthday" binding:"omitempty"`
}

type ClientResponse struct {
	ID                uuid.UUID  `json:"id"`
	SalonID           uuid.UUID  `json:"salon_id"`
	Name              string     `json:"name"`
	Phone             string     `json:"phone"`
	Email             *string    `json:"email,omitempty"`
	Birthday          *time.Time `json:"birthday,omitempty"`
	TotalAppointments int32      `json:"total_appointments"`
	LastAppointment   *time.Time `json:"last_appointment,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func toClientResponse(c sqlc.Client) ClientResponse {
	var email *string
	if c.Email.Valid {
		email = &c.Email.String
	}

	var birthday *time.Time
	if c.Birthday.Valid {
		birthday = &c.Birthday.Time
	}

	var lastAppointment *time.Time
	if c.LastAppointment.Valid {
		lastAppointment = &c.LastAppointment.Time
	}

	return ClientResponse{
		ID:                c.ID,
		SalonID:           c.SalonID,
		Name:              c.Name,
		Phone:             c.Phone,
		Email:             email,
		Birthday:          birthday,
		TotalAppointments: c.TotalAppointments,
		LastAppointment:   lastAppointment,
		CreatedAt:         c.CreatedAt,
	}
}

func (h *ClientsHandler) CreateClient(ctx *gin.Context) {
	var req CreateClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.service.CreateClient(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return
	}

	ctx.JSON(http.StatusCreated, toClientResponse(client))
}

func (h *ClientsHandler) GetClient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
		return
	}

	client, err := h.service.GetClientByID(ctx.Request.Context(), id)
	if err != nil {
		switch err {
		case ErrClientNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get client"})
		}
		return
	}

	ctx.JSON(http.StatusOK, toClientResponse(client))
}

func (h *ClientsHandler) ListClients(ctx *gin.Context) {
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

	clients, err := h.service.ListClientsBySalonID(ctx.Request.Context(), salonID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list clients"})
		return
	}

	response := make([]ClientResponse, len(clients))
	for i, c := range clients {
		response[i] = toClientResponse(c)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *ClientsHandler) SearchClients(ctx *gin.Context) {
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

	name := ctx.Query("name")
	phone := ctx.Query("phone")

	var clients []sqlc.Client

	if name != "" {
		clients, err = h.service.SearchClientsByName(ctx.Request.Context(), salonID, name)
	} else if phone != "" {
		clients, err = h.service.SearchClientsByPhone(ctx.Request.Context(), salonID, phone)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name or phone query parameter is required"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search clients"})
		return
	}

	response := make([]ClientResponse, len(clients))
	for i, c := range clients {
		response[i] = toClientResponse(c)
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *ClientsHandler) UpdateClient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
		return
	}

	var req UpdateClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == nil && req.Phone == nil && req.Email == nil && req.Birthday == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return
	}

	client, err := h.service.UpdateClient(ctx.Request.Context(), id, req)
	if err != nil {
		switch err {
		case ErrClientNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update client"})
		}
		return
	}

	ctx.JSON(http.StatusOK, toClientResponse(client))
}

func (h *ClientsHandler) DeleteClient(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid client id"})
		return
	}

	if err := h.service.DeleteClient(ctx.Request.Context(), id); err != nil {
		switch err {
		case ErrClientNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete client"})
		}
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *ClientsHandler) GetClientByPhone(ctx *gin.Context) {
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

	client, err := h.service.GetClientByPhone(ctx.Request.Context(), salonID, phone)
	if err != nil {
		switch err {
		case ErrClientNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get client"})
		}
		return
	}

	ctx.JSON(http.StatusOK, toClientResponse(client))
}
