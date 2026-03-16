package instruments

import (
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"roger/internal/sampler"
	"roger/internal/tui/shared"
)

type HomeModel struct {
	multi *shared.MultiSelectModel
}

func NewHomeModel(dirs []string) *HomeModel {
	items := make([]shared.MultiSelectItem, len(dirs))
	for i, d := range dirs {
		items[i] = shared.MultiSelectItem{
			Label: sampler.FormatKitName(filepath.Base(d)),
			Value: d,
		}
	}
	return &HomeModel{multi: shared.NewMultiSelectModel(items)}
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	return m.multi.Update(msg)
}

func (m *HomeModel) View() string {
	return m.multi.View()
}
