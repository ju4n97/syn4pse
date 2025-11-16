package env

import (
	"os"
	"strings"

	"github.com/ju4n97/relic/internal/envvar"
)

// Env represents the environment a service is running in.
type Env int

const (
	// EnvDevelopment is the development environment.
	EnvDevelopment Env = iota

	// EnvProduction is the production environment.
	EnvProduction
)

// String returns the string name of the environment.
func (e Env) String() string {
	switch e {
	case EnvDevelopment:
		return "development"
	case EnvProduction:
		return "production"
	default:
		return "unknown"
	}
}

// FromString converts a string into an Env value.
// Falls back to Development if not recognized.
func FromString(s string) Env {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dev", "development":
		return EnvDevelopment
	case "prod", "production":
		return EnvProduction
	default:
		return EnvDevelopment
	}
}

// FromEnv returns the environment based on the RELIC_ENV environment variable.
func FromEnv() Env {
	val := os.Getenv(envvar.RelicEnv)
	if val == "" {
		return EnvDevelopment
	}

	return FromString(val)
}
