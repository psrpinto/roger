package instruments

import (
	tea "charm.land/bubbletea/v2"

	"roger/internal/tui/shared"
)

// Setup holds the results of instruments-specific initialization.
type Setup struct {
	TopLevelDirs []string
	IsFirstRun   bool
}

// SetupFunc performs instruments-specific initialization
// (directory scanning, example creation) and returns the results.
type SetupFunc func() Setup

type Model struct{}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return nil, shared.Transition{Phase: shared.Next}
	}
	return nil, shared.Transition{}
}

func (m *Model) View() string {
	return shared.Bold.Render("Instruments") + " — Coming soon.\n\nPress any key to exit."
}
