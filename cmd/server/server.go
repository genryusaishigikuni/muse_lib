package server

import (
	"database/sql"
	"github.com/genryusaishigikuni/muse_lib/logger"
	"github.com/genryusaishigikuni/muse_lib/services/song"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type Server struct {
	addr string
	db   *sql.DB
}

func NewServer(addr string, db *sql.DB) *Server {
	return &Server{addr: addr, db: db}
}

func (s *Server) Start() error {
	logs := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	router := mux.NewRouter()
	router.Use(logger.New(logs))
	subRouter := router.PathPrefix("/api").Subrouter()

	songStore := song.NewStore(s.db, "local")
	songHandler := song.NewHandler(songStore)
	songHandler.RegisterRoutes(subRouter)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	log.Println("Listening on", s.addr)

	return http.ListenAndServe(s.addr, router)

}
