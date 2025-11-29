package rbac

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	helper "github.com/parlorhub/api-core/internal/helper/user"

	"github.com/parlorhub/api-core/internal/database/sqlc"
)

type RbacMiddleware struct {
	Queries *sqlc.Queries
}

func NewRbacMiddleware(db *sql.DB) *RbacMiddleware {
	return &RbacMiddleware{
		Queries: sqlc.New(db),
	}
}

func (rm *RbacMiddleware) RBACMiddleware(permission sqlc.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exists := ctx.Get("user_id")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user, err := rm.Queries.GetUserByID(ctx, userID.(uuid.UUID))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userAllowed, err := helper.CompareUserRolePrivileges(user.Role, permission)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if !userAllowed {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		ctx.Next()
	}
}
