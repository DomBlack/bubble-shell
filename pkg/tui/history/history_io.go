package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	. "github.com/DomBlack/bubble-shell/pkg/modelid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
)

// readCompletedMsg is sent when the history file has been loaded
// from disk
type readCompletedMsg struct {
	ID    ID
	Items []Item
}

func (msg readCompletedMsg) ForModelID() ID {
	return msg.ID
}

// ReadHistory reads the history file and then
// returns a message to update the history model
func (m Model) ReadHistory() tea.Cmd {
	readHistory := func() ([]Item, error) {
		if m.cfg.HistoryFile == "" {
			return nil, nil
		}

		// Get the path to the history file
		historyFileLocation, err := getHistoryFileLocation(m.cfg.HistoryFile)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get history file location")
		}

		// Open the history file
		file, err := os.Open(historyFileLocation)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				return nil, nil
			default:
				return nil, errors.Wrap(err, "unable to open history file")
			}
		}
		defer func() { _ = file.Close() }()

		// Read each line of the history file and unmarshal it
		var history []Item
		in := bufio.NewReader(file)
		for {
			line, part, err := in.ReadLine()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if part {
				if m.cfg.InlineShell {
					continue
				} else {
					return nil, errors.New("line too long")
				}
			}
			if !utf8.Valid(line) {
				return nil, errors.New("line isn't valid utf8")
			}

			item := Item{}
			if err := json.Unmarshal(line, &item); err != nil {
				return nil, errors.Wrap(err, "unable to unmarshal history item")
			}
			item.LoadedHistory = true

			history = append(history, item)
			if len(history) > Limit {
				history = history[1:]
			}
		}

		// Mark the history as restored if there is any history
		if len(history) > 0 {
			item := NewItem("", "restored history from previous session", SuccessStatus)
			item.LoadedHistory = true
			item.ItemType = HistoryRestored
			history = append(history, item)
		}

		return history, nil
	}

	return func() tea.Msg {
		history, err := readHistory()
		if err != nil {
			item := NewItem("", "error loading history file", ErrorStatus)
			item.ItemType = InternalError
			item.Error = err

			history = []Item{item}
		}

		return readCompletedMsg{
			ID:    m.id,
			Items: history,
		}
	}
}

// SaveHistory saves the history file to disk
//
// We pass in the items to save as a parameter so that we can
// asynchronously save the history file without worrying about
// commands which might be added to the queue after this one
func (m Model) SaveHistory(items []Item) tea.Cmd {
	saveHistory := func() error {
		if m.cfg.HistoryFile == "" {
			return nil
		}

		// Get the path to the history file
		historyFileLocation, err := getHistoryFileLocation(m.cfg.HistoryFile)
		if err != nil {
			return errors.Wrap(err, "unable to get history file location")
		}

		// Open the history file
		file, err := os.Create(historyFileLocation)
		defer func() { _ = file.Close() }()

		for _, item := range items {
			// Don't save internal errors or items marked as don't save
			if item.ItemType != Command {
				continue
			}

			bytes, err := json.Marshal(item)
			if err != nil {
				return errors.Wrap(err, "unable to marshal history item")
			}

			_, err = fmt.Fprintln(file, string(bytes))
			if err != nil {
				return errors.Wrap(err, "unable to write item tohistory file")
			}
		}

		return errors.Wrap(err, "unable to write history file")
	}

	return func() tea.Msg {
		err := saveHistory()
		if err != nil {
			item := NewItem("", "error saving history file", ErrorStatus)
			item.ItemType = InternalError
			item.Error = err

			return addItemMsg{
				ID:   m.id,
				Item: item,
			}
		}
		return nil
	}
}

// getHistoryFileLocation returns the location of the history file
func getHistoryFileLocation(historyFilename string) (string, error) {
	if ext := filepath.Ext(historyFilename); ext == "" {
		// we use .jsonl as we save each line as a json object
		historyFilename += ".jsonl"
	}

	if !filepath.IsAbs(historyFilename) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "unable to get user home directory")
		}

		historyFilename = filepath.Clean(filepath.Join(homeDir, historyFilename))
	}

	// Create the directory if it doesn't exist
	dirName := filepath.Dir(historyFilename)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.MkdirAll(dirName, 0755)
		if err != nil {
			return "", errors.Wrap(err, "unable to create history directory")
		}
	} else if err != nil {
		return "", errors.Wrap(err, "unable to stat history directory")
	}

	return historyFilename, nil
}
