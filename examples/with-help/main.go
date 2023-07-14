package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	shell "github.com/DomBlack/bubble-shell"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

func main() {
	// Build your cobra command as normal
	rootCmd := &cobra.Command{}
	curl := &cobra.Command{
		Use:     "curl [url]",
		Short:   "curl is a tool to transfer data from or to a server",
		Example: "curl https://www.google.com/humans.txt",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := bufio.NewWriter(cmd.OutOrStdout())
			_, _ = fmt.Fprintf(w, "Calling GET %s\n", args[0])
			if err := w.Flush(); err != nil {
				return errors.Wrap(err, "failed to flush")
			}

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, args[0], nil)
			if err != nil {
				return errors.Wrap(err, "failed to create request")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return errors.Wrap(err, "failed to make request")
			}
			defer resp.Body.Close()

			_, err = fmt.Fprintf(w, "Response Status: %s\n\n", resp.Status)
			if err := w.Flush(); err != nil {
				return errors.Wrap(err, "failed to flush")
			}

			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "failed to read response body")
			}

			_, err = fmt.Fprintf(w, "%s\n", string(bytes))
			if err := w.Flush(); err != nil {
				return errors.Wrap(err, "failed to flush")
			}
			return nil
		},
	}
	rootCmd.AddCommand(curl)

	reportError := &cobra.Command{
		Use:   "report-error [error]",
		Short: "report-error is a tool to report an error to the user",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New(strings.Join(args, " "))
		},
	}
	rootCmd.AddCommand(reportError)

	// Create a new bubble-shell program and run it
	p := tea.NewProgram(
		NewWrapperWithHelp(
			shell.New(rootCmd,
				shell.WithHistoryFile(".bubble-shell-help-example"),
			),
		),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	if err != nil {
		panic(err)
	}
}

// Model is a single model in which we've embedded the shell in another model
// which also displays help
type Model struct {
	Shell tea.Model
	Help  help.Model

	width     int
	helpStyle lipgloss.Style
}

func NewWrapperWithHelp(shell tea.Model) Model {
	helpModel := help.New()
	helpModel.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	return Model{
		Shell: shell,
		Help:  helpModel,
		helpStyle: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder(), true, false, false).
			BorderForeground(lipgloss.Color("#4E4E4E")).
			Align(lipgloss.Right),
	}
}

var _ tea.Model = Model{}

func (m Model) Init() tea.Cmd {
	return m.Shell.Init()
}

func (m Model) Update(msg tea.Msg) (rtn tea.Model, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

		m.Shell, cmd = m.Shell.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height - 2, // Leave space for the help and it's border
		})
	default:
		m.Shell, cmd = m.Shell.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		m.Shell.View(),
		m.helpStyle.Copy().Width(m.width).Render(
			m.Help.View(
				m.Shell.(help.KeyMap), // The shell implements the help.KeyMap interface
			),
		),
	)
}
