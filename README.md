# Bubble Shell

This library provides you with a interactive shell powered by [bubbletea](https://github.com/charmbracelet/bubbletea) 
by executing [cobra](https://github.com/spf13/cobra), allowing you to build powerful interactive terminal based shell
applications.

## Usage

Create a root command as you would normally with cobra, attaching your various commands to it.
Then pass it to the `shell.New` function to create a new shell.

```go
package main

import (
	tea "github.com/charmbracelet/bubbletea"
    "github.com/DomBlack/bubble-shell"
    "github.com/spf13/cobra"
)

func main() {
	// Create the root command which we can attach everything to
    rootCmd := &cobra.Command{}
	
	// Create a command to add to the root command
	myCmd := &cobra.Command{
		Use: "hi",
        RunE: func(cmd *cobra.Command, args []string) error {
			cmd.OutOrStdout().Write([]byte("Hello World\n"))
			return nil
        },
	}
	rootCmd.AddCommand(myCmd)
    
	// Create the shell
	shell := shell.New(rootCmd)
	
	// Run the shell
	p := tea.NewProgram(shell)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
```
                                               
In this example, a user will be given an interactive shell with a single command `hi` which will print `Hello World` when
run.

### Options

The `shell.New` function takes a number of options to configure the shell, which are detailed below. Full documentaton
for them can be seen in [options.go](./options.go).

#### `shell.WithHistoryFile` / `shell.WithNoHistory`

By default the shell will save the history of commands to `.bubble-shell-history` in the user's home directory, however
using these two options you can change this behaviour, either providing your own filename or disabling history entirely.

#### `shell.WithKeyMap`

You can use this option to customise the key bindings used by the shell. The default key bindings are located in
[the keymap package](./pkg/config/keymap/keymap.go).

#### `shell.WithStyles`

You can use this option to customise the styles used by the shell. The default styles are located in
[the styles package](./pkg/config/styles/styles.go).

#### `shell.WithStackTraceFilters` / `shell.WithAdditionalStackTraceFilters`

By default the shell will filter out stack traces from the `github.com/DomBlack/bubble-shell` package and other related
packages when rendering the stack traces of errors to the user. You can use these options to customise this behaviour,
either by providing your own list of packages to filter out, or by adding to the default list.

### Guidelines for building commands

1. The shell supports autocompletion of commands and arguments, so ideally implement a `ValidArgsFunction` function or
    `ValidArgs` property on your command.
2. If you implement your commands using `RunE` rather than `Run` you can then return an error to bubble-shell which will
    be displayed to the user. If the error carries a stack trace, it will be displayed to the user. (I recommend using
    [cockroachdb/errors](https://github.com/cockroachdb/errors) to create errors with stack traces by default).
3. If you need a context use `cmd.Context()` rather than creating your own. This is because the shell will cancel the
    context when the user presses `Ctrl+C`.
4. If you want to display a message to the user, use `cmd.OutOrStdout()` rather than `fmt.Println("Hello World")`.  While
    both will be captured and returned to the user, the former will be streamed to the user as it is written, while the
    `os.Stdout` and `os.Stderr` streams are buffered and only displayed when the command completes.


## Example Apps

- [Example which shows the active keybindings](./examples/with-help/main.go)
