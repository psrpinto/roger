package examples

import (
	"fmt"
	"os"
	"path/filepath"
)

var exampleFileNames = []string{
	"Kick.wav", "Snare.wav", "Closed Hat.wav", "Open Hat.wav",
	"Clap.wav", "Rim.wav", "Tom Lo.wav", "Tom Mid.wav",
	"Tom Hi.wav", "Cymbal.wav", "Cowbell.wav", "Shaker.wav",
	"Tambourine.wav", "Percussion.wav", "Kick 2.wav", "Snare 2.wav",
}

func createExampleKit(kitDir string) error {
	if err := os.MkdirAll(kitDir, 0o755); err != nil {
		return err
	}
	for _, name := range exampleFileNames {
		if err := createDummyWav(filepath.Join(kitDir, name)); err != nil {
			return err
		}
	}
	return nil
}

// CreateExampleDirs creates example input directories with dummy WAV files.
func CreateExampleDirs(inputDir string) {
	// Single group example: Kits/ExamplePack/Kit 1, Kit 2
	singleCat := filepath.Join(inputDir, "ExamplePack")
	for _, kit := range []string{"Kit 1", "Kit 2"} {
		if err := createExampleKit(filepath.Join(singleCat, kit)); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating example directory: %s\n", err)
			os.Exit(1)
		}
	}
	if err := createDummyPng(filepath.Join(singleCat, "cover.png")); err != nil {
		fmt.Fprintf(os.Stderr, "error: creating example image: %s\n", err)
	}

	// Multi-group example: Kits/ExampleGroupedPack/Group A/Kit 1, Kit 2, Group B/Kit 1, Kit 2
	multiCat := filepath.Join(inputDir, "ExampleGroupedPack")
	for _, cat := range []string{"Group A", "Group B"} {
		for _, kit := range []string{"Kit 1", "Kit 2"} {
			if err := createExampleKit(filepath.Join(multiCat, cat, kit)); err != nil {
				fmt.Fprintf(os.Stderr, "error: creating example directory: %s\n", err)
				os.Exit(1)
			}
		}
	}
	if err := createDummyPng(filepath.Join(multiCat, "cover.png")); err != nil {
		fmt.Fprintf(os.Stderr, "error: creating example image: %s\n", err)
	}
}
