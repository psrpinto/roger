package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
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
	stateModeActive
	stateHelp
	stateDone
)

// modeModel is implemented by per-mode orchestrators.
type modeModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Cmd, shared.Transition)
	View() string
	Breadcrumb() []string
}

// subModel is implemented by sub-models used directly by the root (e.g. help).
type subModel interface {
	Update(tea.Msg) (tea.Cmd, shared.Transition)
	View() string
}

// Model is the thin root orchestrator.
type Model struct {
	state   appState
	mode    Mode
	cliMode bool
	cfg     *config.Config
	baseDir    string
	kitsSrcDir string
	instSrcDir string
	destDir    string
	width      int
	height     int

	home            *HomeModel
	activeModeModel modeModel

	// help is used only for the root home screen help.
	help     subModel
	helpPrev appState

	// final results (read after Tea exits)
	KitCount    int
	SampleCount int
	TotalSize   int64
	Aborted     bool
}

func NewModel(baseDir, kitsSrcDir, instSrcDir, destDir string, cfg *config.Config, mode Mode, packArgs []string) *Model {
	m := &Model{
		mode:       mode,
		cliMode:    mode != "",
		cfg:        cfg,
		baseDir:    baseDir,
		kitsSrcDir: kitsSrcDir,
		instSrcDir: instSrcDir,
		destDir:    destDir,
	}

	switch mode {
	case ModeKits:
		m.initKits(packArgs)
	case ModeInstruments:
		m.initInstruments(packArgs)
	default:
		m.state = stateHome
		m.home = NewHomeModel()
	}

	return m
}

func (m *Model) initKits(packArgs []string) {
	m.activeModeModel = kits.NewModel(m.baseDir, m.kitsSrcDir, m.destDir, packArgs, m.cfg, m.cliMode)
	m.state = stateModeActive
}

func (m *Model) initInstruments(packArgs []string) {
	m.activeModeModel = instruments.NewModel(m.baseDir, m.instSrcDir, packArgs, m.cliMode)
	m.state = stateModeActive
}

func (m *Model) Init() tea.Cmd {
	if m.state == stateModeActive && m.activeModeModel != nil {
		return m.activeModeModel.Init()
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		if m.home != nil {
			m.home.Resize(ws.Width, ws.Height)
		}
		if m.activeModeModel != nil {
			m.activeModeModel.Update(ws)
		}
		return m, nil
	}

	if _, ok := msg.(shared.ErrMsg); ok {
		m.state = stateDone
		return m, tea.Quit
	}

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

	if m.state == stateHome {
		if kp, ok := msg.(tea.KeyPressMsg); ok && kp.String() == "?" {
			m.help = NewHelpModel(m.baseDir)
			m.helpPrev = stateHome
			m.state = stateHelp
			return m, nil
		}

		cmd, tr := m.home.Update(msg)
		switch tr.Phase {
		case shared.Abort:
			m.Aborted = true
			return m, tea.Quit
		case shared.Next:
			selectedMode := tr.Data.(Mode)
			m.mode = selectedMode
			switch selectedMode {
			case ModeKits:
				m.initKits(nil)
			case ModeInstruments:
				m.initInstruments(nil)
			}
			return m, m.activeModeModel.Init()
		}
		return m, cmd
	}

	if m.state == stateModeActive {
		cmd, tr := m.activeModeModel.Update(msg)
		switch tr.Phase {
		case shared.Abort:
			m.Aborted = true
			return m, tea.Quit
		case shared.Back:
			m.activeModeModel = nil
			m.state = stateHome
			m.home = NewHomeModel()
			return m, nil
		case shared.Next:
			if r, ok := tr.Data.(kits.Result); ok {
				m.KitCount = r.KitCount
				m.SampleCount = r.SampleCount
				m.TotalSize = r.TotalSize
			}
			m.state = stateDone
			return m, tea.Quit
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) breadcrumb() []string {
	switch m.state {
	case stateHome:
		return []string{"roger"}
	case stateModeActive:
		return m.activeModeModel.Breadcrumb()
	case stateHelp:
		return []string{"roger", "Help"}
	}
	return nil
}

func (m *Model) View() tea.View {
	var s string
	switch m.state {
	case stateHome:
		s = m.home.View()
	case stateModeActive:
		s = m.activeModeModel.View()
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
