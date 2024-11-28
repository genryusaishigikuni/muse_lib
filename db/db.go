package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgresSQL driver
)

// NewPostgresStorage initializes a new PostgresSQL connection without SSL mode.
func NewPostgresStorage(user, password, address, dbname string, sslMode string) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		user, password, address, dbname, sslMode,
	)

	// Open the connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open connection:", err)
		return nil, err
	}

	return db, nil
}
