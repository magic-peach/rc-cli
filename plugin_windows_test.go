//go:build windows

package rc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNpmStyleCmdShimsAreFoundAndRun(t *testing.T) {
	dir := t.TempDir()
	shim := "@echo off\r\necho args: %*\r\nexit /b 4\r\n"
	if err := os.WriteFile(filepath.Join(dir, "runtimeconditions-typescript-profiler.cmd"), []byte(shim), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	var exit *ExitError
	out, _, err := rc("profile", "typescript", "generate", "--name", "orders")
	if !errors.As(err, &exit) || exit.Code != 4 || !strings.Contains(out, "args: generate --name orders") {
		t.Fatalf("%v %q", err, out)
	}
}
