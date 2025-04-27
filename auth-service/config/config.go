package config

import "os"
import "log"

type Config struct {
	JWTSecretKey string
	Port 		 string
	DatabaseURL  string
}

func New() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	JWT := os.Getenv("JWT_SECRET_KEY")
	if JWT == "" {
		log.Fatal("JWT_SECRET_KEY environment variable is required")
	}

	return &Config{
		JWTSecretKey: JWT,
		Port:         port,
		DatabaseURL:  dbURL,
	}
}