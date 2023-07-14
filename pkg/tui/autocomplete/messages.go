package autocomplete

import (
	"github.com/DomBlack/bubble-shell/pkg/modelid"
)

// SingleAutoCompleteOptionMsg is a message that is sent by
// the autocomplete model to indicate only one option has
// been found and should just be accepted
type SingleAutoCompleteOptionMsg struct {
	ID modelid.ID
}

func (msg SingleAutoCompleteOptionMsg) ForModelID() modelid.ID {
	return msg.ID
}

// autoCompleteMsg is a message that is sent to the
// autocomplete model to start the autocomplete process
type autoCompleteMsg struct {
	ID       modelid.ID
	Line     string
	Position int
}

func (msg autoCompleteMsg) ForModelID() modelid.ID {
	return msg.ID
}

// clearMsg is a message that is sent to the autocomplete
// model to clear the current autocomplete options
type clearMsg struct {
	ID modelid.ID
}

func (msg clearMsg) ForModelID() modelid.ID {
	return msg.ID
}

// moveOption is a message that is sent to the autocomplete
// model to move the currently selected option
type moveOption struct {
	ID    modelid.ID
	Delta int
}

func (msg moveOption) ForModelID() modelid.ID {
	return msg.ID
}
