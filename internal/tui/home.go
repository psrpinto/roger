package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/tui/shared"
)

const modeExit Mode = "exit"

var modeOptions = []struct {
	mode Mode
	item shared.SelectItem
}{
	{ModeKits, shared.SelectItem{Label: "Kits", Description: "Create Drum programs"}},
	{ModeInstruments, shared.SelectItem{Label: "Instruments", Description: "Create Keygroup programs"}},
	{modeExit, shared.SelectItem{Label: "Quit", Description: "Quit roger"}},
}

const (
	minWidth  = 140
	minHeight = 40
)

type HomeModel struct {
	sel    *shared.SelectModel
	width  int
	height int
}

func NewHomeModel() *HomeModel {
	items := make([]shared.SelectItem, len(modeOptions))
	for i, opt := range modeOptions {
		items[i] = opt.item
	}
	return &HomeModel{sel: shared.NewSelectModel(items)}
}

func (m *HomeModel) Resize(w, h int) {
	m.width = w
	m.height = h
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	if kp, ok := msg.(tea.KeyPressMsg); ok && kp.String() == "ctrl+c" {
		return nil, shared.Transition{Phase: shared.Abort}
	}
	cmd, tr := m.sel.Update(msg)
	if tr.Phase == shared.Next {
		idx := tr.Data.(int)
		if modeOptions[idx].mode == modeExit {
			return nil, shared.Transition{Phase: shared.Abort}
		}
		return nil, shared.Transition{Phase: shared.Next, Data: modeOptions[idx].mode}
	}
	return cmd, tr
}

func (m *HomeModel) View() string {
	var b strings.Builder
	logoStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#c070d0"))
	logo := logoStyle.Render(
		"_ __ ___   __ _  ___ _ __\n" +
			"| '__/ _ \\ / _` |/ _ \\ '__|\n" +
			"| | | (_) | (_| |  __/ |\n" +
			"|_|  \\___/ \\__, |\\___|_|\n" +
			"           |___/")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "%s\n", logo)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "What would you like to create?")
	fmt.Fprintln(&b)
	fmt.Fprint(&b, m.sel.View())
	if (m.width > 0 && m.width < minWidth) || (m.height > 0 && m.height < minHeight) {
		fmt.Fprintln(&b)
		if m.width > 0 && m.width < minWidth {
			fmt.Fprintf(&b, "%s Window is too narrow (%d columns). Widen to at least %d for best results.\n",
				shared.Yellow.Render("warning:"), m.width, minWidth)
		}
		if m.height > 0 && m.height < minHeight {
			fmt.Fprintf(&b, "%s Window is too short (%d rows). Increase to at least %d for best results.\n",
				shared.Yellow.Render("warning:"), m.height, minHeight)
		}
	}
	hints := []string{
		"Use ↑/↓ or j/k to navigate, and Enter to select.",
		"Press Esc to go back to the previous screen.",
		"Scroll long screens with ↑/↓ or your scroll wheel.",
		"Press ? on any screen to open its help page.",
		"Press Ctrl-c at any time to quit.",
	}
	maxLen := 0
	for _, h := range hints {
		if l := len([]rune(h)); l > maxLen {
			maxLen = l
		}
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, shared.Dim.Render(strings.Repeat("─", maxLen)))
	fmt.Fprintln(&b)
	for _, h := range hints {
		fmt.Fprintln(&b, shared.Dim.Render(h))
	}
	return b.String()
}
