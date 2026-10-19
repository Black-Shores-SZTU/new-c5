package logging

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWritesFile(t *testing.T) {
	previous := slog.Default()
	defer slog.SetDefault(previous)

	dir := t.TempDir()
	logger, closeLog, err := Init(Config{
		Level:      "debug",
		Dir:        dir,
		File:       "app.log",
		MaxSizeMB:  1,
		MaxBackups: 1,
		MaxAgeDays: 1,
		Env:        "dev",
	})
	if err != nil {
		t.Fatalf("init logger: %v", err)
	}
	logger.Debug("debug-message", "request_id", "abc")
	logger.Info("info-message")
	if err := closeLog(); err != nil {
		t.Fatalf("close logger: %v", err)
	}

	payload, err := os.ReadFile(filepath.Join(dir, "app.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content := string(payload)
	if !strings.Contains(content, "debug-message") || !strings.Contains(content, "info-message") {
		t.Fatalf("log file missing messages: %s", content)
	}
	if !strings.Contains(content, "level=DEBUG") {
		t.Fatalf("debug level not written: %s", content)
	}
}

func TestInitProdJSON(t *testing.T) {
	previous := slog.Default()
	defer slog.SetDefault(previous)

	dir := t.TempDir()
	logger, closeLog, err := Init(Config{
		Level:      "info",
		Dir:        dir,
		File:       "prod.log",
		MaxSizeMB:  1,
		MaxBackups: 1,
		MaxAgeDays: 1,
		Env:        "prod",
	})
	if err != nil {
		t.Fatalf("init logger: %v", err)
	}
	logger.Warn("prod-message", "request_id", "xyz")
	if err := closeLog(); err != nil {
		t.Fatalf("close logger: %v", err)
	}

	payload, err := os.ReadFile(filepath.Join(dir, "prod.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(lines) != 1 {
		t.Fatalf("unexpected log lines: %d", len(lines))
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("prod log is not JSON: %v (%s)", err, lines[0])
	}
	if entry["msg"] != "prod-message" || entry["level"] != "WARN" || entry["request_id"] != "xyz" {
		t.Fatalf("unexpected prod entry: %+v", entry)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_DIR", "custom-logs")
	t.Setenv("LOG_FILE", "custom.log")
	t.Setenv("LOG_MAX_SIZE_MB", "32")
	t.Setenv("LOG_MAX_BACKUPS", "3")
	t.Setenv("LOG_MAX_AGE_DAYS", "5")
	t.Setenv("LOG_COMPRESS", "false")
	t.Setenv("APP_ENV", "prod")

	cfg := ConfigFromEnv()
	if cfg.Level != "warn" || cfg.Dir != "custom-logs" || cfg.File != "custom.log" {
		t.Fatalf("unexpected string config: %+v", cfg)
	}
	if cfg.MaxSizeMB != 32 || cfg.MaxBackups != 3 || cfg.MaxAgeDays != 5 {
		t.Fatalf("unexpected numeric config: %+v", cfg)
	}
	if cfg.Compress || cfg.Env != "prod" {
		t.Fatalf("unexpected bool/env config: %+v", cfg)
	}
}

func TestParseLevelFallback(t *testing.T) {
	if got := parseLevel("unknown"); got != slog.LevelInfo {
		t.Fatalf("unknown level should fall back to info, got %v", got)
	}
	if got := parseLevel("ERROR"); got != slog.LevelError {
		t.Fatalf("error level parse failed: %v", got)
	}
}
