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

var keys apiKeys

func Init() {
	_ = godotenv.Load()

	keys.weather = os.Getenv("WEATHER_API_KEY")
	keys.abs = os.Getenv("ABS_API_KEY")

	if keys.weather == "" {
		log.Fatal("Missing WEATHER_API_KEY in environment!")
	}
	if keys.abs == "" {
		log.Fatal("Missing ABS_API_KEY in environment!")
	}
}

// exported getter functions
func GetABSKey() string {
	return keys.abs
}

func GetWeatherKey() string {
	return keys.weather
}
