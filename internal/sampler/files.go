package sampler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// ListSubdirs returns the full paths of immediate subdirectories of dir.
func ListSubdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading directory %s: %s\n", dir, err)
		os.Exit(1)
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(dir, entry.Name()))
		}
	}
	return dirs
}

// FindKitDirs walks root recursively and returns sorted paths of directories containing .wav files.
func FindKitDirs(root string) []string {
	var kitDirs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}
		for _, e := range entries {
			if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".wav") {
				kitDirs = append(kitDirs, path)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil
	}
	sort.Strings(kitDirs)
	return kitDirs
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func ReadSampleInfo(path string) (frameCount int, sampleRate int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return 0, 0, err
	}

	return buffer.NumFrames(), int(decoder.SampleRate), nil
}

// ReadPCMData reads a WAV file and returns the full PCM data along with format info.
func ReadPCMData(path string) (*audio.IntBuffer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}
	return buf, nil
}
