package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSslMode       string
	RedisAddr       string
	JWTSecret       string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
	ProfilePicDir   string
	ProfilePicRoute string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Ignore error if .env file is not found, for production environments
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	accessTokenExpMin, _ := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_EXP_MIN", "10"))
	refreshTokenExpHour, _ := strconv.Atoi(getEnv("JWT_REFRESH_TOKEN_EXP_HOUR", "8"))

	cfg := &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "user"),
		DBPassword:      getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "quikchat"),
		DBSslMode:       getEnv("DB_SSLMODE", "disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:       getEnv("JWT_SECRET", "a_very_secret_key"),
		AccessTokenExp:  time.Duration(accessTokenExpMin) * time.Minute,
		RefreshTokenExp: time.Duration(refreshTokenExpHour) * time.Hour,
		ProfilePicDir:   getEnv("PROFILE_PIC_DIR", "./uploads/profile_pics"),
		ProfilePicRoute: getEnv("PROFILE_PIC_ROUTE", "/static/profile_pics/"),
	}

	// Ensure profile pic directory exists
	if err := os.MkdirAll(cfg.ProfilePicDir, os.ModePerm); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

