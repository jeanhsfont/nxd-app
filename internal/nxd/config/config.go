package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var JWTKey []byte

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		JWTKey = []byte(jwtSecret)
	}
}
