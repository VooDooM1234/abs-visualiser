package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type apiKeys struct {
	weather string
	abs     string
}

type Config struct {
	postgresURL string
}

var keys apiKeys
var config Config

func Init() {
	_ = godotenv.Load("../.env")

	config.postgresURL = os.Getenv("DATABASE_URL")
	keys.weather = os.Getenv("WEATHER_API_KEY")
	keys.abs = os.Getenv("ABS_API_KEY")

	if keys.weather == "" {
		log.Fatal("Missing WEATHER_API_KEY in environment!")
	}
	if keys.abs == "" {
		log.Fatal("Missing ABS_API_KEY in environment!")
	}

	if config.postgresURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

}

// exported getter functions
func GetABSKey() string {
	return keys.abs
}

func GetWeatherKey() string {
	return keys.weather
}

func GetPostgresURL() string {
	return config.postgresURL
}
