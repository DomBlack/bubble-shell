package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap is a collection of all the key bindings used by the shell
//
// A [Default] is provided and will be used by the shell if no other KeyMap is provided
// when creating a new shell.
type KeyMap struct {
	Up             key.Binding // Up is a binding for the user to look back in their command history
	Down           key.Binding // Down is a binding for the user to look forward in their command history
	ExecuteCommand key.Binding // ExecuteCommand is a binding for the user to execute the current command

	// Cancel is a binding for the user to cancel their current command
	//
	// It's behaviour changes depending on what the user is doing:
	//
	// 1. If a command is being run, it will cancel the command's context
	// 2. If the user is searching their command history, it will cancel the search
	// 3. If the user has a partially typed in command, it will clear the input
	// 4. If none of the above, it will exit the shell
	Cancel key.Binding

	SearchHistoryBackwards key.Binding // SearchHistoryBackwards is a binding for the user to search their command history backwards
	SearchHistoryForwards  key.Binding // SearchHistoryForwards is a binding for the user to search their command history forwards

	AutoComplete         key.Binding // AutoComplete is a binding for the user to autocomplete their current command or cycle through autocompletions
	PreviousAutoComplete key.Binding // PreviousAutoComplete is a binding for the user to cycle through previous autocompletions
}

// Default is the default keymap for the prompt
var Default = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "history up"),
	),

	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "history down"),
	),

	ExecuteCommand: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "execute command"),
	),

	Cancel: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "cancel"),
	),

	SearchHistoryBackwards: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "history search"),
	),

	SearchHistoryForwards: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "history search (forward)"),
	),

	AutoComplete: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "autocomplete"),
	),

	PreviousAutoComplete: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous autocomplete"),
	),
}
