package examples

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateExampleInstrumentDirs creates example instrument directories with dummy WAV files.
func CreateExampleInstrumentDirs(inputDir string) {
	pianoDir := filepath.Join(inputDir, "ExamplePiano")
	if err := os.MkdirAll(pianoDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: creating example directory: %s\n", err)
		os.Exit(1)
	}
	for _, note := range []string{"C3.wav", "D3.wav", "E3.wav", "F3.wav", "G3.wav", "A3.wav", "B3.wav", "C4.wav"} {
		if err := createDummyWav(filepath.Join(pianoDir, note)); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating example file: %s\n", err)
			os.Exit(1)
		}
	}
}
