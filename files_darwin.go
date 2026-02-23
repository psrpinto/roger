package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func desktopDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %s\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, "Desktop")
}
