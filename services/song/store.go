package song

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/logger" // Import the logger package
	"github.com/genryusaishigikuni/muse_lib/types"
	"github.com/lib/pq"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"
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

	baseQuery := `SELECT s.id, s.songName, g.groupName, s.songLyrics, s.published, s.link 
                  FROM songs s
                  JOIN groups g ON s.songGroupId = g.id`
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	filterMappings := map[string]string{
		"song":      "s.songName",
		"group":     "g.groupName",
		"published": "s.published",
		"lyrics":    "s.songLyrics",
		"link":      "s.link",
		"id":        "s.id", // This is an integer, so no LOWER()
	}

	for key, values := range filters {
		columnName, ok := filterMappings[key]
		if !ok {
			continue // Skip unsupported filters
		}

		// If the column is a string, apply LOWER() for case-insensitive comparison
		if columnName == "s.songName" || columnName == "g.groupName" || columnName == "s.link" || columnName == "s.songLyrics" {
			// Support single and multiple values per filter
			if len(values) > 1 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("LOWER(%s) IN (%s)", columnName, strings.Join(placeholders, ", ")))
			} else if len(values) == 1 {
				whereClauses = append(whereClauses, fmt.Sprintf("LOWER(%s) = LOWER($%d)", columnName, argIndex))
				args = append(args, values[0])
				argIndex++
			}
		} else {
			// For non-string columns, just use the value directly (e.g., ID)
			if len(values) > 1 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", columnName, strings.Join(placeholders, ", ")))
			} else if len(values) == 1 {
				whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", columnName, argIndex))
				args = append(args, values[0])
				argIndex++
			}
		}
	}

	// Build the WHERE clause
	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add LIMIT and OFFSET
	limit := 10
	offset := 0
	if l, ok := filters["limit"]; ok && len(l) > 0 {
		limit, _ = strconv.Atoi(l[0])
	}
	if o, ok := filters["offset"]; ok && len(o) > 0 {
		offset, _ = strconv.Atoi(o[0])
	}
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	s.log.Debug("Executing query", "operation", op, "query", baseQuery, "args", args)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		s.log.Error("Error executing query", "operation", op, logger.Err(err))
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var songs []types.Song
	for rows.Next() {
		var song types.Song
		var groupName string
		err := rows.Scan(&song.ID, &song.SongName, &groupName, pq.Array(&song.SongLyrics), &song.Published, &song.Link)
		if err != nil {
			s.log.Error("Error scanning song", "operation", op, logger.Err(err))
			return nil, err
		}
		song.Group = groupName
		songs = append(songs, song)
	}

	s.log.Debug("Fetched songs", "operation", op, "songs_count", len(songs))
	return songs, nil
}

// DeleteSong removes a song from the database by name and group.
func (s *Store) DeleteSong(name, group, link string, id int) error {
	const op = "song.DeleteSong"
	s.log.Info("Deleting song", "operation", op, "name", name, "group", group, "link", link, "id", id)

	var groupId int
	// If group is provided, get the group ID
	if group != "" {
		queryGroup := `SELECT id FROM groups WHERE groupName = $1`
		err := s.db.QueryRow(queryGroup, group).Scan(&groupId)
		if err != nil {
			s.log.Error("Error finding group ID", "operation", op, "group", group, logger.Err(err))
			return err
		}
	}

	// Construct the DELETE query based on provided parameters
	var query string
	var args []interface{}

	// Determine which filters to apply for deletion
	if id != 0 {
		// Delete by ID
		query = `DELETE FROM songs WHERE id = $1`
		args = append(args, id)
	} else {
		// Otherwise, use name and group or link
		query = `DELETE FROM songs WHERE songName = $1 AND songGroupId = $2`
		args = append(args, name)

		// Only use group ID if group is provided
		if group != "" {
			args = append(args, groupId)
		} else {
			// If no group, match by link (assuming link is unique)
			query = `DELETE FROM songs WHERE link = $1`
			args = append(args, link)
		}
	}

	// Execute the delete query
	result, err := s.db.Exec(query, args...)
	if err != nil {
		s.log.Error("Error deleting song", "operation", op, "name", name, "group", group, "link", link, logger.Err(err))
		return err
	}

	// Check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("Error retrieving rows affected", "operation", op, logger.Err(err))
		return err
	}

	if rowsAffected == 0 {
		s.log.Warn("No song found to delete", "operation", op, "name", name, "group", group, "link", link)
		return errors.New("song not found")
	}

	// Check if the group has no more songs associated with it
	queryCount := `SELECT COUNT(*) FROM songs WHERE songGroupId = $1`
	var count int
	err = s.db.QueryRow(queryCount, groupId).Scan(&count)
	if err != nil {
		s.log.Error("Error counting songs in group", "operation", op, "groupId", groupId, logger.Err(err))
		return err
	}

	// If no more songs in the group, delete the group
	if count == 0 && groupId != 0 {
		queryDeleteGroup := `DELETE FROM groups WHERE id = $1`
		_, err = s.db.Exec(queryDeleteGroup, groupId)
		if err != nil {
			s.log.Error("Error deleting group", "operation", op, "groupId", groupId, logger.Err(err))
			return err
		}
		s.log.Info("Group deleted successfully", "operation", op, "groupId", groupId)
	}

	s.log.Info("Song deleted successfully", "operation", op, "name", name, "group", group, "link", link)
	return nil
}

// UpdateSongInfo updates a song's details in the database.
func (s *Store) UpdateSongInfo(id int, name, group string, lyrics interface{}, published time.Time, link string) error {
	const op = "song.UpdateSongInfo"
	s.log.Info("Updating song info", "operation", op, "id", id, "name", name, "group", group)

	// Check if the song exists by ID
	var songExists bool
	query := `SELECT EXISTS(SELECT 1 FROM songs WHERE id = $1)`
	err := s.db.QueryRow(query, id).Scan(&songExists)
	if err != nil {
		s.log.Error("Error checking song existence", "operation", op, logger.Err(err))
		return err
	}

	if !songExists {
		s.log.Warn("Song not found", "operation", op, "id", id)
		return fmt.Errorf("song with ID %d not found", id)
	}

	// Fetch groupId for the given group name
	var groupId int
	query = `SELECT id FROM groups WHERE groupName = $1`
	err = s.db.QueryRow(query, group).Scan(&groupId)
	if err != nil {
		s.log.Error("Error fetching groupId", "operation", op, "group", group, logger.Err(err))
		return fmt.Errorf("group '%s' not found", group)
	}

	// Log current song info before update
	var currentSong types.Song
	query = `SELECT s.id, s.songName, g.groupName, s.songLyrics, s.published, s.link
			 FROM songs s
			 JOIN groups g ON s.songGroupId = g.id
			 WHERE s.id = $1`
	err = s.db.QueryRow(query, id).Scan(&currentSong.ID, &currentSong.SongName, &currentSong.Group, pq.Array(&currentSong.SongLyrics), &currentSong.Published, &currentSong.Link)
	if err != nil {
		s.log.Error("Error fetching current song", "operation", op, "id", id, logger.Err(err))
	} else {
		s.log.Info("Current song info before update", "operation", op, "song", currentSong)
	}

	// Build the query dynamically based on provided fields
	query = `UPDATE songs SET `
	var args []interface{}
	argIndex := 1

	// Handle lyrics if provided and cast to []string if it's an array
	if l, ok := lyrics.([]string); ok && len(l) > 0 {
		query += fmt.Sprintf("songLyrics = $%d, ", argIndex)
		args = append(args, pq.Array(l)) // Use pq.Array to convert slice to a suitable format
		argIndex++
	}

	// Add songName to update if provided
	if name != "" {
		query += fmt.Sprintf("songName = $%d, ", argIndex)
		args = append(args, name)
		argIndex++
	}

	// Add groupId to update if provided (after fetching the correct groupId)
	if groupId > 0 {
		query += fmt.Sprintf("songGroupId = $%d, ", argIndex)
		args = append(args, groupId)
		argIndex++
	}

	// Add published date to update if provided
	if !published.IsZero() {
		query += fmt.Sprintf("published = $%d, ", argIndex)
		args = append(args, published)
		argIndex++
	}

	// Add link to update if provided
	if link != "" {
		query += fmt.Sprintf("link = $%d, ", argIndex)
		args = append(args, link)
		argIndex++
	}

	// If no fields are provided to update, return an error
	if len(args) == 0 {
		s.log.Warn("No fields to update", "operation", op)
		return fmt.Errorf("no fields to update")
	}

	// Trim the trailing comma and space
	query = query[:len(query)-2]

	// Add the WHERE condition to update the song by ID
	query += ` WHERE id = $` + fmt.Sprintf("%d", argIndex)
	args = append(args, id)

	// Log the full query for debugging
	s.log.Info("Executing update query", "operation", op, "query", query, "args", args)

	// Execute the update query
	result, err := s.db.Exec(query, args...)
	if err != nil {
		s.log.Error("Error executing update query", "operation", op, "query", query, "args", args, logger.Err(err))
		return err
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("Error retrieving rows affected", "operation", op, logger.Err(err))
		return err
	}

	// Log the result to ensure that it actually updated
	s.log.Info("Rows affected", "operation", op, "rows_affected", rowsAffected)

	// If no rows were affected, it means the song was not found
	if rowsAffected == 0 {
		s.log.Warn("No changes made to the song info", "operation", op, "id", id)
		return fmt.Errorf("no changes made to the song with ID %d", id)
	}

	s.log.Info("Song info updated successfully", "operation", op, "id", id)
	return nil
}

func (s *Store) AddSong(song, group string, songDetails *types.SongDetail, songLyrics []string) error {
	const op = "song.AddSong"
	s.log.Info("Adding new song", "operation", op, "name", song, "group", group)

	// Check if the group exists, if not, create it
	var groupID int
	err := s.db.QueryRow(`SELECT id FROM groups WHERE groupName = $1`, group).Scan(&groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Group doesn't exist, insert it
			err = s.db.QueryRow(`INSERT INTO groups (groupName) VALUES ($1) RETURNING id`, group).Scan(&groupID)
			if err != nil {
				s.log.Error("Error creating group", "operation", op, "group", group, logger.Err(err))
				return err
			}
		} else {
			s.log.Error("Error checking group", "operation", op, "group", group, logger.Err(err))
			return err
		}
	}

	// Insert the song with the obtained groupID
	query := `INSERT INTO songs (songName, songGroupId, songLyrics, published, link) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err = s.db.Exec(query, song, groupID, pq.Array(songLyrics), songDetails.ReleaseDate, songDetails.Link)
	if err != nil {
		s.log.Error("Error adding song", "operation", op, "name", song, "group", group, logger.Err(err))
		return err
	}

	s.log.Info("Song added successfully", "operation", op, "name", song, "group", group)
	return nil
}
