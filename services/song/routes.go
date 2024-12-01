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

	err = WriteJSON(w, http.StatusOK, map[string]string{"status": "song added"})
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
