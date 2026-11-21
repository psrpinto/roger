package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"roger/internal/examples"
	"roger/internal/sampler"
)

// onboarding is the first-run flow that guides the user through choosing a workspace.
type onboarding struct {
	app       *App
	container *fyne.Container
}

func (a *App) showOnboarding() {
	ob := &onboarding{
		app:       a,
		container: container.NewStack(),
	}
	ob.showWelcome()
	a.window.SetContent(container.NewPadded(ob.container))
}

func (ob *onboarding) setStep(title string, body fyne.CanvasObject, buttons ...fyne.CanvasObject) {
	heading := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	content := container.NewBorder(
		heading,
		container.NewHBox(buttons...),
		nil, nil,
		container.NewVBox(body),
	)
	ob.container.Objects = []fyne.CanvasObject{content}
	ob.container.Refresh()
}

func wrappedLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func (ob *onboarding) showWelcome() {
	body := wrappedLabel(
		"roger organizes drum sample packs into MPC-ready kits.\n\n" +
			"Let's get you set up. First, you'll choose a workspace folder where " +
			"roger keeps your sample packs, its configuration and the generated output.",
	)
	startBtn := widget.NewButton("Get started", ob.showChooseWorkspace)
	startBtn.Importance = widget.HighImportance

	ob.setStep("Welcome to roger", body, startBtn)
}

func (ob *onboarding) showChooseWorkspace() {
	var chosen string

	pathLabel := widget.NewLabel("No folder selected")
	pathLabel.Wrapping = fyne.TextWrapBreak

	continueBtn := widget.NewButton("Continue", func() {
		if err := ob.app.openWorkspace(chosen); err != nil {
			dialog.NewError(err, ob.app.window).Show()
			return
		}
		ob.showExamples()
	})
	continueBtn.Importance = widget.HighImportance
	continueBtn.Disable()

	chooseBtn := widget.NewButton("Choose folder…", func() {
		startDir := chosen
		if startDir == "" {
			startDir = sampler.DesktopDir()
		}
		ob.app.chooseFolder(startDir, func(dir string) {
			chosen = dir
			pathLabel.SetText(dir)
			continueBtn.Enable()
		})
	})

	body := container.NewVBox(
		wrappedLabel(
			"Choose the folder roger should use as its workspace. You can create a new "+
				"folder from the folder picker.\n\n"+
				"roger will create the following inside it:\n\n"+
				"  Kits/          Your drum sample packs\n"+
				"  Output/        Generated MPC expansions\n"+
				"  config.yaml    Drum type detection and pad layout\n"+
				"  kit.xpm        Program template",
		),
		container.NewBorder(nil, nil, chooseBtn, nil, pathLabel),
	)
	backBtn := widget.NewButton("Back", ob.showWelcome)

	ob.setStep("Choose your workspace", body, backBtn, continueBtn)
}

func (ob *onboarding) showExamples() {
	kitsSrcDir := ob.app.kitsSrcDir

	if len(sampler.ListSubdirs(kitsSrcDir)) > 0 {
		finishBtn := widget.NewButton("Finish", ob.app.showMain)
		finishBtn.Importance = widget.HighImportance
		ob.setStep("You're all set",
			wrappedLabel("This workspace already contains sample packs in Kits/."),
			finishBtn,
		)
		return
	}

	generateBtn := widget.NewButton("Generate example kits", func() {
		examples.CreateExampleDirs(kitsSrcDir)
		ob.app.showMain()
	})
	generateBtn.Importance = widget.HighImportance
	skipBtn := widget.NewButton("Skip", ob.app.showMain)

	body := wrappedLabel(
		"Place your drum sample packs in the Kits/ folder of your workspace.\n\n" +
			"Would you like roger to generate some example packs, so you can see how " +
			"packs and kits should be organized?",
	)
	ob.setStep("Example kits", body, generateBtn, skipBtn)
}
