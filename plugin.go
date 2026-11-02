package rc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func findExecutable(env string, names ...string) (string, error) {
	if override := os.Getenv(env); override != "" {
		names = []string{override}
	}
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s not found: install it on PATH or set %s", names[0], env)
}

// runPlugin runs a separate executable with the terminal streams. Ctrl-C reaches it through the
// terminal, so rc only forwards other termination signals and returns the plugin's exit status.
func runPlugin(c *cobra.Command, path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, c.OutOrStdout(), c.ErrOrStderr()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case s := <-signals:
				if s != os.Interrupt {
					_ = cmd.Process.Signal(s)
				}
			case <-done:
				return
			}
		}
	}()
	var exit *exec.ExitError
	if err := cmd.Wait(); errors.As(err, &exit) {
		return &ExitError{exit.ExitCode()}
	} else {
		return err
	}
}
