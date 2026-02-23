package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func desktopDir() string {
	dir, err := windows.KnownFolderPath(windows.FOLDERID_Desktop, 0)
	if err == nil && dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %s\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, "Desktop")
}
