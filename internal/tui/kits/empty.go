package kits

import (
	tea "charm.land/bubbletea/v2"

	"roger/internal/examples"
	"roger/internal/tui/shared"
)

type EmptyModel struct {
	baseDir string
	srcDir  string
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
	case "y", "Y", "enter":
		examples.CreateExampleDirs(m.srcDir)
		return nil, shared.Transition{Phase: shared.Next}
	case "n", "N":
		return nil, shared.Transition{Phase: shared.Next}
	case "esc", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Back}
	}
	return nil, shared.Transition{}
}

func (m *EmptyModel) View() string {
	return "Workspace: " + shared.Cyan.Render(m.baseDir) + "\n\n" +
		"No kits found. Would you like example directories to be created in Kits/? [Y/n] "
}
