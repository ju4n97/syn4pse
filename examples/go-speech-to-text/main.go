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
	audioPath := filepath.Join(dir, "audio.mp3")
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		panic(err)
	}

	transcript, err := client.TranscribeAudio(context.Background(), audioData,
		relic.WithProvider("whisper.cpp"),
		relic.WithModelID("whisper-cpp-tiny"),
		relic.WithParameter("language", "en"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(transcript)
}
