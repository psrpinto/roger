package kits

import (
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
	case "esc":
		return nil, shared.Transition{Phase: shared.Back}
	case "n", "N", "q", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Abort}
	}
	return nil, shared.Transition{}
}

func (m *FirstRunModel) View() string {
	return RenderHelp(m.baseDir) + "\n" +
		"Example directories have been created in Kits/ to demonstrate the structure.\n" +
		"Preview example kits? [Y/n] "
}
