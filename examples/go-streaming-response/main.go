package main

import (
	"context"
	"fmt"
	"log"

	syn4pse "github.com/ju4n97/syn4pse/sdk-go"
)

func main() {
	client, err := syn4pse.NewClient("localhost:50051")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	messages := []syn4pse.Message{
		syn4pse.NewSystemMessage("You are a creative storyteller."),
		syn4pse.NewUserMessage("Tell me a short story about a robot learning to paint."),
	}

	stream := client.GenerateStream(context.Background(), messages,
		syn4pse.WithProvider("llama.cpp"),
		syn4pse.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
		syn4pse.WithParameter("temperature", 0.7),
	)

	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatal(chunk.Error)
		}

		fmt.Print(chunk.Content)
	}
}
