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
	RedisPassword   string
	JWTSecret       string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
	ProfilePicDir   string
	ProfilePicRoute string
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Ignore error if .env file is not found, for production environments
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	serverPort := getEnv("SERVER_PORT", "8080")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "quikchat")
	dbSslMode := getEnv("DB_SSLMODE", "disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	jwtSecret := getEnv("JWT_SECRET", "a-very-secret-key-that-is-long-enough")
	profilePicDir := getEnv("PROFILE_PIC_DIR", "./uploads/profile_pics")
	profilePicRoute := getEnv("PROFILE_PIC_ROUTE", "/static/profile_pics")

	accessExpMin, _ := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_EXP_MIN", "10"))
	refreshExpHour, _ := strconv.Atoi(getEnv("JWT_REFRESH_TOKEN_EXP_HOUR", "8"))

	cfg := &Config{
		ServerPort:      serverPort,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBName:          dbName,
		DBSslMode:       dbSslMode,
		RedisAddr:       redisAddr,
		RedisPassword:   redisPassword,
		JWTSecret:       jwtSecret,
		AccessTokenExp:  time.Duration(accessExpMin) * time.Minute,
		RefreshTokenExp: time.Duration(refreshExpHour) * time.Hour,
		ProfilePicDir:   profilePicDir,
		ProfilePicRoute: profilePicRoute,
	}

	if err := os.MkdirAll(cfg.ProfilePicDir, os.ModePerm); err != nil {
		return nil, err
	}

	return cfg, nil
}

