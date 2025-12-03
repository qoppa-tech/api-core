package server

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/parlorhub/api-core/internal/logger"
	"github.com/parlorhub/api-core/internal/middleware/auth"
	"github.com/parlorhub/api-core/internal/middleware/origin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()
	r.Use(logger.SetupLogger())
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("LANDING_PAGE_URL"), os.Getenv("FRONTEND_URL")},
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

	r.GET("/profile", auth.AuthMiddleware(s.authHandler.GetService()), s.profileHandler.GetProfile)
	r.PATCH("/profile", auth.AuthMiddleware(s.authHandler.GetService()), s.profileHandler.UpdateProfile)

	r.GET("/onboarding", auth.AuthMiddleware(s.authHandler.GetService()), s.onboardingHandler.GetOnboarding)
	r.POST("/onboarding", auth.AuthMiddleware(s.authHandler.GetService()), s.onboardingHandler.SaveOnboarding)
	r.POST("/onboarding/step", auth.AuthMiddleware(s.authHandler.GetService()), s.onboardingHandler.SaveOnboardingStep)

	landingPage := r.Group("")
	landingPage.Use(origin.LandingPageOnly())
	{
		landingPage.POST("/contact-form", s.contactFormHandler.CreateContactForm)
	}

	r.GET("/contact-form", auth.AuthMiddleware(s.authHandler.GetService()), s.contactFormHandler.ListContactForms)

	r.Static("/swagger", "./docs")
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/swagger.html")
	})

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
