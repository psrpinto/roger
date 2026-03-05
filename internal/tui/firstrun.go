package tui

import tea "charm.land/bubbletea/v2"

type firstRunModel struct {
	baseDir string
}

func newFirstRunModel(baseDir string) *firstRunModel {
	return &firstRunModel{baseDir: baseDir}
}

func (m *firstRunModel) update(msg tea.Msg) (tea.Cmd, transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, transition{}
	}
	switch kp.String() {
	case "y", "Y", "enter":
		return nil, transition{phase: phaseNext}
	case "n", "N", "esc", "q", "ctrl+c":
		return nil, transition{phase: phaseAbort}
	}
	return nil, transition{}
}

func (m *firstRunModel) view() string {
	return RenderUsage(m.baseDir) + "\n" +
		"Example directories have been created in Input/ to demonstrate the structure.\n" +
		"Preview example kits? [Y/n] "
}
