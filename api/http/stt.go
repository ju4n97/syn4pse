package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/ju4n97/relic/internal/backend"
	"github.com/ju4n97/relic/internal/backend/whisper"
	"github.com/ju4n97/relic/internal/model"
	"github.com/ju4n97/relic/internal/service"
)

type (
	// TranscribeRequestDTO is the request body for the Transcribe operation.
	TranscribeRequestDTO struct {
		AudioFile  huma.FormFile `contentType:"audio/*,application/octet-stream" form:"file" required:"true"`
		ModelID    string        `form:"model_id" minLength:"1" required:"true"`
		Parameters string        `form:"parameters"` // JSON-encoded optional parameters
	}

	// TranscribeResponseDTO is the response body for the Transcribe operation.
	TranscribeResponseDTO struct {
		Metadata *backend.ResponseMetadata `json:"metadata,omitempty"`
		Text     string                    `json:"text"`
	}
)

type (
	// TranscribeInput is the huma input for the Transcribe operation.
	TranscribeInput struct {
		RawBody huma.MultipartFormFiles[TranscribeRequestDTO]
	}

	// TranscribeOutput is the huma output for the Transcribe operation.
	TranscribeOutput struct {
		Body TranscribeResponseDTO
	}
)

// STTHandler handles HTTP requests for STT.
type STTHandler struct {
	service *service.STT
}

// NewSTTHandler creates a new STTHandler instance.
func NewSTTHandler(api huma.API, svc *service.STT) *STTHandler {
	h := &STTHandler{service: svc}

	huma.Register(api, huma.Operation{
		OperationID:   "transcribe",
		Method:        "POST",
		Path:          "/stt",
		Summary:       "Transcribe speech from an audio file",
		Tags:          []string{"stt"},
		DefaultStatus: http.StatusOK,
	}, h.handleTranscribe)

	return h
}

// handleTranscribe handles the transcribe operation.
func (h *STTHandler) handleTranscribe(ctx context.Context, input *TranscribeInput) (*TranscribeOutput, error) {
	formData := input.RawBody.Data()
	audioFile := formData.AudioFile

	if !audioFile.IsSet {
		return nil, huma.Error400BadRequest("audio file is required", nil)
	}

	audioBytes, err := io.ReadAll(audioFile)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read audio file", err)
	}

	var parameters map[string]any
	if formData.Parameters != "" {
		if err := json.Unmarshal([]byte(formData.Parameters), &parameters); err != nil {
			return nil, huma.Error400BadRequest("invalid parameters JSON", err)
		}
	}

	provider := whisper.BackendName

	resp, err := h.service.Transcribe(
		ctx,
		provider,
		formData.ModelID,
		&backend.Request{
			Input:      bytes.NewReader(audioBytes),
			Parameters: parameters,
		},
	)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, huma.Error404NotFound("model not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to transcribe", err)
	}

	transcribedBytes, err := io.ReadAll(resp.Output)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read model output", err)
	}

	return &TranscribeOutput{
		Body: TranscribeResponseDTO{
			Text:     string(transcribedBytes),
			Metadata: resp.Metadata,
		},
	}, nil
}
