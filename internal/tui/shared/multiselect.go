package shared

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type MultiSelectItem struct {
	Label string
	Value string
}

type MultiSelectModel struct {
	items    []MultiSelectItem
	selected []bool
	cursor   int
}

func NewMultiSelectModel(items []MultiSelectItem) *MultiSelectModel {
	selected := make([]bool, len(items))
	for i := range selected {
		selected[i] = true
	}
	return &MultiSelectModel{
		items:    items,
		selected: selected,
	}
}

func (m *MultiSelectModel) Update(msg tea.Msg) (tea.Cmd, Transition) {
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
	case "space":
		m.selected[m.cursor] = !m.selected[m.cursor]
	case "enter":
		var chosen []string
		for i, sel := range m.selected {
			if sel {
				chosen = append(chosen, m.items[i].Value)
			}
		}
		if len(chosen) == 0 {
			return nil, Transition{}
		}
		return nil, Transition{Phase: Next, Data: chosen}
	case "esc":
		return nil, Transition{Phase: Back}
	}
	return nil, Transition{}
}

func (m *MultiSelectModel) View() string {
	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Select packs to generate:")
	fmt.Fprintln(&b)
	for i, item := range m.items {
		var check string
		if m.selected[i] {
			check = Cyan.Render("[✓]")
		} else {
			check = Dim.Render("[✗]")
		}
		if i == m.cursor {
			fmt.Fprintf(&b, "%s %s %s\n", Cyan.Render("▸"), check, Bold.Render(item.Label))
		} else {
			fmt.Fprintf(&b, "  %s %s\n", check, Dim.Render(item.Label))
		}
	}
	selCount := 0
	for _, s := range m.selected {
		if s {
			selCount++
		}
	}
	fmt.Fprintln(&b)
	h := NewHelpModel()
	hint := h.ShortHelpView([]key.Binding{KeyConfirm, KeyToggle})
	fmt.Fprintf(&b, "  %s\n", Dim.Render(fmt.Sprintf("%d of %d selected", selCount, len(m.items)))+"  "+hint)
	return b.String()
}
