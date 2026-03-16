package shared

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// breadcrumbKeys implements help.KeyMap for the breadcrumb bar.
type breadcrumbKeys struct{}

func (breadcrumbKeys) ShortHelp() []key.Binding {
	return []key.Binding{KeyHelp, KeyBack, KeyQuit}
}

func (breadcrumbKeys) FullHelp() [][]key.Binding { return nil }

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
	h := NewHelpModel()
	hint := h.View(breadcrumbKeys{})
	// Calculate visible width of left side (sum of segment lengths + separators + indent)
	visibleLeft := 2 // indent
	for i, s := range segments {
		if i > 0 {
			visibleLeft += 3 // " › "
		}
		visibleLeft += len([]rune(s))
	}
	hintLen := lipgloss.Width(hint)
	gap := width - visibleLeft - hintLen
	if gap < 2 {
		gap = 2
	}
	line := left + strings.Repeat(" ", gap) + hint
	rule := Dim.Render(strings.Repeat("─", width))
	return line + "\n" + rule + "\n\n"
}
