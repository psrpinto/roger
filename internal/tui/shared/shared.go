package shared

import (
	"strings"

	"charm.land/lipgloss/v2"
)

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

// RenderBreadcrumb renders a breadcrumb bar from the given path segments,
// with a separator line underneath and a back hint on the right.
func RenderBreadcrumb(segments []string, width int) string {
	parts := make([]string, len(segments))
	for i, s := range segments {
		if i == len(segments)-1 {
			parts[i] = Bold.Render(s)
		} else {
			parts[i] = Dim.Render(s)
		}
	}
	sep := Dim.Render(" › ")
	left := "  " + strings.Join(parts, sep)
	hint := Dim.Render("Esc ← back  Ctrl-c quit")
	// Calculate visible width of left side (sum of segment lengths + separators + indent)
	visibleLeft := 2 // indent
	for i, s := range segments {
		if i > 0 {
			visibleLeft += 3 // " › "
		}
		visibleLeft += len([]rune(s))
	}
	hintLen := 23 // "Esc ← back  Ctrl-c quit"
	gap := width - visibleLeft - hintLen
	if gap < 2 {
		gap = 2
	}
	line := left + strings.Repeat(" ", gap) + hint
	rule := Dim.Render(strings.Repeat("─", width))
	return line + "\n" + rule + "\n\n"
}

// Shared styles
var (
	Bold   = lipgloss.NewStyle().Bold(true)
	Dim    = lipgloss.NewStyle().Faint(true)
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)
