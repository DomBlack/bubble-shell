package shell

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode is an interface which represents an input mode the shell can be in
// and is used to handle input
type Mode interface {
	// Enter is called when the mode is entered
	Enter(m Model) (Model, tea.Cmd)

	// Leave is called when the mode is left, before the next mode is entered
	Leave(m Model) (Model, tea.Cmd)

	// Update is called when a message is received for the mode
	Update(m Model, msg tea.Msg) (Model, tea.Cmd)

	// AdditionalView is called to render additional output
	// at the bottom of the screen
	AdditionalView(m Model) string

	// ShortHelp returns a list of key bindings and their descriptions
	ShortHelp(m Model, keyMap KeyMap) []key.Binding

	// FullHelp returns a list of key bindings and their descriptions
	FullHelp(m Model, keyMap KeyMap) [][]key.Binding
}

// Enter tells the mode to enter the given mode leaving the old mode
func (m Model) Enter(mode Mode) tea.Cmd {
	return func() tea.Msg {
		return enterModeMsg{
			ID:   m.id,
			Mode: mode,
		}
	}
}
