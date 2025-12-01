package origin

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func LandingPageOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		landingPageURL := os.Getenv("LANDING_PAGE_URL")
		origin := c.GetHeader("Origin")
		referer := c.GetHeader("Referer")

		if origin != "" && origin != landingPageURL {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		if origin == "" && referer != "" {
			if len(referer) < len(landingPageURL) || referer[:len(landingPageURL)] != landingPageURL {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}

		c.Next()
	}
}
