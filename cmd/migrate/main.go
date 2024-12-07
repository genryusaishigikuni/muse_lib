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
	var env = config.Envs.Environment
	log := logger.SetupLogger(env)

	log.Info("Starting migration process")

	log.Debug("Connecting to database", "user", config.Envs.DBUser, "address", config.Envs.DBAddress, "database", config.Envs.DBName)

	database, err := db.NewPostgresStorage(config.Envs.DBUser, config.Envs.DBPassword, config.Envs.DBAddress, config.Envs.DBName, "disable")
	if err != nil {
		log.Error("Failed to connect to database", logger.Err(err))
		return
	}

	log.Debug("Connected to database successfully")

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		log.Error("Failed to create migration driver", logger.Err(err))
		return
	}

	log.Info("Initializing migrations")

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Error("Failed to initialize migrations", logger.Err(err))
		return
	}

	v, d, _ := m.Version()
	log.Info("Current migration version", "version", v, "dirty", d)

	cmd := os.Args[len(os.Args)-1]
	if cmd == "up" {
		log.Info("Applying migrations up")
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Error("Migration up failed", logger.Err(err))
			return
		}
	} else if cmd == "down" {
		log.Info("Applying migrations down")
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Error("Migration down failed", logger.Err(err))
			return
		}
	} else {
		log.Warn("Invalid command", "command", cmd)
	}
}
