package shell

import (
	"context"

	"github.com/DomBlack/bubble-shell/pkg/modelid"
)

// This message is sent when a new command is being run
// back to the main model, which allows the main model to
// trigger a cancel on the context of the command which
// is currently running.
type currentCmdContextCancelFuncMsg struct {
	id     modelid.ID
	cancel context.CancelFunc
}

func (msg currentCmdContextCancelFuncMsg) ForModelID() modelid.ID {
	return msg.id
}

// enterModeMsg tells the shell mode to enter into given mode
type enterModeMsg struct {
	ID   modelid.ID
	Mode Mode
}

func (msg enterModeMsg) ForModelID() modelid.ID {
	return msg.ID
}

// ShutdownMsg is sent by the shell when it wants to shutdown
// the application.
//
// It is used to allow for one render cycle to occur before
// the [tea.Quit] is sent, allowing us to remove the prompt
type ShutdownMsg struct {
	ID modelid.ID
}

func (msg ShutdownMsg) ForModelID() modelid.ID {
	return msg.ID
}
