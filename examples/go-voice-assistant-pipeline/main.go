package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	syn4pse "github.com/ju4n97/syn4pse/sdk-go"
)

func main() {
	client, err := syn4pse.NewClient("localhost:50051")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	conversation := []syn4pse.Message{
		syn4pse.NewSystemMessage("You are a helpful voice assistant. Keep responses under 20 words and natural."),
	}

	audioInputPath := filepath.Join(dir, "user_input.wav")
	audioInput, err := os.ReadFile(audioInputPath)
	if err != nil {
		panic(err)
	}

	// 1. transcribe user's voice
	transcript, err := client.TranscribeAudio(context.Background(), audioInput,
		syn4pse.WithProvider("whisper.cpp"),
		syn4pse.WithModelID("whisper-cpp-tiny"),
		syn4pse.WithParameter("language", "en"),
	)
	if err != nil {
		log.Fatal(err)
	}
	conversation = append(conversation, syn4pse.NewUserMessage(transcript))
	fmt.Println("You:\n", transcript)

	// 2. generate AI response
	response, err := client.Generate(context.Background(), conversation,
		syn4pse.WithProvider("llama.cpp"),
		syn4pse.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
		syn4pse.WithParameter("max_tokens", 150),
	)
	if err != nil {
		log.Fatal(err)
	}
	conversation = append(conversation, syn4pse.NewAssistantMessage(response))
	fmt.Println("Assistant:\n", response)

	// 3. synthesize AI response
	audioOutput, err := client.SynthesizeSpeech(context.Background(), response,
		syn4pse.WithProvider("piper"),
		syn4pse.WithModelID("piper-en-us-lessac-high"),
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
