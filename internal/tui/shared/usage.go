package shared

import tea "charm.land/bubbletea/v2"

type UsageModel struct {
	content string
}

func NewUsageModel(content string) *UsageModel {
	return &UsageModel{content: content}
}

func (m *UsageModel) Update(msg tea.Msg) (tea.Cmd, Transition) {
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

func (m *UsageModel) View() string {
	return m.content
}
