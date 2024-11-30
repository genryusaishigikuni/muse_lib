package song

import (
	"database/sql"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/lib/pq"
	"log"
)

type Store struct {
	db *sql.DB
}

// NewStore creates a new Store instance with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetSongByName fetches a song by its name.
func (s *Store) GetSongByName(name string) (*types.Song, error) {
	var song types.Song
	query := `SELECT id, songName, songGroup, songLyrics, published, link FROM songs WHERE songName = $1`
	err := s.db.QueryRow(query, name).Scan(&song.ID, &song.SongName, &song.Group, pq.Array(&song.SongLyrics), &song.Published, &song.Link)
	if err != nil {
		log.Printf("Error retrieving song: %v", err)
		return nil, err
	}
	return &song, nil
}

// DeleteSong removes a song from the database by name and group.
func (s *Store) DeleteSong(name, group string) error {
	query := `DELETE FROM songs WHERE songName = $1 AND songGroup = $2`
	_, err := s.db.Exec(query, name, group)
	if err != nil {
		log.Printf("Error deleting song: %v", err)
		return err
	}
	return nil
}

// UpdateSongInfo updates a song's details in the database.
func (s *Store) UpdateSongInfo(name, group string, lyrics interface{}, published string, link string) error {
	query := `UPDATE songs SET songLyrics = $1, published = $2, link = $3 WHERE songName = $4 AND songGroup = $5`
	_, err := s.db.Exec(query, lyrics, published, link, name, group)
	if err != nil {
		log.Printf("Error updating song: %v", err)
		return err
	}
	return nil
}

// AddSong adds a new song to the database with the provided details.
func (s *Store) AddSong(name, group string, songDetails *types.SongDetail) error {
	query := `INSERT INTO songs (songName, songGroup, songLyrics, published, link) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(query, name, group, songDetails.Text, songDetails.ReleaseDate, songDetails.Link)
	if err != nil {
		log.Printf("Error adding song: %v", err)
		return err
	}
	return nil
}
