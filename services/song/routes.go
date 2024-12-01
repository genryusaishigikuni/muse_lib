package song

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/logger"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
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

	songDetails, err := h.fetchSongDetailsFromAPI(payload.Group, payload.SongName)
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

	songs, err := h.store.GetSongs(r.URL.Query())
	if err != nil {
		h.logs.Error("Error fetching songs", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if len(songs) == 0 {
		h.logs.Warn("No songs found", "operation", op)
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

	if err := h.store.DeleteSong(payload.SongName, payload.Group); err != nil {
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

	var payload types.SongUpdatePayload
	if err := ParseJson(r, &payload); err != nil {
		h.logs.Error("Invalid input", "operation", op, logger.Err(err))
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.UpdateSongInfo(payload.SongName, payload.Group, pq.Array(payload.SongLyrics), payload.Published, payload.Link); err != nil {
		h.logs.Error("Error updating song", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.logs.Info("Song updated successfully", "operation", op, "song_name", payload.SongName)
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

func (h *Handler) fetchSongDetailsFromAPI(group, song string) (*types.SongDetail, error) {
	const op = "Handler.fetchSongDetailsFromAPI"

	// Build the external API URL with query parameters
	apiURL := fmt.Sprintf("http://localhost:8081/info?group=%s&song=%s", url.QueryEscape(group), url.QueryEscape(song))

	// Make the GET request to the external API
	resp, err := h.apiClient.R().Get(apiURL)
	if err != nil {
		log.Printf("Error calling external API: %v", err)
		return nil, errors.New("failed to fetch song details from external API")
	}

	// Check if the response status is not 200 OK
	if resp.StatusCode() != http.StatusOK {
		log.Printf("External API returned non-200 status: %d", resp.StatusCode())
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode())
	}

	// Parse the response body into SongDetail
	var songDetails types.SongDetail
	if err := json.Unmarshal(resp.Body(), &songDetails); err != nil {
		h.logs.Error("Error parsing external API response", "operation", op, logger.Err(err))
		return nil, errors.New("failed to parse external API response")
	}

	return &songDetails, nil
}

func splitLyrics(text string) []string {
	// Split the lyrics into lines by newline character
	lines := strings.Split(text, "\n")
	// Remove empty lines (if any)
	var cleanedLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			cleanedLines = append(cleanedLines, trimmedLine)
		}
	}
	return cleanedLines
}
