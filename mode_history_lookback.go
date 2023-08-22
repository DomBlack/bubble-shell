package shell

import (
	"github.com/DomBlack/bubble-shell/pkg/tui/history"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type HistoryLookbackMode struct {
	TriggerMsg tea.Msg
}

var _ Mode = (*HistoryLookbackMode)(nil)

func (h *HistoryLookbackMode) Enter(m Model) (Model, tea.Cmd) {
	m.lookBack = 0
	return m, func() tea.Msg {
		return h.TriggerMsg
	}
}

func (h *HistoryLookbackMode) Leave(m Model) (Model, tea.Cmd) {
	m.input.SetValue("")
	m.input.Blur()
	return m, nil
}

func (h *HistoryLookbackMode) Update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.cfg.KeyMap.ExecuteCommand):
			line := m.input.Value()
			historyItem := history.NewItem(m.input.Prompt, line, history.RunningStatus)

			return m, tea.Sequence(
				m.Enter(&CommandRunningMode{}),
				m.history.AppendItem(historyItem),
				m.ExecuteCommand(historyItem),
			)

		case key.Matches(msg, m.cfg.KeyMap.Up):
			if len(m.history.Items) == 0 {
				return m, nil
			}

			previousLine := m.history.Lookback(m.lookBack).Line
			possibleLine := previousLine

			for previousLine == possibleLine {
				m.lookBack++
				if m.lookBack > len(m.history.Items) {
					m.lookBack = len(m.history.Items)
					break
				}
				item := m.history.Lookback(m.lookBack)
				if item.ItemType == history.Command {
					possibleLine = item.Line
				}
			}
			m.input.SetValue(possibleLine)
			m.input.CursorEnd()

			return m, nil

		case key.Matches(msg, m.cfg.KeyMap.Down):
			if len(m.history.Items) == 0 {
				return m, nil
			}

			previousLine := m.history.Lookback(m.lookBack).Line
			possibleLine := previousLine

			for previousLine == possibleLine {
				m.lookBack--
				if m.lookBack <= 0 {
					break
				}
				item := m.history.Lookback(m.lookBack)
				if item.ItemType == history.Command {
					possibleLine = item.Line
				}
			}

			if m.lookBack <= 0 {
				return m, m.Enter(&CommandEntryMode{})
			} else {
				m.input.SetValue(possibleLine)
			}

			return m, nil

		case key.Matches(msg, m.cfg.KeyMap.Cancel):
			return m, m.Enter(&CommandEntryMode{})

		case msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight || msg.Type == tea.KeyTab:
			line := m.history.Lookback(m.lookBack).Line
			m.input.SetValue(line)

			if msg.Type == tea.KeyLeft {
				return m, tea.Sequence(
					m.Enter(&CommandEntryMode{KeepInputContent: true}),
					func() tea.Msg { return msg },
				)
			} else {
				return m, m.Enter(&CommandEntryMode{KeepInputContent: true})
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (h *HistoryLookbackMode) AdditionalView(m Model) string {
	return ""
}

func (h *HistoryLookbackMode) ShortHelp(m Model, keyMap KeyMap) []key.Binding {
	return []key.Binding{
		keyMap.Up, keyMap.Down, keyMap.ExecuteCommand, keyMap.Cancel,
	}
}

func (h *HistoryLookbackMode) FullHelp(m Model, keyMap KeyMap) [][]key.Binding {
	return [][]key.Binding{
		h.ShortHelp(m, keyMap),
	}
}
