package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")

	if group == "" || song == "" {
		http.Error(w, "Bad Request: Missing 'group' or 'song' parameter", http.StatusBadRequest)
		return
	}

	songDetails := SongDetail{
		ReleaseDate: "2006-07-16",
		Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(songDetails)
	if err != nil {
		http.Error(w, "Internal Server Error: Unable to encode response", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/info", infoHandler)

	// Start the server
	port := "8081"
	fmt.Printf("Mock API running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
