package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/parlorhub/api-core/internal/database"
	"github.com/parlorhub/api-core/internal/logger"
	"github.com/parlorhub/api-core/internal/middleware/rbac"
	"github.com/parlorhub/api-core/internal/modules/auth"
	"github.com/parlorhub/api-core/internal/modules/auth/session"
	contactform "github.com/parlorhub/api-core/internal/modules/contact_form"
	"github.com/parlorhub/api-core/internal/modules/dashboard"
	"github.com/parlorhub/api-core/internal/modules/managment/appointment"
	"github.com/parlorhub/api-core/internal/modules/managment/clients"
	"github.com/parlorhub/api-core/internal/modules/managment/salons"
	"github.com/parlorhub/api-core/internal/modules/managment/services"
	"github.com/parlorhub/api-core/internal/modules/onboarding"
	"github.com/parlorhub/api-core/internal/modules/profile"
	"github.com/parlorhub/api-core/internal/modules/sso"
)

type Server struct {
	port int

	db                 database.Service
	authHandler        *auth.AuthHandler
	profileHandler     *profile.ProfileHandler
	contactFormHandler *contactform.ContactFormHandler
	onboardingHandler  *onboarding.OnboardingHandler
	rbacMiddleware     *rbac.RbacMiddleware
	ssoHandler         *sso.SSOHandler
	clientsHandler     *clients.ClientsHandler
	appointmentHandler *appointment.AppointmentHandler
	salonsHandler      *salons.SalonsHandler
	servicesHandler    *services.ServicesHandler
	dashboardHandler   *dashboard.Handler
}

func NewServer() *http.Server {
	logger.Init()

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()

	sessionService, err := session.NewSessionService()
	if err != nil {
		logger.Warn("Failed to initialize session service", logger.Err(err))
	}

	authHandler := auth.NewAuthHandler(db.GetDB(), sessionService)
	profileHandler := profile.NewProfileHandler(db.GetDB())
	contactFormHandler := contactform.NewContactFormHandler(db.GetDB())
	onboardingHandler := onboarding.NewOnboardingHandler(db.GetDB())
	rbacMiddleware := rbac.NewRbacMiddleware(db.GetDB())
	clientsHandler := clients.NewClientsHandler(db.GetDB())
	appointmentHandler := appointment.NewAppointmentHandler(db.GetDB())
	salonsHandler := salons.NewSalonsHandler(db.GetDB())
	servicesHandler := services.NewServicesHandler(db.GetDB())
	dashboardHandler := dashboard.NewHandler(db.GetDB())

	NewServer := &Server{
		port: port,

		db:                 db,
		authHandler:        authHandler,
		profileHandler:     profileHandler,
		rbacMiddleware:     rbacMiddleware,
		contactFormHandler: contactFormHandler,
		onboardingHandler:  onboardingHandler,
		ssoHandler:         sso.NewSSOHandler(db.GetDB(), authHandler.GetService()),
		clientsHandler:     clientsHandler,
		appointmentHandler: appointmentHandler,
		salonsHandler:      salonsHandler,
		servicesHandler:    servicesHandler,
		dashboardHandler:   dashboardHandler,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("Server initialized", logger.F("port", port))

	return server
}
