package server

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/parlorhub/api-core/internal/modules/auth"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/health", s.healthHandler)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", s.authHandler.RegisterHandler)
		authGroup.POST("/login", s.authHandler.LoginHandler)
		authGroup.POST("/refresh", s.authHandler.RefreshHandler)

		authGroup.GET("/google", s.ssoHandler.GoogleLoginHandler)
		authGroup.GET("/google/callback", s.ssoHandler.GoogleCallbackHandler)
	}

	authProtected := r.Group("/auth")
	authProtected.Use(auth.AuthMiddleware(s.authHandler.GetService()))
	{
		authProtected.GET("/me", s.authHandler.MeHandler)
		authProtected.POST("/logout", s.authHandler.LogoutHandler)
	}

	contactForm := r.Group("/contact-form")
	{
		contactForm.POST("/", s.contactFormHandler.CreateContactForm)
		contactForm.GET("/", s.contactFormHandler.ListContactForms)
	}

	r.Static("/swagger", "./docs")
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/swagger.html")
	})

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
