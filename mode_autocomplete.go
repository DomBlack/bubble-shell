package shell

import (
	"strings"

	"github.com/DomBlack/bubble-shell/pkg/tui/autocomplete"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type AutoCompleteMode struct{}

var _ Mode = (*AutoCompleteMode)(nil)

func (a *AutoCompleteMode) Enter(m Model) (Model, tea.Cmd) {
	m.input.Blur()
	return m, m.autocomplete.AutoComplete(m.input.Value(), m.input.Position())
}

func (a *AutoCompleteMode) Leave(m Model) (Model, tea.Cmd) {
	return m, m.autocomplete.Clear()
}

func (a *AutoCompleteMode) Update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case autocomplete.SingleAutoCompleteOptionMsg:
		if m.id.Matches(msg) {
			return a.AcceptOption(m)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.cfg.KeyMap.AutoComplete):
			return m, m.autocomplete.NextResult()

		case key.Matches(msg, m.cfg.KeyMap.PreviousAutoComplete):
			return m, m.autocomplete.PreviousResult()

		case key.Matches(msg, m.cfg.KeyMap.Up):
			return m, m.autocomplete.PreviousRow()

		case key.Matches(msg, m.cfg.KeyMap.Down):
			return m, m.autocomplete.NextRow()

		case key.Matches(msg, m.cfg.KeyMap.Left):
			return m, m.autocomplete.PreviousColumn()

		case key.Matches(msg, m.cfg.KeyMap.Right):
			return m, m.autocomplete.NextColumn()

		case key.Matches(msg, m.cfg.KeyMap.Cancel):
			return m, m.Enter(&CommandEntryMode{KeepInputContent: true})

		case key.Matches(msg, m.cfg.KeyMap.ExecuteCommand):
			return a.AcceptOption(m)

		case msg.Type == tea.KeySpace:
			return a.AcceptOption(m)

		default:
			// default is to exit the mode and re-send the key message
			return m, tea.Sequence(
				m.Enter(&CommandEntryMode{KeepInputContent: true}),
				func() tea.Msg { return msg }, // re-send the key message
			)
		}
	}

	return m, nil
}

func (a *AutoCompleteMode) AdditionalView(m Model) string {
	return m.autocomplete.View()
}

func (a *AutoCompleteMode) AcceptOption(m Model) (Model, tea.Cmd) {
	if suggestion := m.autocomplete.Accept(); suggestion != "" {
		input := m.input.Value()

		// Cut off the last word so we can replace it with the autocomplete
		if idx := strings.LastIndexByte(input, ' '); idx == -1 {
			input = ""
		} else {
			input = input[:idx+1]
		}

		m.input.SetValue(input + suggestion + " ")
		m.input.CursorEnd()
	}

	return m, m.Enter(&CommandEntryMode{KeepInputContent: true})
}

func (a *AutoCompleteMode) ShortHelp(m Model, keyMap KeyMap) []key.Binding {
	execute := keyMap.ExecuteCommand
	execute.SetHelp(execute.Help().Key, "Accept")

	return []key.Binding{
		keyMap.AutoComplete, keyMap.PreviousAutoComplete,
		execute,
		keyMap.Cancel,
	}
}

func (a *AutoCompleteMode) FullHelp(m Model, keyMap KeyMap) [][]key.Binding {
	return [][]key.Binding{
		a.ShortHelp(m, keyMap),
	}
}
