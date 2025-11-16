package relic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"

	inferencev1 "github.com/ju4n97/relic/sdk-go/pb/inference/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client is a client for RELIC inference services.
type Client struct {
	conn            *grpc.ClientConn
	inferenceClient inferencev1.InferenceServiceClient
}

// NewClient creates a new Client instance.
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("relic: failed to create gRPC client: %w", err)
	}

	return &Client{
		conn:            conn,
		inferenceClient: inferencev1.NewInferenceServiceClient(conn),
	}, nil
}

// Close closes the client gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Generate calls the LLM inference service.
//
// Example:
//
//	output, err := client.Generate(ctx, messages, opts...)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(output)
func (c *Client) Generate(ctx context.Context, messages []Message, options ...Option) (string, error) {
	cfg := c.applyOptions(options...)

	req, err := c.buildLLMRequest(messages, cfg)
	if err != nil {
		return "", fmt.Errorf("relic: failed to build generate request: %w", err)
	}

	resp, err := c.inferenceClient.Infer(ctx, req)
	if err != nil {
		return "", fmt.Errorf("relic: failed to generate: %w", err)
	}

	return string(resp.Output), nil
}

// GenerateStream calls the LLM inference service with streaming support.
// The channel is closed when streaming completes or an error occurs.
//
// Example:
//
//	ch := client.GenerateStream(ctx, messages, opts...)
//
//	for chunk := range ch {
//		if chunk.Error != nil {
//			log.Fatal(chunk.Error)
//		}
//
//		fmt.Println(chunk.Content)
//	}
func (c *Client) GenerateStream(ctx context.Context, messages []Message, options ...Option) <-chan StreamChunk {
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)

		cfg := c.applyOptions(options...)

		req, err := c.buildLLMRequest(messages, cfg)
		if err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("relic: failed to build generate request: %w", err)}
			return
		}

		stream, err := c.inferenceClient.InferStream(ctx)
		if err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("relic: failed to create stream: %w", err)}
			return
		}

		if err := stream.Send(req); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("relic: failed to send initial request: %w", err)}
			return
		}

		if err := stream.CloseSend(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("relic: failed to close stream: %w", err)}
			return
		}

		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				ch <- StreamChunk{Error: fmt.Errorf("relic: stream received error: %w", err)}
				return
			}

			select {
			case ch <- StreamChunk{Content: string(chunk.Data)}:
			case <-ctx.Done():
				ch <- StreamChunk{Error: ctx.Err()}
				return
			}
		}
	}()

	return ch
}

// TranscribeAudio calls the STT inference service.
//
// Example:
//
//	output, err := client.TranscribeAudio(ctx, audio, opts...)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(output)
func (c *Client) TranscribeAudio(ctx context.Context, audio []byte, options ...Option) (string, error) {
	cfg := c.applyOptions(options...)

	parameters, err := c.buildParameters(cfg.Parameters)
	if err != nil {
		return "", fmt.Errorf("relic: failed to build parameters: %w", err)
	}

	req := &inferencev1.InferenceRequest{
		Provider:   cfg.Provider,
		ModelId:    cfg.ModelID,
		Parameters: parameters,
		Input:      audio,
	}

	resp, err := c.inferenceClient.Infer(ctx, req)
	if err != nil {
		return "", fmt.Errorf("relic: failed to transcribe audio: %w", err)
	}

	return string(resp.Output), nil
}

// SynthesizeSpeech converts text to speech using the TTS inference service.
//
// Example:
//
//	audio, err := client.SynthesizeSpeech(ctx, "Hello, world!", opts...)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(audio)
func (c *Client) SynthesizeSpeech(ctx context.Context, text string, options ...Option) ([]byte, error) {
	cfg := c.applyOptions(options...)

	parameters, err := c.buildParameters(cfg.Parameters)
	if err != nil {
		return nil, fmt.Errorf("relic: failed to build parameters: %w", err)
	}

	req := &inferencev1.InferenceRequest{
		Provider:   cfg.Provider,
		ModelId:    cfg.ModelID,
		Parameters: parameters,
		Input:      []byte(text),
	}

	resp, err := c.inferenceClient.Infer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("relic: failed to synthesize speech: %w", err)
	}

	return resp.Output, nil
}

// applyOptions applies all options and returns a configured Config.
func (c *Client) applyOptions(options ...Option) *Config {
	cfg := &Config{
		Parameters: map[string]any{},
	}

	for _, option := range options {
		option(cfg)
	}

	return cfg
}

// buildLLMRequest constructs an InferenceRequest for LLM operations.
func (c *Client) buildLLMRequest(messages []Message, cfg *Config) (*inferencev1.InferenceRequest, error) {
	if len(messages) == 0 {
		return nil, errors.New("relic: messages cannot be empty")
	}

	chatMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		chatMessages[i] = map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		}
	}

	messagesJSON, err := json.Marshal(chatMessages)
	if err != nil {
		return nil, fmt.Errorf("relic: failed to marshal messages: %w", err)
	}

	parametersMap := make(map[string]any, len(cfg.Parameters)+1)
	parametersMap["messages"] = string(messagesJSON)
	maps.Copy(parametersMap, cfg.Parameters)

	parameters, err := c.buildParameters(parametersMap)
	if err != nil {
		return nil, fmt.Errorf("relic: failed to build parameters: %w", err)
	}

	return &inferencev1.InferenceRequest{
		Provider:   cfg.Provider,
		ModelId:    cfg.ModelID,
		Input:      []byte(""),
		Parameters: parameters,
	}, nil
}

// buildParameters converts Go native types to protobuf Value map.
func (c *Client) buildParameters(params map[string]any) (*structpb.Struct, error) {
	if params == nil {
		return &structpb.Struct{Fields: map[string]*structpb.Value{}}, nil
	}

	fields := make(map[string]*structpb.Value, len(params))
	for k, v := range params {
		val, err := structpb.NewValue(v)
		if err != nil {
			return nil, fmt.Errorf("relic: failed to set parameter %s: %w", k, err)
		}

		fields[k] = val
	}

	return &structpb.Struct{Fields: fields}, nil
}
