package core

import (
	"errors"
	"os"
	"time"
)

type Auth struct {
	JWTsecret     string
	RefreshSecret string
	JWTexpiry     time.Duration
	RefreshExpiry time.Duration
}

type Config struct {
	Auth        Auth
	DatabaseURL string
}

func InitConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("Environment settings: DATABASE_URL doesn't exist")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("Environment settings: JWT_SECRET doesn't exist")
	}

	refreshSecret := os.Getenv("REFRESH_SECRET")
	if refreshSecret == "" {
		return nil, errors.New("Environment settings: REFRESH_SECRET doesn't exist")
	}

	return &Config{
		DatabaseURL: dbURL,
		Auth: Auth{
			JWTsecret:     jwtSecret,
			RefreshSecret: refreshSecret,
			JWTexpiry:     15 * time.Minute,
			RefreshExpiry: 24 * time.Hour,
		},
	}, nil
}
