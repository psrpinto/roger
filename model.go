package main

import (
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type appState int

const (
	stateFirstRun appState = iota
	stateScanning
	statePreview
	stateGenerating
	stateDone
)

// Messages

type scanProgressMsg struct {
	done, total int
}

type scanDoneMsg struct {
	packs            []pack
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

// Model

type model struct {
	state appState

	// config
	baseDir, srcDir, destDir string
	topLevelDirs             []string
	padStyles                [16]lipgloss.Style

	// components
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
	footer        string

	// scanning
	scanCh       chan scanProgressMsg
	scanProgress int
	scanTotal    int

	// scan results
	packs            []pack
	emptyPacks       []string
	wrongSampleCount []string

	// generation
	genCh       chan genProgressMsg
	genProgress int
	genTotal    int

	// final result (read after Tea exits)
	kitCount    int
	sampleCount int
	totalSize   int64
	aborted     bool

	// terminal size
	width, height int
}

func newModel(baseDir, srcDir, destDir string, topLevelDirs []string, isFirstRun bool, padStyles [16]lipgloss.Style) model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("6"))),
	)

	state := stateScanning
	if isFirstRun {
		state = stateFirstRun
	}

	m := model{
		state:        state,
		baseDir:      baseDir,
		srcDir:       srcDir,
		destDir:      destDir,
		topLevelDirs: topLevelDirs,
		padStyles:    padStyles,
		spinner:      s,
	}

	if state == stateScanning {
		m.scanCh = make(chan scanProgressMsg, 1)
	}

	return m
}

func (m model) Init() tea.Cmd {
	switch m.state {
	case stateFirstRun:
		return nil
	case stateScanning:
		return tea.Batch(
			m.spinner.Tick,
			scanPacksCmd(m.scanCh, m.topLevelDirs, m.srcDir),
			waitForScanProgress(m.scanCh),
		)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.viewportReady {
			footerHeight := strings.Count(m.footer, "\n") + 1
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - footerHeight)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch m.state {
		case stateFirstRun:
			switch msg.String() {
			case "y", "Y", "enter":
				m.state = stateScanning
				m.scanCh = make(chan scanProgressMsg, 1)
				return m, tea.Batch(
					m.spinner.Tick,
					scanPacksCmd(m.scanCh, m.topLevelDirs, m.srcDir),
					waitForScanProgress(m.scanCh),
				)
			case "n", "N", "esc", "q":
				m.aborted = true
				return m, tea.Quit
			case "ctrl+c":
				m.aborted = true
				return m, tea.Quit
			}

		case statePreview:
			switch msg.String() {
			case "y", "Y", "enter":
				m.state = stateGenerating
				m.genCh = make(chan genProgressMsg, 1)
				return m, tea.Batch(
					m.spinner.Tick,
					generatePacksCmd(m.genCh, m.packs, m.destDir),
					waitForGenProgress(m.genCh),
				)
			case "n", "N", "esc", "q":
				m.aborted = true
				return m, tea.Quit
			case "ctrl+c":
				m.aborted = true
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}

		case stateScanning, stateGenerating:
			if msg.String() == "ctrl+c" {
				m.aborted = true
				return m, tea.Quit
			}
		}

	case scanProgressMsg:
		m.scanProgress = msg.done
		m.scanTotal = msg.total
		return m, waitForScanProgress(m.scanCh)

	case scanDoneMsg:
		m.packs = msg.packs
		m.emptyPacks = msg.emptyPacks
		m.wrongSampleCount = msg.wrongSampleCount

		if len(m.packs) == 0 {
			m.state = stateDone
			return m, tea.Quit
		}

		m.footer = "\n" + renderLegend() + renderWarnings(m.packs, m.emptyPacks, m.wrongSampleCount) + "\nGenerate output files? [Y/n] "
		footerHeight := strings.Count(m.footer, "\n") + 1

		content := renderGrids(m.packs, m.padStyles)
		vp := viewport.New(
			viewport.WithWidth(m.width),
			viewport.WithHeight(m.height-footerHeight),
		)
		vp.SetContent(content)
		m.viewport = vp
		m.viewportReady = true

		m.state = statePreview
		return m, nil

	case genProgressMsg:
		m.genProgress = msg.done
		m.genTotal = msg.total
		return m, waitForGenProgress(m.genCh)

	case genDoneMsg:
		m.kitCount = msg.kitCount
		m.sampleCount = msg.sampleCount
		m.totalSize = msg.totalSize
		m.state = stateDone
		return m, tea.Quit

	case errMsg:
		m.state = stateDone
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	var s string
	switch m.state {
	case stateFirstRun:
		s = renderUsage(m.baseDir) + "\n" +
			"Example directories have been created in Input/ to demonstrate the structure.\n" +
			"Preview example kits? [Y/n] "
	case stateScanning:
		s = m.spinner.View() + scanningStatus(m.scanProgress, m.scanTotal)
	case statePreview:
		s = m.viewport.View() + m.footer
	case stateGenerating:
		s = m.spinner.View() + generatingStatus(m.genProgress, m.genTotal)
	case stateDone:
		s = ""
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
