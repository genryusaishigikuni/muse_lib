package main

import (
	"database/sql"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/cmd/server"
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/db"
	"github.com/genryusaishigikuni/muse_lib/logger"
	"log/slog"
)

func main() {
	// Set up logger
	var env = config.Envs.Environment
	logs := logger.SetupLogger(env)
	logs.Info("starting muse_lib", slog.String("env", "local"))
	logs.Debug("debug mode is on")

	// Retrieve PostgresSQL configuration from environment variables or config
	dbUser := config.Envs.DBUser
	dbPassword := config.Envs.DBPassword
	dbAddress := config.Envs.DBAddress
	dbName := config.Envs.DBName
	sslMode := "disable"

	// Initialize PostgresSQL storage
	postgresDB, err := db.NewPostgresStorage(dbUser, dbPassword, dbAddress, dbName, sslMode)
	if err != nil {
		logs.Error("Failed to initialize a new connection to PostgresSQL", logger.Err(err))
		return
	}

	// Verify database connection
	initStorage(postgresDB, logs)

	// Start server
	newServer := server.NewServer(fmt.Sprintf(":%s", config.Envs.Port), postgresDB)
	err = newServer.Start()
	if err != nil {
		logs.Error("Failed to start server", logger.Err(err))
	}
}

func initStorage(db *sql.DB, logs *slog.Logger) {
	err := db.Ping()
	if err != nil {
		logs.Error("Failed to verify connection to PostgresSQL", logger.Err(err))
		return
	}
	logs.Info("Successfully connected to PostgresSQL")
}
