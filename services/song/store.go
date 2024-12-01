package song

import (
	"database/sql"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/logger" // Import the logger package
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/lib/pq"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
)

type Store struct {
	db  *sql.DB
	log *slog.Logger // Add a logger field to the Store
}

// NewStore creates a new Store instance with the given database connection and logger.
func NewStore(db *sql.DB, env string) *Store {
	log := logger.SetupLogger(env) // Create a logger instance based on environment
	const op = "song.NewStore"
	log.Debug("Initializing new store", "operation", op)
	return &Store{db: db, log: log}
}

func (s *Store) GetSongs(filters url.Values) ([]types.Song, error) {
	const op = "song.GetSongs"
	s.log.Debug("Fetching songs with filters", "operation", op, "filters", filters)

	baseQuery := `SELECT id, songName, songGroup, songLyrics, published, link FROM songs`
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Handle filters
	for key, values := range filters {
		switch key {
		case "limit", "offset": // Skip pagination params in where clause
			continue
		default:
			if len(values) > 1 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", key, strings.Join(placeholders, ", ")))
			} else if len(values) == 1 {
				whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", key, argIndex))
				args = append(args, values[0])
				argIndex++
			}
		}
	}

	// Add WHERE clauses if filters exist
	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add pagination (limit and offset)
	limit := 10 // Default limit
	offset := 0 // Default offset

	if l, ok := filters["limit"]; ok && len(l) > 0 {
		limit, _ = strconv.Atoi(l[0])
	}
	if o, ok := filters["offset"]; ok && len(o) > 0 {
		offset, _ = strconv.Atoi(o[0])
	}

	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	// Log the query being executed with its parameters
	s.log.Debug("Executing query", "operation", op, "query", baseQuery, "args", args)

	// Execute the query
	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		s.log.Error("Error executing query", "operation", op, logger.Err(err))
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			s.log.Warn("Error closing rows", "operation", op, logger.Err(err))
		}
	}(rows)

	var songs []types.Song
	for rows.Next() {
		var song types.Song
		err := rows.Scan(&song.ID, &song.SongName, &song.Group, pq.Array(&song.SongLyrics), &song.Published, &song.Link)
		if err != nil {
			s.log.Error("Error scanning song", "operation", op, logger.Err(err))
			return nil, err
		}
		songs = append(songs, song)
	}

	s.log.Debug("Fetched songs", "operation", op, "songs_count", len(songs))
	return songs, nil
}

// DeleteSong removes a song from the database by name and group.
func (s *Store) DeleteSong(name, group string) error {
	const op = "song.DeleteSong"
	s.log.Info("Deleting song", "operation", op, "name", name, "group", group)

	query := `DELETE FROM songs WHERE songName = $1 AND songGroup = $2`
	_, err := s.db.Exec(query, name, group)
	if err != nil {
		s.log.Error("Error deleting song", "operation", op, "name", name, "group", group, logger.Err(err))
		return err
	}

	s.log.Info("Song deleted successfully", "operation", op, "name", name, "group", group)
	return nil
}

// UpdateSongInfo updates a song's details in the database.
func (s *Store) UpdateSongInfo(name, group string, lyrics interface{}, published string, link string) error {
	const op = "song.UpdateSongInfo"
	s.log.Info("Updating song info", "operation", op, "name", name, "group", group)

	query := `UPDATE songs SET songLyrics = $1, published = $2, link = $3 WHERE songName = $4 AND songGroup = $5`
	_, err := s.db.Exec(query, lyrics, published, link, name, group)
	if err != nil {
		s.log.Error("Error updating song", "operation", op, "name", name, "group", group, logger.Err(err))
		return err
	}

	s.log.Info("Song info updated", "operation", op, "name", name, "group", group)
	return nil
}

// AddSong adds a new song to the database with the provided details.
func (s *Store) AddSong(name, group string, songDetails *types.SongDetail) error {
	const op = "song.AddSong"
	s.log.Info("Adding new song", "operation", op, "name", name, "group", group)

	query := `INSERT INTO songs (songName, songGroup, songLyrics, published, link) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(query, name, group, songDetails.Text, songDetails.ReleaseDate, songDetails.Link)
	if err != nil {
		s.log.Error("Error adding song", "operation", op, "name", name, "group", group, logger.Err(err))
		return err
	}

	s.log.Info("Song added successfully", "operation", op, "name", name, "group", group)
	return nil
}
