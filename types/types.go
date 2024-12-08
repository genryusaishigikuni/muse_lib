package types

import (
	"net/url"
	"time"
)

type SongAddPayload struct {
	SongName string `json:"song"`
	Group    string `json:"group"`
}

type SongDeletePayload struct {
	ID int `json:"id,omitempty"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type Song struct {
	ID         int       `json:"id"`
	SongName   string    `json:"song"`
	Group      string    `json:"group"`
	SongLyrics []string  `json:"songLyrics"`
	Published  time.Time `json:"published"`
	Link       string    `json:"link"`
}

type SongStore interface {
	GetSongs(filters url.Values) ([]Song, error) // Updated signature
	DeleteSong(id int) error
	UpdateSongInfo(id int, name, group string, lyrics interface{}, published time.Time, link string) error
	AddSong(name, group string, songDetails *SongDetail, text []string) error
}
