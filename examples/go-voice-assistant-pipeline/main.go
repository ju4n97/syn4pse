package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	relic "github.com/ju4n97/relic/sdk-go"
)

func main() {
	client, err := relic.NewClient("localhost:50051")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	conversation := []relic.Message{
		relic.NewSystemMessage("You are a helpful voice assistant. Keep responses under 20 words and natural."),
	}

	audioInputPath := filepath.Join(dir, "user_input.wav")
	audioInput, err := os.ReadFile(audioInputPath)
	if err != nil {
		panic(err)
	}

	// 1. transcribe user's voice
	transcript, err := client.TranscribeAudio(context.Background(), audioInput,
		relic.WithProvider("whisper.cpp"),
		relic.WithModelID("whisper-cpp-tiny"),
		relic.WithParameter("language", "en"),
	)
	if err != nil {
		log.Fatal(err)
	}
	conversation = append(conversation, relic.NewUserMessage(transcript))
	fmt.Println("You:\n", transcript)

	// 2. generate AI response
	response, err := client.Generate(context.Background(), conversation,
		relic.WithProvider("llama.cpp"),
		relic.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
		relic.WithParameter("max_tokens", 150),
	)
	if err != nil {
		log.Fatal(err)
	}
	conversation = append(conversation, relic.NewAssistantMessage(response))
	fmt.Println("Assistant:\n", response)

	// 3. synthesize AI response
	audioOutput, err := client.SynthesizeSpeech(context.Background(), response,
		relic.WithProvider("piper"),
		relic.WithModelID("piper-en-us-lessac-high"),
	)
	if err != nil {
		log.Fatal(err)
	}

	audioOutputPath := filepath.Join(dir, "assistant_output.wav")
	if err := os.WriteFile(audioOutputPath, audioOutput, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nGenerated %d bytes of audio saved to %s\n", len(audioOutput), audioOutputPath)
}
