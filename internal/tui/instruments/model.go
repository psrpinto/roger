package instruments

import (
	tea "charm.land/bubbletea/v2"

	"roger/internal/tui/shared"
)

type Model struct{}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, shared.Transition{}
	}
	switch kp.String() {
	case "esc":
		return nil, shared.Transition{Phase: shared.Back}
	case "ctrl+c":
		return nil, shared.Transition{Phase: shared.Abort}
	default:
		return nil, shared.Transition{Phase: shared.Next}
	}
}

func (m *Model) View() string {
	return shared.Bold.Render("Instruments") + " — Coming soon.\n\nPress any key to exit."
}
