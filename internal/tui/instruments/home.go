package instruments

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
	case "esc":
		return nil, shared.Transition{Phase: shared.Back}
	case "ctrl+c":
		return nil, shared.Transition{Phase: shared.Abort}
	}
	return nil, shared.Transition{}
}

func (m *HomeModel) View() string {
	return shared.Bold.Render("Instruments") + " — Coming soon."
}
