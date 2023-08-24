package chanwriter

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Read reads from the channel and calls the onNewData function with the data
func Read(w *ChanWriter, onNewData func(bytes []byte) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		buffer, ok := <-w.channel
		updateMsg := onNewData(buffer)
		updateCmd := func() tea.Msg {
			return updateMsg
		}

		// Channel is closed
		if !ok {
			if len(buffer) > 0 {
				return updateCmd
			} else {
				return nil
			}
		}

		return tea.Batch(
			updateCmd,
			Read(w, onNewData),
		)()
	}
}
