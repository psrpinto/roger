package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showHelp(window fyne.Window, baseDir string) {
	text := appHelpText(baseDir) + "\n" + kitsHelpText(baseDir)
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(label)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	dialog.NewCustom("Help", "Close", scroll, window).Show()
}

func appHelpText(baseDir string) string {
	var b strings.Builder

	fmt.Fprintln(&b, "roger organizes samples into MPC-ready kits and instruments.")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", baseDir)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "  Kits/          Drum sample packs")
	fmt.Fprintln(&b, "  Instruments/   Instrument sample directories")
	fmt.Fprintln(&b, "  Output/        Generated output (shared)")

	return b.String()
}

func kitsHelpText(baseDir string) string {
	var b strings.Builder

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "roger organizes drum sample WAV files into 16-pad MPC kits.")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", baseDir)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "  Kits/     Place your drum sample packs here")
	fmt.Fprintln(&b, "  Output/   Generated MPC kits and program files appear here")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Organize your samples in Kits/ like this:")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "  PackName/")
	fmt.Fprintln(&b, "    Kit 1/")
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b, "    Kit 2/")
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Kits can also be grouped:")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "  PackName/")
	fmt.Fprintln(&b, "    Group A/")
	fmt.Fprintln(&b, "      Kit 1/")
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b, "    Group B/")
	fmt.Fprintln(&b, "      Kit 1/")
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Samples are auto-detected by type (kick, snare, hat, etc.)")
	fmt.Fprintln(&b, "from their filenames and assigned to pads.")

	return b.String()
}
