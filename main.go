package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	baseDir := filepath.Join(desktopDir(), "roger")
	srcDir := filepath.Join(baseDir, "Input")
	destDir := filepath.Join(baseDir, "Output")

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Print(renderUsage(baseDir))
		return
	}

	for _, dir := range []string{baseDir, srcDir, destDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", dir, err)
			os.Exit(1)
		}
	}

	templatePath := filepath.Join(baseDir, "template.xpm")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		os.WriteFile(templatePath, programTemplate, 0o644)
	}
	expansionPath := filepath.Join(baseDir, "expansion.xml")
	if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
		os.WriteFile(expansionPath, expansionTemplate, 0o644)
	}

	loadCustomTemplate(baseDir)
	loadCustomExpansionTemplate(baseDir)
	loadOrCreateConfig(baseDir)
	padStyles := extractPadStyles()

	var topLevelDirs []string
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			packDir := filepath.Join(srcDir, arg)
			if info, err := os.Stat(packDir); err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "error: pack directory not found: %s\n", packDir)
				os.Exit(1)
			}
			topLevelDirs = append(topLevelDirs, packDir)
		}
	} else {
		topLevelDirs = listSubdirs(srcDir)
	}

	isFirstRun := false
	if len(os.Args) <= 1 && len(topLevelDirs) == 0 {
		createExampleDirs(srcDir)
		topLevelDirs = listSubdirs(srcDir)
		isFirstRun = true
	}

	m := newModel(baseDir, srcDir, destDir, topLevelDirs, isFirstRun, padStyles)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fm := finalModel.(model)
	if fm.aborted {
		fmt.Println("Aborted.")
	} else if fm.kitCount > 0 {
		fmt.Printf("\n%d kits, %d samples, %s\n", fm.kitCount, fm.sampleCount, formatSize(fm.totalSize))
	}
}
