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

	text := "Soy un asistente virtual diseñado para ayudar al usuario de manera eficiente."

	audioData, err := client.SynthesizeSpeech(context.Background(), text,
		syn4pse.WithProvider("piper"),
		syn4pse.WithModelID("piper-es-ar-daniela-high"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	audioPath := filepath.Join(dir, "output.wav")

	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated %d bytes of audio saved to %s\n", len(audioData), audioPath)
}
