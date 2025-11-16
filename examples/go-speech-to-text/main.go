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
	audioPath := filepath.Join(dir, "audio.mp3")
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		panic(err)
	}

	transcript, err := client.TranscribeAudio(context.Background(), audioData,
		syn4pse.WithProvider("whisper.cpp"),
		syn4pse.WithModelID("whisper-cpp-tiny"),
		syn4pse.WithParameter("language", "en"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(transcript)
}
