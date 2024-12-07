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
	log *slog.Logger
}

func NewStore(db *sql.DB, env string) *Store {
	log := logger.SetupLogger(env)
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
		"id":        "s.id",
	}

	for key, values := range filters {
		columnName, ok := filterMappings[key]
		if !ok {
			continue
		}

		switch columnName {
		case "s.songName", "g.groupName", "s.link":
			// String fields: Case-insensitive match
			if len(values) > 1 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("LOWER(%s) IN (%s)", columnName, strings.Join(placeholders, ", ")))
			} else {
				whereClauses = append(whereClauses, fmt.Sprintf("LOWER(%s) = LOWER($%d)", columnName, argIndex))
				args = append(args, values[0])
				argIndex++
			}

		case "s.published":
			// Time field: Parse and handle date comparison
			if len(values) == 1 {
				publishedTime, err := time.Parse("2006-01-02", values[0])
				if err != nil {
					return nil, fmt.Errorf("invalid date format for 'published': %v", err)
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", columnName, argIndex))
				args = append(args, publishedTime)
				argIndex++
			}

		case "s.songLyrics":
			// Array field: Match at least one of the lyrics
			if len(values) > 0 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s && ARRAY[%s]::text[]", columnName, strings.Join(placeholders, ", ")))
			}

		case "s.id":
			// Ensure numeric comparisons for IDs
			if len(values) == 1 {
				id, err := strconv.Atoi(values[0])
				if err != nil {
					return nil, fmt.Errorf("invalid id format: %v", err)
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", columnName, argIndex))
				args = append(args, id)
				argIndex++
			} else if len(values) > 1 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					id, err := strconv.Atoi(v)
					if err != nil {
						return nil, fmt.Errorf("invalid id format: %v", err)
					}
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, id)
					argIndex++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", columnName, strings.Join(placeholders, ", ")))
			}

		}
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

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

func (s *Store) DeleteSong(name, group, link string, id int) error {
	const op = "song.DeleteSong"
	s.log.Info("Deleting song", "operation", op, "name", name, "group", group, "link", link, "id", id)

	var groupId int
	if group != "" {
		queryGroup := `SELECT id FROM groups WHERE groupName = $1`
		err := s.db.QueryRow(queryGroup, group).Scan(&groupId)
		if err != nil {
			s.log.Error("Error finding group ID", "operation", op, "group", group, logger.Err(err))
			return err
		}
	}

	var query string
	var args []interface{}

	if id != 0 {
		query = `DELETE FROM songs WHERE id = $1`
		args = append(args, id)
	} else {
		query = `DELETE FROM songs WHERE songName = $1 AND songGroupId = $2`
		args = append(args, name)

		if group != "" {
			args = append(args, groupId)
		} else {
			query = `DELETE FROM songs WHERE link = $1`
			args = append(args, link)
		}
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		s.log.Error("Error deleting song", "operation", op, "name", name, "group", group, "link", link, logger.Err(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("Error retrieving rows affected", "operation", op, logger.Err(err))
		return err
	}

	if rowsAffected == 0 {
		s.log.Warn("No song found to delete", "operation", op, "name", name, "group", group, "link", link)
		return errors.New("song not found")
	}

	queryCount := `SELECT COUNT(*) FROM songs WHERE songGroupId = $1`
	var count int
	err = s.db.QueryRow(queryCount, groupId).Scan(&count)
	if err != nil {
		s.log.Error("Error counting songs in group", "operation", op, "groupId", groupId, logger.Err(err))
		return err
	}

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

func (s *Store) UpdateSongInfo(id int, name, group string, lyrics interface{}, published time.Time, link string) error {
	const op = "song.UpdateSongInfo"
	s.log.Info("Updating song info", "operation", op, "id", id, "name", name, "group", group)

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

	groupId := -1
	var oldGroupId int

	query = `SELECT songGroupId FROM songs WHERE id = $1`
	err = s.db.QueryRow(query, id).Scan(&oldGroupId)
	if err != nil {
		s.log.Error("Error fetching old groupId", "operation", op, "id", id, logger.Err(err))
		return err
	}

	if group != "" {
		query = `SELECT id FROM groups WHERE groupName = $1`
		err = s.db.QueryRow(query, group).Scan(&groupId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.log.Error("Error fetching groupId", "operation", op, "group", group, logger.Err(err))
			return fmt.Errorf("group '%s' not found", group)
		}

		if errors.Is(err, sql.ErrNoRows) {
			query = `INSERT INTO groups (groupName) VALUES ($1) RETURNING id`
			err = s.db.QueryRow(query, group).Scan(&groupId)
			if err != nil {
				s.log.Error("Error creating new group", "operation", op, "group", group, logger.Err(err))
				return fmt.Errorf("could not create group '%s'", group)
			}
			s.log.Info("Created new group", "operation", op, "group", group, "groupId", groupId)
		}
	}

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

	query = `UPDATE songs SET `
	var args []interface{}
	argIndex := 1

	if l, ok := lyrics.([]string); ok && len(l) > 0 {
		query += fmt.Sprintf("songLyrics = $%d, ", argIndex)
		args = append(args, pq.Array(l))
		argIndex++
	}

	if name != "" {
		query += fmt.Sprintf("songName = $%d, ", argIndex)
		args = append(args, name)
		argIndex++
	}

	if groupId > -1 {
		query += fmt.Sprintf("songGroupId = $%d, ", argIndex)
		args = append(args, groupId)
		argIndex++
	}

	if !published.IsZero() {
		query += fmt.Sprintf("published = $%d, ", argIndex)
		args = append(args, published)
		argIndex++
	}

	if link != "" {
		query += fmt.Sprintf("link = $%d, ", argIndex)
		args = append(args, link)
		argIndex++
	}

	if len(args) == 0 {
		s.log.Warn("No fields to update", "operation", op)
		return fmt.Errorf("no fields to update")
	}

	query = query[:len(query)-2]

	query += ` WHERE id = $` + fmt.Sprintf("%d", argIndex)
	args = append(args, id)

	s.log.Info("Executing update query", "operation", op, "query", query, "args", args)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		s.log.Error("Error executing update query", "operation", op, "query", query, "args", args, logger.Err(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.log.Error("Error retrieving rows affected", "operation", op, logger.Err(err))
		return err
	}

	s.log.Info("Rows affected", "operation", op, "rows_affected", rowsAffected)

	if rowsAffected == 0 {
		s.log.Warn("No changes made to the song info", "operation", op, "id", id)
		return fmt.Errorf("no changes made to the song with ID %d", id)
	}

	if oldGroupId > -1 {
		var count int
		query = `SELECT COUNT(*) FROM songs WHERE songGroupId = $1`
		err = s.db.QueryRow(query, oldGroupId).Scan(&count)
		if err != nil {
			s.log.Error("Error counting songs for old group", "operation", op, "oldGroupId", oldGroupId, logger.Err(err))
		} else if count == 0 {
			// No songs left in the old group, delete the group
			query = `DELETE FROM groups WHERE id = $1`
			_, err := s.db.Exec(query, oldGroupId)
			if err != nil {
				s.log.Error("Error deleting old group", "operation", op, "oldGroupId", oldGroupId, logger.Err(err))
				return err
			}
			s.log.Info("Deleted old group", "operation", op, "oldGroupId", oldGroupId)
		}
	}

	s.log.Info("Song info updated successfully", "operation", op, "id", id)
	return nil
}

func (s *Store) AddSong(song, group string, songDetails *types.SongDetail, songLyrics []string) error {
	const op = "song.AddSong"
	s.log.Info("Adding new song", "operation", op, "name", song, "group", group)

	var groupID int
	err := s.db.QueryRow(`SELECT id FROM groups WHERE groupName = $1`, group).Scan(&groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
