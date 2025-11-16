package backend

import (
	"context"
	"io"
	"time"
)

// Backend defines the core interface for all inference backends.
type Backend interface {
	// Name returns the backend identifier.
	Provider() string

	// Infer executes inference and returns complete result.
	Infer(ctx context.Context, req *Request) (*Response, error)

	// Close cleans up resources.
	Close() error
}

// StreamingBackend is an optional interface for backends that support streaming.
type StreamingBackend interface {
	Backend

	// InferStream executes inference and streams results as they're produced.
	InferStream(ctx context.Context, req *Request) (<-chan StreamChunk, error)
}

// Request encapsulates all parameters for an inference call.
type Request struct {
	Input      io.Reader
	Parameters map[string]any
	ModelPath  string
}

// Response contains the result of an inference operation.
type Response struct {
	// Output is the raw output data.
	Output io.Reader

	// Metadata contains backend-specific information.
	Metadata *ResponseMetadata
}

// ResponseMetadata contains metadata about the response.
type ResponseMetadata struct {
	Timestamp       time.Time      `json:"timestamp"`
	BackendSpecific map[string]any `json:"backend_specific,omitempty"`
	Provider        string         `json:"provider"`
	Model           string         `json:"model"`
	DurationSeconds float64        `json:"inference_time_seconds"`
	OutputSizeBytes int64          `json:"output_size_bytes"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	Error error  `json:"error,omitempty"`
	Data  []byte `json:"data,omitempty"`
	Done  bool   `json:"done,omitempty"`
}
