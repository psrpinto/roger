package tui

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
)

type previewModel struct {
	viewport viewport.Model
	footer   string
	packs    []kit.Pack
}

func newPreviewModel(packs []kit.Pack, emptyPacks, wrongSampleCount []string, padStyles [16]lipgloss.Style, cfg *config.Config, width, height int) *previewModel {
	footer := "\n" + renderLegend() + renderWarnings(packs, emptyPacks, wrongSampleCount) + "\nGenerate output files? [Y/n] "
	footerHeight := strings.Count(footer, "\n") + 1

	content := renderGrids(packs, padStyles, cfg.DrumTypes, cfg.PadLayout)
	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height-footerHeight),
	)
	vp.SetContent(content)

	return &previewModel{
		viewport: vp,
		footer:   footer,
		packs:    packs,
	}
}

func (m *previewModel) resize(width, height int) {
	footerHeight := strings.Count(m.footer, "\n") + 1
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height - footerHeight)
}

func (m *previewModel) update(msg tea.Msg) (tea.Cmd, transition) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, transition{}
	}
	switch kp.String() {
	case "y", "Y", "enter":
		return nil, transition{phase: phaseNext, data: m.packs}
	case "n", "N", "esc", "q", "ctrl+c":
		return nil, transition{phase: phaseAbort}
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return cmd, transition{}
	}
}

func (m *previewModel) view() string {
	return m.viewport.View() + m.footer
}
