package shell

import (
	"github.com/DomBlack/bubble-shell/pkg/tui/history"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type HistorySearchMode struct{}

var _ Mode = (*HistorySearchMode)(nil)

func (h *HistorySearchMode) Enter(m Model) (Model, tea.Cmd) {
	m.searchDirBackwards = true
	m.searchInput.Prompt = m.searchInputPrompt(true)
	m.searchInput.SetValue("")
	m.searchInput.CursorEnd()
	m.searchInput.Focus()

	return m, nil
}

func (h *HistorySearchMode) Leave(m Model) (Model, tea.Cmd) {
	m.searchInput.SetValue("")
	m.searchInput.Blur()
	return m, nil
}

func (h *HistorySearchMode) Update(m Model, msg tea.Msg) (Model, tea.Cmd) {
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

		case key.Matches(msg, m.cfg.KeyMap.SearchHistoryBackwards):
			foundIdx, found := m.history.Search(m.searchInput.Value(), m.lookBack, 1)
			m.searchDirBackwards = true
			m.searchInput.Prompt = m.searchInputPrompt(found)
			return m.updateSearchResult(foundIdx), nil

		case key.Matches(msg, m.cfg.KeyMap.SearchHistoryForwards):
			foundIdx, found := m.history.Search(m.searchInput.Value(), m.lookBack, -1)
			m.searchDirBackwards = false
			m.searchInput.Prompt = m.searchInputPrompt(found)
			return m.updateSearchResult(foundIdx), nil

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

		default:
			// If the search term has changed, reset the search
			searchTerms := m.searchInput.Value()
			if searchTerms != m.lastSearch {
				foundIdx, found := m.history.Search(m.searchInput.Value(), 0, 1)
				m.searchInput.Prompt = m.searchInputPrompt(found)

				return m.updateSearchResult(foundIdx), nil
			}
		}

		return m, nil
	}

	return m, nil
}

func (h *HistorySearchMode) AdditionalView(m Model) string {
	return m.searchInput.View()
}

func (h *HistorySearchMode) ShortHelp(m Model, keyMap KeyMap) []key.Binding {
	backwards := keyMap.SearchHistoryBackwards
	backwards.SetHelp(backwards.Help().Key, "search backwards")
	forwards := keyMap.SearchHistoryForwards
	forwards.SetHelp(forwards.Help().Key, "search forwards")

	return []key.Binding{
		backwards, forwards, keyMap.ExecuteCommand, keyMap.Cancel,
	}
}

func (h *HistorySearchMode) FullHelp(m Model, keyMap KeyMap) [][]key.Binding {
	return [][]key.Binding{
		h.ShortHelp(m, keyMap),
	}
}

func (m Model) updateSearchResult(idx int) Model {
	m.lookBack = idx
	if m.lookBack == 0 {
		m.input.SetValue("")
	} else {
		m.input.SetValue(m.history.Lookback(m.lookBack).Line)
	}
	return m
}

func (m Model) searchInputPrompt(foundResult bool) string {
	if foundResult {
		if m.searchDirBackwards {
			return "bck-i-search: "
		} else {
			return "fwd-i-search: "
		}
	} else {
		if m.searchDirBackwards {
			return "failing bck-i-search: "
		} else {
			return "failing fwd-i-search: "
		}
	}
}
