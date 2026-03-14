package kits

import "charm.land/lipgloss/v2"

// Setup holds the results of kits-specific initialization.
type Setup struct {
	TopLevelDirs []string
	PadStyles    [16]lipgloss.Style
	IsFirstRun   bool
}

// SetupFunc performs kits-specific initialization (template loading,
// directory scanning, example creation) and returns the results.
type SetupFunc func() Setup
