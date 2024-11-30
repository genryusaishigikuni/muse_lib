package main

import (
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // PostgresSQL driver
	"log"
	"os"
)

func main() {
	// Connect to the PostgresSQL database
	db, err := db.NewPostgresStorage(config.Envs.DBUser, config.Envs.DBPassword, config.Envs.DBAddress, config.Envs.DBName, "disable")
	if err != nil {
		log.Fatal(err)
	}

	// Create a migration driver for PostgresSQL
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize migrations
	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations", // Path to migrations
		"postgres",                      // Database type for PostgresSQL
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get current migration version
	v, d, _ := m.Version()
	log.Printf("Version: %d, dirty: %v", v, d)

	// Command-line argument for migration direction (up or down)
	cmd := os.Args[len(os.Args)-1]
	if cmd == "up" {
		// Apply migrations up
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
	if cmd == "down" {
		// Apply migrations down
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}
}
