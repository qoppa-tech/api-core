package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the access token from cookies or Authorization header
func AuthMiddleware(authService *AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		InitJWTConfig()

		token := GetAccessTokenFromRequest(ctx)
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Check if token is blacklisted
		if authService != nil {
			blacklisted, err := authService.IsTokenBlacklisted(ctx.Request.Context(), token)
			if err == nil && blacklisted {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
				return
			}
		}

		// Parse and validate token
		claims, err := authService.ParseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Set user info in context
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("role", claims.Role)
		if claims.SalonID != nil {
			ctx.Set("salon_id", *claims.SalonID)
		}

		ctx.Next()
	}
}
