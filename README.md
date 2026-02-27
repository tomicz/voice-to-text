# Voice to Text

Record audio from your microphone and transcribe it to text using [whisper.cpp](https://github.com/ggerganov/whisper.cpp). The program records until you press Enter, then runs the recording through Whisper and prints the transcription.

**Platform:** macOS (capture uses AVFoundation). Linux/Windows would require changing the ffmpeg input in code.

## Prerequisites

- **Go** 1.25 or later
- **ffmpeg** (install via Homebrew: `brew install ffmpeg`)
- **whisper.cpp** (included as a submodule; must be built and the model downloaded)

## Setup

1. **Clone with submodules**

   ```bash
   git clone --recurse-submodules <repo-url>
   cd voice-to-text
   ```

   If the repo is already cloned:

   ```bash
   git submodule update --init --recursive
   ```

2. **Build whisper.cpp**

   From the project root:

   ```bash
   cd third_party/whisper
   mkdir -p build && cd build
   cmake ..
   make
   cd ../../..
   ```

   Ensure the CLI binary is at `third_party/whisper/build/bin/whisper-cli`. If your build puts the binary elsewhere (e.g. `third_party/whisper/main`), copy or symlink it to that path.

3. **Download a Whisper model**

   The app expects the English base model at:

   `third_party/whisper/models/ggml-base.en.bin`

   From the whisper.cpp repo you can use their script, or download manually, for example:

   ```bash
   mkdir -p third_party/whisper/models
   curl -L -o third_party/whisper/models/ggml-base.en.bin \
     https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin
   ```

   Model files are large (e.g. 150MB+); they are gitignored.

## Usage

Run from the **project root** (paths to the CLI and model are relative to it):

```bash
go run ./cmd/vtt
```

Or build and run the binary:

```bash
go build -o vtt ./cmd/vtt
./vtt
```

1. The program starts and prints "Recording... Press ENTER to stop."
2. Speak; when done, press Enter to stop recording.
3. The audio is saved to `cmd/vtt/input.wav` and then transcribed.
4. The transcription is printed under "--- TRANSCRIPTION ---".

Recordings and generated `.wav` files are gitignored.

## Project layout

- `cmd/vtt/main.go` – entrypoint: recording via ffmpeg, then transcription via whisper-cli
- `third_party/whisper` – [whisper.cpp](https://github.com/ggerganov/whisper.cpp) as a git submodule (build output and models are local only)

## How it works

- **Recording:** ffmpeg captures from the default system microphone (`:0`) with AVFoundation. Output is 16 kHz, mono, which matches Whisper’s expected input.
- **Transcription:** The saved WAV is passed to `whisper-cli` with the base English model. The `-nt` and `-np` flags restrict output to the transcribed text only.
