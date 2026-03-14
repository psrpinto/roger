package tui

import tea "charm.land/bubbletea/v2"

type instrumentsModel struct{}

func newInstrumentsModel() *instrumentsModel {
	return &instrumentsModel{}
}

func (m *instrumentsModel) update(msg tea.Msg) (tea.Cmd, transition) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return nil, transition{phase: phaseNext}
	}
	return nil, transition{}
}

func (m *instrumentsModel) view() string {
	return styleBold.Render("Instruments") + " — Coming soon.\n\nPress any key to exit."
}
