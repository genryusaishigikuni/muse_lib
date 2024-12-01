package server

import (
	"database/sql"
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/logger"
	"github.com/genryusaishigikuni/muse_lib/services/song"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
)

type Server struct {
	addr string
	db   *sql.DB
}

// NewServer creates a new instance of the Server
func NewServer(addr string, db *sql.DB) *Server {
	return &Server{addr: addr, db: db}
}

func (s *Server) Start() error {
	const op = "server.Start" // Operation context for logging

	// Initialize logger
	env := config.Envs.Environment
	logs := logger.SetupLogger(env)
	logs.Info("Starting HTTP server", slog.String("address", s.addr), slog.String("operation", op))

	// Set up router with logging middleware
	router := mux.NewRouter()
	router.Use(logger.New(logs)) // Custom middleware for request logging
	logs.Debug("Router and middleware initialized", slog.String("operation", op))

	// Initialize song service and routes
	songStore := song.NewStore(s.db, env)
	songHandler := song.NewHandler(songStore, env)
	songHandler.RegisterRoutes(router.PathPrefix("/api").Subrouter())
	logs.Debug("Song routes registered", slog.String("operation", op))

	// Static file handler
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	logs.Info("Static file handler configured", slog.String("operation", op))

	// Start listening for requests
	logs.Info("Listening for incoming connections", slog.String("address", s.addr), slog.String("operation", op))
	err := http.ListenAndServe(s.addr, router)
	if err != nil {
		logs.Error("Server failed to start", logger.Err(err), slog.String("operation", op))
		return err
	}

	logs.Info("Server stopped", slog.String("operation", op))
	return nil
}
