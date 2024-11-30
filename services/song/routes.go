package song

import (
	"encoding/json"
	"errors"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type Handler struct {
	store     types.SongStore
	apiClient *resty.Client // HTTP client to interact with external API
}

func NewHandler(songStore types.SongStore) *Handler {
	return &Handler{
		store:     songStore,
		apiClient: resty.New(), // Initialize the Resty HTTP client
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/songs/add", h.HandleAddSong).Methods("POST")
	router.HandleFunc("/songs/get", h.HandleGetSong).Methods("GET")
	router.HandleFunc("/songs/update", h.HandleUpdateSong).Methods("PUT")
	router.HandleFunc("/songs/delete", h.HandleDeleteSong).Methods("DELETE")
}

func (h *Handler) HandleAddSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongAddPayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Call external API to get song details
	songDetails, err := h.fetchSongDetailsFromAPI(payload.Group, payload.SongName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Add the song to the database
	err = h.store.AddSong(payload.SongName, payload.Group, songDetails)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "song added"})
}

func (h *Handler) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongGetPayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Retrieve the song from the database
	song, err := h.store.GetSongByName(payload.SongName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusOK, song)
}

func (h *Handler) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongDeletePayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Delete the song from the database
	err = h.store.DeleteSong(payload.SongName, payload.Group)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "song deleted"})
}

func (h *Handler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {
	var payload types.SongUpdatePayload
	err := ParseJson(r, &payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Update the song information in the database with all parameters
	err = h.store.UpdateSongInfo(payload.SongName, payload.Group, pq.Array(payload.SongLyrics), payload.Published, payload.Link)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "song updated"})
}

func (h *Handler) fetchSongDetailsFromAPI(group, song string) (*types.SongDetail, error) {
	// Make the request to the external API described by the Swagger specification
	resp, err := h.apiClient.R().
		SetQueryParams(map[string]string{
			"group": group,
			"song":  song,
		}).
		Get("http://external-api.com/info")

	if err != nil {
		log.Printf("Failed to fetch song details: %v", err)
		return nil, err
	}

	// Parse the response into the SongDetail structure
	var songDetail types.SongDetail
	if err := json.Unmarshal(resp.Body(), &songDetail); err != nil {
		log.Printf("Failed to unmarshal song details: %v", err)
		return nil, err
	}

	return &songDetail, nil
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
