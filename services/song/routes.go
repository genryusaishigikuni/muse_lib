package song

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/logger"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	store     types.SongStore
	apiClient *resty.Client
	logs      *slog.Logger
}

func NewHandler(songStore types.SongStore, env string) *Handler {
	return &Handler{
		store:     songStore,
		apiClient: resty.New(),
		logs:      logger.SetupLogger(env),
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/songs/add", h.HandleAddSong).Methods("POST")
	router.HandleFunc("/songs/get", h.HandleGetSong).Methods("GET")
	router.HandleFunc("/songs/update", h.HandleUpdateSong).Methods("PUT")
	router.HandleFunc("/songs/delete", h.HandleDeleteSong).Methods("DELETE")
}

func (h *Handler) HandleAddSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleAddSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method)

	var payload types.SongAddPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logs.Error("Invalid input", "operation", op, logger.Err(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	h.logs.Debug("Payload decoded", "operation", op, "payload", payload)

	externalAPI := config.Envs.ExtApi
	songDetails, err := h.fetchSongDetailsFromAPI(payload.Group, payload.SongName, externalAPI)
	if err != nil {
		h.logs.Error("Error fetching song details", "operation", op, logger.Err(err))
		http.Error(w, "Failed to fetch song details", http.StatusInternalServerError)
		return
	}

	songLyrics := splitLyrics(songDetails.Text)
	h.logs.Debug("Song lyrics processed", "operation", op, "lyrics_lines", len(songLyrics))

	if err := h.store.AddSong(payload.SongName, payload.Group, songDetails, songLyrics); err != nil {
		h.logs.Error("Error adding song to DB", "operation", op, logger.Err(err))
		http.Error(w, "Failed to add song to the database", http.StatusInternalServerError)
		return
	}

	h.logs.Info("Song added successfully", "operation", op, "song_name", payload.SongName)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("Song added successfully"))
}

func (h *Handler) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleGetSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method, "query_params", r.URL.Query())

	// Initialize filters from query parameters
	filters := r.URL.Query()

	// Check if no filters were provided in the query
	if len(filters) == 0 && r.Body != nil {
		h.logs.Debug("No query parameters provided, checking request body", "operation", op)
		var bodyFilters map[string]string
		if err := json.NewDecoder(r.Body).Decode(&bodyFilters); err != nil {
			h.logs.Error("Error decoding JSON body", "operation", op, logger.Err(err))
			WriteError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
			return
		}

		// Convert body filters to url.Values for compatibility
		for key, value := range bodyFilters {
			filters.Add(key, value)
		}
	}

	if len(filters) == 0 {
		h.logs.Warn("No filters provided in query or body", "operation", op)
		WriteError(w, http.StatusBadRequest, errors.New("no filters provided"))
		return
	}

	// Fetch songs based on filters
	songs, err := h.store.GetSongs(filters)
	if err != nil {
		h.logs.Error("Error fetching songs", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Handle the case where no songs match the filters
	if len(songs) == 0 {
		h.logs.Warn("No songs found matching the criteria", "operation", op)
		WriteError(w, http.StatusNotFound, errors.New("no songs found matching the criteria"))
		return
	}

	h.logs.Debug("Songs retrieved", "operation", op, "count", len(songs))
	if err := WriteJSON(w, http.StatusOK, songs); err != nil {
		h.logs.Error("Error writing response", "operation", op, logger.Err(err))
	}
}

func (h *Handler) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleDeleteSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method)

	var payload types.SongDeletePayload
	if err := ParseJson(r, &payload); err != nil {
		h.logs.Error("Invalid input", "operation", op, logger.Err(err))
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Validate input parameters: we need at least one of these parameters
	if payload.SongName == "" && payload.Group == "" && payload.Link == "" && payload.ID == 0 {
		h.logs.Error("No identifier provided", "operation", op)
		WriteError(w, http.StatusBadRequest, errors.New("either song name, group, link, or ID must be provided"))
		return
	}

	// Call the store to delete the song
	if err := h.store.DeleteSong(payload.SongName, payload.Group, payload.Link, payload.ID); err != nil {
		h.logs.Error("Error deleting song", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.logs.Info("Song deleted successfully", "operation", op, "song_name", payload.SongName)
	if err := WriteJSON(w, http.StatusOK, map[string]string{"status": "song deleted"}); err != nil {
		h.logs.Error("Error writing response", "operation", op, logger.Err(err))
	}
}

func (h *Handler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleUpdateSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method)

	// Parse the incoming request payload
	var payload types.Song
	if err := ParseJson(r, &payload); err != nil {
		h.logs.Error("Invalid input", "operation", op, logger.Err(err))
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Ensure that the ID is not being updated and is positive
	if payload.ID <= 0 {
		h.logs.Error("Invalid ID", "operation", op, "error", "ID must be positive")
		WriteError(w, http.StatusBadRequest, fmt.Errorf("ID must be positive"))
		return
	}

	// Log the payload to see what is being sent
	h.logs.Info("Received payload", "operation", op, "payload", payload)

	// Update the song info based on the ID
	if err := h.store.UpdateSongInfo(payload.ID, payload.SongName, payload.Group, payload.SongLyrics, payload.Published, payload.Link); err != nil {
		h.logs.Error("Error updating song", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Respond with success message
	h.logs.Info("Song updated successfully", "operation", op, "song_id", payload.ID)
	if err := WriteJSON(w, http.StatusOK, map[string]string{"status": "song updated"}); err != nil {
		h.logs.Error("Error writing response", "operation", op, logger.Err(err))
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

func (h *Handler) fetchSongDetailsFromAPI(group, song, externalAPI string) (*types.SongDetail, error) {
	const op = "Handler.fetchSongDetailsFromAPI"

	// Build the external API URL with query parameters
	apiURL := fmt.Sprintf("%s/info?group=%s&song=%s", externalAPI, url.QueryEscape(group), url.QueryEscape(song))

	h.logs.Debug("Making API request", "operation", op, "url", apiURL)

	// Make the request using the Resty client
	resp, err := h.apiClient.R().Get(apiURL)
	if err != nil {
		h.logs.Error("Error fetching from API", "operation", op, logger.Err(err))
		return nil, err
	}

	h.logs.Debug("API response received", "operation", op, "status_code", resp.StatusCode())
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", resp.StatusCode())
	}

	var songDetails types.SongDetail
	if err := json.Unmarshal(resp.Body(), &songDetails); err != nil {
		h.logs.Error("Error unmarshalling API response", "operation", op, logger.Err(err))
		return nil, err
	}

	return &songDetails, nil
}

func splitLyrics(text string) []string {
	if text == "" {
		return nil
	}
	// Split text into verses
	return strings.Split(text, "\n\n")
}
