CREATE TABLE IF NOT EXISTS songs (
                       id SERIAL PRIMARY KEY,
                       songName VARCHAR(255) NOT NULL,
                       songGroup VARCHAR(255) NOT NULL,
                       songLyrics TEXT[],
                       published TIMESTAMP,
                       link VARCHAR(255)
);
