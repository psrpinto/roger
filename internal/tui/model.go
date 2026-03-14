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
	stateHome appState = iota
	stateKitsFirstRun
	stateKitsScanning
	stateKitsPreview
	stateKitsGenerating
	stateInstrumentsFirstRun
	stateInstruments
	stateHelp
	stateDone
)

// subModel is implemented by all sub-models (help, first-run, etc.).
type subModel interface {
	Update(tea.Msg) (tea.Cmd, shared.Transition)
	View() string
}

const breadcrumbHeight = 3

// Model is the thin orchestrator.
type Model struct {
	state      appState
	mode       Mode
	cliMode    bool // true if mode was passed via CLI
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
	home          *HomeModel
	kitsFirstRun        *kits.FirstRunModel
	kitsScan            *kits.ScanModel
	kitsPreview         *kits.PreviewModel
	kitsGen             *kits.GenModel
	instrumentsFirstRun *instruments.FirstRunModel
	instrumentsModel    *instruments.Model

	// help
	help     subModel
	helpPrev appState

	// final results (read after Tea exits)
	KitCount    int
	SampleCount int
	TotalSize   int64
	Aborted     bool
}

func NewModel(baseDir, kitsSrcDir, instSrcDir, destDir string, cfg *config.Config, mode Mode, kitsSetupFn kits.SetupFunc, instrumentsSetupFn instruments.SetupFunc) *Model {
	m := &Model{
		mode:               mode,
		cliMode:            mode != "",
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
		m.state = stateHome
		m.home = NewHomeModel()
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

func (m *Model) Init() tea.Cmd {
	switch m.state {
	case stateKitsScanning:
		return m.kitsScan.Init()
	}
	return nil
}

// newHelpForMode returns the help model for the current mode.
func (m *Model) newHelpForMode() subModel {
	switch m.mode {
	case ModeKits:
		return kits.NewHelpModel(m.baseDir)
	case ModeInstruments:
		return instruments.NewHelpModel(m.baseDir)
	default:
		return NewHelpModel(m.baseDir)
	}
}

// canShowHelp returns true for interactive states where help makes sense.
func (m *Model) canShowHelp() bool {
	switch m.state {
	case stateHome, stateKitsFirstRun, stateKitsPreview, stateInstrumentsFirstRun, stateInstruments:
		return true
	}
	return false
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.home != nil {
			m.home.Resize(msg.Width, msg.Height)
		}
		if m.kitsPreview != nil {
			m.kitsPreview.Resize(msg.Width, msg.Height-breadcrumbHeight)
		}
		return m, nil
	case shared.ErrMsg:
		m.state = stateDone
		return m, tea.Quit
	}

	// Handle help state
	if m.state == stateHelp {
		cmd, tr := m.help.Update(msg)
		switch tr.Phase {
		case shared.Abort:
			m.Aborted = true
			return m, tea.Quit
		case shared.Back:
			m.state = m.helpPrev
			m.help = nil
			return m, nil
		}
		return m, cmd
	}

	// Intercept "?" for help in interactive states
	if kp, ok := msg.(tea.KeyPressMsg); ok && kp.String() == "?" && m.canShowHelp() {
		m.help = m.newHelpForMode()
		m.helpPrev = m.state
		m.state = stateHelp
		return m, nil
	}

	var cmd tea.Cmd
	var tr shared.Transition

	switch m.state {
	case stateHome:
		cmd, tr = m.home.Update(msg)
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
	case shared.Back:
		return m.retreatPhase()
	case shared.Next:
		return m.advancePhase(tr.Data)
	}

	return m, cmd
}

func (m *Model) advancePhase(data any) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateHome:
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
		m.kitsPreview = kits.NewPreviewModel(d.Packs, d.EmptyPacks, d.WrongSampleCount, m.padStyles, m.cfg, m.width, m.height-breadcrumbHeight)
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

func (m *Model) retreatPhase() (tea.Model, tea.Cmd) {
	if m.cliMode {
		m.Aborted = true
		return m, tea.Quit
	}

	switch m.state {
	case stateKitsFirstRun, stateKitsPreview:
		m.kitsFirstRun = nil
		m.kitsScan = nil
		m.kitsPreview = nil
		m.state = stateHome
		m.home = NewHomeModel()
		return m, nil
	case stateInstrumentsFirstRun, stateInstruments:
		m.instrumentsFirstRun = nil
		m.instrumentsModel = nil
		m.state = stateHome
		m.home = NewHomeModel()
		return m, nil
	}

	return m, nil
}

func (m *Model) breadcrumb() []string {
	switch m.state {
	case stateKitsFirstRun:
		return []string{"roger", "Kits", "Setup"}
	case stateKitsScanning:
		return []string{"roger", "Kits", "Scanning"}
	case stateKitsPreview:
		return []string{"roger", "Kits", "Preview"}
	case stateKitsGenerating:
		return []string{"roger", "Kits", "Generating"}
	case stateInstrumentsFirstRun:
		return []string{"roger", "Instruments", "Setup"}
	case stateInstruments:
		return []string{"roger", "Instruments"}
	case stateHelp:
		switch m.mode {
		case ModeKits:
			return []string{"roger", "Kits", "Help"}
		case ModeInstruments:
			return []string{"roger", "Instruments", "Help"}
		default:
			return []string{"roger", "Help"}
		}
	case stateHome:
		return []string{"roger"}
	default:
		return nil
	}
}

func (m *Model) View() tea.View {
	var s string
	switch m.state {
	case stateHome:
		s = m.home.View()
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
	case stateHelp:
		s = m.help.View()
	case stateDone:
		s = ""
	}

	padding := lipgloss.NewStyle().PaddingLeft(2)
	s = padding.Render(s)

	if segments := m.breadcrumb(); segments != nil {
		s = shared.RenderBreadcrumb(segments, m.width) + s
	}

	v := tea.NewView(s)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
