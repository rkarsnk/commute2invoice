// Package config は、アプリケーション全体の設定値を管理します。
package config

import (
	"os"
	"strconv"
)

// Config は、アプリケーション設定を集約する構造体です。
type Config struct {
	ServerPort int
	ServerHost string
	DBPath     string
	LogLevel   string
}

// Load は、環境変数から設定を読み込みます。
func Load() *Config {
	cfg := &Config{
		ServerPort: 8080,
		ServerHost: "0.0.0.0",
		DBPath:     "/data/commute2invoice.db",
		LogLevel:   "info",
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.ServerPort = p
		}
	}

	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.ServerHost = host
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	return cfg
}
