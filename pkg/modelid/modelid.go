// Package modelid provides a way to identify a specific model instance
// by providing a global counter that is incremented each time a new
// model is created.
package modelid

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// ID is a unique identifier for a model
//
// It can be used to identify a specific instance
// of a model within a collection of models, which is useful
// when you need to update a specific model using a message
//
// Create a new ID using [Next]
type ID uint

// Matches returns true if the given message is for the model with the given [ID]
func (id ID) Matches(msg MsgForModelID) bool {
	return msg.ForModelID() == id
}

var (
	idMu sync.Mutex
	id   ID
)

// Next returns the next available [ID]
//
// The ID generated is unique within the process,
// but is not guaranteed to be unique across restarts
func Next() ID {
	idMu.Lock()
	defer idMu.Unlock()

	id++
	return id
}

// MsgForModelID is a message that is associated with a [ModelID]
//
// Implement this interface on your [tea.Msg] to allow it to work
// with [ModelID.Matches] method.
type MsgForModelID interface {
	tea.Msg
	ForModelID() ID
}
