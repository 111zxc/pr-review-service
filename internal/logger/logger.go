package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var globalLogger *slog.Logger

type Config struct {
	Level  string
	Format string
}

func Init(cfg Config) error {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	globalLogger = slog.New(handler)
	return nil
}

func InitWithEnv(env string) error {
	var cfg Config

	switch env {
	case "production":
		cfg = Config{Level: "info", Format: "json"}
	case "test":
		cfg = Config{Level: "warn", Format: "json"}
	default:
		cfg = Config{Level: "debug", Format: "text"}
	}

	return Init(cfg)
}

func Get() *slog.Logger {
	if globalLogger == nil {
		err := InitWithEnv("development")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
			globalLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		}
	}
	return globalLogger
}

func Info(msg string, args ...interface{}) {
	Get().Info(msg, args...)
}

func Error(msg string, args ...interface{}) {
	Get().Error(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	Get().Debug(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	Get().Warn(msg, args...)
}

func WithError(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func WithField(key string, value interface{}) slog.Attr {
	return slog.Any(key, value)
}
