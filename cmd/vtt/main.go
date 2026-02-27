package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Define paths relative to the project root
const (
	whisperBin   = "./third_party/whisper/build/bin/whisper-cli"
	whisperModel = "./third_party/whisper/models/ggml-base.en.bin"
	outputFile   = "cmd/vtt/input.wav"
)

func StartRecordingAudio() {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("Could not find ffmpeg package.")
		return
	}

	// -ar 16000 is vital for Whisper
	cmd := exec.Command("ffmpeg", "-y", "-f", "avfoundation", "-i", ":0", "-ar", "16000", "-ac", "1", outputFile)

	fmt.Println("Recording... Press ENTER to stop.")
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start ffmpeg")
		return
	}

	fmt.Scanln() // Wait for user input

	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
	fmt.Printf("Finished recording. Saved to: %s\n", outputFile)
}

func TranscribeAudio() {
	fmt.Println("Transcribing...")

	// -nt: no timestamps
	// -np: no prints (results only)
	cmd := exec.Command(whisperBin, "-m", whisperModel, "-f", outputFile, "-nt", "-np")

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Transcription failed: %v\n", err)
		return
	}

	text := strings.TrimSpace(string(out))
	fmt.Printf("\n--- TRANSCRIPTION ---\n%s\n---------------------\n", text)
}

func main() {
	fmt.Println("STARTED")

	StartRecordingAudio()
	TranscribeAudio()
}
