package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/ju4n97/relic/internal/envvar"
)

// DefaultHTTPPort returns the default HTTP port.
// Precedence:
// 1. RELIC_SERVER_HTTP_PORT environment variable.
// 2. 8080.
func DefaultHTTPPort() int {
	if p := os.Getenv(envvar.RelicServerHTTPPort); p != "" {
		value, err := strconv.ParseInt(p, 10, 32)
		if err == nil {
			return int(value)
		}
	}

	return 8080
}

// DefaultGRPCPort returns the default gRPC port.
// Precedence:
// 1. RELIC_SERVER_GRPC_PORT environment variable.
// 2. 50051.
func DefaultGRPCPort() int {
	if p := os.Getenv(envvar.RelicServerGRPCPort); p != "" {
		value, err := strconv.ParseInt(p, 10, 32)
		if err == nil {
			return int(value)
		}
	}

	return 50051
}

// DefaultConfigPath returns the default path for RELIC config directory.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "relic", "config")
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "relic")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "relic")
	default: // Linux, BSD, etc.
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "relic")
		}
		return filepath.Join(home, ".config", "relic")
	}
}

// DefaultModelsPath returns the default path for RELIC models directory.
func DefaultModelsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "relic", "models")
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local", "relic", "models")
	case "darwin":
		return filepath.Join(home, "Library", "Caches", "relic", "models")
	default: // Linux, BSD, etc.
		if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
			return filepath.Join(xdg, "relic", "models")
		}
		return filepath.Join(home, ".cache", "relic", "models")
	}
}
