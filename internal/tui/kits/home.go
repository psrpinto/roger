package kits

import (
	tea "charm.land/bubbletea/v2"

	"roger/internal/tui/shared"
)

type HomeModel struct{}

func NewHomeModel() *HomeModel {
	return &HomeModel{}
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, shared.Transition{}
	}
	switch kp.String() {
	case "enter":
		return nil, shared.Transition{Phase: shared.Next}
	case "esc", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Back}
	}
	return nil, shared.Transition{}
}

func (m *HomeModel) View() string {
	return "\nPress Enter to scan.\n"
}
