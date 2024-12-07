package server

import (
	"database/sql"
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/logger"
	"github.com/genryusaishigikuni/muse_lib/services/song"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
)

type Server struct {
	addr string
	db   *sql.DB
}

func NewServer(addr string, db *sql.DB) *Server {
	return &Server{addr: addr, db: db}
}

func (s *Server) Start() error {
	const op = "server.Start"

	env := config.Envs.Environment
	logs := logger.SetupLogger(env)
	logs.Info("Starting HTTP server", slog.String("address", s.addr), slog.String("operation", op))

	router := mux.NewRouter()
	router.Use(handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	))
	router.Use(logger.New(logs))
	logs.Debug("Router and middleware initialized", slog.String("operation", op))

	songStore := song.NewStore(s.db, env)
	songHandler := song.NewHandler(songStore, env)
	songHandler.RegisterRoutes(router.PathPrefix("/api").Subrouter())
	logs.Debug("Song routes registered", slog.String("operation", op))

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	logs.Info("Static file handler configured", slog.String("operation", op))

	logs.Info("Listening for incoming connections", slog.String("address", s.addr), slog.String("operation", op))
	err := http.ListenAndServe(s.addr, router)
	if err != nil {
		logs.Error("Server failed to start", logger.Err(err), slog.String("operation", op))
		return err
	}

	logs.Info("Server stopped", slog.String("operation", op))
	return nil
}
