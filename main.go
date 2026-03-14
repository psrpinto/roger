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
	"roger/internal/tui/instruments"
	"roger/internal/tui/kits"
)

func main() {
	baseDir := filepath.Join(sampler.DesktopDir(), "roger")
	kitsSrcDir := filepath.Join(baseDir, "Kits")
	instSrcDir := filepath.Join(baseDir, "Instruments")
	destDir := filepath.Join(baseDir, "Output")

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Print(tui.RenderGeneralUsage(baseDir))
		return
	}

	for _, dir := range []string{baseDir, kitsSrcDir, instSrcDir, destDir} {
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
		// Resolve pack args against mode-specific input directory
		modeSrcDir := kitsSrcDir
		if mode == tui.ModeInstruments {
			modeSrcDir = instSrcDir
		}
		for _, arg := range args {
			packDir := filepath.Join(modeSrcDir, arg)
			if info, err := os.Stat(packDir); err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "error: pack directory not found: %s\n", packDir)
				os.Exit(1)
			}
			packArgs = append(packArgs, packDir)
		}
	}

	kitsSetupFn := func() kits.Setup {
		templatePath := filepath.Join(kitsSrcDir, "template.xpm")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			os.WriteFile(templatePath, mpc.ProgramTemplate, 0o644)
		}
		expansionPath := filepath.Join(kitsSrcDir, "expansion.xml")
		if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
			os.WriteFile(expansionPath, mpc.ExpansionTemplate, 0o644)
		}

		mpc.LoadCustomTemplate(kitsSrcDir)
		mpc.LoadCustomExpansionTemplate(kitsSrcDir)

		topLevelDirs := packArgs
		if len(topLevelDirs) == 0 {
			topLevelDirs = sampler.ListSubdirs(kitsSrcDir)
		}

		isFirstRun := false
		if len(packArgs) == 0 && len(topLevelDirs) == 0 {
			examples.CreateExampleDirs(kitsSrcDir)
			topLevelDirs = sampler.ListSubdirs(kitsSrcDir)
			isFirstRun = true
		}

		return kits.Setup{
			TopLevelDirs: topLevelDirs,
			PadStyles:    mpc.ExtractPadStyles(),
			IsFirstRun:   isFirstRun,
		}
	}

	instrumentsSetupFn := func() instruments.Setup {
		topLevelDirs := packArgs
		if len(topLevelDirs) == 0 {
			topLevelDirs = sampler.ListSubdirs(instSrcDir)
		}

		isFirstRun := false
		if len(packArgs) == 0 && len(topLevelDirs) == 0 {
			examples.CreateExampleInstrumentDirs(instSrcDir)
			topLevelDirs = sampler.ListSubdirs(instSrcDir)
			isFirstRun = true
		}

		return instruments.Setup{
			TopLevelDirs: topLevelDirs,
			IsFirstRun:   isFirstRun,
		}
	}

	m := tui.NewModel(baseDir, kitsSrcDir, instSrcDir, destDir, cfg, mode, kitsSetupFn, instrumentsSetupFn)
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
