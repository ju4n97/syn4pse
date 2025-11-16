package env_test

import (
	"testing"

	"github.com/ju4n97/relic/internal/env"
	"github.com/ju4n97/relic/internal/envvar"
	"github.com/stretchr/testify/assert"
)

func TestEnv_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
		env  env.Env
	}{
		{
			name: "returns development for EnvDevelopment",
			env:  env.EnvDevelopment,
			want: "development",
		},
		{
			name: "returns production for EnvProduction",
			env:  env.EnvProduction,
			want: "production",
		},
		{
			name: "returns unknown for unrecognized env",
			env:  env.Env(999),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.env.String())
		})
	}
}

func TestFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  env.Env
	}{
		{
			name:  "returns EnvDevelopment for dev",
			input: "dev",
			want:  env.EnvDevelopment,
		},
		{
			name:  "returns EnvDevelopment for development",
			input: "development",
			want:  env.EnvDevelopment,
		},
		{
			name:  "returns EnvProduction for prod",
			input: "prod",
			want:  env.EnvProduction,
		},
		{
			name:  "returns EnvProduction for production",
			input: "production",
			want:  env.EnvProduction,
		},
		{
			name:  "is case insensitive",
			input: "DEV",
			want:  env.EnvDevelopment,
		},
		{
			name:  "trims whitespace",
			input: " dev ",
			want:  env.EnvDevelopment,
		},
		{
			name:  "falls back to development for unknown values",
			input: "unknown",
			want:  env.EnvDevelopment,
		},
		{
			name:  "falls back to development for empty string",
			input: "",
			want:  env.EnvDevelopment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, env.FromString(tt.input))
		})
	}
}

func TestFromEnv(t *testing.T) {
	t.Run("returns development when RELIC_ENV is not set", func(t *testing.T) {
		t.Setenv(envvar.RelicEnv, "")
		assert.Equal(t, env.EnvDevelopment, env.FromEnv())
	})

	t.Run("returns correct environment when RELIC_ENV is set", func(t *testing.T) {
		t.Setenv(envvar.RelicEnv, "production")
		assert.Equal(t, env.EnvProduction, env.FromEnv())
	})

	t.Run("falls back to development for invalid RELIC_ENV values", func(t *testing.T) {
		t.Setenv(envvar.RelicEnv, "invalid")
		assert.Equal(t, env.EnvDevelopment, env.FromEnv())
	})
}
