package styles

import (
	. "github.com/charmbracelet/lipgloss"
)

// Styles is the set of styles which will be used to render the shell.
//
// A [Default] is provided and will be used by the shell if no other Styles is provided
// when creating a new shell.
type Styles struct {
	// Styles for User Input
	Placeholder Style // The style for placeholder text in both the command & search input boxes
	Cursor      Style // The style for the cursor in both inputs

	CommandPrompt Style // The style for the command prompt
	Command       Style // The style for the inputted command test
	SearchPrompt  Style // The style for the search prompt
	Search        Style // The style for the text in the search input

	// Styles for the history
	HistoricPrompt Style // The style for the prompt in the history
	HistoricLine   Style // The style for a historic command executed by the user
	HistoricTime   Style // The style for the time a command was executed

	// Styles for errors being printed
	ErrorTitle         Style // The style for the title of an error
	ErrorMessage       Style // The style for the message of an error
	ErrorDetails       Style // The style for the details of an error
	StackFramePackage  Style // The style for the name of a module in an error
	StackFrameFunction Style // The style for the name of a function in an error

	// Misc Styles
	InternalError Style // The styling for an internal error
}

// Default is the default styles for the shell
var Default = Styles{
	Placeholder: NewStyle().Foreground(Color("240")),
	Cursor:      NewStyle(),

	CommandPrompt: NewStyle().Foreground(Color("205")),
	Command:       NewStyle(),
	SearchPrompt:  NewStyle(),
	Search:        NewStyle(),

	HistoricPrompt: NewStyle().Foreground(Color("91")),
	HistoricLine:   NewStyle().Foreground(Color("244")),
	HistoricTime:   NewStyle().Foreground(Color("240")).Align(Right),

	ErrorTitle:         NewStyle().Foreground(Color("#FF0000")).Bold(true),
	ErrorMessage:       NewStyle().Foreground(Color("#FF8888")),
	ErrorDetails:       NewStyle().Foreground(Color("#D3D3D3")).PaddingLeft(2),
	StackFramePackage:  NewStyle().Foreground(Color("90")),
	StackFrameFunction: NewStyle().Foreground(Color("35")),

	InternalError: NewStyle().Foreground(Color("196")).Bold(true).Blink(true),
}
