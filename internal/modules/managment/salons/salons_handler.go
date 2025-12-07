package salons

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/parlorhub/api-core/internal/database/sqlc"
	"github.com/sqlc-dev/pqtype"
)

type SalonsHandler struct {
	queries *sqlc.Queries
}

func NewSalonsHandler(db *sql.DB) *SalonsHandler {
	return &SalonsHandler{
		queries: sqlc.New(db),
	}
}

type CreateSalonRequest struct {
	Name          string          `json:"name" binding:"required,min=1"`
	Slug          string          `json:"slug" binding:"required,min=1"`
	OwnerID       uuid.UUID       `json:"owner_id" binding:"required"`
	Address       string          `json:"address" binding:"required"`
	Whatsapp      *string         `json:"whatsapp" binding:"omitempty"`
	BusinessHours json.RawMessage `json:"business_hours" binding:"omitempty"`
	LogoUrl       *string         `json:"logo_url" binding:"omitempty"`
}

type UpdateSalonRequest struct {
	Name          *string         `json:"name" binding:"omitempty,min=1"`
	Slug          *string         `json:"slug" binding:"omitempty,min=1"`
	Address       *string         `json:"address" binding:"omitempty"`
	Whatsapp      *string         `json:"whatsapp" binding:"omitempty"`
	BusinessHours json.RawMessage `json:"business_hours" binding:"omitempty"`
	LogoUrl       *string         `json:"logo_url" binding:"omitempty"`
}

type SalonResponse struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Slug          string          `json:"slug"`
	OwnerID       uuid.UUID       `json:"owner_id"`
	Address       string          `json:"address"`
	Whatsapp      *string         `json:"whatsapp,omitempty"`
	Timezone      string          `json:"timezone"`
	BusinessHours json.RawMessage `json:"business_hours,omitempty"`
	LogoUrl       *string         `json:"logo_url,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func toSalonResponse(s sqlc.Salon) SalonResponse {
	var whatsapp *string
	if s.Whatsapp.Valid {
		whatsapp = &s.Whatsapp.String
	}

	var logoUrl *string
	if s.LogoUrl.Valid {
		logoUrl = &s.LogoUrl.String
	}

	return SalonResponse{
		ID:            s.ID,
		Name:          s.Name,
		Slug:          s.Slug,
		OwnerID:       s.OwnerID,
		Address:       s.Address,
		Whatsapp:      whatsapp,
		Timezone:      s.Timezone,
		BusinessHours: s.BusinessHours,
		LogoUrl:       logoUrl,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

// CreateSalon creates a new salon
// @Summary Create a new salon
// @Tags salons
// @Accept json
// @Produce json
// @Param request body CreateSalonRequest true "Salon data"
// @Success 201 {object} SalonResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /salons [post]
func (h *SalonsHandler) CreateSalon(ctx *gin.Context) {
	var req CreateSalonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.CreateSalonParams{
		Name:    req.Name,
		Slug:    req.Slug,
		OwnerID: req.OwnerID,
		Address: req.Address,
	}

	if req.Whatsapp != nil {
		params.Whatsapp = sql.NullString{String: *req.Whatsapp, Valid: true}
	}

	if req.BusinessHours != nil {
		params.BusinessHours = req.BusinessHours
	} else {
		params.BusinessHours = json.RawMessage(`{}`)
	}

	if req.LogoUrl != nil {
		params.LogoUrl = sql.NullString{String: *req.LogoUrl, Valid: true}
	}

	salon, err := h.queries.CreateSalon(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create salon"})
		return
	}

	ctx.JSON(http.StatusCreated, toSalonResponse(salon))
}

// GetSalon gets a salon by ID
// @Summary Get a salon by ID
// @Tags salons
// @Produce json
// @Param id path string true "Salon ID"
// @Success 200 {object} SalonResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /salons/{id} [get]
func (h *SalonsHandler) GetSalon(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon id"})
		return
	}

	salon, err := h.queries.GetSalonByID(ctx.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "salon not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get salon"})
		return
	}

	ctx.JSON(http.StatusOK, toSalonResponse(salon))
}

// GetSalonBySlug gets a salon by slug
// @Summary Get a salon by slug
// @Tags salons
// @Produce json
// @Param slug query string true "Salon slug"
// @Success 200 {object} SalonResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /salons/by-slug [get]
func (h *SalonsHandler) GetSalonBySlug(ctx *gin.Context) {
	slug := ctx.Query("slug")
	if slug == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	salon, err := h.queries.GetSalonBySlug(ctx.Request.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "salon not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get salon"})
		return
	}

	ctx.JSON(http.StatusOK, toSalonResponse(salon))
}

// GetSalonByOwner gets a salon by owner ID
// @Summary Get a salon by owner ID
// @Tags salons
// @Produce json
// @Param owner_id query string true "Owner ID"
// @Success 200 {object} SalonResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /salons/by-owner [get]
func (h *SalonsHandler) GetSalonByOwner(ctx *gin.Context) {
	ownerIDStr := ctx.Query("owner_id")
	if ownerIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
		return
	}

	salon, err := h.queries.GetSalonByOwnerID(ctx.Request.Context(), ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "salon not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get salon"})
		return
	}

	ctx.JSON(http.StatusOK, toSalonResponse(salon))
}

// ListSalons lists all salons
// @Summary List all salons
// @Tags salons
// @Produce json
// @Success 200 {array} SalonResponse
// @Failure 500 {object} map[string]string
// @Router /salons [get]
func (h *SalonsHandler) ListSalons(ctx *gin.Context) {
	salons, err := h.queries.ListSalons(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list salons"})
		return
	}

	response := make([]SalonResponse, len(salons))
	for i, s := range salons {
		response[i] = toSalonResponse(s)
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateSalon updates a salon
// @Summary Update a salon
// @Tags salons
// @Accept json
// @Produce json
// @Param id path string true "Salon ID"
// @Param request body UpdateSalonRequest true "Salon data to update"
// @Success 200 {object} SalonResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /salons/{id} [patch]
func (h *SalonsHandler) UpdateSalon(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon id"})
		return
	}

	var req UpdateSalonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.UpdateSalonParams{
		ID: id,
	}

	if req.Name != nil {
		params.Name = sql.NullString{String: *req.Name, Valid: true}
	}

	if req.Slug != nil {
		params.Slug = sql.NullString{String: *req.Slug, Valid: true}
	}

	if req.Address != nil {
		params.Address = sql.NullString{String: *req.Address, Valid: true}
	}

	if req.Whatsapp != nil {
		params.Whatsapp = sql.NullString{String: *req.Whatsapp, Valid: true}
	}

	if req.BusinessHours != nil {
		params.BusinessHours = pqtype.NullRawMessage{RawMessage: req.BusinessHours, Valid: true}
	}

	if req.LogoUrl != nil {
		params.LogoUrl = sql.NullString{String: *req.LogoUrl, Valid: true}
	}

	salon, err := h.queries.UpdateSalon(ctx.Request.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "salon not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update salon"})
		return
	}

	ctx.JSON(http.StatusOK, toSalonResponse(salon))
}

// DeleteSalon deletes a salon
// @Summary Delete a salon
// @Tags salons
// @Param id path string true "Salon ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /salons/{id} [delete]
func (h *SalonsHandler) DeleteSalon(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid salon id"})
		return
	}

	err = h.queries.DeleteSalon(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete salon"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
