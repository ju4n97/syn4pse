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
		syn4pse.NewSystemMessage("You are a helpful assistant."),
		syn4pse.NewUserMessage("Why do some insects glow at night?"),
	}

	response, err := client.Generate(context.Background(), messages,
		syn4pse.WithProvider("llama.cpp"),
		syn4pse.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
		syn4pse.WithParameter("temperature", 0.7),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}
