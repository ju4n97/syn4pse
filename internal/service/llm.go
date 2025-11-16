package service

import (
	"context"
	"log/slog"

	"github.com/ju4n97/relic/internal/backend"
	"github.com/ju4n97/relic/internal/model"
)

// LLM is a service abstraction for large language models.
type LLM struct {
	backends *backend.Registry
	models   *model.Registry
}

// NewLLM creates a new LLM service.
func NewLLM(backends *backend.Registry, models *model.Registry) *LLM {
	return &LLM{
		backends: backends,
		models:   models,
	}
}

// Generate generates text using a large language model.
func (s *LLM) Generate(ctx context.Context, provider, modelID string, req *backend.Request) (*backend.Response, error) {
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
		slog.Error("Failed to generate text", "error", err)
		return nil, err
	}

	return resp, nil
}

// GenerateStream generates streamed text using a large language model.
func (s *LLM) GenerateStream(ctx context.Context, provider, modelID string, req *backend.Request) (<-chan backend.StreamChunk, error) {
	b, ok := s.backends.Get(provider)
	if !ok {
		return nil, backend.ErrNotFound
	}

	bs, ok := b.(backend.StreamingBackend)
	if !ok {
		return nil, backend.ErrNotStreamable
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

	resp, err := bs.InferStream(ctx, breq)
	if err != nil {
		slog.Error("Failed to generate streamed text", "error", err)
		return nil, err
	}

	return resp, nil
}
