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

const (
	minWidth  = 140
	minHeight = 40
)

type HomeModel struct {
	cursor int
	width  int
	height int
}

func NewHomeModel() *HomeModel {
	return &HomeModel{}
}

func (m *HomeModel) Resize(w, h int) {
	m.width = w
	m.height = h
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
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

func (m *HomeModel) View() string {
	var b strings.Builder
	logo := shared.Bold.Render(
		"  _ __ ___   __ _  ___ _ __\n" +
			"  | '__/ _ \\ / _` |/ _ \\ '__|\n" +
			"  | | | (_) | (_| |  __/ |\n" +
			"  |_|  \\___/ \\__, |\\___|_|\n" +
			"             |___/")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "%s\n", logo)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "  What would you like to create?")
	fmt.Fprintln(&b)
	for i, opt := range modeOptions {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", shared.Cyan.Render("▸"), shared.Bold.Render(opt.label))
			fmt.Fprintf(&b, "    %s\n", shared.Dim.Render(opt.description))
		} else {
			fmt.Fprintf(&b, "    %s\n", shared.Dim.Render(opt.label))
			fmt.Fprintln(&b)
		}
	}
	if (m.width > 0 && m.width < minWidth) || (m.height > 0 && m.height < minHeight) {
		fmt.Fprintln(&b)
		if m.width > 0 && m.width < minWidth {
			fmt.Fprintf(&b, "  %s Window is too narrow (%d columns). Widen to at least %d for best results.\n",
				shared.Yellow.Render("warning:"), m.width, minWidth)
		}
		if m.height > 0 && m.height < minHeight {
			fmt.Fprintf(&b, "  %s Window is too short (%d rows). Increase to at least %d for best results.\n",
				shared.Yellow.Render("warning:"), m.height, minHeight)
		}
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, shared.Dim.Render("  ↑/↓ to navigate, Enter to select, Esc to quit"))
	return b.String()
}
