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

	conversation := []syn4pse.Message{
		syn4pse.NewSystemMessage("You are a helpful coding assistant."),
	}

	// first question
	question1 := syn4pse.NewUserMessage("How do I read a file in Go?")
	conversation = append(conversation, question1)
	fmt.Println("\nYou:\n", question1.Content)

	// first response
	response1, err := client.Generate(context.Background(), conversation,
		syn4pse.WithProvider("llama.cpp"),
		syn4pse.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nAssistant:\n", response1)

	// second question
	question2 := syn4pse.NewUserMessage("What about handling errors?")
	conversation = append(conversation, question2)
	fmt.Println("\nYou:\n", question2.Content)

	// second response
	response2, err := client.Generate(context.Background(), conversation,
		syn4pse.WithProvider("llama.cpp"),
		syn4pse.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nAssistant:\n", response2)
}
