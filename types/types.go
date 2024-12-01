package types

import "net/url"

type SongAddPayload struct {
	SongName string `json:"song"`
	Group    string `json:"group"`
}

type SongDeletePayload struct {
	SongName string `json:"songName"`
	Group    string `json:"songGroup"`
}

type SongUpdatePayload struct {
	SongName   string   `json:"songName"`
	Group      string   `json:"songGroup"`
	SongLyrics []string `json:"songLyrics"`
	Published  string   `json:"published"`
	Link       string   `json:"link"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type Song struct {
	ID         int      `json:"id"`
	SongName   string   `json:"songName"`
	Group      string   `json:"songGroup"`
	SongLyrics []string `json:"songLyrics"`
	Published  string   `json:"published"`
	Link       string   `json:"link"`
}

type SongStore interface {
	GetSongs(filters url.Values) ([]Song, error) // Updated signature
	DeleteSong(name, group string) error
	UpdateSongInfo(name, group string, lyrics interface{}, published string, link string) error
	AddSong(name, group string, songDetails *SongDetail, text []string) error
}
