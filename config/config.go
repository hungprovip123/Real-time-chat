package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	RedisURL          string
	JWTSecret         string
	RateLimitRequests int
	RateLimitWindow   int
}

func Load() *Config {
	// Try to load .env file, but don't panic if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("Không tìm thấy file .env, sử dụng biến môi trường")
	}

	// Parse rate limit requests
	rateLimitRequests, err := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "10"))
	if err != nil {
		rateLimitRequests = 10
	}

	// Parse rate limit window
	rateLimitWindow, err := strconv.Atoi(getEnv("RATE_LIMIT_WINDOW", "60"))
	if err != nil {
		rateLimitWindow = 60
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:         getEnv("JWT_SECRET", "week3-secret-key"),
		RateLimitRequests: rateLimitRequests,
		RateLimitWindow:   rateLimitWindow,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
