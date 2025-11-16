package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/ju4n97/relic/internal/env"
	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is a wrapper around slog.Logger.
type Logger = *slog.Logger

// Config holds configuration for the logger.
type Config struct {
	LogFile     string
	MaxSizeMB   int
	MaxBackups  int
	MaxAge      int
	LogToFile   bool
	LogToStdout bool
	Compress    bool
}

// DefaultConfig returns the default configuration for the logger.
func DefaultConfig() Config {
	return Config{
		LogFile:     "logs/relic.log",
		LogToFile:   false,
		LogToStdout: true,
		MaxSizeMB:   100,  // 100 MB
		MaxBackups:  3,    // Keep 3 old log files
		MaxAge:      28,   // Keep files for 28 days
		Compress:    true, // Compress old log files
	}
}

// Option is a function that configures the logger.
type Option func(*Config)

// WithLogFile sets the log file for the logger.
func WithLogFile(logFile string) Option {
	return func(c *Config) {
		c.LogFile = logFile
	}
}

// WithLogToFile sets whether to log to a file.
func WithLogToFile(logToFile bool) Option {
	return func(c *Config) {
		c.LogToFile = logToFile
	}
}

// WithLogToStdout sets whether to log to stdout.
func WithLogToStdout(logToStdout bool) Option {
	return func(c *Config) {
		c.LogToStdout = logToStdout
	}
}

// WithMaxSizeMB sets the maximum size in megabytes before rotation.
func WithMaxSizeMB(maxSizeMB int) Option {
	return func(c *Config) {
		c.MaxSizeMB = maxSizeMB
	}
}

// WithMaxBackups sets the maximum number of old log files to retain.
func WithMaxBackups(maxBackups int) Option {
	return func(c *Config) {
		c.MaxBackups = maxBackups
	}
}

// WithMaxAge sets the maximum number of days to retain old log files.
func WithMaxAge(maxAge int) Option {
	return func(c *Config) {
		c.MaxAge = maxAge
	}
}

// WithCompress sets whether to compress old log files.
func WithCompress(compress bool) Option {
	return func(c *Config) {
		c.Compress = compress
	}
}

// New returns a new logger based on the environment.
func New(e env.Env, opts ...Option) Logger {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	var writers []io.Writer

	if cfg.LogToStdout {
		writers = append(writers, os.Stdout)
	}

	if cfg.LogToFile {
		rotatingWriter := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, rotatingWriter)
	}

	var output io.Writer
	if len(writers) == 0 {
		output = io.Discard
	} else if len(writers) == 1 {
		output = writers[0]
	} else {
		output = io.MultiWriter(writers...)
	}

	var handler slog.Handler
	switch e {
	case env.EnvDevelopment:
		handler = tint.NewHandler(output, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	case env.EnvProduction:
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}
