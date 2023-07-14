package cobrautils

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	stdoutCaptureMu sync.Mutex
)

// InitRootCmd sets up the root command for the shell and calls
// some of the same functions that would run behind the scenes for
// `rootCmd.Execute()`.
//
// These are needed to do stuff like add the internal autocomplete command
func InitRootCmd(rootCmd *cobra.Command) {
	addDefaultShellCommands(rootCmd)

	rootCmd.SilenceErrors = true
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.Hidden = true

	for _, subCmd := range rootCmd.Commands() {
		subCmd.SilenceUsage = true
	}

	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
}

// ExecuteCmd executes a command and captures the output
func ExecuteCmd(ctx context.Context, rootCmd *cobra.Command, line string, out io.Writer) error {
	// Capture stdout and stderr and then restore them when we leave here
	stdoutCaptureMu.Lock()
	originalStdOut := os.Stdout
	originalStdErr := os.Stderr
	defer func() {
		os.Stdout = originalStdOut
		os.Stderr = originalStdErr
		stdoutCaptureMu.Unlock()
	}()

	// Set up our stdout and stderr pipes
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Stderr = w

	stdOutBuf := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		stdOutBuf <- buf.Bytes()
	}()

	// Now execute the command
	args := strings.Split(line, " ")

	if cmd, _, err := rootCmd.Find(args); err == nil {
		// Reset the context to nil
		cmd.SetContext(nil)

		// Reset flag values between runs due to a limitation in Cobra
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if val, ok := flag.Value.(pflag.SliceValue); ok {
				_ = val.Replace([]string{})
			} else {
				_ = flag.Value.Set(flag.DefValue)
			}
		})

		cmd.InitDefaultHelpFlag()
		cmd.InitDefaultVersionFlag()
	}

	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetContext(ctx)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	// Close the pipe
	_ = w.Close()
	_, _ = out.Write(<-stdOutBuf)

	return errors.WithStack(err)
}

// addDefaultShellCommands adds the default shell commands to the root command
func addDefaultShellCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:     "exit",
		Short:   "Exit the shell",
		Aliases: []string{"quit"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// no-op which is needed otherwise Cobra won't list the help for it
			// when the user runs `help exit`
			//
			// However the shell will intercept this command and exit the shell
			// before running this function
			return nil
		},
	})
}
