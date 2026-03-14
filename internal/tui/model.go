package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
)

// Mode represents the top-level operating mode.
type Mode string

const (
	ModeKits        Mode = "kits"
	ModeInstruments Mode = "instruments"
)

type appState int

const (
	stateModeSelect appState = iota
	stateFirstRun
	stateScanning
	statePreview
	stateGenerating
	stateInstruments
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

// KitsSetup holds the results of kits-specific initialization.
type KitsSetup struct {
	TopLevelDirs []string
	PadStyles    [16]lipgloss.Style
	IsFirstRun   bool
}

// KitsSetupFunc performs kits-specific initialization (template loading,
// directory scanning, example creation) and returns the results.
type KitsSetupFunc func() KitsSetup

// Model is the thin orchestrator.
type Model struct {
	state    appState
	mode     Mode
	cfg      *config.Config
	baseDir  string
	srcDir   string
	destDir  string
	width    int
	height   int

	// kits-specific state
	kitsSetupFn  KitsSetupFunc
	topLevelDirs []string
	padStyles    [16]lipgloss.Style
	isFirstRun   bool

	modeSelect  *modeSelectModel
	firstRun    *firstRunModel
	scan        *scanModel
	preview     *previewModel
	gen         *genModel
	instruments *instrumentsModel

	// final results (read after Tea exits)
	KitCount    int
	SampleCount int
	TotalSize   int64
	Aborted     bool
}

func NewModel(baseDir, srcDir, destDir string, cfg *config.Config, mode Mode, kitsSetupFn KitsSetupFunc) Model {
	m := Model{
		mode:        mode,
		cfg:         cfg,
		baseDir:     baseDir,
		srcDir:      srcDir,
		destDir:     destDir,
		kitsSetupFn: kitsSetupFn,
	}

	switch mode {
	case ModeKits:
		m.initKits()
	case ModeInstruments:
		m.state = stateInstruments
		m.instruments = newInstrumentsModel()
	default:
		m.state = stateModeSelect
		m.modeSelect = newModeSelectModel()
	}

	return m
}

func (m *Model) initKits() {
	ks := m.kitsSetupFn()
	m.topLevelDirs = ks.TopLevelDirs
	m.padStyles = ks.PadStyles
	m.isFirstRun = ks.IsFirstRun

	if m.isFirstRun {
		m.state = stateFirstRun
		m.firstRun = newFirstRunModel(m.baseDir)
	} else {
		m.state = stateScanning
		m.scan = newScanModel(m.topLevelDirs, m.srcDir, m.cfg)
	}
}

func (m Model) Init() tea.Cmd {
	switch m.state {
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
	case stateModeSelect:
		cmd, tr = m.modeSelect.update(msg)
	case stateFirstRun:
		cmd, tr = m.firstRun.update(msg)
	case stateScanning:
		cmd, tr = m.scan.update(msg)
	case statePreview:
		cmd, tr = m.preview.update(msg)
	case stateGenerating:
		cmd, tr = m.gen.update(msg)
	case stateInstruments:
		cmd, tr = m.instruments.update(msg)
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
	case stateModeSelect:
		m.mode = data.(Mode)
		switch m.mode {
		case ModeKits:
			m.initKits()
			if m.state == stateScanning {
				return m, m.scan.init()
			}
			return m, nil
		case ModeInstruments:
			m.state = stateInstruments
			m.instruments = newInstrumentsModel()
			return m, nil
		}
		return m, nil

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

	case stateInstruments:
		m.state = stateDone
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) View() tea.View {
	var s string
	switch m.state {
	case stateModeSelect:
		s = m.modeSelect.view()
	case stateFirstRun:
		s = m.firstRun.view()
	case stateScanning:
		s = m.scan.view()
	case statePreview:
		s = m.preview.view()
	case stateGenerating:
		s = m.gen.view()
	case stateInstruments:
		s = m.instruments.view()
	case stateDone:
		s = ""
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
