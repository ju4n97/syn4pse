package main

import (
	"context"
	"fmt"
	"log"

	relic "github.com/ju4n97/relic/sdk-go"
)

func main() {
	client, err := relic.NewClient("localhost:50051")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	messages := []relic.Message{
		relic.NewSystemMessage("You are a creative storyteller."),
		relic.NewUserMessage("Tell me a short story about a robot learning to paint."),
	}

	stream := client.GenerateStream(context.Background(), messages,
		relic.WithProvider("llama.cpp"),
		relic.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
		relic.WithParameter("temperature", 0.7),
	)

	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatal(chunk.Error)
		}

		fmt.Print(chunk.Content)
	}
}
