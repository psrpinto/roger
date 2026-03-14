package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type modeOption struct {
	label       string
	description string
	mode        Mode
}

var modeOptions = []modeOption{
	{"Kits", "Create Drum programs", ModeKits},
	{"Instruments", "Create Keygroup programs", ModeInstruments},
}

type modeSelectModel struct {
	cursor int
}

func newModeSelectModel() *modeSelectModel {
	return &modeSelectModel{}
}

func (m *modeSelectModel) update(msg tea.Msg) (tea.Cmd, transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, transition{}
	}
	switch kp.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(modeOptions)-1 {
			m.cursor++
		}
	case "enter":
		return nil, transition{phase: phaseNext, data: modeOptions[m.cursor].mode}
	case "esc", "ctrl+c":
		return nil, transition{phase: phaseAbort}
	}
	return nil, transition{}
}

func (m *modeSelectModel) view() string {
	var b strings.Builder
	fmt.Fprintln(&b, styleBold.Render("Select mode:"))
	fmt.Fprintln(&b)
	for i, opt := range modeOptions {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", styleBold.Render(">"), opt.label)
			fmt.Fprintf(&b, "    %s\n", styleCyan.Faint(true).Render(opt.description))
		} else {
			fmt.Fprintf(&b, "    %s\n", styleDim.Render(opt.label))
		}
	}
	return b.String()
}
