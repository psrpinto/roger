package kits

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/lipgloss/v2"

	"roger/internal/examples"
	"roger/internal/mpc"
	"roger/internal/sampler"
)

// Setup holds the results of kits-specific initialization.
type Setup struct {
	TopLevelDirs   []string
	PadStyles      [16]lipgloss.Style
	IsFirstRun     bool
	CreateExamples func() []string // non-nil on first run; call to create example dirs and return them
}

// SetupFunc performs kits-specific initialization (template loading,
// directory scanning, example creation) and returns the results.
type SetupFunc func() Setup

// NewSetupFunc returns a SetupFunc that initializes the kits mode.
func NewSetupFunc(baseDir, srcDir string, packArgs []string) SetupFunc {
	return func() Setup {
		if err := os.MkdirAll(srcDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", srcDir, err)
			os.Exit(1)
		}

		templatePath := filepath.Join(baseDir, "kit.xpm")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			os.WriteFile(templatePath, mpc.ProgramTemplate, 0o644)
		}
		expansionPath := filepath.Join(baseDir, "expansion.xml")
		if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
			os.WriteFile(expansionPath, mpc.ExpansionTemplate, 0o644)
		}

		mpc.LoadCustomTemplate(baseDir)
		mpc.LoadCustomExpansionTemplate(baseDir)

		topLevelDirs := packArgs
		if len(topLevelDirs) == 0 {
			topLevelDirs = sampler.ListSubdirs(srcDir)
		}

		var createExamples func() []string
		if len(packArgs) == 0 && len(topLevelDirs) == 0 {
			createExamples = func() []string {
				examples.CreateExampleDirs(srcDir)
				return sampler.ListSubdirs(srcDir)
			}
		}

		return Setup{
			TopLevelDirs:   topLevelDirs,
			PadStyles:      mpc.ExtractPadStyles(),
			IsFirstRun:     createExamples != nil,
			CreateExamples: createExamples,
		}
	}
}
