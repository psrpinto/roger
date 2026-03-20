package gui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"roger/internal/config"
	"roger/internal/mpc"
	"roger/internal/sampler"
)

// App is the main GUI application orchestrator.
type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	cfg        *config.Config
	baseDir    string
	kitsSrcDir string
	instSrcDir string
	destDir    string
	padColors  [16]color.NRGBA
}

// NewApp creates and initializes the Fyne application.
func NewApp() *App {
	a := &App{}

	a.fyneApp = app.New()
	a.window = a.fyneApp.NewWindow("roger")

	a.baseDir = filepath.Join(sampler.DesktopDir(), "roger")
	a.kitsSrcDir = filepath.Join(a.baseDir, "Kits")
	a.instSrcDir = filepath.Join(a.baseDir, "Instruments")
	a.destDir = filepath.Join(a.baseDir, "Output")

	for _, dir := range []string{a.baseDir, a.destDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", dir, err)
			os.Exit(1)
		}
	}

	cfg, err := config.LoadOrCreate(a.baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	a.cfg = cfg

	// Write default templates if they don't exist
	templatePath := filepath.Join(a.baseDir, "kit.xpm")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		os.WriteFile(templatePath, mpc.ProgramTemplate, 0o644)
	}
	expansionPath := filepath.Join(a.baseDir, "expansion.xml")
	if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
		os.WriteFile(expansionPath, mpc.ExpansionTemplate, 0o644)
	}
	mpc.LoadCustomTemplate(a.baseDir)
	mpc.LoadCustomExpansionTemplate(a.baseDir)

	a.padColors = ExtractPadColors()

	// Set up main menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Help", func() {
			showHelp(a.window, a.baseDir)
		}),
	)
	a.window.SetMainMenu(fyne.NewMainMenu(helpMenu))

	return a
}

// Run builds the tabbed layout and starts the application.
func (a *App) Run() {
	kitsTab := newKitsTab(a)
	instrumentsTab := newInstrumentsTab(a)

	tabs := container.NewAppTabs(kitsTab, instrumentsTab)

	a.window.SetContent(tabs)
	a.window.Resize(fyne.NewSize(900, 700))
	a.window.ShowAndRun()
}
