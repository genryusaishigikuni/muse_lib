package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func NewPostgresStorage(user, password, address, dbname string, sslMode string) (*sql.DB, error) {
	const op = "db.NewPostgresStorage"
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		user, password, address, dbname, sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open connection: %w", op, err)
	}

	return db, nil
}
