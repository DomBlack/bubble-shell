package shell

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type CommandRunningMode struct{ KeepInputContent bool }

var _ Mode = (*CommandRunningMode)(nil)

func (c *CommandRunningMode) Enter(m Model) (Model, tea.Cmd) {
	return m, nil
}

func (c *CommandRunningMode) Leave(m Model) (Model, tea.Cmd) {
	return m, nil
}

func (c *CommandRunningMode) Update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.cfg.KeyMap.Cancel):
			if m.currentCmdCancel != nil {
				m.currentCmdCancel()
				m.currentCmdCancel = nil
				return m, nil
			} else {
				return m, m.Shutdown
			}
		}
	}

	return m, nil
}

func (c *CommandRunningMode) AdditionalView(m Model) string {
	return ""
}

func (c *CommandRunningMode) ShortHelp(m Model, keyMap KeyMap) []key.Binding {
	cancel := keyMap.Cancel

	if m.currentCmdCancel != nil {
		cancel.SetHelp(cancel.Help().Key, "stop executing command")
	} else {
		cancel.SetHelp(cancel.Help().Key, "quit")
	}

	return []key.Binding{
		cancel,
	}
}

func (c *CommandRunningMode) FullHelp(m Model, keyMap KeyMap) [][]key.Binding {
	return [][]key.Binding{
		c.ShortHelp(m, keyMap),
	}
}
