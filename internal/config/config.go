package config

import (
	"github.com/DomBlack/bubble-shell/pkg/config/keymap"
	"github.com/DomBlack/bubble-shell/pkg/config/styles"
)

// Config is the configuration for the shell
type Config struct {
	// The file to store the history in (if not absolute will be relative to $HOME)
	//
	// If blank no history will be stored
	HistoryFile string

	KeyMap keymap.KeyMap // The key map to use
	Styles styles.Styles // The styles to use

	MaxStackFrames int // The maximum number of stack frames to show in errors

	// Packages to filter from the stack trace rendering of errors
	//
	// By default the shell will filter out packages related to running
	// the shell itself as well as the Go runtime package.
	//
	// If empty no filtering will be done
	PackagesToFilterFromStack []string
}

// Default returns a default configuration for the shell
func Default() *Config {
	return &Config{
		HistoryFile:    ".bubble-shell-history",
		KeyMap:         keymap.Default,
		Styles:         styles.Default,
		MaxStackFrames: 8,
		PackagesToFilterFromStack: []string{
			"runtime",
			"testing",
			"github.com/spf13/cobra",
			"github.com/cockroachdb/errors",
			"github.com/charmbracelet/bubbletea",
			"github.com/DomBlack/bubble-shell",
		},
	}
}
