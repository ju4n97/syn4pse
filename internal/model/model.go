package model

import (
	"time"

	"github.com/ju4n97/relic/internal/config"
)

// Type is the type of a model.
type Type string

const (
	// TypeLLM is the type of a large language model.
	TypeLLM Type = "llm"

	// TypeNLU is the type of a natural language understanding model.
	TypeNLU Type = "nlu"

	// TypeSTT is the type of a speech-to-text model.
	TypeSTT Type = "stt"

	// TypeTTS is the type of a text-to-speech model.
	TypeTTS Type = "tts"

	// TypeEmbedding is the type of an embedding model.
	TypeEmbedding Type = "embedding"

	// TypeVision is the type of a vision model.
	TypeVision Type = "vision"
)

// Status is the current loading status of a model.
type Status string

const (
	// StatusUnloaded indicates that the model is not loaded.
	StatusUnloaded Status = "unloaded"

	// StatusLoading indicates that the model is being loaded.
	StatusLoading Status = "loading"

	// StatusLoaded indicates that the model is loaded.
	StatusLoaded Status = "loaded"

	// StatusFailed indicates that the model failed to load.
	StatusFailed Status = "failed"

	// StatusUnloading indicates that the model is being unloaded.
	StatusUnloading Status = "unloading"
)

// Instance represents a loaded model instance.
type Instance struct {
	Config   *config.ModelConfig `json:"config"`
	LoadedAt *time.Time          `json:"loaded_at,omitempty"`
	ID       string              `json:"id"`
	Path     string              `json:"-"`
	Status   Status              `json:"status"`
	Error    string              `json:"error,omitempty"`
}

// NewModelInstance creates a new model instance.
func NewModelInstance(cfg *config.ModelConfig, id, path string) *Instance {
	return &Instance{
		ID:     id,
		Path:   path,
		Config: cfg,
		Status: StatusUnloaded,
	}
}

// SetStatus sets the status of the model instance.
func (mi *Instance) SetStatus(status Status) {
	mi.Status = status
	if status == StatusLoaded {
		now := time.Now()
		mi.LoadedAt = &now
	}
}

// SetError sets the error associated with the model instance.
func (mi *Instance) SetError(err error) {
	mi.Error = err.Error()
}
