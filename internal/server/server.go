package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/parlorhub/api-core/internal/database"
	"github.com/parlorhub/api-core/internal/middleware/rbac"
	"github.com/parlorhub/api-core/internal/modules/auth"
	"github.com/parlorhub/api-core/internal/modules/auth/session"
	contactform "github.com/parlorhub/api-core/internal/modules/contact_form"
	"github.com/parlorhub/api-core/internal/modules/sso"
)

type Server struct {
	port int

	db                 database.Service
	authHandler        *auth.AuthHandler
	contactFormHandler *contactform.ContactFormHandler
	rbacMiddleware     *rbac.RbacMiddleware
	ssoHandler         *sso.SSOHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()

	sessionService, err := session.NewSessionService()
	if err != nil {
		log.Printf("Warning: Failed to initialize session service: %v", err)
	}

	authHandler := auth.NewAuthHandler(db.GetDB(), sessionService)
	contactFormHandler := contactform.NewContactFormHandler(db.GetDB())
	rbacMiddleware := rbac.NewRbacMiddleware(db.GetDB())

	NewServer := &Server{
		port: port,

		db:                 db,
		authHandler:        authHandler,
		rbacMiddleware:     rbacMiddleware,
		contactFormHandler: contactFormHandler,
		ssoHandler:         sso.NewSSOHandler(db.GetDB(), authHandler.GetService()),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
