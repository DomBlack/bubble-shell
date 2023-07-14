package autocomplete

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DomBlack/bubble-shell/internal/cobrautils"
	"github.com/DomBlack/bubble-shell/pkg/modelid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

const maxLinesToShow = 8

type Model struct {
	id            modelid.ID
	parent        modelid.ID
	rootCmd       *cobra.Command
	width, height int

	optionStyle         lipgloss.Style
	selectedOptionStyle lipgloss.Style
	descriptionStyle    lipgloss.Style

	init bool // init only happens after the first window sizing

	selectedOption  int
	input           string
	options         []Option
	longestOption   int
	hasDescriptions bool
	directive       cobra.ShellCompDirective
}

func New(rootCmd *cobra.Command, parent modelid.ID) Model {
	return Model{
		id:      modelid.Next(),
		parent:  parent,
		rootCmd: rootCmd,

		optionStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		selectedOptionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFFFF")),
		descriptionStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.init = true

	case clearMsg:
		if m.id.Matches(msg) {
			m.options = nil
			m.input = ""
			m.selectedOption = 0
			return m, nil
		}

	case moveOption:
		if m.id.Matches(msg) {
			m.selectedOption += msg.Delta
			if m.selectedOption >= len(m.options) {
				m.selectedOption = 0
			} else if m.selectedOption < 0 {
				m.selectedOption = len(m.options) - 1
			}
		}

	case autoCompleteMsg:
		if m.id.Matches(msg) {
			if m.input == msg.Line {
				// Otherwise we're just moving the cursor around
				m.selectedOption++
				if m.selectedOption >= len(m.options) {
					m.selectedOption = 0
				}
			} else {
				// Input changed; new search
				directive, options, err := m.computeOptions(msg.Line)
				if err != nil {
					// FIXME: handle this error somehow?
					return m, nil
				}

				length := 0
				hasDescriptions := false
				for _, option := range options {
					if len(option.Name) > length {
						length = len(option.Name)
					}
					if option.Description != "" {
						hasDescriptions = true
					}
				}

				m.input = msg.Line
				m.selectedOption = 0
				m.options = options
				m.directive = directive
				m.longestOption = length
				m.hasDescriptions = hasDescriptions

				if len(options) == 1 {
					return m, func() tea.Msg {
						return SingleAutoCompleteOptionMsg{m.parent}
					}
				}
			}
		}

		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if !m.init {
		return ""
	}

	if len(m.options) == 0 {
		return ""
	}

	// Don't show more than maxLinesToShow or the height in the terminal
	maxLines := maxLinesToShow
	if maxLines > m.height-2 {
		maxLines = m.height - 2
	}

	// If we've got too many options or none of the options have a description
	// list them in columns with no descriptions
	if len(m.options) > maxLines || !m.hasDescriptions {
		const colPadding = 3
		// Render in columns
		numColumns := m.width / (m.longestOption + colPadding)
		if numColumns == 0 {
			numColumns = 1
		}

		perColumn := len(m.options) / numColumns
		if perColumn == 0 {
			perColumn = 1
		}

		columns := make([][]string, numColumns)
		for i, option := range m.options {
			colNum := i / perColumn

			if i == m.selectedOption {
				columns[colNum] = append(columns[colNum], m.selectedOptionStyle.Render(option.Name))
			} else {
				columns[colNum] = append(columns[colNum], m.optionStyle.Render(option.Name))
			}
		}

		// Format each column
		cols := make([]string, numColumns)
		for i, col := range columns {
			style := lipgloss.NewStyle()
			if i > 0 {
				style = style.PaddingLeft(colPadding)
			}

			cols[i] = style.Render(
				lipgloss.JoinVertical(lipgloss.Top, col...),
			)
		}
		return lipgloss.JoinHorizontal(lipgloss.Left, cols...)

	} else {
		// Otherwise render them in a list with the descriptions
		var lines []string

		for i, option := range m.options {
			var line string

			if i == m.selectedOption {
				line += m.selectedOptionStyle.Render(option.Name)
			} else {
				line += m.optionStyle.Render(option.Name)
			}

			line += strings.Repeat(" ", m.longestOption-len(option.Name))
			line += " - "
			line += m.descriptionStyle.Render(option.Description)

			lines = append(lines, line)
		}

		return lipgloss.JoinVertical(lipgloss.Top, lines...)
	}
}

func (m Model) computeOptions(input string) (cobra.ShellCompDirective, []Option, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	cmd := fmt.Sprintf("%s %s", cobra.ShellCompRequestCmd, input)

	var sb strings.Builder

	err := cobrautils.ExecuteCmd(ctx, m.rootCmd, cmd, &sb)
	if err != nil {
		return cobra.ShellCompDirectiveError, nil, errors.Wrap(err, "failed to execute shell completion")
	}

	// Grab the output
	return parseOptions(sb.String())
}

// Accept returns the currently selected suggestion, or an empty string
func (m Model) Accept() string {
	if len(m.options) == 0 {
		return ""
	}

	return m.options[m.selectedOption].Name
}

// Clear clears the suggestions
func (m Model) Clear() tea.Cmd {
	return func() tea.Msg {
		return clearMsg{
			ID: m.id,
		}
	}
}

// AutoComplete returns a command that will start the auto-complete process
func (m Model) AutoComplete(input string, cursorPosition int) tea.Cmd {
	return func() tea.Msg {
		return autoCompleteMsg{
			ID:       m.id,
			Line:     input,
			Position: cursorPosition,
		}
	}
}

// Next returns a command that will select the next option
func (m Model) Next() tea.Cmd {
	return func() tea.Msg {
		return moveOption{
			ID:    m.id,
			Delta: 1,
		}
	}
}

// Previous returns a command that will select the previous option
func (m Model) Previous() tea.Cmd {
	return func() tea.Msg {
		return moveOption{
			ID:    m.id,
			Delta: -1,
		}
	}
}
