package service

import (
	"context"
	"log/slog"

	"github.com/ju4n97/relic/internal/backend"
	"github.com/ju4n97/relic/internal/model"
)

// TTS is a service abstraction for text-to-speech.
type TTS struct {
	backends *backend.Registry
	models   *model.Registry
}

// NewTTS creates a new TTS service.
func NewTTS(backends *backend.Registry, models *model.Registry) *TTS {
	return &TTS{
		backends: backends,
		models:   models,
	}
}

// Synthesize synthesizes speech using a text-to-speech model.
func (s *TTS) Synthesize(ctx context.Context, provider, modelID string, req *backend.Request) (*backend.Response, error) {
	b, ok := s.backends.Get(provider)
	if !ok {
		return nil, backend.ErrNotFound
	}

	m, ok := s.models.Get(modelID)
	if !ok {
		return nil, model.ErrNotFound
	}

	breq := &backend.Request{
		ModelPath:  m.Path,
		Input:      req.Input,
		Parameters: req.Parameters,
	}

	resp, err := b.Infer(ctx, breq)
	if err != nil {
		slog.Error("Failed to synthesize speech", "error", err)
		return nil, err
	}

	return resp, nil
}
