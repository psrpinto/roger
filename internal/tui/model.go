package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
)

type appState int

const (
	stateFirstRun appState = iota
	stateScanning
	statePreview
	stateGenerating
	stateDone
)

// Messages (inter-phase contracts)

type scanProgressMsg struct {
	done, total int
}

type scanDoneMsg struct {
	packs            []kit.Pack
	emptyPacks       []string
	wrongSampleCount []string
}

type genProgressMsg struct {
	done, total int
}

type genDoneMsg struct {
	kitCount    int
	sampleCount int
	totalSize   int64
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string { return e.err.Error() }

// phaseTransition signals from sub-models to the parent.
type phaseTransition int

const (
	phaseStay phaseTransition = iota
	phaseNext
	phaseAbort
)

type transition struct {
	phase phaseTransition
	data  any
}

// Model is the thin orchestrator.
type Model struct {
	state        appState
	cfg          *config.Config
	baseDir      string
	srcDir       string
	destDir      string
	topLevelDirs []string
	padStyles    [16]lipgloss.Style
	width        int
	height       int

	firstRun *firstRunModel
	scan     *scanModel
	preview  *previewModel
	gen      *genModel

	// final results (read after Tea exits)
	KitCount    int
	SampleCount int
	TotalSize   int64
	Aborted     bool
}

func NewModel(baseDir, srcDir, destDir string, topLevelDirs []string, isFirstRun bool, padStyles [16]lipgloss.Style, cfg *config.Config) Model {
	state := stateScanning
	if isFirstRun {
		state = stateFirstRun
	}

	m := Model{
		state:        state,
		cfg:          cfg,
		baseDir:      baseDir,
		srcDir:       srcDir,
		destDir:      destDir,
		topLevelDirs: topLevelDirs,
		padStyles:    padStyles,
	}

	switch state {
	case stateFirstRun:
		m.firstRun = newFirstRunModel(baseDir)
	case stateScanning:
		m.scan = newScanModel(topLevelDirs, srcDir, cfg)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case stateFirstRun:
		return nil
	case stateScanning:
		return m.scan.init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.preview != nil {
			m.preview.resize(msg.Width, msg.Height)
		}
		return m, nil
	case errMsg:
		m.state = stateDone
		return m, tea.Quit
	}

	var cmd tea.Cmd
	var tr transition

	switch m.state {
	case stateFirstRun:
		cmd, tr = m.firstRun.update(msg)
	case stateScanning:
		cmd, tr = m.scan.update(msg)
	case statePreview:
		cmd, tr = m.preview.update(msg)
	case stateGenerating:
		cmd, tr = m.gen.update(msg)
	}

	switch tr.phase {
	case phaseAbort:
		m.Aborted = true
		return m, tea.Quit
	case phaseNext:
		return m.advancePhase(tr.data)
	}

	return m, cmd
}

func (m Model) advancePhase(data any) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateFirstRun:
		m.state = stateScanning
		m.scan = newScanModel(m.topLevelDirs, m.srcDir, m.cfg)
		return m, m.scan.init()

	case stateScanning:
		d := data.(scanDoneMsg)
		if len(d.packs) == 0 {
			m.state = stateDone
			return m, tea.Quit
		}
		m.state = statePreview
		m.preview = newPreviewModel(d.packs, d.emptyPacks, d.wrongSampleCount, m.padStyles, m.cfg, m.width, m.height)
		return m, nil

	case statePreview:
		packs := data.([]kit.Pack)
		m.state = stateGenerating
		m.gen = newGenModel(packs, m.destDir, m.cfg.PadLayout)
		return m, m.gen.init()

	case stateGenerating:
		d := data.(genDoneMsg)
		m.KitCount = d.kitCount
		m.SampleCount = d.sampleCount
		m.TotalSize = d.totalSize
		m.state = stateDone
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) View() tea.View {
	var s string
	switch m.state {
	case stateFirstRun:
		s = m.firstRun.view()
	case stateScanning:
		s = m.scan.view()
	case statePreview:
		s = m.preview.view()
	case stateGenerating:
		s = m.gen.view()
	case stateDone:
		s = ""
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
