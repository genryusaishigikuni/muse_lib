package song

import (
	"encoding/json"
	"errors"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Handler struct {
	store types.SongStore
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/songs", h.HandleAddSong).Methods("POST")
	router.HandleFunc("/songs", h.HandleGetSong).Methods("GET")
	router.HandleFunc("/songs", h.HandleUpdateSong).Methods("PUT")
	router.HandleFunc("/songs", h.HandleDeleteSong).Methods("DELETE")
}

func (h *Handler) HandleAddSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongAddPayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
	}
}

func (h *Handler) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongGetPayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
	}
}

func (h *Handler) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongDeletePayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
	}
}

func (h *Handler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongUpdatePayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
	}
}

func ParseJson(r *http.Request, payload any) error {
	if r.Body == nil {
		return errors.New("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	err = WriteJSON(w, status, map[string]string{"error": err.Error()})
	if err != nil {
		log.Println(err)
	}
}
