package backend

import (
	"log/slog"
	"sync"
)

// Registry manages backend instances.
type Registry struct {
	backends map[string]Backend
	mu       sync.RWMutex
}

// NewRegistry creates a new backend registry.
func NewRegistry() *Registry {
	return &Registry{
		backends: map[string]Backend{},
	}
}

// Register adds a backend to the registry.
func (r *Registry) Register(b Backend) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.backends[b.Provider()]; ok {
		return ErrAlreadyRegistered
	}

	r.backends[b.Provider()] = b

	slog.Info("Backend registered", "provider", b.Provider())

	return nil
}

// Get retrieves a backend by provider.
func (r *Registry) Get(provider string) (Backend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.backends[provider]
	return b, ok
}

// Close closes all registered backends.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, b := range r.backends {
		if err := b.Close(); err != nil {
			slog.Error("Failed to close backend", "provider", b.Provider(), "error", err)
			return err
		}

		slog.Info("Backend closed", "provider", b.Provider())
	}

	return nil
}
