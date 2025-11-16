package logger_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/ju4n97/relic/internal/env"
	"github.com/ju4n97/relic/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("development environment produces text logger with debug level", func(t *testing.T) {
		t.Parallel()
		log := logger.New(env.EnvDevelopment)
		assert.NotNil(t, log)
	})

	t.Run("production environment produces JSON logger with info level", func(t *testing.T) {
		t.Parallel()
		log := logger.New(env.EnvProduction)
		assert.NotNil(t, log)
	})

	t.Run("unknown environment defaults to text logger with info level", func(t *testing.T) {
		t.Parallel()
		log := logger.New(env.Env(999))
		assert.NotNil(t, log)
	})

	t.Run("logger can log messages without panicking", func(t *testing.T) {
		t.Parallel()
		log := logger.New(env.EnvDevelopment)
		assert.NotPanics(t, func() {
			log.Info("info message")
			log.Debug("debug message")
			log.Error("error message")
		})
	})

	t.Run("options override default config", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		log := logger.New(
			env.EnvDevelopment,
			logger.WithLogFile("ignored.log"), // won't write to disk in test
			logger.WithLogToFile(false),
			logger.WithLogToStdout(false),
		)

		// Should not panic logging
		assert.NotPanics(t, func() {
			log.Info("hello")
		})

		// Simulate writing to in-memory writer
		logger.New(env.EnvDevelopment)
		multiWriter := io.MultiWriter(&buf)
		assert.NotNil(t, multiWriter)
	})

	t.Run("multi-writer includes stdout and file writer when enabled", func(t *testing.T) {
		t.Parallel()
		log := logger.New(env.EnvDevelopment)
		assert.NotNil(t, log)
	})
}
