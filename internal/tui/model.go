package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
	"roger/internal/tui/instruments"
	"roger/internal/tui/kits"
	"roger/internal/tui/shared"
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
	stateKitsFirstRun
	stateKitsScanning
	stateKitsPreview
	stateKitsGenerating
	stateInstrumentsFirstRun
	stateInstruments
	stateDone
)

// Model is the thin orchestrator.
type Model struct {
	state      appState
	mode       Mode
	cfg        *config.Config
	baseDir    string
	kitsSrcDir string
	instSrcDir string
	destDir    string
	width      int
	height     int

	// kits-specific state
	kitsSetupFn  kits.SetupFunc
	topLevelDirs []string
	padStyles    [16]lipgloss.Style

	// instruments-specific state
	instrumentsSetupFn instruments.SetupFunc

	// sub-models
	modeSelect          *modeSelectModel
	kitsFirstRun        *kits.FirstRunModel
	kitsScan            *kits.ScanModel
	kitsPreview         *kits.PreviewModel
	kitsGen             *kits.GenModel
	instrumentsFirstRun *instruments.FirstRunModel
	instrumentsModel    *instruments.Model

	// final results (read after Tea exits)
	KitCount    int
	SampleCount int
	TotalSize   int64
	Aborted     bool
}

func NewModel(baseDir, kitsSrcDir, instSrcDir, destDir string, cfg *config.Config, mode Mode, kitsSetupFn kits.SetupFunc, instrumentsSetupFn instruments.SetupFunc) Model {
	m := Model{
		mode:               mode,
		cfg:                cfg,
		baseDir:            baseDir,
		kitsSrcDir:         kitsSrcDir,
		instSrcDir:         instSrcDir,
		destDir:            destDir,
		kitsSetupFn:        kitsSetupFn,
		instrumentsSetupFn: instrumentsSetupFn,
	}

	switch mode {
	case ModeKits:
		m.initKits()
	case ModeInstruments:
		m.initInstruments()
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

	if ks.IsFirstRun {
		m.state = stateKitsFirstRun
		m.kitsFirstRun = kits.NewFirstRunModel(m.baseDir)
	} else {
		m.state = stateKitsScanning
		m.kitsScan = kits.NewScanModel(m.topLevelDirs, m.kitsSrcDir, m.cfg)
	}
}

func (m *Model) initInstruments() {
	is := m.instrumentsSetupFn()
	if is.IsFirstRun {
		m.state = stateInstrumentsFirstRun
		m.instrumentsFirstRun = instruments.NewFirstRunModel(m.baseDir)
	} else {
		m.state = stateInstruments
		m.instrumentsModel = instruments.NewModel()
	}
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case stateKitsScanning:
		return m.kitsScan.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.kitsPreview != nil {
			m.kitsPreview.Resize(msg.Width, msg.Height)
		}
		return m, nil
	case shared.ErrMsg:
		m.state = stateDone
		return m, tea.Quit
	}

	var cmd tea.Cmd
	var tr shared.Transition

	switch m.state {
	case stateModeSelect:
		cmd, tr = m.modeSelect.update(msg)
	case stateKitsFirstRun:
		cmd, tr = m.kitsFirstRun.Update(msg)
	case stateKitsScanning:
		cmd, tr = m.kitsScan.Update(msg)
	case stateKitsPreview:
		cmd, tr = m.kitsPreview.Update(msg)
	case stateKitsGenerating:
		cmd, tr = m.kitsGen.Update(msg)
	case stateInstrumentsFirstRun:
		cmd, tr = m.instrumentsFirstRun.Update(msg)
	case stateInstruments:
		cmd, tr = m.instrumentsModel.Update(msg)
	}

	switch tr.Phase {
	case shared.Abort:
		m.Aborted = true
		return m, tea.Quit
	case shared.Next:
		return m.advancePhase(tr.Data)
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
			if m.state == stateKitsScanning {
				return m, m.kitsScan.Init()
			}
			return m, nil
		case ModeInstruments:
			m.initInstruments()
			return m, nil
		}
		return m, nil

	case stateKitsFirstRun:
		m.state = stateKitsScanning
		m.kitsScan = kits.NewScanModel(m.topLevelDirs, m.kitsSrcDir, m.cfg)
		return m, m.kitsScan.Init()

	case stateKitsScanning:
		d := data.(kits.ScanDoneMsg)
		if len(d.Packs) == 0 {
			m.state = stateDone
			return m, tea.Quit
		}
		m.state = stateKitsPreview
		m.kitsPreview = kits.NewPreviewModel(d.Packs, d.EmptyPacks, d.WrongSampleCount, m.padStyles, m.cfg, m.width, m.height)
		return m, nil

	case stateKitsPreview:
		packs := data.([]kit.Pack)
		m.state = stateKitsGenerating
		m.kitsGen = kits.NewGenModel(packs, m.destDir, m.cfg.PadLayout)
		return m, m.kitsGen.Init()

	case stateKitsGenerating:
		d := data.(kits.GenDoneMsg)
		m.KitCount = d.KitCount
		m.SampleCount = d.SampleCount
		m.TotalSize = d.TotalSize
		m.state = stateDone
		return m, tea.Quit

	case stateInstrumentsFirstRun:
		m.state = stateInstruments
		m.instrumentsModel = instruments.NewModel()
		return m, nil

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
	case stateKitsFirstRun:
		s = m.kitsFirstRun.View()
	case stateKitsScanning:
		s = m.kitsScan.View()
	case stateKitsPreview:
		s = m.kitsPreview.View()
	case stateKitsGenerating:
		s = m.kitsGen.View()
	case stateInstrumentsFirstRun:
		s = m.instrumentsFirstRun.View()
	case stateInstruments:
		s = m.instrumentsModel.View()
	case stateDone:
		s = ""
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
