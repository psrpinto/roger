package shared

import "charm.land/lipgloss/v2"

// Phase transition types used by sub-models to signal the orchestrator.
type Phase int

const (
	Stay Phase = iota
	Next
	Back
	Abort
)

type Transition struct {
	Phase Phase
	Data  any
}

// ErrMsg is sent through tea.Msg when a sub-model encounters an error.
type ErrMsg struct {
	Err error
}

func (e ErrMsg) Error() string { return e.Err.Error() }

// Shared styles
var (
	Bold   = lipgloss.NewStyle().Bold(true)
	Dim    = lipgloss.NewStyle().Faint(true)
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)
