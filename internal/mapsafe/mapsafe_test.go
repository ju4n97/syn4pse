package mapsafe_test

import (
	"testing"

	"github.com/ju4n97/relic/internal/mapsafe"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	tests := []struct {
		defaultValue any
		expected     any
		m            map[string]any
		name         string
		key          string
	}{
		{
			name:         "existing int",
			m:            map[string]any{"count": 42},
			key:          "count",
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "float stored as int",
			m:            map[string]any{"height": 180},
			key:          "height",
			defaultValue: 0.0,
			expected:     180.0,
		},
		{
			name:         "int stored as float64",
			m:            map[string]any{"age": 25.0},
			key:          "age",
			defaultValue: 0,
			expected:     25,
		},
		{
			name:         "existing string",
			m:            map[string]any{"name": "juan"},
			key:          "name",
			defaultValue: "unknown",
			expected:     "juan",
		},
		{
			name:         "missing string key",
			m:            map[string]any{},
			key:          "name",
			defaultValue: "unknown",
			expected:     "unknown",
		},
		{
			name:         "existing bool",
			m:            map[string]any{"active": true},
			key:          "active",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "type mismatch returns default",
			m:            map[string]any{"enabled": "true"},
			key:          "enabled",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "float value when expecting int returns coerced int",
			m:            map[string]any{"ratio": 99.9},
			key:          "ratio",
			defaultValue: 0,
			expected:     99, // truncates float
		},
		{
			name:         "untyped nil map",
			m:            nil,
			key:          "anything",
			defaultValue: 5,
			expected:     5,
		},
		{
			name:         "unsupported type uses default",
			m:            map[string]any{"custom": []int{1, 2, 3}},
			key:          "custom",
			defaultValue: 0,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch def := tt.defaultValue.(type) {
			case int:
				assert.Equal(t, tt.expected, mapsafe.Get(tt.m, tt.key, def))
			case float64:
				assert.Equal(t, tt.expected, mapsafe.Get(tt.m, tt.key, def))
			case string:
				assert.Equal(t, tt.expected, mapsafe.Get(tt.m, tt.key, def))
			case bool:
				assert.Equal(t, tt.expected, mapsafe.Get(tt.m, tt.key, def))
			default:
				t.Fatalf("unsupported default type %T in test %q", def, tt.name)
			}
		})
	}
}
