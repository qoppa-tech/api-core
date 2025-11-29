package server

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/parlorhub/api-core/internal/database/sqlc"
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

	auth := r.Group("/auth")
	{
		auth.POST("/register", s.authHandler.RegisterHandler)
		auth.POST("/login", s.authHandler.LoginHandler)
		auth.GET("/me", s.rbacMiddleware.RBACMiddleware(sqlc.UserRoleCustomer), s.authHandler.MeHandler)
		auth.POST("/logout", s.authHandler.LogoutHandler)

		auth.GET("/google", s.ssoHandler.GoogleLoginHandler)
		auth.GET("/google/callback", s.ssoHandler.GoogleCallbackHandler)
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
