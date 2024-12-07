package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Environment string
	PublicHost  string
	Port        string

	DBUser     string
	DBPassword string
	DBAddress  string
	DBName     string

	ExtApi string
}

var Envs = initConfig()

func initConfig() Config {

	err := godotenv.Load("./.env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	return Config{
		Environment: getEnv("ENVIRONMENT", "local"),
		PublicHost:  getEnv("PUBLIC_HOST", "http://localhost"),
		Port:        getEnv("PORT", ":8080"),
		DBUser:      getEnv("DB_USER", "tamerlan"),
		DBPassword:  getEnv("DB_PASSWORD", "tamerlan123"),
		DBAddress:   fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "5432")),
		DBName:      getEnv("DB_NAME", "muse_lib"),
		ExtApi:      getEnv("EXT_API", "http://localhost:8081/info"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
