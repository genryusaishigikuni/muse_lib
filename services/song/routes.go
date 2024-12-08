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
	"strconv"
	"strings"
	"time"
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

// RegisterRoutes registers the song-related routes.
//
// @Summary Register song routes
// @Description Adds routes for managing songs to the given router.
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/songs/add", h.HandleAddSong).Methods("POST")
	router.HandleFunc("/songs/get", h.HandleGetSong).Methods("GET")
	router.HandleFunc("/songs/update", h.HandleUpdateSong).Methods("PUT")
	router.HandleFunc("/songs/delete", h.HandleDeleteSong).Methods("DELETE")
}

// HandleAddSong adds a new song to the database.
//
// @Summary Add a new song
// @Description Adds a new song with details retrieved from an external API.
// @Tags songs
// @Accept  json
// @Produce json
// @Param payload body types.SongAddPayload true "Song data to add"
// @Success 201 {string} string "Song added successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to add song"
// @Router /songs/add [post]
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

// HandleGetSong retrieves songs based on query parameters.
//
// @Summary Retrieve songs
// @Description Retrieves songs matching specified criteria through query parameters.
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int false "ID of the song"
// @Param song query string false "Name of the song"
// @Param group query string false "Group name"
// @Param link query string false "Link to the song"
// @Param time query string false "Time (RFC3339 format)"
// @Param lyrics query string false "Lyrics as a JSON array"
// @Param limit query int false "Maximum number of results to return"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} types.Song "Songs retrieved successfully"
// @Failure 400 {string} string "Invalid query parameter"
// @Failure 404 {string} string "No songs found"
// @Failure 500 {string} string "Failed to fetch songs"
// @Router /songs/get [get]
func (h *Handler) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleGetSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method, "query_params", r.URL.Query())

	// Expected types for validation
	expectedTypes := map[string]string{
		"id":     "int",
		"song":   "string",
		"group":  "string",
		"link":   "string",
		"time":   "time",
		"lyrics": "array",
	}

	// Retrieve and validate query parameters
	normalizedFilters := url.Values{}
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue
		}

		value := values[0]
		if expectedType, exists := expectedTypes[key]; exists {
			switch expectedType {
			case "int":
				if _, err := strconv.Atoi(value); err != nil {
					h.logs.Error("Invalid integer value for filter", "key", key, "value", value, logger.Err(err))
					WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid value for %s: must be an integer", key))
					return
				}
			case "time":
				if _, err := time.Parse(time.RFC3339, value); err != nil {
					h.logs.Error("Invalid time value for filter", "key", key, "value", value, logger.Err(err))
					WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid value for %s: must be a valid timestamp", key))
					return
				}
			case "array":
				var arr []string
				if err := json.Unmarshal([]byte(value), &arr); err != nil {
					h.logs.Error("Invalid array value for filter", "key", key, "value", value, logger.Err(err))
					WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid value for %s: must be a valid array", key))
					return
				}
			}
		}
		normalizedFilters.Add(key, value)
	}

	// Fetch songs from storage
	songs, err := h.store.GetSongs(normalizedFilters)
	if err != nil {
		h.logs.Error("Error fetching songs", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if len(songs) == 0 {
		h.logs.Warn("No songs found matching the criteria", "operation", op, "filters", normalizedFilters)
		WriteError(w, http.StatusNotFound, errors.New("no songs found matching the criteria"))
		return
	}

	h.logs.Debug("Songs retrieved", "operation", op, "count", len(songs))
	if err := WriteJSON(w, http.StatusOK, songs); err != nil {
		h.logs.Error("Error writing response", "operation", op, logger.Err(err))
	}
}

// HandleDeleteSong deletes a song from the database.
//
// @Summary Delete a song
// @Description Deletes a song based on its name and group.
// @Tags songs
// @Accept  json
// @Param payload body types.SongDeletePayload true "delete the song based on ID"
// @Produce json
// @Success 200 {string} string "Song deleted successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to delete song"
// @Router /songs/delete [delete]
func (h *Handler) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleDeleteSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method)

	var payload types.SongDeletePayload
	if err := ParseJson(r, &payload); err != nil {
		h.logs.Error("Invalid input", "operation", op, logger.Err(err))
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.ID == 0 {
		h.logs.Error("No identifier provided", "operation", op)
		WriteError(w, http.StatusBadRequest, errors.New("either song name, group, link, or ID must be provided"))
		return
	}

	if err := h.store.DeleteSong(payload.ID); err != nil {
		h.logs.Error("Error deleting song", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	h.logs.Info("Song deleted successfully", "operation", op, "song_name")
	if err := WriteJSON(w, http.StatusOK, map[string]string{"status": "song deleted"}); err != nil {
		h.logs.Error("Error writing response", "operation", op, logger.Err(err))
	}
}

// HandleUpdateSong updates song information.
//
// @Summary Update song
// @Description Updates existing song details.
// @Tags songs
// @Accept json
// @Param payload body types.Song true "update the song"
// @Produce json
// @Success 200 {string} string "Song updated successfully"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to update song"
// @Router /songs/update [put]
func (h *Handler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.HandleUpdateSong"
	h.logs.Info("Starting request", "operation", op, "method", r.Method)

	var payload types.Song
	if err := ParseJson(r, &payload); err != nil {
		h.logs.Error("Invalid input", "operation", op, logger.Err(err))
		WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.ID <= 0 {
		h.logs.Error("Invalid ID", "operation", op, "error", "ID must be positive")
		WriteError(w, http.StatusBadRequest, fmt.Errorf("ID must be positive"))
		return
	}

	h.logs.Info("Received payload", "operation", op, "payload", payload)

	if err := h.store.UpdateSongInfo(payload.ID, payload.SongName, payload.Group, payload.SongLyrics, payload.Published, payload.Link); err != nil {
		h.logs.Error("Error updating song", "operation", op, logger.Err(err))
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

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

	apiURL := fmt.Sprintf("%s/info?group=%s&song=%s", externalAPI, url.QueryEscape(group), url.QueryEscape(song))

	h.logs.Debug("Making API request", "operation", op, "url", apiURL)

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
	return strings.Split(text, "\n\n")
}
