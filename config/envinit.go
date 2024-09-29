package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var Conf Config

func init() {
	// loads values from .env into the system
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Необходим файл .env с токенами, консоль закроется сама через 10 секунд.")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
	Conf = New()
}

func New() Config {
	return Config{
		SpotifyClientID:     getEnv("SPOTIFY_CLIENT_ID", ""),
		SpotifyClientSecret: getEnv("SPOTIFY_CLIENT_SECRET", ""),
		SpotifyTokenURL:     "https://accounts.spotify.com/api/token",
		RedirectURL:         "http://localhost:8888/",
		SpotifyRedirectURI:  "http://localhost:8888/callback",
		TelegramAPI_id:      getEnvAsInt("TELEGRAM_API_ID", 0),
		TelegramAPI_hash:    getEnv("TELEGRAM_API_HASH", ""),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int32 {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return int32(value)
	}

	return int32(defaultVal)
}
