package cobrautils

import (
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
func ExecuteCmd(ctx context.Context, rootCmd *cobra.Command, line string, in io.Reader, stdout io.Writer, stderr io.Writer) error {
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
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Copy the output to the correct places
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, _ = io.Copy(stdout, stdoutR)
		wg.Done()
	}()
	go func() {
		_, _ = io.Copy(stderr, stderrR)
		wg.Done()
	}()

	// Reset the internal state of the command
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

	// Setup the cobra command to output to the write places
	rootCmd.SetIn(in)
	rootCmd.SetOut(stdoutW)
	rootCmd.SetErr(stderrW)
	rootCmd.SetContext(ctx)
	rootCmd.SetArgs(args)

	// Finally execute it!
	err := rootCmd.Execute()

	// Close the pipes for all the IO copies to finish
	_ = stderrW.Close()
	_ = stdoutW.Close()
	wg.Wait()

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
