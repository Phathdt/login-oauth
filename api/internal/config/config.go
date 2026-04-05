package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	Env                    string
	DatabaseURL            string
	JWTSecret              string
	FrontendURL            string
	FirebaseProjectID      string
	FirebaseCredentialsJSON string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := &Config{
		Port:                    getEnv("PORT", "3000"),
		Env:                     getEnv("ENV", "development"),
		DatabaseURL:             getEnv("DATABASE_URL", ""),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		FrontendURL:             getEnv("FRONTEND_URL", "http://localhost:5173"),
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredentialsJSON: getEnv("FIREBASE_CREDENTIALS_JSON", ""),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
