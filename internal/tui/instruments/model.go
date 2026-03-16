package instruments

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"roger/internal/sampler"
	"roger/internal/tui/shared"
)

type instState int

const (
	stateHome instState = iota
	stateEmpty
	stateHelp
)

// Model is the instruments orchestrator.
type Model struct {
	state instState

	baseDir      string
	instSrcDir   string
	packArgs     []string
	topLevelDirs []string

	home     *HomeModel
	empty    *EmptyModel
	help     *HelpModel
	helpPrev instState
}

func NewModel(baseDir, instSrcDir string, packArgs []string) *Model {
	m := &Model{
		baseDir:    baseDir,
		instSrcDir: instSrcDir,
		packArgs:   packArgs,
	}
	m.init()
	return m
}

func (m *Model) init() {
	if err := os.MkdirAll(m.instSrcDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", m.instSrcDir, err)
		os.Exit(1)
	}

	m.topLevelDirs = m.packArgs
	if len(m.topLevelDirs) == 0 {
		m.topLevelDirs = sampler.ListSubdirs(m.instSrcDir)
	}

	if len(m.packArgs) == 0 && len(m.topLevelDirs) == 0 {
		m.state = stateEmpty
		m.empty = NewEmptyModel(m.baseDir, m.instSrcDir)
	} else {
		m.state = stateHome
		m.home = NewHomeModel(m.topLevelDirs)
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) canShowHelp() bool {
	switch m.state {
	case stateHome, stateEmpty:
		return true
	}
	return false
}

// Breadcrumb returns the breadcrumb path segments for the current state.
func (m *Model) Breadcrumb() []string {
	switch m.state {
	case stateHome:
		return []string{"roger", "Instruments"}
	case stateEmpty:
		return []string{"roger", "Instruments", "No kits found"}
	case stateHelp:
		return []string{"roger", "Instruments", "Help"}
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	if m.state == stateHelp {
		cmd, tr := m.help.Update(msg)
		switch tr.Phase {
		case shared.Abort:
			return nil, shared.Transition{Phase: shared.Abort}
		case shared.Back:
			m.state = m.helpPrev
			m.help = nil
			return nil, shared.Transition{}
		}
		return cmd, shared.Transition{}
	}

	if kp, ok := msg.(tea.KeyPressMsg); ok && kp.String() == "?" && m.canShowHelp() {
		m.help = NewHelpModel(m.baseDir)
		m.helpPrev = m.state
		m.state = stateHelp
		return nil, shared.Transition{}
	}

	var cmd tea.Cmd
	var tr shared.Transition

	switch m.state {
	case stateHome:
		cmd, tr = m.home.Update(msg)
	case stateEmpty:
		cmd, tr = m.empty.Update(msg)
	}

	switch tr.Phase {
	case shared.Abort:
		return nil, shared.Transition{Phase: shared.Abort}
	case shared.Back:
		return m.retreatPhase()
	case shared.Next:
		return m.advancePhase(tr.Data)
	case shared.ShowHelp:
		m.help = NewHelpModel(m.baseDir)
		m.helpPrev = m.state
		m.state = stateHelp
		return nil, shared.Transition{}
	}

	return cmd, shared.Transition{}
}

func (m *Model) advancePhase(data any) (tea.Cmd, shared.Transition) {
	switch m.state {
	case stateEmpty:
		m.empty = nil
		m.state = stateHome
		m.home = NewHomeModel(m.topLevelDirs)
	}
	return nil, shared.Transition{}
}

func (m *Model) retreatPhase() (tea.Cmd, shared.Transition) {
	switch m.state {
	case stateEmpty:
		// Go to instruments home (not root home), matching original behavior.
		m.empty = nil
		m.state = stateHome
		m.home = NewHomeModel(m.topLevelDirs)
		return nil, shared.Transition{}
	case stateHome:
		return nil, shared.Transition{Phase: shared.Back}
	}

	return nil, shared.Transition{}
}

// View returns the content string for the current state (no breadcrumb, no padding).
func (m *Model) View() string {
	switch m.state {
	case stateHome:
		return m.home.View()
	case stateEmpty:
		return m.empty.View()
	case stateHelp:
		return m.help.View()
	}
	return ""
}
