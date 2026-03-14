package kits

import (
	"fmt"
	"strings"

	"roger/internal/tui/shared"
)

func RenderHelp(baseDir string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s organizes drum sample WAV files into 16-pad MPC kits.\n", shared.Bold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", shared.Cyan.Render(baseDir))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s   Place your drum sample packs here\n", shared.Green.Render("Kits/"))
	fmt.Fprintf(&b, "  %s  Generated MPC kits and program files appear here\n", shared.Yellow.Render("Output/"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Organize your samples in Kits/ like this:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", shared.Bold.Render("PackName/"))
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Kit 1/"))
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Kit 2/"))
	fmt.Fprintln(&b, "      Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Kits can also be grouped:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", shared.Bold.Render("PackName/"))
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Group A/"))
	fmt.Fprintf(&b, "      %s\n", shared.Dim.Render("Kit 1/"))
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintf(&b, "    %s\n", shared.Dim.Render("Group B/"))
	fmt.Fprintf(&b, "      %s\n", shared.Dim.Render("Kit 1/"))
	fmt.Fprintln(&b, "        Kick.wav, Snare.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Samples are auto-detected by type (kick, snare, hat, etc.) from their")
	fmt.Fprintln(&b, "filenames and assigned to pads.")

	return b.String()
}
