package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	ACCESS_TTL    = 30 // days
	REFRESH_TTL = 1  // day
)

// Read variable from dot env by key
func GetEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
