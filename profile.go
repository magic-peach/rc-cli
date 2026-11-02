package rc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	goprofiler "github.com/runtimeconditions/go-rc-profiler/cli"
	"github.com/spf13/cobra"
)

// CoreSchema is the core profile schema version this release of rc targets.
const CoreSchema = "0.4.0"

var languageName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

var publishedProfilers = map[string]string{"python": "runtimeconditions-python-profiler", "typescript": "runtimeconditions-typescript-profiler"}

func profileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "profile <language> [args...]",
		Short: "Profile an application with an installed language profiler",
		Long: `Runs the profiler for <language> and passes the remaining arguments to it unchanged.

The Go profiler is built into rc; RC_GO_PROFILER_BIN runs an external one instead. Any other profiler is
a plugin: the executable named by RC_<LANGUAGE>_PROFILER_BIN, else rc-profiler-<language> on PATH. The python
and typescript profilers are also found under their own published executable names.

A plugin runs only when it is installed and its language runtime is on PATH. When either is missing, rc
reports what is missing and how to install it, and does not run the plugin.`,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
				return c.Help()
			}
			if args[0] == "go" && os.Getenv(profilerEnv("go")) == "" {
				return runBuiltinGoProfiler(c, args[1:])
			}
			path, err := profilerPath(args[0])
			if _, known := knownPlugins[args[0]]; err != nil && !known {
				return err
			}
			if err = checkPlugin(args[0], path, c.ErrOrStderr()); err != nil {
				return err
			}
			if runsProfiler(args[1:]) {
				if err = checkCoreSchema(c.Context(), path, args[0], c.ErrOrStderr()); err != nil {
					return err
				}
			}
			return runPlugin(c, path, args[1:])
		},
	}
}

func runBuiltinGoProfiler(c *cobra.Command, args []string) error {
	if code := goprofiler.Exit(c.ErrOrStderr(), goprofiler.Run(c.Context(), args, c.OutOrStdout(), c.ErrOrStderr())); code != 0 {
		return &ExitError{code}
	}
	return nil
}

func profilerPath(language string) (string, error) {
	if !languageName.MatchString(language) {
		return "", fmt.Errorf("invalid language %q", language)
	}
	names := []string{"rc-profiler-" + language}
	if published, ok := publishedProfilers[language]; ok {
		names = append(names, published)
	}
	path, err := findExecutable(profilerEnv(language), names...)
	if err != nil {
		return "", fmt.Errorf("no %s profiler found: install rc-profiler-%s on PATH or set %s", language, language, profilerEnv(language))
	}
	return path, nil
}

func profilerEnv(language string) string {
	return "RC_" + strings.ToUpper(strings.ReplaceAll(language, "-", "_")) + "_PROFILER_BIN"
}

func runsProfiler(args []string) bool {
	return len(args) > 0 && !strings.HasPrefix(args[0], "-") && args[0] != "version"
}

type profilerInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	CoreSchemas []string `json:"coreSchemas"`
}

// queryInfo asks a profiler for --version --json; ok is false when it cannot answer that way.
func queryInfo(ctx context.Context, path string) (info profilerInfo, ok bool) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version", "--json").Output()
	return info, err == nil && json.Unmarshal(out, &info) == nil && len(info.CoreSchemas) > 0
}

func checkCoreSchema(ctx context.Context, path, language string, stderr io.Writer) error {
	info, ok := queryInfo(ctx, path)
	if !ok {
		fmt.Fprintf(stderr, "note: the %s profiler does not report its core schemas; skipping the compatibility check\n", language)
		return nil
	}
	if !slices.Contains(info.CoreSchemas, CoreSchema) {
		return fmt.Errorf("the %s profiler %s supports core schema %s, but this rc targets %s; install a compatible profiler", language, info.Version, strings.Join(info.CoreSchemas, ", "), CoreSchema)
	}
	return nil
}
