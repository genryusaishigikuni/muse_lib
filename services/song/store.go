package song

import (
	"database/sql"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetSongByName(name string) (*types.Song, error) {
	rows, err := s.db.Query("SELECT * FROM songs WHERE name = ?", name)
	if err != nil {
		return nil, err
	}

	song := new(types.Song)
	for rows.Next() {
		song, err = ScanRowsIntoSongs(rows)
		if err != nil {
			return nil, err
		}
	}
	if song.ID == 0 {
		return nil, fmt.Errorf("no song found with name %s", name)
	}
	return song, nil
}

func ScanRowsIntoSongs(rows *sql.Rows) (*types.Song, error) {
	song := new(types.Song)

	err := rows.Scan(
		&song.ID,
		&song.SongName,
		&song.SongLyrics,
		&song.Group,
		&song.Published,
		&song.Link)

	if err != nil {
		return nil, err
	}

	return song, nil
}

func (s *Store) AddSong(name, group string) error {
	return nil
}

func (s *Store) UpdateSongInfo(name string) error {
	return nil
}

func (s *Store) DeleteSong(name, group string) error {
	return nil
}
