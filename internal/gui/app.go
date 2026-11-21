package gui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/ncruces/zenity"

	"roger/internal/config"
	"roger/internal/mpc"
	"roger/internal/sampler"
)

const (
	appID            = "io.github.psrpinto.roger"
	workspacePrefKey = "workspace"
)

// App is the main GUI application orchestrator.
type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	cfg        *config.Config
	baseDir    string
	kitsSrcDir string
	destDir    string
	padColors  [16]color.NRGBA
}

// NewApp creates and initializes the Fyne application.
func NewApp() *App {
	a := &App{}

	a.fyneApp = app.NewWithID(appID)
	a.window = a.fyneApp.NewWindow("roger")

	a.padColors = ExtractPadColors()

	// Set up main menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Change Workspace…", func() {
			a.changeWorkspace()
		}),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Help", func() {
			showHelp(a.window, a.baseDir)
		}),
	)
	a.window.SetMainMenu(fyne.NewMainMenu(fileMenu, helpMenu))

	return a
}

// Run shows the main UI, or onboarding if no workspace has been chosen yet,
// and starts the application.
func (a *App) Run() {
	dir := a.fyneApp.Preferences().String(workspacePrefKey)
	if dir != "" && dirExists(dir) && a.openWorkspace(dir) == nil {
		a.showMain()
	} else {
		a.showOnboarding()
	}

	a.window.Resize(fyne.NewSize(900, 700))
	a.window.ShowAndRun()
}

// showMain builds the tabbed layout for the current workspace.
func (a *App) showMain() {
	kitsTab := newKitsTab(a)
	tabs := container.NewAppTabs(kitsTab)

	a.window.SetContent(tabs)
}

// openWorkspace makes dir the current workspace, creating its folders,
// config and templates if needed.
func (a *App) openWorkspace(dir string) error {
	kitsSrcDir := filepath.Join(dir, "Kits")
	destDir := filepath.Join(dir, "Output")

	for _, d := range []string{dir, kitsSrcDir, destDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	cfg, err := config.LoadOrCreate(dir)
	if err != nil {
		return err
	}

	// Write default templates if they don't exist
	templatePath := filepath.Join(dir, "kit.xpm")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		os.WriteFile(templatePath, mpc.ProgramTemplate, 0o644)
	}
	expansionPath := filepath.Join(dir, "expansion.xml")
	if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
		os.WriteFile(expansionPath, mpc.ExpansionTemplate, 0o644)
	}
	mpc.LoadCustomTemplate(dir)
	mpc.LoadCustomExpansionTemplate(dir)

	a.cfg = cfg
	a.baseDir = dir
	a.kitsSrcDir = kitsSrcDir
	a.destDir = destDir
	a.fyneApp.Preferences().SetString(workspacePrefKey, dir)

	return nil
}

// chooseFolder shows the native folder picker starting at startDir and calls
// onChosen with the selected path. If no native picker is available, it falls
// back to Fyne's folder dialog.
func (a *App) chooseFolder(startDir string, onChosen func(dir string)) {
	go func() {
		dir, err := zenity.SelectFile(
			zenity.Directory(),
			zenity.Title("Choose workspace folder"),
			zenity.Filename(startDir+string(filepath.Separator)),
		)
		fyne.Do(func() {
			switch {
			case err == zenity.ErrCanceled:
			case err != nil:
				a.chooseFolderFallback(startDir, onChosen)
			default:
				onChosen(dir)
			}
		})
	}()
}

func (a *App) chooseFolderFallback(startDir string, onChosen func(dir string)) {
	dlg := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.NewError(err, a.window).Show()
			return
		}
		if uri == nil {
			return
		}
		onChosen(uri.Path())
	}, a.window)

	if startDir != "" {
		if lister, err := storage.ListerForURI(storage.NewFileURI(startDir)); err == nil {
			dlg.SetLocation(lister)
		}
	}
	dlg.Show()
}

func (a *App) changeWorkspace() {
	startDir := a.baseDir
	if startDir == "" {
		startDir = sampler.DesktopDir()
	}
	a.chooseFolder(startDir, func(dir string) {
		if err := a.openWorkspace(dir); err != nil {
			dialog.NewError(err, a.window).Show()
			return
		}
		a.showMain()
	})
}

func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}
