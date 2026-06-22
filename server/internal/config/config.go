package config

import (
	"lensamity/internal/auth"
	"lensamity/internal/storage"
	"log"
	"os"
	"time"
)

type Config struct {
	Auth        auth.Config
	DatabaseURL string
	S3          storage.Config
}

func Load() *Config {
	return &Config{
		Auth:        LoadAuth(),
		DatabaseURL: LoadDB(),
		S3:          LoadS3(),
	}
}

func LoadAuth() auth.Config {
	return auth.Config{
		SessionSecret:   required("SESSION_SECRET"),
		IdleTimeout:     2 * 24 * time.Hour,
		AbsoluteTimeout: 15 * 24 * time.Hour,
		TouchInterval:   15 * time.Minute,
	}
}

func LoadDB() string {
	return required("DATABASE_URL")
}

func LoadS3() storage.Config {
	return storage.Config{
		Region:           required("S3_REGION"),
		AccessKeyID:      required("S3_ACCESS_KEY_ID"),
		SecretAccessKey:  required("S3_SECRET_ACCESS_KEY"),
		InternalEndpoint: required("S3_INTERNAL_ENDPOINT"),
		Backet:           required("S3_BUCKET"),
		UsePathStyle:     withDefault("S3_FORCE_PATH_STYLE", "true") != "false",
	}
}

func required(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %s doesn't exist", key)
	}
	return val
}

func withDefault(key, defaultKey string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultKey
	}
	return val
}
