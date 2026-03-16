package instruments

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"roger/internal/examples"
	"roger/internal/tui/shared"
)

type EmptyModel struct {
	sel     *shared.SelectModel
	baseDir string
	srcDir  string
}

func NewEmptyModel(baseDir, srcDir string) *EmptyModel {
	items := []shared.SelectItem{
		{Label: "Generate example files", Description: "Create example instrument directories in Instruments/"},
		{Label: "Show instructions", Description: "Open the help screen"},
	}
	return &EmptyModel{
		sel:     shared.NewSelectModel(items),
		baseDir: baseDir,
		srcDir:  srcDir,
	}
}

func (m *EmptyModel) Update(msg tea.Msg) (tea.Cmd, shared.Transition) {
	cmd, tr := m.sel.Update(msg)
	if tr.Phase == shared.Next {
		idx := tr.Data.(int)
		if idx == 0 {
			examples.CreateExampleInstrumentDirs(m.srcDir)
			return nil, shared.Transition{Phase: shared.Next}
		}
		return nil, shared.Transition{Phase: shared.ShowHelp}
	}
	return cmd, tr
}

func (m *EmptyModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Workspace: %s\n\n", shared.Cyan.Render(m.baseDir))
	fmt.Fprintln(&b, "No instruments found. What would you like to do?")
	fmt.Fprintln(&b)
	fmt.Fprint(&b, m.sel.View())
	return b.String()
}
