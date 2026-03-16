package shared

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// Common key bindings used across the TUI.
var (
	KeyNav = key.NewBinding(
		key.WithKeys("up", "down", "k", "j"),
		key.WithHelp("↑/↓", "navigate"),
	)
	KeySelect = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	)
	KeyBack = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	)
	KeyHelp = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)
	KeyQuit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	KeyScroll = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "scroll"),
	)
	KeyToggle = key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "toggle"),
	)
	KeyConfirm = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)
)

// NewHelpView returns a help.Model styled with dim text.
func NewHelpView() help.Model {
	m := help.New()
	dimKey := lipgloss.NewStyle().Faint(true)
	dimDesc := lipgloss.NewStyle().Faint(true)
	dimSep := lipgloss.NewStyle().Faint(true)
	m.Styles.ShortKey = dimKey
	m.Styles.ShortDesc = dimDesc
	m.Styles.ShortSeparator = dimSep
	m.Styles.FullKey = dimKey
	m.Styles.FullDesc = dimDesc
	m.Styles.FullSeparator = dimSep
	m.Styles.Ellipsis = dimSep
	return m
}
