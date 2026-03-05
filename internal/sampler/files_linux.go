package sampler

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DesktopDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %s\n", err)
		os.Exit(1)
	}

	// XDG_DESKTOP_DIR env var takes precedence.
	if dir := os.Getenv("XDG_DESKTOP_DIR"); dir != "" {
		return dir
	}

	// Parse ~/.config/user-dirs.dirs for XDG_DESKTOP_DIR.
	if dir := readXDGDesktopDir(home); dir != "" {
		return dir
	}

	return filepath.Join(home, "Desktop")
}

// readXDGDesktopDir parses the XDG user-dirs.dirs file and returns the
// value of XDG_DESKTOP_DIR, or "" if not found.
func readXDGDesktopDir(home string) string {
	path := filepath.Join(home, ".config", "user-dirs.dirs")
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "XDG_DESKTOP_DIR") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.Trim(parts[1], "\"")
		val = strings.ReplaceAll(val, "$HOME", home)
		return val
	}
	return ""
}
