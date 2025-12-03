package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	a "github.com/parlorhub/api-core/internal/modules/auth"
)

func AuthMiddleware(authService *a.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		a.InitJWTConfig()

		token := a.GetAccessTokenFromRequest(ctx)
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if authService != nil {
			blacklisted, err := authService.IsTokenBlacklisted(ctx.Request.Context(), token)
			if err == nil && blacklisted {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
				return
			}
		}

		claims, err := authService.ParseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("role", claims.Role)
		if claims.SalonID != nil {
			ctx.Set("salon_id", *claims.SalonID)
		}

		ctx.Next()
	}
}
