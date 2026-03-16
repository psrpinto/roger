package shared

import tea "charm.land/bubbletea/v2"

type HelpModel struct {
	content string
}

func NewHelpModel(content string) *HelpModel {
	return &HelpModel{content: content}
}

func (m *HelpModel) Update(msg tea.Msg) (tea.Cmd, Transition) {
	if kp, ok := msg.(tea.KeyPressMsg); ok {
		switch kp.String() {
		case "ctrl+c":
			return nil, Transition{Phase: Abort}
		case "esc":
			return nil, Transition{Phase: Back}
		}
	}
	return nil, Transition{}
}

func (m *HelpModel) View() string {
	return m.content
}
