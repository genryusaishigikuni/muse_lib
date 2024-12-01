package main

import (
	"errors"
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/db"
	"github.com/genryusaishigikuni/muse_lib/logger" // Import your logger package
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // PostgresSQL driver
	"os"
)

func main() {
	// Set up the logger based on the environment
	var env = config.Envs.Environment
	log := logger.SetupLogger(env)

	// Info: Starting migration process
	log.Info("Starting migration process")

	// Debug: Database connection parameters (ensure not to log sensitive info in production)
	log.Debug("Connecting to database", "user", config.Envs.DBUser, "address", config.Envs.DBAddress, "database", config.Envs.DBName)

	// Connect to the PostgresSQL database
	database, err := db.NewPostgresStorage(config.Envs.DBUser, config.Envs.DBPassword, config.Envs.DBAddress, config.Envs.DBName, "disable")
	if err != nil {
		// Error: Failed to connect to the database
		log.Error("Failed to connect to database", logger.Err(err))
		return
	}

	// Debug: Successfully connected to database
	log.Debug("Connected to database successfully")

	// Create a migration driver for PostgresSQL
	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		// Error: Failed to create migration driver
		log.Error("Failed to create migration driver", logger.Err(err))
		return
	}

	// Info: Initializing migrations
	log.Info("Initializing migrations")

	// Initialize migrations
	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations", // Path to migrations
		"postgres",                      // Database type for PostgresSQL
		driver,
	)
	if err != nil {
		// Error: Failed to initialize migrations
		log.Error("Failed to initialize migrations", logger.Err(err))
		return
	}

	// Get current migration version
	v, d, _ := m.Version()
	// Info: Current migration version
	log.Info("Current migration version", "version", v, "dirty", d)

	// Command-line argument for migration direction (up or down)
	cmd := os.Args[len(os.Args)-1]
	if cmd == "up" {
		// Info: Applying migrations up
		log.Info("Applying migrations up")
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			// Error: Migration up failed
			log.Error("Migration up failed", logger.Err(err))
			return
		}
	} else if cmd == "down" {
		// Info: Applying migrations down
		log.Info("Applying migrations down")
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			// Error: Migration down failed
			log.Error("Migration down failed", logger.Err(err))
			return
		}
	} else {
		// Warn: Invalid command
		log.Warn("Invalid command", "command", cmd)
	}
}
