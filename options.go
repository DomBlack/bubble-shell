package shell

import (
	"context"

	"github.com/DomBlack/bubble-shell/internal/config"
	"github.com/DomBlack/bubble-shell/pkg/config/keymap"
	"github.com/DomBlack/bubble-shell/pkg/config/styles"
)

// Option is a function that configures the shell.
type Option func(*config.Config)

// WithHistoryFile sets the history file to be used by the shell
// if not absolute will be relative to $HOME
func WithHistoryFile(fileName string) Option {
	return func(o *config.Config) {
		o.HistoryFile = fileName
	}
}

// WithNoHistory disables history for the shell
func WithNoHistory() Option {
	return func(o *config.Config) {
		o.HistoryFile = ""
	}
}

// KeyMap is a collection of all the key bindings used by the shell
//
// A default is provided and will be used by the shell if no other KeyMap is provided
// when creating a new shell.
type KeyMap = keymap.KeyMap

// WithKeyMap sets the keymap to be used by the shell
func WithKeyMap(keyMap KeyMap) Option {
	return func(o *config.Config) {
		o.KeyMap = keyMap
	}
}

// Styles is the set of styles which will be used to render the shell.
//
// A default is provided and will be used by the shell if no other Styles is provided
// when creating a new shell.
type Styles = styles.Styles

// WithStyles sets the styles to be used by the shell
func WithStyles(styles Styles) Option {
	return func(o *config.Config) {
		o.Styles = styles
	}
}

// WithMaxStackFrames sets the maximum number of stack frames to show in errors
// when rendered to the user.
//
// The shell defaults to 8 frames.
func WithMaxStackFrames(frames int) Option {
	return func(o *config.Config) {
		o.MaxStackFrames = frames
	}
}

// WithAdditionalStackTraceFilters adds additional packages to filter from the stack traces
// of errors when rendered to the user.
func WithAdditionalStackTraceFilters(packages ...string) Option {
	return func(o *config.Config) {
		o.PackagesToFilterFromStack = append(packages, o.PackagesToFilterFromStack...)
	}
}

// WithStackTraceFilters sets the packages to filter from the stack traces
// of errors when rendered to the user.
//
// By default the shell will filter out packages related to running the shell itself,
// if you want to keep these filters, then use [WithAdditionalStackTraceFilters].
//
// If you want no filtering to be done, then call this with no packages listed.
func WithStackTraceFilters(packages ...string) Option {
	return func(o *config.Config) {
		o.PackagesToFilterFromStack = packages
	}
}

// WithInlineShell sets the shell to be inline rather than trying to render full screen
//
// This means that recovered history will not be shown, however your terminals own render
// will be incharge of scrolling.
func WithInlineShell() Option {
	return func(o *config.Config) {
		o.InlineShell = true
	}
}

// WithBaseContext sets the context that commands will be run with
// when they are executed by users.
//
// By default [context.Background] will be used
func WithBaseContext(ctx context.Context) Option {
	return func(o *config.Config) {
		o.RootContext = ctx
	}
}

// WithPromptFunc sets the function for rendering the prompt
//
// By default a function will be provided that returns "> "
func WithPromptFunc(promptFunc func() string) Option {
	if promptFunc == nil {
		panic("promptFunc cannot be nil")
	}

	return func(o *config.Config) {
		o.PromptFunc = promptFunc
	}
}
