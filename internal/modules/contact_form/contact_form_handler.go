package contactform

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type ContactFormHandler struct {
	Queries *sqlc.Queries
}

type CreateContactFormRequest struct {
	FullName    string `json:"full_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	PhoneNumber string `json:"phone_number,omitempty"`
	SalonName   string `json:"salon_name,omitempty"`
	Subject     string `json:"subject" binding:"required" oneof:"support sales partnership"`
	Message     string `json:"message" binding:"required"`
}

type CreateContactFormResponse struct {
	ContactForm sqlc.ContactForm `json:"contact_form"`
}

type FindContactFormParams struct {
	Page            int32 `json:"page,omitempty"`
	Limit           int32 `json:"limit,omitempty"`
	HasNotResponded bool  `json:"has_not_responded,omitempty"`
}

func NewContactFormHandler(db *sql.DB) *ContactFormHandler {
	return &ContactFormHandler{
		Queries: sqlc.New(db),
	}
}

func (h *ContactFormHandler) CreateContactForm(ctx *gin.Context) {
	var req CreateContactFormRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cf, err := h.Queries.CreateContactForm(ctx, sqlc.CreateContactFormParams{
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		SalonName:   sql.NullString{String: req.SalonName, Valid: req.SalonName != ""},
		Subject:     sqlc.ContactFormSubject(req.Subject),
		Message:     req.Message,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, CreateContactFormResponse{ContactForm: cf})
}

func (h *ContactFormHandler) ListContactForms(ctx *gin.Context) {
	var paginationQuery FindContactFormParams
	if err := ctx.ShouldBindQuery(&paginationQuery); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := paginationQuery.Page
	limit := paginationQuery.Limit

	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = 20
	}

	offset := (page - 1) * limit

	if paginationQuery.HasNotResponded {
		cfs, err := h.Queries.ListContactFormsHasNotResponded(ctx, sqlc.ListContactFormsHasNotRespondedParams{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"contact_forms": cfs})
		return
	}

	cfs, err := h.Queries.ListContactForms(ctx, sqlc.ListContactFormsParams{
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"contact_forms": cfs})
}
