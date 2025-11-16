package syn4pse

import "maps"

// Config represents the configuration for inference operations.
type Config struct {
	Provider   string
	ModelID    string
	Parameters map[string]any
}

// Option is a function that configures inference operations.
type Option func(*Config)

// WithProvider sets the provider for inference operations.
func WithProvider(provider string) Option {
	return func(c *Config) {
		c.Provider = provider
	}
}

// WithModelID sets the model ID for inference operations.
func WithModelID(modelID string) Option {
	return func(c *Config) {
		c.ModelID = modelID
	}
}

// WithParameters merges the provided parameters with existing ones.
func WithParameters(parameters map[string]any) Option {
	return func(c *Config) {
		if c.Parameters == nil {
			c.Parameters = map[string]any{}
		}

		maps.Copy(c.Parameters, parameters)
	}
}

// WithParameter sets a single parameter.
func WithParameter(key string, value any) Option {
	return func(c *Config) {
		if c.Parameters == nil {
			c.Parameters = map[string]any{}
		}

		c.Parameters[key] = value
	}
}
