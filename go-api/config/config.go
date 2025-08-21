package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type apiKeys struct {
	weather string
	abs     string
}

type Config struct {
	postgresURL       string
	PythonPath        string `json:"python_path"`
	DefaultChart      string `json:"default_chart"`
	DataSource        string `json:"data_source"`
	PlotServicePort   string `json:"plot_service_port"`
	PlotServiceScript string `json:"plot_service_script"`
	Host              string `json:"host"`
	Port              string `json:"port"`
	HTMLTemplates     string `json:"HTMLTemplates"`
}

var keys apiKeys
var config Config

func Init() (*Config, error) {
	_ = godotenv.Load(".env")

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

	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	err = json.NewDecoder(file).Decode(&cfg)
	return &cfg, err

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
