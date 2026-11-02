package rc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/runtimeconditions/extensions/tooling/extension-bindings/orchestrator"
	goprofiler "github.com/runtimeconditions/go-rc-profiler/cli"
	"github.com/spf13/cobra"
)

func doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Show the installed bindings tool and profilers, their versions and core schemas",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			out := tabwriter.NewWriter(c.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintf(out, "rc\t%s\ttargets core schema %s\n", c.Root().Version, CoreSchema)
			fmt.Fprintln(out, "bindings\tbuilt in")
			incompatible := 0
			for _, language := range doctorLanguages() {
				var info profilerInfo
				var ok bool
				path, err := profilerPath(language)
				switch {
				case language == "go" && os.Getenv(profilerEnv(language)) == "":
					path, info, ok = "built in", builtinGoInfo(), true
				case err != nil:
					fmt.Fprintf(out, "%s\tnot found\n", language)
					continue
				default:
					info, ok = queryInfo(c.Context(), path)
				}
				status := "compatible"
				switch {
				case !ok:
					info.Version, status = versionLine(path), "unchecked: core schemas not reported"
				case !slices.Contains(info.CoreSchemas, CoreSchema):
					incompatible, status = incompatible+1, "incompatible: supports "+strings.Join(info.CoreSchemas, ", ")
				}
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", language, path, info.Version, status)
			}
			if err := out.Flush(); err != nil {
				return err
			}
			if incompatible > 0 {
				return fmt.Errorf("%d profiler(s) do not support core schema %s", incompatible, CoreSchema)
			}
			return nil
		},
	}
}

func doctorLanguages() []string {
	languages := []string{"go", "python", "typescript"}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			if language, found := strings.CutPrefix(name, "rc-profiler-"); found && languageName.MatchString(language) && !slices.Contains(languages, language) {
				languages = append(languages, language)
			}
		}
	}
	return languages
}

func versionLine(path string) string {
	out, err := exec.Command(path, "--version").Output()
	if line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n"); err == nil && line != "" {
		return line
	}
	return "unknown"
}

func builtinGoInfo() (info profilerInfo) {
	var out bytes.Buffer
	if goprofiler.Run(context.Background(), []string{"--version", "--json"}, &out, io.Discard) == nil {
		_ = json.Unmarshal(out.Bytes(), &info)
	}
	if identity, err := orchestrator.ModuleIdentity(goProfilerModule, ""); err == nil {
		info.Version = identity.Version
	}
	return info
}
