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

	text := "Soy un asistente virtual diseñado para ayudar al usuario de manera eficiente."

	audioData, err := client.SynthesizeSpeech(context.Background(), text,
		relic.WithProvider("piper"),
		relic.WithModelID("piper-es-ar-daniela-high"),
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
