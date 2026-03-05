package sampler

import (
	"fmt"
	"os"
	"path/filepath"
)

func DesktopDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %s\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, "Desktop")
}
