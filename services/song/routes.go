package song

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/songs", h.HandleAddSong).Methods("POST")
	router.HandleFunc("/songs", h.HandleGetSong).Methods("GET")
	router.HandleFunc("/songs", h.HandleUpdateSong).Methods("POST")
	router.HandleFunc("/songs", h.HandleDeleteSong).Methods("POST")
}

func (h *Handler) HandleAddSong(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) HandleGetSong(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {}
