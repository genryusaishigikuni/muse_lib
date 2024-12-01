package song

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	const op = "Handler.HandleAddSong"

	// Parse the incoming JSON payload
	var payload types.SongAddPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch enriched song details from the external API
	songDetails, err := h.fetchSongDetailsFromAPI(payload.Group, payload.SongName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching song details: %v", err), http.StatusInternalServerError)
		return
	}

	// Split the song lyrics into a slice
	songLyrics := splitLyrics(songDetails.Text)

	// Add the song to the database
	err = h.store.AddSong(payload.SongName, payload.Group, songDetails, songLyrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error adding song to the database: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Song added successfully"))
	if err != nil {
		log.Printf("Error while writing response: %v", err)
	}

}

func (h *Handler) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// Retrieve songs based on filters
	songs, err := h.store.GetSongs(queryParams)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if len(songs) == 0 {
		WriteError(w, http.StatusNotFound, errors.New("no songs found matching the criteria"))
		return
	}

	// Extract the 'fields' parameter
	fields := queryParams.Get("fields")
	if fields != "" {
		selectedFields := strings.Split(fields, ",")
		var response []map[string]interface{}

		for _, song := range songs {
			data := make(map[string]interface{})
			for _, field := range selectedFields {
				switch strings.TrimSpace(field) {
				case "id":
					data["id"] = song.ID
				case "songName":
					data["songName"] = song.SongName
				case "songGroup":
					data["songGroup"] = song.Group
				case "songLyrics":
					data["songLyrics"] = song.SongLyrics
				case "published":
					data["published"] = song.Published
				case "link":
					data["link"] = song.Link
				// Handle additional fields here if necessary
				default:
					// Optionally log or handle invalid fields
					continue
				}
			}
			response = append(response, data)
		}

		err := WriteJSON(w, http.StatusOK, response)
		if err != nil {
			log.Printf("Error while writing response: %v", err)
		}
		return
	}

	// Default full response if 'fields' is not specified
	err = WriteJSON(w, http.StatusOK, songs)
	if err != nil {
		log.Printf("Error while writing response: %v", err)
	}

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

	err = WriteJSON(w, http.StatusOK, map[string]string{"status": "song deleted"})
	if err != nil {
		log.Printf("Error while writing response: %v", err)
	}
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

	err = WriteJSON(w, http.StatusOK, map[string]string{"status": "song updated"})
	if err != nil {
		log.Printf("Error while writing response: %v", err)
		return
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
		log.Printf("Error parsing external API response: %v", err)
		return nil, errors.New("failed to parse external API response")
	}

	return &songDetails, nil
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
