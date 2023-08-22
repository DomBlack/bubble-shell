package shell

import (
	"strings"

	"github.com/DomBlack/bubble-shell/pkg/tui/history"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type CommandEntryMode struct{ KeepInputContent bool }

var _ Mode = (*CommandEntryMode)(nil)

func (c *CommandEntryMode) Enter(m Model) (Model, tea.Cmd) {
	if !c.KeepInputContent {
		m.input.SetValue(m.lookBackPartial)
	}

	m.lookBackPartial = ""
	m.input.Prompt = m.cfg.PromptFunc()
	m.input.CursorEnd()
	m.input.Focus()

	return m, nil
}

func (c *CommandEntryMode) Leave(m Model) (Model, tea.Cmd) {
	m.lookBackPartial = m.input.Value()
	m.input.Blur()
	return m, nil
}

func (c *CommandEntryMode) Update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		m.height = msg.Height
		m.width = msg.Width

		m.input.Width = m.width

		m.history, cmd = m.history.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height - 2, // one for the prompt, one for search and autocomplete
		})
		cmds = append(cmds, cmd)

		// We only want to init once we have a window size
		m.init = true

		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.cfg.KeyMap.ExecuteCommand):
			line := strings.TrimSpace(m.input.Value())
			if line == "" {
				return m, nil
			}

			if line == "exit" || line == "quit" {
				return m, tea.Quit
			}

			historyItem := history.NewItem(m.input.Prompt, line, history.RunningStatus)
			m.input.SetValue("")
			m.input.CursorEnd()

			return m, tea.Sequence(
				m.Enter(&CommandRunningMode{}),
				m.history.AppendItem(historyItem),
				m.ExecuteCommand(historyItem),
			)

		case key.Matches(msg, m.cfg.KeyMap.Up) && len(m.history.Items) > 0:
			return m, m.Enter(&HistoryLookbackMode{TriggerMsg: msg})

		case key.Matches(msg, m.cfg.KeyMap.SearchHistoryBackwards) && len(m.history.Items) > 0:
			return m, m.Enter(&HistorySearchMode{})

		case key.Matches(msg, m.cfg.KeyMap.AutoComplete):
			return m, m.Enter(&AutoCompleteMode{})

		case key.Matches(msg, m.cfg.KeyMap.Cancel):
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.input.CursorEnd()
				return m, nil
			} else {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (c *CommandEntryMode) AdditionalView(m Model) string {
	return ""
}

func (c *CommandEntryMode) ShortHelp(m Model, keyMap KeyMap) []key.Binding {
	cancel := keyMap.Cancel
	if m.input.Value() != "" {
		cancel.SetHelp(cancel.Help().Key, "clear input")
	} else {
		cancel.SetHelp(cancel.Help().Key, "exit")
	}

	return []key.Binding{
		keyMap.Up, keyMap.SearchHistoryBackwards, keyMap.AutoComplete, keyMap.ExecuteCommand, cancel,
	}
}

func (c *CommandEntryMode) FullHelp(m Model, keyMap KeyMap) [][]key.Binding {
	return [][]key.Binding{
		c.ShortHelp(m, keyMap),
	}
}
