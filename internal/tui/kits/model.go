package kits

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"roger/internal/config"
	"roger/internal/kit"
	"roger/internal/mpc"
	"roger/internal/sampler"
	"roger/internal/tui/shared"
)

type kitState int

const (
	stateHome kitState = iota
	stateFirstRun
	stateScanning
	statePreview
	stateGenerating
	stateHelp
)

// Result is the data returned to the root model after generation completes.
type Result struct {
	KitCount    int
	SampleCount int
	TotalSize   int64
}

// Model is the kits orchestrator.
type Model struct {
	state   kitState
	cliMode bool

	baseDir    string
	kitsSrcDir string
	destDir    string
	packArgs   []string
	cfg        *config.Config
	width      int
	height     int

	topLevelDirs []string
	padStyles    [16]lipgloss.Style

	home     *HomeModel
	firstRun *FirstRunModel
	scan     *ScanModel
	preview  *PreviewModel
	gen      *GenModel
	help     *HelpModel
	helpPrev kitState
}

func NewModel(baseDir, kitsSrcDir, destDir string, packArgs []string, cfg *config.Config, cliMode bool) *Model {
	m := &Model{
		baseDir:    baseDir,
		kitsSrcDir: kitsSrcDir,
		destDir:    destDir,
		packArgs:   packArgs,
		cfg:        cfg,
		cliMode:    cliMode,
	}
	m.init()
	return m
}

func (m *Model) init() {
	if err := os.MkdirAll(m.kitsSrcDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: creating directory %s: %s\n", m.kitsSrcDir, err)
		os.Exit(1)
	}

	templatePath := filepath.Join(m.baseDir, "kit.xpm")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		os.WriteFile(templatePath, mpc.ProgramTemplate, 0o644)
	}
	expansionPath := filepath.Join(m.baseDir, "expansion.xml")
	if _, err := os.Stat(expansionPath); os.IsNotExist(err) {
		os.WriteFile(expansionPath, mpc.ExpansionTemplate, 0o644)
	}
	mpc.LoadCustomTemplate(m.baseDir)
	mpc.LoadCustomExpansionTemplate(m.baseDir)

	m.topLevelDirs = m.packArgs
	if len(m.topLevelDirs) == 0 {
		m.topLevelDirs = sampler.ListSubdirs(m.kitsSrcDir)
	}
	m.padStyles = mpc.ExtractPadStyles()

	if len(m.packArgs) == 0 && len(m.topLevelDirs) == 0 {
		m.state = stateFirstRun
		m.firstRun = NewFirstRunModel(m.baseDir, m.kitsSrcDir)
	} else {
		m.state = stateHome
		m.home = NewHomeModel()
	}
}

func (m *Model) Init() tea.Cmd {
	if m.state == stateScanning && m.scan != nil {
		return m.scan.Init()
	}
	return nil
}

func (m *Model) canShowHelp() bool {
	switch m.state {
	case stateHome, stateFirstRun, statePreview:
		return true
	}
	return false
}

// Breadcrumb returns the breadcrumb path segments for the current state.
func (m *Model) Breadcrumb() []string {
	switch m.state {
	case stateHome:
		return []string{"roger", "Kits"}
	case stateFirstRun:
		return []string{"roger", "Kits", "Setup"}
	case stateScanning:
		return []string{"roger", "Kits", "Scanning"}
	case statePreview:
		return []string{"roger", "Kits", "Preview"}
	case stateGenerating:
		return []string{"roger", "Kits", "Generating"}
	case stateHelp:
		return []string{"roger", "Kits", "Help"}
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		if m.state == statePreview && m.preview != nil {
			m.preview.Resize(ws.Width, ws.Height-shared.BreadcrumbHeight)
		}
		return nil, shared.Transition{}
	}

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
	case stateFirstRun:
		cmd, tr = m.firstRun.Update(msg)
	case stateScanning:
		cmd, tr = m.scan.Update(msg)
	case statePreview:
		cmd, tr = m.preview.Update(msg)
	case stateGenerating:
		cmd, tr = m.gen.Update(msg)
	}

	switch tr.Phase {
	case shared.Abort:
		return nil, shared.Transition{Phase: shared.Abort}
	case shared.Back:
		return m.retreatPhase()
	case shared.Next:
		return m.advancePhase(tr.Data)
	}

	return cmd, shared.Transition{}
}

func (m *Model) advancePhase(data any) (tea.Cmd, shared.Transition) {
	switch m.state {
	case stateFirstRun:
		m.firstRun = nil
		// Re-scan in case firstRun created new directories.
		if len(m.packArgs) == 0 {
			m.topLevelDirs = sampler.ListSubdirs(m.kitsSrcDir)
		}
		m.state = stateHome
		m.home = NewHomeModel()
		return nil, shared.Transition{}

	case stateHome:
		m.home = nil
		m.state = stateScanning
		m.scan = NewScanModel(m.topLevelDirs, m.kitsSrcDir, m.cfg)
		return m.scan.Init(), shared.Transition{}

	case stateScanning:
		d := data.(ScanDoneMsg)
		if len(d.Packs) == 0 {
			return nil, shared.Transition{Phase: shared.Next, Data: Result{}}
		}
		m.scan = nil
		m.state = statePreview
		m.preview = NewPreviewModel(d.Packs, d.EmptyPacks, d.WrongSampleCount, m.padStyles, m.cfg, m.width, m.height-shared.BreadcrumbHeight)
		return nil, shared.Transition{}

	case statePreview:
		packs := data.([]kit.Pack)
		m.preview = nil
		m.state = stateGenerating
		m.gen = NewGenModel(packs, m.destDir, m.cfg.PadLayout)
		return m.gen.Init(), shared.Transition{}

	case stateGenerating:
		d := data.(GenDoneMsg)
		return nil, shared.Transition{Phase: shared.Next, Data: Result{
			KitCount:    d.KitCount,
			SampleCount: d.SampleCount,
			TotalSize:   d.TotalSize,
		}}
	}

	return nil, shared.Transition{}
}

func (m *Model) retreatPhase() (tea.Cmd, shared.Transition) {
	if m.cliMode {
		return nil, shared.Transition{Phase: shared.Abort}
	}

	switch m.state {
	case stateHome, stateFirstRun:
		return nil, shared.Transition{Phase: shared.Back}
	case statePreview:
		m.preview = nil
		m.state = stateHome
		m.home = NewHomeModel()
		return nil, shared.Transition{}
	}

	return nil, shared.Transition{}
}

// View returns the content string for the current state (no breadcrumb, no padding).
func (m *Model) View() string {
	switch m.state {
	case stateHome:
		return m.home.View()
	case stateFirstRun:
		return m.firstRun.View()
	case stateScanning:
		return m.scan.View()
	case statePreview:
		return m.preview.View()
	case stateGenerating:
		return m.gen.View()
	case stateHelp:
		return m.help.View()
	}
	return ""
}
