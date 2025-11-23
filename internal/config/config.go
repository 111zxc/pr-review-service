package config

import (
	"os"
	"strconv"
)

type Config struct {
	DB     DBConfig
	Logger LoggerConfig
	Server ServerConfig
	Env    string
}

type DBConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	SSLMode  string

	Port int
}

type ServerConfig struct {
	Port int
}

type LoggerConfig struct {
	Level  string
	Format string
}

func Load() *Config {
	env := getEnv("ENV", "development")

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "pr_user"),
			Password: getEnv("DB_PASSWORD", "pr_password"),
			Name:     getEnv("DB_NAME", "pr_review_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", getDefaultLogLevel(env)),
			Format: getEnv("LOG_FORMAT", getDefaultLogFormat(env)),
		},
		Env: env,
	}
}

func getDefaultLogLevel(env string) string {
	switch env {
	case "production":
		return "info"
	case "test":
		return "warn"
	default:
		return "debug"
	}
}

func getDefaultLogFormat(env string) string {
	switch env {
	case "production", "test":
		return "json"
	default:
		return "text"
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
