-- Create the `groups` table
CREATE TABLE IF NOT EXISTS groups (
                                      id SERIAL PRIMARY KEY,
                                      groupName VARCHAR(255) NOT NULL
);

-- Create a unique index on groupName for faster lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_groups_groupName ON groups(groupName);

-- Create the `songs` table
CREATE TABLE IF NOT EXISTS songs (
                                     id SERIAL PRIMARY KEY,
                                     songName VARCHAR(255) NOT NULL,
                                     songGroupId INTEGER NOT NULL,
                                     songLyrics TEXT[],
                                     published TIMESTAMP,
                                     link VARCHAR(255),
                                     FOREIGN KEY (songGroupId) REFERENCES groups(id) ON DELETE CASCADE
);

-- Add indexes for frequent search patterns
-- Index for searching songs by name
CREATE INDEX IF NOT EXISTS idx_songs_songName ON songs(songName);

-- Composite index for searching songs by group and name
CREATE INDEX IF NOT EXISTS idx_songs_songGroup_songName ON songs(songGroupId, songName);

-- Index for joining songs and groups on songGroupId
CREATE INDEX IF NOT EXISTS idx_songs_songGroupID ON songs(songGroupId);
