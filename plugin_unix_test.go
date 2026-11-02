//go:build !windows

package rc

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func buildRC(t *testing.T) string {
	path := filepath.Join(t.TempDir(), "rc")
	if out, err := exec.Command("go", "build", "-o", path, "./cmd/rc").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	return path
}

func script(t *testing.T, name, body string) string {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestStdinReachesTheProfilerAndItsExitStatusIsReturned(t *testing.T) {
	bin := buildRC(t)
	cmd := exec.Command(bin, "profile", "ruby", "generate")
	cmd.Env = append(os.Environ(), "PATH="+script(t, "rc-profiler-ruby", `[ "$1" = --version ] && exit 2
read line
echo "got: $line"
exit 7
`))
	cmd.Stdin = strings.NewReader("hello\n")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 7 || out.String() != "got: hello\n" {
		t.Fatalf("exit %v, output %q", err, out.String())
	}
}

func TestTerminationIsForwardedToTheProfiler(t *testing.T) {
	bin := buildRC(t)
	marker := filepath.Join(t.TempDir(), "terminated")
	cmd := exec.Command(bin, "profile", "ruby", "generate")
	cmd.Env = append(os.Environ(), "MARKER="+marker, "PATH="+script(t, "rc-profiler-ruby", `[ "$1" = --version ] && exit 2
trap 'echo terminated > "$MARKER"; exit 0' TERM
echo ready
while :; do sleep 0.1; done
`))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if n, _ := stdout.Read(make([]byte, 6)); n != 6 {
		t.Fatal("profiler did not start")
	}
	if err = cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err = cmd.Wait(); err != nil {
		t.Fatalf("rc did not return the profiler's clean exit: %v", err)
	}
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(50 * time.Millisecond) {
		if _, err = os.Stat(marker); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the profiler never received SIGTERM")
		}
	}
}
