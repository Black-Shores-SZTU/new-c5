package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置。
type Config struct {
	Level      string
	Dir        string
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	Env        string
}

// ConfigFromEnv 从环境变量读取日志配置。
func ConfigFromEnv() Config {
	return Config{
		Level:      envString("LOG_LEVEL", "info"),
		Dir:        envString("LOG_DIR", "logs"),
		File:       envString("LOG_FILE", "app.log"),
		MaxSizeMB:  envInt("LOG_MAX_SIZE_MB", 100),
		MaxBackups: envInt("LOG_MAX_BACKUPS", 7),
		MaxAgeDays: envInt("LOG_MAX_AGE_DAYS", 14),
		Compress:   envBool("LOG_COMPRESS", true),
		Env:        envString("APP_ENV", "dev"),
	}
}

// Init 初始化全局 slog：同时输出到 stdout 与按大小轮转的日志文件。
func Init(cfg Config) (*slog.Logger, func() error, error) {
	if cfg.Dir == "" {
		cfg.Dir = "logs"
	}
	if cfg.File == "" {
		cfg.File = "app.log"
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	rotation := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, cfg.File),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	writer := io.MultiWriter(os.Stdout, rotation)
	options := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}

	var handler slog.Handler
	if strings.EqualFold(cfg.Env, "prod") {
		handler = slog.NewJSONHandler(writer, options)
	} else {
		handler = slog.NewTextHandler(writer, options)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger, rotation.Close, nil
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func envString(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envBool(name string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}
