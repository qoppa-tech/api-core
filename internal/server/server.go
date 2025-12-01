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
	"github.com/parlorhub/api-core/internal/modules/onboarding"
	"github.com/parlorhub/api-core/internal/modules/sso"
)

type Server struct {
	port int

	db                 database.Service
	authHandler        *auth.AuthHandler
	contactFormHandler *contactform.ContactFormHandler
	onboardingHandler  *onboarding.OnboardingHandler
	rbacMiddleware     *rbac.RbacMiddleware
	ssoHandler         *sso.SSOHandler
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
	contactFormHandler := contactform.NewContactFormHandler(db.GetDB())
	onboardingHandler := onboarding.NewOnboardingHandler(db.GetDB())
	rbacMiddleware := rbac.NewRbacMiddleware(db.GetDB())

	NewServer := &Server{
		port: port,

		db:                 db,
		authHandler:        authHandler,
		rbacMiddleware:     rbacMiddleware,
		contactFormHandler: contactFormHandler,
		onboardingHandler:  onboardingHandler,
		ssoHandler:         sso.NewSSOHandler(db.GetDB(), authHandler.GetService()),
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
