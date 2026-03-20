package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"roger/internal/config"
	"roger/internal/sampler"
)

func Run() int {
	baseDir := filepath.Join(sampler.DesktopDir(), "roger")
	kitsSrcDir := filepath.Join(baseDir, "Kits")
	instSrcDir := filepath.Join(baseDir, "Instruments")
	destDir := filepath.Join(baseDir, "Output")

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Print(RenderHelp(baseDir))
		return 0
	}

	for _, dir := range []string{baseDir, destDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", dir, err)
			return 1
		}
	}

	cfg, err := config.LoadOrCreate(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return 1
	}

	var mode Mode
	var packArgs []string
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "kits":
			mode = ModeKits
			args = args[1:]
		case "instruments":
			mode = ModeInstruments
			args = args[1:]
		}
	}
	if len(args) > 0 {
		if mode == "" {
			mode = ModeKits
		}
		modeSrcDir := kitsSrcDir
		if mode == ModeInstruments {
			modeSrcDir = instSrcDir
		}
		for _, arg := range args {
			packDir := filepath.Join(modeSrcDir, arg)
			if info, err := os.Stat(packDir); err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "error: pack directory not found: %s\n", packDir)
				return 1
			}
			packArgs = append(packArgs, packDir)
		}
	}

	m := NewModel(baseDir, kitsSrcDir, instSrcDir, destDir, cfg, mode, packArgs)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return 1
	}

	fm := finalModel.(*Model)
	if fm.Aborted {
		fmt.Println("Aborted.")
	} else if fm.KitCount > 0 {
		fmt.Printf("\n%d kits, %d samples, %s\n", fm.KitCount, fm.SampleCount, sampler.FormatSize(fm.TotalSize))
	}

	return 0
}
