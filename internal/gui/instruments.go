package gui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"roger/internal/examples"
	"roger/internal/sampler"
)

func newInstrumentsTab(app *App) *container.TabItem {
	content := container.NewStack()

	if err := os.MkdirAll(app.instSrcDir, 0o755); err != nil {
		content.Objects = []fyne.CanvasObject{widget.NewLabel(fmt.Sprintf("Error: %s", err))}
		return container.NewTabItem("Instruments", content)
	}

	topLevelDirs := sampler.ListSubdirs(app.instSrcDir)

	if len(topLevelDirs) == 0 {
		// Empty state
		genBtn := widget.NewButton("Generate example files", func() {
			examples.CreateExampleInstrumentDirs(app.instSrcDir)
			// Refresh: rebuild the tab content
			dirs := sampler.ListSubdirs(app.instSrcDir)
			if len(dirs) > 0 {
				content.Objects = []fyne.CanvasObject{buildInstrumentsList(dirs)}
				content.Refresh()
			}
		})
		instrBtn := widget.NewButton("Show instructions", func() {
			showHelp(app.window, app.baseDir)
		})

		empty := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Workspace: %s", app.baseDir)),
			widget.NewLabel(""),
			widget.NewLabel("No instruments found. What would you like to do?"),
			widget.NewLabel(""),
			genBtn,
			instrBtn,
		)
		content.Objects = []fyne.CanvasObject{empty}
	} else {
		content.Objects = []fyne.CanvasObject{buildInstrumentsList(topLevelDirs)}
	}

	return container.NewTabItem("Instruments", content)
}

func buildInstrumentsList(dirs []string) fyne.CanvasObject {
	labels := make([]string, len(dirs))
	for i, d := range dirs {
		labels[i] = sampler.FormatKitName(filepath.Base(d))
	}

	checkGroup := widget.NewCheckGroup(labels, nil)
	checkGroup.SetSelected(labels)

	return container.NewBorder(
		widget.NewLabel("Instruments (coming soon):"),
		nil, nil, nil,
		container.NewVScroll(checkGroup),
	)
}
