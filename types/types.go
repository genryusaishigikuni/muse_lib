package types

import (
	"time"
)

type SongAddPayload struct {
	SongName   string    `json:"songName"`
	SongLyrics []string  `json:"songLyrics"`
	Group      string    `json:"group"`
	Published  time.Time `json:"published"`
	Link       string    `json:"link"`
}

type SongDeletePayload struct {
	SongName string `json:"songName"`
	Group    string `json:"group"`
}

type SongUpdatePayload struct {
	SongName   string   `json:"songName"`
	Group      string   `json:"group"`
	SongLyrics []string `json:"songLyrics"`
}

type SongGetPayload struct {
	SongName string `json:"songName"`
	Group    string `json:"group"`
}

type Song struct {
	ID         int       `json:"id"`
	SongName   string    `json:"songName"`
	SongLyrics []string  `json:"songLyrics"`
	Group      string    `json:"group"`
	Published  time.Time `json:"published"`
	Link       string    `json:"link"`
}

type SongStore interface {
	GetSongByName(name string) (*Song, error)
	DeleteSong(name, group string) error
	UpdateSongInfo(name string) error
	AddSong(name, group string) error
}
