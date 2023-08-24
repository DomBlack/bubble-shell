package history

import (
	"fmt"
	"strings"
	"time"

	"github.com/DomBlack/bubble-shell/internal/config"
	"github.com/DomBlack/bubble-shell/pkg/tui/errdisplay"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/xid"
)

type Status uint8

const (
	UnknownStatus Status = iota
	RunningStatus
	SuccessStatus
	ErrorStatus
)

type ItemType uint8

const (
	Command         ItemType = iota // A command entered by the user
	InternalError                   // An internal error encountered by the shell
	HistoryRestored                 // A message indicating that the history was restored from disk and everything above it is from a previous session
)

// Item represents a single entry in the command history list
//
// It is both a tea.Model as well as a serializable struct
// to be stored within the history file.
type Item struct {
	// This group of fields are serialized and stored
	ID       xid.ID    `json:"id"`       // The unique ID of the command
	Prompt   string    `json:"prompt"`   // The prompt that was displayed when the command was executed
	Line     string    `json:"line"`     // The command executed
	Started  time.Time `json:"started"`  // The time the command was executed
	Finished time.Time `json:"finished"` // The time the command finished executing
	Status   Status    `json:"status"`   // The status of the command
	Output   string    `json:"output"`   // The output of the command

	// This group of fields are not serialized and are only used
	// for rendering the UI during the current shell session
	StreamedOutput []byte   `json:"-"` // The output of the command as it is streamed
	ItemType       ItemType `json:"-"` // If true then this item is an internal error item and not a user command
	Error          error    `json:"-"` // The error returned from the command
	LoadedHistory  bool     `json:"-"` // If true then this item is a history restored item and not a user command
}

// NewItem creates a new history item with the given line and status
func NewItem(prompt, line string, status Status) Item {
	return Item{
		ID:       xid.New(),
		Prompt:   prompt,
		Line:     line,
		Started:  time.Now(),
		Status:   status,
		ItemType: Command,
	}
}

// Init implements tea.Model init function
func (i Item) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model update function
//
// Note we return [Item] here rather than [tea.Model] as we
// save on type casting in the history model.
func (i Item) Update(_ tea.Msg) (Item, tea.Cmd) {
	return i, nil
}

// View implements tea.Model view function
func (i Item) View(cfg *config.Config, width int) string {
	lines := make([]string, 1, 6)

	// Generate the time string
	timeStr := i.Started.Format("15:04:05")
	if i.Started.Before(time.Now().Add(-24 * time.Hour)) {
		timeStr = i.Started.Format("2006-01-02 15:04:05")
	}

	finished := i.Finished
	if finished.IsZero() && i.Status == RunningStatus {
		finished = time.Now()
	}
	if !finished.IsZero() {
		dur := finished.Sub(i.Started)
		switch {
		case dur < 1*time.Second:
			timeStr = fmt.Sprintf("(%dms) %s", dur.Milliseconds(), timeStr)
		case dur < 10*time.Second:
			timeStr = fmt.Sprintf("(%.1fs) %s", dur.Seconds(), timeStr)
		case dur < 1*time.Minute:
			timeStr = fmt.Sprintf("(%.0fs) %s", dur.Seconds(), timeStr)
		case dur < 2*time.Minute:
			timeStr = fmt.Sprintf("(%.1fm) %s", dur.Minutes(), timeStr)
		default:
			timeStr = fmt.Sprintf("(%.0fm) %s", dur.Minutes(), timeStr)
		}
	}

	// Render the input line
	lines[0] = cfg.Styles.HistoricLine.Render(i.Line)
	switch i.ItemType {
	case Command:
		prompt := i.Prompt
		if prompt == "" {
			prompt = "> "
		}

		lines[0] = cfg.Styles.HistoricPrompt.Render(prompt) + lines[0]

	case InternalError:
		lines[0] = cfg.Styles.InternalError.Render("!! ") + lines[0]

	case HistoryRestored:
		lines[0] = cfg.Styles.HistoricLine.Render("-- ") + lines[0] + " "
		lines[0] += cfg.Styles.HistoricLine.Render(strings.Repeat("-", width-lipgloss.Width(lines[0])))
	}

	if i.ItemType != HistoryRestored {
		// Add the time to the line
		lineWidth := lipgloss.Width(lines[0])
		timeStr = cfg.Styles.HistoricTime.Copy().Width(width - lineWidth).Render(timeStr)
		lines[0] = lipgloss.JoinHorizontal(lipgloss.Left, lines[0], timeStr)
	}

	// Render the output if we have any
	if i.Output != "" {
		lines = append(lines, i.Output)
	} else if len(i.StreamedOutput) > 0 && i.Status == RunningStatus {
		// If we're running, then render the streamed output
		lines = append(lines, strings.TrimSpace(string(i.StreamedOutput)))
	}

	// Render the error if we have any
	if i.Error != nil {
		errView := lipgloss.NewStyle().Width(width).Render(errdisplay.New(cfg, i.Error).View())
		if errView != "" {
			lines = append(lines, errView)
		}
	}

	// Add a spacer line between items
	lines = append(lines, "")

	// Now join all the lines together
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
