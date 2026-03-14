package instruments

import (
	"fmt"
	"os"

	"roger/internal/examples"
	"roger/internal/sampler"
)

// Setup holds the results of instruments-specific initialization.
type Setup struct {
	TopLevelDirs   []string
	IsFirstRun     bool
	CreateExamples func() []string // non-nil on first run; call to create example dirs and return them
}

// SetupFunc performs instruments-specific initialization
// (directory scanning, example creation) and returns the results.
type SetupFunc func() Setup

// NewSetupFunc returns a SetupFunc that initializes the instruments mode.
func NewSetupFunc(srcDir string, packArgs []string) SetupFunc {
	return func() Setup {
		if err := os.MkdirAll(srcDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", srcDir, err)
			os.Exit(1)
		}

		topLevelDirs := packArgs
		if len(topLevelDirs) == 0 {
			topLevelDirs = sampler.ListSubdirs(srcDir)
		}

		var createExamples func() []string
		if len(packArgs) == 0 && len(topLevelDirs) == 0 {
			createExamples = func() []string {
				examples.CreateExampleInstrumentDirs(srcDir)
				return sampler.ListSubdirs(srcDir)
			}
		}

		return Setup{
			TopLevelDirs:   topLevelDirs,
			IsFirstRun:     createExamples != nil,
			CreateExamples: createExamples,
		}
	}
}
