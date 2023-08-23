package shell

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/DomBlack/bubble-shell/internal/cobrautils"
	"github.com/DomBlack/bubble-shell/internal/config"
	"github.com/DomBlack/bubble-shell/pkg/modelid"
	"github.com/DomBlack/bubble-shell/pkg/tui/autocomplete"
	"github.com/DomBlack/bubble-shell/pkg/tui/history"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Model is the main shell model
type Model struct {
	id  modelid.ID
	cfg *config.Config

	init          bool
	height, width int

	rootCmd          *cobra.Command
	currentCmdCancel context.CancelFunc

	history      history.Model
	autocomplete autocomplete.Model
	input        textinput.Model

	searchInput        textinput.Model
	lastSearch         string
	searchDirBackwards bool

	lookBack        int
	lookBackPartial string

	shuttingDown bool

	mode Mode
}

// The shell model implements the [help.KeyMap] interface so that you can
// embed it somewhere and use the [help] package to render help.
//
// The key mapping will be updated depending on the [Mode] the shell is in.
var _ help.KeyMap = Model{}

func New(rootCmd *cobra.Command, options ...Option) tea.Model {
	// Create the config we need
	cfg := config.Default()
	for _, option := range options {
		option(cfg)
	}

	input := textinput.New()
	input.Placeholder = "Enter your command here..."
	input.TextStyle = cfg.Styles.Command
	input.PromptStyle = cfg.Styles.CommandPrompt
	input.PlaceholderStyle = cfg.Styles.Placeholder
	input.Cursor.Style = cfg.Styles.Cursor
	input.Focus()

	searchInput := textinput.New()
	searchInput.Prompt = "search: "
	searchInput.TextStyle = cfg.Styles.Search
	searchInput.PromptStyle = cfg.Styles.SearchPrompt
	searchInput.PlaceholderStyle = cfg.Styles.Placeholder

	// Reroute cobra to output via our logs
	cobrautils.InitRootCmd(rootCmd)

	id := modelid.Next()
	return Model{
		id:  id,
		cfg: cfg,

		rootCmd: rootCmd,

		history:      history.New(cfg),
		autocomplete: autocomplete.New(rootCmd, id),
		input:        input,
		searchInput:  searchInput,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.Enter(&CommandEntryMode{}),
		m.history.Init(),
		m.autocomplete.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Pass all messages to the two inputs
	// before we do anything else
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.searchInput, cmd = m.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	// Now handle the messages we care about
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		m.input.Width = m.width

		m.history, cmd = m.history.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height - 1, // one for the prompt
		})
		cmds = append(cmds, cmd)

		m.autocomplete, cmd = m.autocomplete.Update(msg)
		cmds = append(cmds, cmd)

		// We only want to init once we have a window size
		m.init = true

		return m, tea.Batch(cmds...)

	case ShutdownMsg:
		if m.id.Matches(msg) {
			m.shuttingDown = true
			return m, tea.Quit
		}

	case enterModeMsg:
		if m.id.Matches(msg) {
			if m.mode == nil {
				m.mode = msg.Mode
				return m.mode.Enter(m)
			}

			left, leaveCmds := m.mode.Leave(m)
			left.mode = msg.Mode
			entered, enterCmds := left.mode.Enter(left)

			return entered, tea.Batch(leaveCmds, enterCmds)
		}
		return m, nil

	case currentCmdContextCancelFuncMsg:
		if m.id.Matches(msg) {
			m.currentCmdCancel = msg.cancel
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.cfg.KeyMap.Quit):
			return m, m.Shutdown
		}
	}

	if m.mode != nil {
		m, cmd = m.mode.Update(m, msg)
		cmds = append(cmds, cmd)
	}

	m.history, cmd = m.history.Update(msg)
	cmds = append(cmds, cmd)

	m.autocomplete, cmd = m.autocomplete.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.mode == nil {
		// we've not finished init yet
		return ""
	}

	if !m.init {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("Waiting for window size...")
	}

	historyView := m.history.View()
	input := m.input.View()
	modeView := m.mode.AdditionalView(m)

	if !m.cfg.InlineShell {
		// Fit the history to the screen based on the output of the autocomplete and if we're showing the search input
		neededHistoryHeight := m.height - 1 // one for the prompt
		if modeView != "" {
			neededHistoryHeight -= lipgloss.Height(modeView)
		}

		for lipgloss.Height(historyView) > neededHistoryHeight {
			firstNewLine := strings.IndexRune(historyView, '\n')
			historyView = historyView[firstNewLine+1:]
		}
	}

	// If we're an inline shell and the previous command is still running, wait for it to finish
	lastCmd := m.history.Lookback(1)
	if m.cfg.InlineShell && !lastCmd.LoadedHistory && !lastCmd.Started.IsZero() && lastCmd.Finished.IsZero() {
		input = ""
	}

	// Now render all the parts vertically
	parts := make([]string, 0, 3)
	parts = append(parts, historyView)
	if !m.shuttingDown { // on shutdown we don't want to render the input or other mode views
		if input != "" {
			parts = append(parts, input)
		}
		if modeView != "" {
			parts = append(parts, modeView)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Top, parts...)
}

func (m Model) ExecuteCommand(cmd history.Item) tea.Cmd {
	ctx, cancel := context.WithCancel(m.cfg.RootContext)

	return tea.Batch(
		func() tea.Msg {
			return currentCmdContextCancelFuncMsg{m.id, cancel}
		},
		func() tea.Msg {
			defer cancel()
			stdoutBuffer := new(bytes.Buffer)

			err := cobrautils.ExecuteCmd(ctx, m.rootCmd, cmd.Line, stdoutBuffer)
			cmd.Finished = time.Now()
			if err != nil {
				cmd.Status = history.ErrorStatus
				cmd.Error = err
			}
			cmd.Status = history.SuccessStatus

			// Capture the stdout
			cmd.Output = strings.TrimSpace(string(stdoutBuffer.Bytes()))

			return tea.Sequence(
				m.history.UpdateItem(cmd),    // Create an UpdateItem [tea.Cmd]
				m.Enter(&CommandEntryMode{}), // Then switch back to command entry mode
			)()
		},
	)
}

// Shutdown is a [tea.Cmd] to shutdown the shell cleanly
func (m Model) Shutdown() tea.Msg {
	return ShutdownMsg{ID: m.id}
}

func (m Model) ShortHelp() []key.Binding {
	if m.mode == nil {
		return nil
	}
	return m.mode.ShortHelp(m, m.cfg.KeyMap)
}

func (m Model) FullHelp() [][]key.Binding {
	if m.mode == nil {
		return nil
	}
	return m.mode.FullHelp(m, m.cfg.KeyMap)
}
