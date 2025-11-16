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
		relic.NewSystemMessage("You are a helpful assistant."),
		relic.NewUserMessage("Why do some insects glow at night?"),
	}

	response, err := client.Generate(context.Background(), messages,
		relic.WithProvider("llama.cpp"),
		relic.WithModelID("llama-cpp-qwen2.5-1.5b-instruct-q4_k_m"),
		relic.WithParameter("temperature", 0.7),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}
