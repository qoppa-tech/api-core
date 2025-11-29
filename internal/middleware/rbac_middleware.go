package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	helper "github.com/parlorhub/api-core/internal/helper/user"

	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type RbacMiddleware struct {
	queries *sqlc.Queries
}

func (rm *RbacMiddleware) RBACMiddleware(permission sqlc.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			ctx.Abort()
		}

		user, err := rm.queries.GetUserByID(ctx, userID.(uuid.UUID))
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			ctx.Abort()
		}

		userAllowed, err := helper.CompareUserRolePrivileges(user.Role, permission)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			ctx.Abort()
		}

		if !userAllowed {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			ctx.Abort()
		}
	}
}
