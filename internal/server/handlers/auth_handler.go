package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type AuthHandler struct {
	queries *sqlc.Queries
}

func (ah *AuthHandler) RegisterHandler(ctx gin.Context) {

}
