package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/tui/shared"
)

type UsageModel struct {
	inner *shared.UsageModel
}

func NewUsageModel(baseDir string) *UsageModel {
	return &UsageModel{inner: shared.NewUsageModel(RenderHelp(baseDir))}
}

func (m *UsageModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	return m.inner.Update(msg)
}

func (m *UsageModel) View() string {
	return m.inner.View()
}

func RenderHelp(baseDir string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s organizes samples into MPC-ready kits and instruments.\n", shared.Bold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Usage: %s [kits|instruments] [PackName ...]\n", shared.Bold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Modes:")
	fmt.Fprintln(&b, "  kits          Process drum sample packs into MPC kits (default)")
	fmt.Fprintln(&b, "  instruments   Instrument mode (coming soon)")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "With no mode, you will be prompted to choose.")
	fmt.Fprintln(&b, "With no pack names, all packs in the mode's input directory are processed.")
	fmt.Fprintln(&b, "Pass one or more pack names to process only those.")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", shared.Cyan.Render(baseDir))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s         Drum sample packs\n", shared.Green.Render("Kits/"))
	fmt.Fprintf(&b, "  %s  Instrument sample directories\n", shared.Green.Render("Instruments/"))
	fmt.Fprintf(&b, "  %s       Generated output (shared)\n", shared.Yellow.Render("Output/"))

	return b.String()
}
