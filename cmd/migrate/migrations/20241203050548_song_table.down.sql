-- Drop indexes before dropping tables
DROP INDEX IF EXISTS idx_songs_songgroupid;
DROP INDEX IF EXISTS idx_songs_songgroup_songname;
DROP INDEX IF EXISTS idx_songs_songname;
DROP INDEX IF EXISTS idx_groups_groupname;

-- Drop tables
DROP TABLE IF EXISTS songs;
DROP TABLE IF EXISTS groups;
