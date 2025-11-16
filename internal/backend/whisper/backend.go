package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/ju4n97/syn4pse/internal/backend"
	"github.com/ju4n97/syn4pse/internal/mapsafe"
)

const (
	// BackendName is the name of the backend.
	BackendName = "whisper.cpp"

	// BackendPort is the default port for the backend server.
	BackendPort = 8082
)

// Backend implements backend.Backend for whisper.cpp.
type Backend struct {
	serverManager *backend.ServerManager
	client        *http.Client
	binPath       string
	port          int
}

// TranscriptionRequest represents a request to the whisper-server API.
type TranscriptionRequest struct {
	Language     string  `json:"language,omitempty"`
	Prompt       string  `json:"prompt,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	BeamSize     int     `json:"beam_size,omitempty"`
	BestOf       int     `json:"best_of,omitempty"`
	Translate    bool    `json:"translate,omitempty"`
	NoTimestamps bool    `json:"no_timestamps,omitempty"`
}

// TranscriptionResponse represents a response from the whisper-server API.
type TranscriptionResponse struct {
	LanguageProbabilities       map[string]float64  `json:"language_probabilities,omitempty"`
	Task                        string              `json:"task,omitempty"`
	Language                    string              `json:"language,omitempty"`
	Text                        string              `json:"text,omitempty"`
	DetectedLanguage            string              `json:"detected_language,omitempty"`
	Segments                    []TranscriptSegment `json:"segments,omitempty"`
	Duration                    float64             `json:"durationm,omitempty"`
	DetectedLanguageProbability float64             `json:"detected_language_probability,omitempty"`
}

// TranscriptSegment represents a single segment in the transcription.
type TranscriptSegment struct {
	Text         string                     `json:"text"`
	Tokens       []int                      `json:"tokens,omitempty"`
	Words        []TranscriptionSegmentWord `json:"words,omitempty"`
	ID           int                        `json:"id"`
	Start        float64                    `json:"start"`
	End          float64                    `json:"end"`
	Temperature  float64                    `json:"temperature,omitempty"`
	AvgLogprob   float64                    `json:"avg_logprob,omitempty"`
	NoSpeechProb float64                    `json:"no_speech_prob,omitempty"`
}

// TranscriptionSegmentWord represents a word in the transcription segment.
type TranscriptionSegmentWord struct {
	Word        string  `json:"word"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
	DTW         float64 `json:"t_dtw"`
	Probability float64 `json:"probability"`
}

// NewBackend creates a new Backend instance.
func NewBackend(binPath string, serverManager *backend.ServerManager) (backend.Backend, error) {
	return &Backend{
		binPath:       binPath,
		serverManager: serverManager,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		port: BackendPort,
	}, nil
}

// Close implements backend.Backend.
func (b *Backend) Close() error {
	return b.serverManager.StopServer(BackendName, b.port)
}

// Provider implements backend.Backend.
func (b *Backend) Provider() string {
	return BackendName
}

// Infer implements backend.Backend.
func (b *Backend) Infer(ctx context.Context, req *backend.Request) (*backend.Response, error) {
	args := []string{
		"--model", req.ModelPath,
		"--port", strconv.Itoa(b.port),
		"--host", "127.0.0.1",
	}

	if err := b.serverManager.StartServer(backend.ServerConfig{
		Name:       BackendName,
		BinPath:    b.binPath,
		Args:       args,
		Port:       b.port,
		HealthPath: "/",
	}); err != nil {
		return nil, fmt.Errorf("manager: failed to start server: %w", err)
	}

	audioData, err := io.ReadAll(req.Input)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to read audio input: %w", err)
	}

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add audio file
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, fmt.Errorf("manager: failed to create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return nil, fmt.Errorf("manager: failed to write audio data: %w", err)
	}

	// Add parameters to form
	transcriptionReq := b.buildTranscriptionRequest(req)
	if err := b.addTranscriptionParams(writer, transcriptionReq); err != nil {
		return nil, fmt.Errorf("manager: failed to add parameters: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("manager: failed to close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		fmt.Sprintf("http://localhost:%d/inference", b.port),
		&requestBody,
	)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	start := time.Now()

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Seconds()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("manager: failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("manager: request failed with status code %d: %s", resp.StatusCode, body)
	}

	var transcriptionResp TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResp); err != nil {
		return nil, fmt.Errorf("manager: failed to decode response: %w", err)
	}

	return &backend.Response{
		Output: bytes.NewReader([]byte(transcriptionResp.Text)),
		Metadata: &backend.ResponseMetadata{
			Provider:        b.Provider(),
			Model:           req.ModelPath,
			Timestamp:       time.Now(),
			DurationSeconds: elapsed,
			OutputSizeBytes: int64(len(transcriptionResp.Text)),
			BackendSpecific: map[string]any{
				"response": transcriptionResp,
			},
		},
	}, nil
}

// buildTranscriptionRequest builds a TranscriptionRequest from a backend.Request.
func (b *Backend) buildTranscriptionRequest(req *backend.Request) *TranscriptionRequest {
	p := req.Parameters
	if p == nil {
		p = map[string]any{}
	}

	return &TranscriptionRequest{
		Language:     mapsafe.Get(p, "language", ""),
		Temperature:  mapsafe.Get(p, "temperature", 0.0),
		Translate:    mapsafe.Get(p, "translate", false),
		NoTimestamps: mapsafe.Get(p, "no_timestamps", false),
		Prompt:       mapsafe.Get(p, "prompt", ""),
		BeamSize:     mapsafe.Get(p, "beam_size", -1),
		BestOf:       mapsafe.Get(p, "best_of", 2),
	}
}

// addTranscriptionParams adds transcription parameters to the multipart writer.
func (b *Backend) addTranscriptionParams(w *multipart.Writer, req *TranscriptionRequest) error {
	params := map[string]string{
		"language":        req.Language,
		"response_format": "verbose_json",
		"temperature":     fmt.Sprintf("%.2f", req.Temperature),
		"translate":       strconv.FormatBool(req.Translate),
		"no_timestamps":   strconv.FormatBool(req.NoTimestamps),
	}

	if req.BeamSize >= 0 {
		params["beam_size"] = strconv.Itoa(req.BeamSize)
	}

	if req.BestOf > 0 {
		params["best_of"] = strconv.Itoa(req.BestOf)
	}

	if req.Prompt != "" {
		params["prompt"] = req.Prompt
	}

	for key, value := range params {
		if err := w.WriteField(key, value); err != nil {
			return fmt.Errorf("manager: failed to write field %s: %w", key, err)
		}
	}

	return nil
}
