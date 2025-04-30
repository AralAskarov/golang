package config

import "os"
import "log"

type Config struct {
	Port 		 string
	DatabaseURL  string
}

func New() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}


	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	return &Config{
		Port:         port,
		DatabaseURL:  dbURL,
	}
}