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


	dbURLMYSQL := os.Getenv("DATABASE_URL_MYSQL")
	if dbURLMYSQL == "" {
		log.Fatal("DATABASE_URL_MYSQL environment variable is required")
	}

	return &Config{
		Port:         port,
		DatabaseURL:  dbURLMYSQL,
	}
}