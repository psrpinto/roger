package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/tui/shared"
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

func (m *modeSelectModel) update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, shared.Transition{}
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
		return nil, shared.Transition{Phase: shared.Next, Data: modeOptions[m.cursor].mode}
	case "esc", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Abort}
	}
	return nil, shared.Transition{}
}

func (m *modeSelectModel) view() string {
	var b strings.Builder
	fmt.Fprintln(&b, shared.Bold.Render("Select mode:"))
	fmt.Fprintln(&b)
	for i, opt := range modeOptions {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", shared.Bold.Render(">"), opt.label)
			fmt.Fprintf(&b, "    %s\n", shared.Cyan.Faint(true).Render(opt.description))
		} else {
			fmt.Fprintf(&b, "    %s\n", shared.Dim.Render(opt.label))
		}
	}
	return b.String()
}
