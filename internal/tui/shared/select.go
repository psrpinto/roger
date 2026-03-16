package shared

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type SelectItem struct {
	Label       string
	Description string
}

type SelectModel struct {
	items  []SelectItem
	cursor int
}

func NewSelectModel(items []SelectItem) *SelectModel {
	return &SelectModel{items: items}
}

func (m *SelectModel) Update(msg tea.Msg) (tea.Cmd, Transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, Transition{}
	}
	switch kp.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter":
		return nil, Transition{Phase: Next, Data: m.cursor}
	case "esc":
		return nil, Transition{Phase: Back}
	}
	return nil, Transition{}
}

func (m *SelectModel) View() string {
	var b strings.Builder
	for i, item := range m.items {
		if i > 0 {
			fmt.Fprintln(&b)
		}
		if i == m.cursor {
			fmt.Fprintf(&b, "%s %s\n", Cyan.Render("▸"), Bold.Render(item.Label))
			if item.Description != "" {
				fmt.Fprintf(&b, "  %s\n", Dim.Render(item.Description))
			}
		} else {
			fmt.Fprintf(&b, "  %s\n", Dim.Render(item.Label))
		}
	}
	return b.String()
}
