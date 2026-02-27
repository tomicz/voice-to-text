package main

import (
	"fmt"
	"os"
	"os/exec"
)

func StartRecordingAudio() {

	path, err := exec.LookPath("ffmpeg")

	if err != nil {
		fmt.Println("Could not find ffmpeg package. Starting installation.")
		return
	}
	fmt.Printf("Found package at path: %s\n", path)

	outputFile := "input.wav"

	// 2. The Command
	// -y: overwrite file if exists
	// -f avfoundation (macOS) / -f pulse (Linux) / -f dshow (Windows)
	// -i ":0": default mic
	// -ar 16000: 16kHz (Whisper's favorite)
	// -ac 1: Mono
	cmd := exec.Command("ffmpeg", "-y", "-f", "avfoundation", "-i", ":0", "-ar", "16000", "-ac", "1", outputFile)

	fmt.Println("Recording has started...")

	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start ffmpeg")
		return
	}

	fmt.Scanln()

	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()

	fmt.Printf("Finished recording and saved the file to: %s\n", outputFile)

}

func main() {
	fmt.Println("STARTED")

	StartRecordingAudio()
}
