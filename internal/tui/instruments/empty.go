package instruments

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/examples"
	"roger/internal/tui/shared"
)

type EmptyModel struct {
	baseDir string
	srcDir  string
	cursor  int
}

func NewEmptyModel(baseDir, srcDir string) *EmptyModel {
	return &EmptyModel{baseDir: baseDir, srcDir: srcDir}
}

func (m *EmptyModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, shared.Transition{}
	}
	switch kp.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}
	case "enter":
		if m.cursor == 0 {
			examples.CreateExampleInstrumentDirs(m.srcDir)
			return nil, shared.Transition{Phase: shared.Next}
		}
		return nil, shared.Transition{Phase: shared.ShowHelp}
	case "esc":
		return nil, shared.Transition{Phase: shared.Back}
	}
	return nil, shared.Transition{}
}

func (m *EmptyModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Workspace: %s\n\n", shared.Cyan.Render(m.baseDir))
	fmt.Fprintln(&b, "No instruments found. What would you like to do?")
	fmt.Fprintln(&b)

	options := []struct{ label, desc string }{
		{"Generate example files", "Create example instrument directories in Instruments/"},
		{"Show instructions", "Open the help screen"},
	}
	for i, opt := range options {
		if i > 0 {
			fmt.Fprintln(&b)
		}
		if i == m.cursor {
			fmt.Fprintf(&b, "%s %s\n", shared.Cyan.Render("▸"), shared.Bold.Render(opt.label))
			fmt.Fprintf(&b, "  %s\n", shared.Dim.Render(opt.desc))
		} else {
			fmt.Fprintf(&b, "  %s\n", shared.Dim.Render(opt.label))
		}
	}
	return b.String()
}
