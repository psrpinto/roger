package kits

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/sampler"
	"roger/internal/tui/shared"
)

type HomeModel struct {
	dirs     []string
	names    []string
	selected []bool
	cursor   int
}

func NewHomeModel(dirs []string) *HomeModel {
	names := make([]string, len(dirs))
	for i, d := range dirs {
		names[i] = sampler.FormatKitName(filepath.Base(d))
	}
	selected := make([]bool, len(dirs))
	for i := range selected {
		selected[i] = true
	}
	return &HomeModel{
		dirs:     dirs,
		names:    names,
		selected: selected,
	}
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
		if m.cursor < len(m.dirs)-1 {
			m.cursor++
		}
	case "space":
		m.selected[m.cursor] = !m.selected[m.cursor]
	case "enter":
		var chosen []string
		for i, sel := range m.selected {
			if sel {
				chosen = append(chosen, m.dirs[i])
			}
		}
		if len(chosen) == 0 {
			return nil, shared.Transition{}
		}
		return nil, shared.Transition{Phase: shared.Next, Data: chosen}
	case "esc", "ctrl+c":
		return nil, shared.Transition{Phase: shared.Back}
	}
	return nil, shared.Transition{}
}

func (m *HomeModel) View() string {
	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Select packs to generate:")
	fmt.Fprintln(&b)
	for i, name := range m.names {
		var check string
		if m.selected[i] {
			check = shared.Cyan.Render("[✓]")
		} else {
			check = shared.Dim.Render("[✗]")
		}
		if i == m.cursor {
			fmt.Fprintf(&b, "%s %s %s\n", shared.Cyan.Render("▸"), check, shared.Bold.Render(name))
		} else {
			fmt.Fprintf(&b, "  %s %s\n", check, shared.Dim.Render(name))
		}
	}
	selCount := 0
	for _, s := range m.selected {
		if s {
			selCount++
		}
	}
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s\n", shared.Dim.Render(fmt.Sprintf("%d of %d selected  ·  Enter to scan, Space to toggle", selCount, len(m.dirs))))
	return b.String()
}
