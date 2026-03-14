package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"roger/internal/config"
	"roger/internal/examples"
	"roger/internal/mpc"
	"roger/internal/sampler"
	"roger/internal/tui"
)

func main() {
	baseDir := filepath.Join(sampler.DesktopDir(), "roger")
	srcDir := filepath.Join(baseDir, "Input")
	destDir := filepath.Join(baseDir, "Output")

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Print(tui.RenderUsage(baseDir))
		return
	}

	for _, dir := range []string{baseDir, srcDir, destDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", dir, err)
			os.Exit(1)
		}
	}

	cfg, err := config.LoadOrCreate(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	// Parse mode and pack arguments
	var mode tui.Mode
	var packArgs []string
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "kits":
			mode = tui.ModeKits
			args = args[1:]
		case "instruments":
			mode = tui.ModeInstruments
			args = args[1:]
		}
	}
	if len(args) > 0 {
		if mode == "" {
			mode = tui.ModeKits
		}
		for _, arg := range args {
			packDir := filepath.Join(srcDir, arg)
			if info, err := os.Stat(packDir); err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "error: pack directory not found: %s\n", packDir)
				os.Exit(1)
			}
			packArgs = append(packArgs, packDir)
		}
	}

	kitsSetupFn := func() tui.KitsSetup {
		templatePath := filepath.Join(baseDir, "template.xpm")
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

		isFirstRun := false
		if len(packArgs) == 0 && len(topLevelDirs) == 0 {
			examples.CreateExampleDirs(srcDir)
			topLevelDirs = sampler.ListSubdirs(srcDir)
			isFirstRun = true
		}

		return tui.KitsSetup{
			TopLevelDirs: topLevelDirs,
			PadStyles:    mpc.ExtractPadStyles(),
			IsFirstRun:   isFirstRun,
		}
	}

	m := tui.NewModel(baseDir, srcDir, destDir, cfg, mode, kitsSetupFn)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fm := finalModel.(tui.Model)
	if fm.Aborted {
		fmt.Println("Aborted.")
	} else if fm.KitCount > 0 {
		fmt.Printf("\n%d kits, %d samples, %s\n", fm.KitCount, fm.SampleCount, sampler.FormatSize(fm.TotalSize))
	}
}
