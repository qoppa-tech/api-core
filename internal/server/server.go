package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/parlorhub/api-core/internal/database"
	"github.com/parlorhub/api-core/internal/modules/auth"
	contactform "github.com/parlorhub/api-core/internal/modules/contact_form"
)

type Server struct {
	port int

	db                 database.Service
	authHandler        *auth.AuthHandler
	contactFormHandler *contactform.ContactFormHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	NewServer := &Server{
		port: port,

		db:          db,
		authHandler: auth.NewAuthHandler(db.GetDB()),
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
