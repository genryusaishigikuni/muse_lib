package main

import (
	"database/sql"
	"fmt"
	"github.com/genryusaishigikuni/muse_lib/cmd/server"
	"github.com/genryusaishigikuni/muse_lib/config"
	"github.com/genryusaishigikuni/muse_lib/db"
	"log"
)

func main() {
	// Retrieve PostgresSQL configuration from environment variables or config
	dbUser := config.Envs.DBUser
	dbPassword := config.Envs.DBPassword
	dbAddress := config.Envs.DBAddress
	dbName := config.Envs.DBName
	sslMode := "disable"

	// Initialize the PostgresSQL storage
	postgresDB, err := db.NewPostgresStorage(dbUser, dbPassword, dbAddress, dbName, sslMode)
	if err != nil {
		log.Fatal("Failed to connect to PostgresSQL:", err)
	}

	initStorage(postgresDB)

	newServer := server.NewServer(fmt.Sprintf(":%s", config.Envs.Port), postgresDB)
	err = newServer.Start()
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}
