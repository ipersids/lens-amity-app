package config

import (
	"lensamity/internal/auth"
	"log"
	"os"
	"time"
)

type Config struct {
	Auth        auth.Config
	DatabaseURL string
}

func LoadAuth() auth.Config {
	return auth.Config{
		JWTsecret:     required("JWT_SECRET"),
		RefreshSecret: required("REFRESH_SECRET"),
		JWTexpiry:     15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}
}

func LoadDB() string {
	return required("DATABASE_URL")
}

func required(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %s doesn't exist", key)
	}
	return val
}

func Load() *Config {
	return &Config{
		Auth:        LoadAuth(),
		DatabaseURL: LoadDB(),
	}
}
