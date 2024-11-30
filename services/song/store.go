package song

import (
	"database/sql"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/lib/pq"
	"log"
	"net/url"
	"strings"
)

type Store struct {
	db *sql.DB
}

// NewStore creates a new Store instance with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetSongs(filters url.Values) ([]types.Song, error) {
	baseQuery := `SELECT id, songName, songGroup, songLyrics, published, link FROM songs`
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	for key, values := range filters {
		// If multiple values for a parameter, use IN clause
		if len(values) > 1 {
			placeholders := make([]string, len(values))
			for i, v := range values {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, v)
				argIndex++
			}
			whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", key, strings.Join(placeholders, ", ")))
		} else if len(values) == 1 { // Single value
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, values[0])
			argIndex++
		}
	}

	// Add WHERE clauses if filters exist
	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []types.Song
	for rows.Next() {
		var song types.Song
		err := rows.Scan(&song.ID, &song.SongName, &song.Group, pq.Array(&song.SongLyrics), &song.Published, &song.Link)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
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
