package server

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/parlorhub/api-core/internal/database/sqlc"
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

	clientsGroup := r.Group("/clients")
	clientsGroup.Use(auth.AuthMiddleware(s.authHandler.GetService()))
	{
		clientsGroup.POST("", s.clientsHandler.CreateClient)
		clientsGroup.GET("", s.clientsHandler.ListClients)
		clientsGroup.GET("/search", s.clientsHandler.SearchClients)
		clientsGroup.GET("/by-phone", s.clientsHandler.GetClientByPhone)
		clientsGroup.GET("/:id", s.clientsHandler.GetClient)
		clientsGroup.PATCH("/:id", s.clientsHandler.UpdateClient)
		clientsGroup.DELETE("/:id", s.clientsHandler.DeleteClient)
	}

	appointmentsGroup := r.Group("/appointments")
	appointmentsGroup.Use(auth.AuthMiddleware(s.authHandler.GetService()))
	{
		appointmentsGroup.POST("", s.appointmentHandler.CreateAppointment)
		appointmentsGroup.GET("", s.appointmentHandler.ListAppointments)
		appointmentsGroup.GET("/by-date", s.appointmentHandler.ListAppointmentsByDate)
		appointmentsGroup.GET("/by-date-range", s.appointmentHandler.ListAppointmentsByDateRange)
		appointmentsGroup.GET("/by-user", s.appointmentHandler.ListAppointmentsByUser)
		appointmentsGroup.GET("/by-status", s.appointmentHandler.ListAppointmentsByStatus)
		appointmentsGroup.GET("/by-client-phone", s.appointmentHandler.ListAppointmentsByClientPhone)
		appointmentsGroup.GET("/count", s.appointmentHandler.CountAppointmentsByUserAndDate)
		appointmentsGroup.GET("/:id", s.appointmentHandler.GetAppointment)
		appointmentsGroup.PATCH("/:id", s.appointmentHandler.UpdateAppointment)
		appointmentsGroup.PATCH("/:id/status", s.appointmentHandler.UpdateAppointmentStatus)
		appointmentsGroup.DELETE("/:id", s.appointmentHandler.DeleteAppointment)
	}

	salonsGroup := r.Group("/salons")
	salonsGroup.Use(auth.AuthMiddleware(s.authHandler.GetService()))
	salonsGroup.Use(s.rbacMiddleware.RBACMiddleware(sqlc.UserRoleOwner))
	{
		salonsGroup.POST("", s.salonsHandler.CreateSalon)
		salonsGroup.GET("", s.salonsHandler.ListSalons)
		salonsGroup.GET("/by-slug", s.salonsHandler.GetSalonBySlug)
		salonsGroup.GET("/by-owner", s.salonsHandler.GetSalonByOwner)
		salonsGroup.GET("/:id", s.salonsHandler.GetSalon)
		salonsGroup.PATCH("/:id", s.salonsHandler.UpdateSalon)
		salonsGroup.DELETE("/:id", s.salonsHandler.DeleteSalon)
	}

	servicesGroup := r.Group("/services")
	servicesGroup.Use(auth.AuthMiddleware(s.authHandler.GetService()))
	servicesGroup.Use(s.rbacMiddleware.RBACMiddleware(sqlc.UserRoleOwner))
	{
		servicesGroup.POST("", s.servicesHandler.CreateService)
		servicesGroup.GET("", s.servicesHandler.ListServices)
		servicesGroup.GET("/active", s.servicesHandler.ListActiveServices)
		servicesGroup.GET("/by-user", s.servicesHandler.ListServicesByUser)
		servicesGroup.GET("/by-salon-and-user", s.servicesHandler.ListServicesBySalonAndUser)
		servicesGroup.GET("/:id", s.servicesHandler.GetService)
		servicesGroup.PATCH("/:id", s.servicesHandler.UpdateService)
		servicesGroup.DELETE("/:id", s.servicesHandler.DeleteService)
	}

	dashboardGroup := r.Group("/dashboard")
	dashboardGroup.Use(auth.AuthMiddleware(s.authHandler.GetService()))
	dashboardGroup.Use(s.rbacMiddleware.RBACMiddleware(sqlc.UserRoleOwner))
	{
		dashboardGroup.GET("/summary", s.dashboardHandler.GetSummary)
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
