package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/ju4n97/relic/internal/backend"
	"github.com/ju4n97/relic/internal/backend/piper"
	"github.com/ju4n97/relic/internal/model"
	"github.com/ju4n97/relic/internal/service"
)

type (
	// SynthesizeRequestDTO is the request body for the Synthesize operation.
	SynthesizeRequestDTO struct {
		Parameters map[string]any `json:"parameters,omitempty"`
		ModelID    string         `json:"model_id" minLength:"1"`
		Text       string         `json:"text" maxLength:"4096" minLength:"1"`
	}
)

type (
	// SynthesizeInput is the input for the Synthesize operation.
	SynthesizeInput struct {
		Body SynthesizeRequestDTO
	}
)

// TTSHandler handles HTTP requests for TTS.
type TTSHandler struct {
	service *service.TTS
}

// NewTTSHandler creates a new TTSHandler instance.
func NewTTSHandler(api huma.API, svc *service.TTS) *TTSHandler {
	h := &TTSHandler{service: svc}

	huma.Register(api, huma.Operation{
		OperationID:   "synthesize",
		Method:        "POST",
		Path:          "/tts",
		Summary:       "Synthesize speech from a text",
		Tags:          []string{"tts"},
		DefaultStatus: http.StatusOK,
	}, h.handleSynthesize)

	return h
}

// handleSynthesize handles the synthesize operation.
func (h *TTSHandler) handleSynthesize(ctx context.Context, input *SynthesizeInput) (*huma.StreamResponse, error) {
	provider := piper.BackendName

	resp, err := h.service.Synthesize(
		ctx,
		provider,
		input.Body.ModelID,
		&backend.Request{
			Input:      strings.NewReader(input.Body.Text),
			Parameters: input.Body.Parameters,
		},
	)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, huma.Error404NotFound("model not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to synthesize", err)
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.SetHeader("Content-Type", "audio/wav")

			writer := ctx.BodyWriter()
			if _, err := io.Copy(writer, resp.Output); err != nil {
				slog.Error("Failed to copy response body", "error", err)
			}

			if f, ok := writer.(http.Flusher); ok {
				f.Flush()
			}
		},
	}, nil
}
