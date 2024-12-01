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

// @title Muse_Library App API
// @version 1.0
// @description API Server for Music Library App

// @host localhost:8080
// @BasePath /api/
func main() {
	const op = "main" // Operation context for consistent logging

	// Set up logger
	env := config.Envs.Environment
	logs := logger.SetupLogger(env)
	logs.Info("Starting muse_lib", slog.String("env", env))
	logs.Debug("Debug mode is enabled", slog.String("operation", op))

	// Retrieve PostgresSQL configuration from environment variables or config
	logs.Info("Loading database configuration", slog.String("operation", op))
	dbUser := config.Envs.DBUser
	dbPassword := config.Envs.DBPassword
	dbAddress := config.Envs.DBAddress
	dbName := config.Envs.DBName
	sslMode := "disable"

	// Log database connection details (excluding sensitive data)
	logs.Debug("Database config loaded",
		slog.String("user", dbUser),
		slog.String("address", dbAddress),
		slog.String("db_name", dbName),
		slog.String("operation", op),
	)

	// Initialize PostgresSQL storage
	logs.Info("Initializing PostgresSQL storage", slog.String("operation", op))
	postgresDB, err := db.NewPostgresStorage(dbUser, dbPassword, dbAddress, dbName, sslMode)
	if err != nil {
		logs.Error("Failed to initialize PostgresSQL connection", logger.Err(err), slog.String("operation", op))
		return
	}

	// Verify database connection
	initStorage(postgresDB, logs)

	// Start the server
	logs.Info("Starting HTTP server", slog.String("port", config.Envs.Port), slog.String("operation", op))
	newServer := server.NewServer(fmt.Sprintf(":%s", config.Envs.Port), postgresDB)
	err = newServer.Start()
	if err != nil {
		logs.Error("Failed to start server", logger.Err(err), slog.String("operation", op))
		return
	}
	logs.Info("Server started successfully", slog.String("operation", op))
}

func initStorage(db *sql.DB, logs *slog.Logger) {
	const op = "main.initStorage"

	logs.Info("Verifying database connection", slog.String("operation", op))
	err := db.Ping()
	if err != nil {
		logs.Error("Failed to verify connection to PostgresSQL", logger.Err(err), slog.String("operation", op))
		return
	}
	logs.Info("Successfully connected to PostgresSQL", slog.String("operation", op))
}
