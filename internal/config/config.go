package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	JWTSecret           string
	Port                string
	HoldDurationMinutes int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	holdMinutes, _ := strconv.Atoi(getEnv("HOLD_DURATION_MINUTES", "10"))

	return &Config{
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", "postgres"),
		DBName:              getEnv("DB_NAME", "movie_reservation"),
		JWTSecret:           getEnv("JWT_SECRET", "secret"),
		Port:                getEnv("PORT", "8080"),
		HoldDurationMinutes: holdMinutes,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
