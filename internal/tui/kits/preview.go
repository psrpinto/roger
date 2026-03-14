package kits

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
	"roger/internal/tui/shared"
)

type PreviewModel struct {
	viewport viewport.Model
	footer   string
	packs    []kit.Pack
}

func NewPreviewModel(packs []kit.Pack, emptyPacks, wrongSampleCount []string, padStyles [16]lipgloss.Style, cfg *config.Config, width, height int) *PreviewModel {
	footer := "\n" + RenderLegend() + RenderWarnings(packs, emptyPacks, wrongSampleCount) + "\nGenerate output files? [Y/n] "
	footerHeight := strings.Count(footer, "\n") + 1

	content := RenderGrids(packs, padStyles, cfg.DrumTypes, cfg.PadLayout)
	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height-footerHeight),
	)
	vp.SetContent(content)

	return &PreviewModel{
		viewport: vp,
		footer:   footer,
		packs:    packs,
	}
}

func (m *PreviewModel) Resize(width, height int) {
	footerHeight := strings.Count(m.footer, "\n") + 1
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height - footerHeight)
}

func (m *PreviewModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	if kp, ok := msg.(tea.KeyPressMsg); ok {
		switch kp.String() {
		case "y", "Y", "enter":
			return nil, shared.Transition{Phase: shared.Next, Data: m.packs}
		case "esc":
			return nil, shared.Transition{Phase: shared.Back}
		case "n", "N", "q", "ctrl+c":
			return nil, shared.Transition{Phase: shared.Abort}
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return cmd, shared.Transition{}
}

func (m *PreviewModel) View() string {
	return m.viewport.View() + m.footer
}
