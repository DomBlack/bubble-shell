package errdisplay

import (
	"fmt"
	"strings"

	"github.com/DomBlack/bubble-shell/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
)

type Model struct {
	cfg         *config.Config
	err         error
	frames      []Frame
	longestFunc int
}

// New creates a new error display model
//
// If cfg is nil then the default config will be used
func New(cfg *config.Config, err error) tea.Model {
	if cfg == nil {
		cfg = config.Default()
	}
	if err == nil {
		err = errors.New("nil error passed to errdisplay.New")
	}

	originalFrames := DeepestStack(err)
	frames := FilterCommonFrames(cfg, originalFrames)

	thisModule, _ := modulePackageAndFilePath()

	// Find the longest function name that we'll render
	// so we can pad the others to match
	longestFunc := 0
	for i, frame := range frames {
		// If the frame package is in this binary, then trim it just to show
		// the user the package path within their project
		if strings.HasPrefix(frame.Package, thisModule+"/") {
			frame.Package = strings.TrimPrefix(frame.Package, thisModule+"/")
			frames[i] = frame
		}

		if frame.Package == "main" {
			// On the main package we don't need to show the package name
			if len(frame.Function) > longestFunc {
				longestFunc = len(frame.Function)
			}
		} else {
			fName := fmt.Sprintf("%s.%s", frame.Package, frame.Function)
			if len(fName) > longestFunc {
				longestFunc = len(fName)
			}
		}
	}

	return Model{
		cfg:         cfg,
		err:         err,
		frames:      frames,
		longestFunc: longestFunc,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	lines := []string{
		m.cfg.Styles.ErrorTitle.Render("Error:") + " " + m.cfg.Styles.ErrorMessage.Render(m.err.Error()),
	}

	for _, detail := range errors.GetAllDetails(m.err) {
		lines = append(lines, m.cfg.Styles.ErrorDetails.Render(detail))
	}

	// Print the stack trace
	if len(m.frames) > 0 {
		lines = append(lines, " ", "Stack trace:")

		count := 0
		for _, frame := range m.frames {
			if frame.Package == "main" {
				lines = append(lines,
					fmt.Sprintf(
						"  at %s%s %s:%d",
						m.cfg.Styles.StackFrameFunction.Render(frame.Function),
						strings.Repeat(" ", m.longestFunc-len(frame.Function)),
						FilePathFromFrame(frame),
						frame.Line,
					),
				)
			} else {
				lines = append(lines,
					fmt.Sprintf(
						"  at %s.%s%s %s:%d",
						m.cfg.Styles.StackFramePackage.Render(frame.Package),
						m.cfg.Styles.StackFrameFunction.Render(frame.Function),
						strings.Repeat(" ", m.longestFunc-(len(frame.Function)+len(frame.Package)+1)),
						FilePathFromFrame(frame),
						frame.Line,
					),
				)
			}

			count++
			if count >= m.cfg.MaxStackFrames {
				lines = append(lines, fmt.Sprintf("... remaining %d frames omitted ...", len(m.frames)-count))
				break
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, lines...)
}
