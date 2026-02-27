package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Paths relative to project root (resolved at startup)
var (
	whisperBin     string
	whisperModel   string
	inputWavFile   string
	outputTextFile string
)

func init() {
	root, err := findProjectRoot()
	if err != nil {
		// Fallback to cwd if we can't find go.mod
		root, _ = os.Getwd()
	}
	whisperBin = filepath.Join(root, "third_party", "whisper", "build", "bin", "whisper-cli")
	modelName := "ggml-large-v3-turbo.bin"
	if m := os.Getenv("WHISPER_MODEL"); m != "" {
		modelName = m
	}
	whisperModel = filepath.Join(root, "third_party", "whisper", "models", modelName)
	inputWavFile = filepath.Join(root, "cmd", "vtt", "input.wav")
	outputTextFile = filepath.Join(root, "output.txt")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func StartRecordingAudio() {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("Could not find ffmpeg package.")
		return
	}

	// -ar 16000 is vital for Whisper
	// afftdn: light noise reduction (helps transcription quality in noisy environments)
	cmd := exec.Command("ffmpeg", "-y", "-f", "avfoundation", "-i", ":0", "-af", "afftdn=nr=15:nf=-25:tn=1", "-ar", "16000", "-ac", "1", inputWavFile)

	// Pipe stdin so we can send 'q' to stop ffmpeg gracefully and get a valid WAV file.
	// SIGINT often leaves the WAV truncated or corrupt, causing poor transcription.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Failed to create stdin pipe: %v\n", err)
		return
	}

	fmt.Println("Recording... Press ENTER to stop.")
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start ffmpeg")
		return
	}

	fmt.Scanln() // Wait for user input

	// Graceful stop: send 'q' so ffmpeg flushes and finalizes the WAV file.
	stdin.Write([]byte("q\n"))
	stdin.Close()
	cmd.Wait()
	// Brief delay so the filesystem has fully flushed the WAV (avoids truncated reads).
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Finished recording. Saved to: %s\n", inputWavFile)
}

func TranscribeAudio() {
	fmt.Println("Transcribing...")

	// -nt: no timestamps  -np: no prints (results only)
	// -sns: suppress non-speech tokens (reduces hallucinations)
	// -nth 0.3: lower no-speech threshold so more audio is decoded as speech (default 0.6 can drop quiet speech)
	// -l en: explicit English (avoids auto-detect mistakes on short clips)
	cmd := exec.Command(whisperBin, "-m", whisperModel, "-f", inputWavFile, "-nt", "-np", "-sns", "-nth", "0.3", "-l", "en")

	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			fmt.Printf("Transcription failed: %v\n%s\n", err, string(ee.Stderr))
		} else {
			fmt.Printf("Transcription failed: %v\n", err)
		}
		return
	}

	text := strings.TrimSpace(string(out))
	fmt.Printf("\n--- TRANSCRIPTION ---\n%s\n---------------------\n", text)

	if err := os.WriteFile(outputTextFile, []byte(text+"\n"), 0644); err != nil {
		fmt.Printf("Failed to write %s: %v\n", outputTextFile, err)
	} else {
		fmt.Printf("Output written to: %s\n", outputTextFile)
	}
}

func main() {
	fmt.Println("STARTED")

	StartRecordingAudio()
	TranscribeAudio()
}
