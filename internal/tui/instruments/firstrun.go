package instruments

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/tui/shared"
)

type FirstRunModel struct {
	baseDir string
}

func NewFirstRunModel(baseDir string) *FirstRunModel {
	return &FirstRunModel{baseDir: baseDir}
}

func (m *FirstRunModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, shared.Transition{}
	}
	switch kp.String() {
	case "y", "Y", "enter":
		return nil, shared.Transition{Phase: shared.Next}
	case "n", "N", "esc", "q", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Abort}
	}
	return nil, shared.Transition{}
}

func (m *FirstRunModel) View() string {
	return renderUsage(m.baseDir) + "\n" +
		"Example directories have been created in Instruments/ to demonstrate the structure.\n" +
		"Continue? [Y/n] "
}

func renderUsage(baseDir string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s organizes instrument samples for MPC instruments.\n", shared.Bold.Render("roger"))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Workspace: %s\n", shared.Cyan.Render(baseDir))
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s  Place your instrument samples here\n", shared.Green.Render("Instruments/"))
	fmt.Fprintf(&b, "  %s       Generated output appears here\n", shared.Yellow.Render("Output/"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Organize your samples in Instruments/ like this:")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", shared.Bold.Render("InstrumentName/"))
	fmt.Fprintln(&b, "    C3.wav, D3.wav, E3.wav, ...")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Each subdirectory becomes an instrument. WAV files should be named")
	fmt.Fprintln(&b, "by their note (e.g., C3.wav, D#4.wav).")

	return b.String()
}
