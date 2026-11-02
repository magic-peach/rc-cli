package rc

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// ExitError carries the exit status of a profiler so the rc executable can return it unchanged.
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

func Run(ctx context.Context, version string, args []string, stdout, stderr io.Writer) error {
	root := &cobra.Command{Use: "rc", Short: "Runtime Conditions command line", Version: version, SilenceUsage: true, SilenceErrors: true}
	root.SetVersionTemplate("rc {{.Version}}\n")
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	root.AddCommand(bindingsCommand(), profileCommand(), doctorCommand())
	return root.ExecuteContext(ctx)
}
