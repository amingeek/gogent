package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	BASE_URL string
	API_KEY  string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	BASE_URL = os.Getenv("BASE_URL")
	API_KEY = os.Getenv("API_KEY")
}

func GetApiKey() string {
	return API_KEY
}

func GetApiBaseUrl() string {
	return BASE_URL
}
