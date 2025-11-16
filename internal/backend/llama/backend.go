package llama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ju4n97/syn4pse/internal/backend"
	"github.com/ju4n97/syn4pse/internal/mapsafe"
)

const (
	// BackendName is the name of the backend.
	BackendName = "llama.cpp"

	// BackendPort is the default port for the backend server.
	BackendPort = 8081
)

// Backend implements backend.Backend for llama.cpp.
type Backend struct {
	serverManager *backend.ServerManager
	client        *http.Client
	binPath       string
	port          int
}

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is a request to the llama-server API.
type ChatCompletionRequest struct {
	Messages         []ChatMessage `json:"messages"`
	Stream           bool          `json:"stream,omitempty"`
	Temperature      float64       `json:"temperature,omitempty"`
	TopK             int           `json:"top_k,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	MinP             float64       `json:"min_p,omitempty"`
	NPredict         int           `json:"n_predict,omitempty"`
	RepeatPenalty    float64       `json:"repeat_penalty,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
}

// ChatCompletionResponse is a response from the llama-server API.
type ChatCompletionResponse struct {
	Timings           map[string]any `json:"timings,omitempty"`
	ID                string         `json:"id,omitempty"`
	Object            string         `json:"object,omitempty"`
	Model             string         `json:"model,omitempty"`
	SystemFingerprint string         `json:"system_fingerprint,omitempty"`
	Choices           []Choice       `json:"choices"`
	Usage             Usage          `json:"usage"`
	Created           int64          `json:"created,omitempty"`
}

// Choice represents a single choice in a response.
type Choice struct {
	Message      Message     `json:"message"`
	Delta        ChoiceDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
	Index        int         `json:"index"`
}

// ChoiceDelta represents the delta of a choice in a response.
type ChoiceDelta struct {
	Content string `json:"content"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage represents the usage information of a response.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewBackend creates a new Backend instance.
func NewBackend(binPath string, serverManager *backend.ServerManager) (backend.StreamingBackend, error) {
	return &Backend{
		binPath:       binPath,
		serverManager: serverManager,
		client: &http.Client{
			Timeout: 2 * time.Minute,
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
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(b.port),
	}

	if err := b.serverManager.StartServer(backend.ServerConfig{
		Name:       BackendName,
		BinPath:    b.binPath,
		Args:       args,
		Port:       b.port,
		HealthPath: "/health",
	}); err != nil {
		return nil, fmt.Errorf("manager: failed to start server: %w", err)
	}

	prompt, err := io.ReadAll(req.Input)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to read input: %w", err)
	}

	const shouldStream = false
	completionReq := b.buildChatCompletionRequest(req, string(prompt), shouldStream)

	jsonData, err := json.Marshal(completionReq)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		fmt.Sprintf("http://localhost:%d/chat/completions", b.port),
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

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

	var completionResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return nil, fmt.Errorf("manager: failed to decode response: %w", err)
	}

	content := ""
	if len(completionResp.Choices) > 0 {
		content = completionResp.Choices[0].Message.Content
	}

	return &backend.Response{
		Output: bytes.NewReader([]byte(content)),
		Metadata: &backend.ResponseMetadata{
			Provider:        b.Provider(),
			Model:           req.ModelPath,
			Timestamp:       time.Now(),
			DurationSeconds: elapsed,
			OutputSizeBytes: int64(len(content)),
			BackendSpecific: map[string]any{
				"response": completionResp,
			},
		},
	}, nil
}

// InferStream implements backend.StreamingBackend.
func (b *Backend) InferStream(ctx context.Context, req *backend.Request) (<-chan backend.StreamChunk, error) {
	args := []string{
		"--model", req.ModelPath,
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(b.port),
	}

	if err := b.serverManager.StartServer(backend.ServerConfig{
		Name:       BackendName,
		BinPath:    b.binPath,
		Args:       args,
		Port:       b.port,
		HealthPath: "/health",
	}); err != nil {
		return nil, fmt.Errorf("manager: failed to start server: %w", err)
	}

	prompt, err := io.ReadAll(req.Input)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to read input: %w", err)
	}

	const shouldStream = true
	completionReq := b.buildChatCompletionRequest(req, string(prompt), shouldStream)

	jsonData, err := json.Marshal(completionReq)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		fmt.Sprintf("http://localhost:%d/chat/completions", b.port),
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("manager: failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		return nil, fmt.Errorf("manager: request failed with status code %d: %s", resp.StatusCode, body)
	}

	chunks := make(chan backend.StreamChunk)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			select {
			case <-ctx.Done():
				chunks <- backend.StreamChunk{Error: ctx.Err(), Done: true}
				return
			default:
			}

			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					chunks <- backend.StreamChunk{Error: err, Done: true}
				} else {
					chunks <- backend.StreamChunk{Done: true}
				}
				return
			}

			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))

			if bytes.Equal(data, []byte("[DONE]")) {
				chunks <- backend.StreamChunk{Done: true}
				return
			}

			var completionResp ChatCompletionResponse
			if err := json.Unmarshal(data, &completionResp); err != nil {
				continue
			}

			if len(completionResp.Choices) > 0 {
				content := completionResp.Choices[0].Delta.Content
				if content != "" {
					chunks <- backend.StreamChunk{
						Data: []byte(content),
						Done: false,
					}
				}

				if completionResp.Choices[0].FinishReason != nil {
					chunks <- backend.StreamChunk{Done: true}
					return
				}
			}
		}
	}()

	return chunks, nil
}

// buildChatCompletionRequest builds a ChatCompletionRequest from a backend.Request.
func (b *Backend) buildChatCompletionRequest(req *backend.Request, prompt string, stream bool) *ChatCompletionRequest {
	p := req.Parameters
	if p == nil {
		p = map[string]any{}
	}

	messages := []ChatMessage{}

	if messagesJSON, ok := p["messages"].(string); ok && messagesJSON != "" {
		var chatMsgs []map[string]string

		if err := json.Unmarshal([]byte(messagesJSON), &chatMsgs); err == nil {
			for _, msg := range chatMsgs {
				messages = append(messages, ChatMessage{
					Role:    msg["role"],
					Content: msg["content"],
				})
			}
		}
	} else {
		messages = []ChatMessage{
			{Role: "user", Content: prompt},
		}

		if sysPrompt, ok := p["system_prompt"].(string); ok && sysPrompt != "" {
			messages = append([]ChatMessage{{Role: "system", Content: sysPrompt}}, messages...)
		}
	}

	return &ChatCompletionRequest{
		Messages:         messages,
		Stream:           stream,
		NPredict:         mapsafe.Get(p, "n_predict", 128),
		Temperature:      mapsafe.Get(p, "temperature", 0.7),
		TopK:             mapsafe.Get(p, "top_k", 40),
		TopP:             mapsafe.Get(p, "top_p", 0.9),
		MinP:             mapsafe.Get(p, "min_p", 0.05),
		RepeatPenalty:    mapsafe.Get(p, "repeat_penalty", 1.1),
		PresencePenalty:  mapsafe.Get(p, "presence_penalty", 0.0),
		FrequencyPenalty: mapsafe.Get(p, "frequency_penalty", 0.0),
	}
}
