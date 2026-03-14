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
		return nil, shared.Transition{Phase: shared.Next, Data: true}
	case "n", "N":
		return nil, shared.Transition{Phase: shared.Next, Data: false}
	case "esc", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Back}
	}
	return nil, shared.Transition{}
}

func (m *FirstRunModel) View() string {
	return "Workspace: " + shared.Cyan.Render(m.baseDir) + "\n\n" +
		"No kits found. Would you like example directories to be created in Kits/? [Y/n] "
}
