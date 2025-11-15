package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDBURI string
	MongoDBDatabase string
    ServerPort string
	GinMode string
}

func Load() (*Config, error) {
	godotenv.Load()
	return &Config{
		MongoDBURI: getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase: getEnv("MONGODB_DATABASE","idol_database"),
		ServerPort: getEnv("SERVER_PORT", "8081"),
		GinMode: getEnv("GIN_MODE", "debug"),
		}, nil
}

func getEnv(key, defaultValue string) string {    
	if value := os.Getenv(key); value != "" {
		return value
    }
	return defaultValue
}